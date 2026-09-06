package filter

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestSanitizeNonStreamingContent(t *testing.T) {
	input := "Hello world! <think>Let me contemplate this.\nStep 1: Calculate.\nStep 2: Conclude.</think> The answer is 42."
	expected := "Hello world!  The answer is 42."
	actual := SanitizeNonStreamingContent(input)
	if actual != expected {
		t.Errorf("expected %q, got %q", expected, actual)
	}

	input2 := "<think>Pondering...</think>Direct answer."
	expected2 := "Direct answer."
	actual2 := SanitizeNonStreamingContent(input2)
	if actual2 != expected2 {
		t.Errorf("expected %q, got %q", expected2, actual2)
	}
}

func TestStreamFilter_SimpleStreaming(t *testing.T) {
	filter := NewStreamFilter()

	chunks := []string{
		"Hello ",
		"<think>internal thoughts</think>",
		"world!",
	}

	var result strings.Builder
	for _, c := range chunks {
		result.WriteString(filter.FilterText(c))
	}

	if result.String() != "Hello world!" {
		t.Errorf("expected 'Hello world!', got %q", result.String())
	}
}

func TestStreamFilter_SplitTagStreaming(t *testing.T) {
	filter := NewStreamFilter()

	chunks := []string{
		"Solution: ",
		"<th",
		"ink>pondering step 1\n",
		"step 2</th",
		"ink>",
		"Result is 56.",
	}

	var result strings.Builder
	for _, c := range chunks {
		result.WriteString(filter.FilterText(c))
	}

	expected := "Solution: Result is 56."
	if result.String() != expected {
		t.Errorf("expected %q, got %q", expected, result.String())
	}
}

func TestStreamFilter_FilterSSEChunk(t *testing.T) {
	filter := NewStreamFilter()

	makeChunk := func(content string) []byte {
		payload := map[string]interface{}{
			"id":      "chatcmpl-123",
			"choices": []map[string]interface{}{{"delta": map[string]string{"content": content}}},
		}
		data, _ := json.Marshal(payload)
		return []byte("data: " + string(data) + "\n\n")
	}

	// Chunk 1: regular text
	chunk1 := filter.FilterSSEChunk(makeChunk("Calculating: "))
	if chunk1 == nil || !strings.Contains(string(chunk1), "Calculating: ") {
		t.Errorf("expected chunk1 to pass through, got %s", string(chunk1))
	}

	// Chunk 2: start think
	chunk2 := filter.FilterSSEChunk(makeChunk("<think>secret thoughts"))
	if chunk2 != nil {
		t.Errorf("expected chunk2 to be suppressed, got %s", string(chunk2))
	}

	// Chunk 3: still inside think
	chunk3 := filter.FilterSSEChunk(makeChunk(" more secret thoughts"))
	if chunk3 != nil {
		t.Errorf("expected chunk3 to be suppressed, got %s", string(chunk3))
	}

	// Chunk 4: end think + answer
	chunk4 := filter.FilterSSEChunk(makeChunk("</think>42"))
	if chunk4 == nil || !strings.Contains(string(chunk4), "42") {
		t.Errorf("expected chunk4 to emit 42, got %s", string(chunk4))
	}
}
