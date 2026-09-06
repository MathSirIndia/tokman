package filter

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
)

var thinkRegex = regexp.MustCompile(`(?s)<think>.*?</think>`)

// SanitizeNonStreamingContent removes all <think>...</think> blocks from text.
func SanitizeNonStreamingContent(content string) string {
	cleaned := thinkRegex.ReplaceAllString(content, "")
	return strings.TrimSpace(cleaned)
}

// ThinkFilterState tracks the state of the SSE stream parser.
type ThinkFilterState int

const (
	StateNormal ThinkFilterState = iota
	StateInsideThink
)

// StreamFilter filters out <think>...</think> tokens from an SSE stream on the fly.
type StreamFilter struct {
	mu         sync.Mutex
	state      ThinkFilterState
	partialTag string
}

// NewStreamFilter creates a new streaming thinking token filter.
func NewStreamFilter() *StreamFilter {
	return &StreamFilter{
		state: StateNormal,
	}
}

// FilterText processes a raw text fragment and returns the sanitized fragment.
func (f *StreamFilter) FilterText(input string) string {
	f.mu.Lock()
	defer f.mu.Unlock()

	combined := f.partialTag + input
	f.partialTag = ""

	var output strings.Builder
	runes := []rune(combined)
	n := len(runes)

	for i := 0; i < n; {
		if f.state == StateNormal {
			// Look for start tag "<think>"
			if runes[i] == '<' {
				remaining := string(runes[i:])
				if strings.HasPrefix("<think>", remaining) {
					// Potential prefix match of "<think>" at end of input
					f.partialTag = remaining
					break
				}
				if strings.HasPrefix(remaining, "<think>") {
					f.state = StateInsideThink
					i += len([]rune("<think>"))
					continue
				}
			}
			output.WriteRune(runes[i])
			i++
		} else { // StateInsideThink
			// Look for end tag "</think>"
			if runes[i] == '<' {
				remaining := string(runes[i:])
				if strings.HasPrefix("</think>", remaining) {
					// Potential prefix match of "</think>" at end of input
					f.partialTag = remaining
					break
				}
				if strings.HasPrefix(remaining, "</think>") {
					f.state = StateNormal
					i += len([]rune("</think>"))
					continue
				}
			}
			// In thinking state, suppress output
			i++
		}
	}

	return output.String()
}

// FilterSSEChunk parses an SSE line (e.g. "data: {...}") and filters any content in delta.
// If the delta content is entirely filtered or empty, it returns nil to suppress the empty chunk.
func (f *StreamFilter) FilterSSEChunk(rawChunk []byte) []byte {
	line := string(rawChunk)
	trimmed := strings.TrimSpace(line)

	if !strings.HasPrefix(trimmed, "data: ") {
		return rawChunk
	}

	dataPayload := strings.TrimPrefix(trimmed, "data: ")
	if dataPayload == "[DONE]" {
		return rawChunk
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal([]byte(dataPayload), &parsed); err != nil {
		// If unmarshal fails, pass through
		return rawChunk
	}

	choices, ok := parsed["choices"].([]interface{})
	if !ok || len(choices) == 0 {
		return rawChunk
	}

	choiceMap, ok := choices[0].(map[string]interface{})
	if !ok {
		return rawChunk
	}

	delta, ok := choiceMap["delta"].(map[string]interface{})
	if !ok {
		return rawChunk
	}

	contentVal, hasContent := delta["content"]
	if !hasContent {
		return rawChunk
	}

	contentStr, ok := contentVal.(string)
	if !ok {
		return rawChunk
	}

	filtered := f.FilterText(contentStr)

	// If filtered content is empty and original wasn't empty, suppress delta emission
	// to avoid spamming empty tokens unless it's a finish chunk
	if filtered == "" && contentStr != "" && choiceMap["finish_reason"] == nil {
		return nil
	}

	delta["content"] = filtered
	choiceMap["delta"] = delta
	choices[0] = choiceMap
	parsed["choices"] = choices

	newJSON, err := json.Marshal(parsed)
	if err != nil {
		return rawChunk
	}

	var buf bytes.Buffer
	buf.WriteString("data: ")
	buf.Write(newJSON)
	buf.WriteString("\n\n")
	return buf.Bytes()
}
