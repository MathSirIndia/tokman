package supervisor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// Standard Intent Constants
const (
	IntentGeneral = "GENERAL"
	IntentCode    = "CODE"
	IntentReason  = "REASON"
	IntentSearch  = "SEARCH"
	IntentDoc     = "DOC"
)

// IntentPoolMap binds classified intents to canonical capability pools.
var IntentPoolMap = map[string]string{
	IntentGeneral: "pool/general",
	IntentCode:    "pool/agent-coding",
	IntentReason:  "pool/deep-reasoning",
	IntentSearch:  "pool/web-research",
	IntentDoc:     "pool/document-analysis",
}

// ClassificationResult encapsulates classification metadata and latency.
type ClassificationResult struct {
	Intent     string        `json:"intent"`
	TargetPool string        `json:"target_pool"`
	Tier       int           `json:"tier"` // 0: Heuristic, 1: LLM 1-token
	Latency    time.Duration `json:"latency"`
	Confidence float64       `json:"confidence"`
	Reasoning  string        `json:"reasoning,omitempty"`
}

// Classifier provides sub-40ms intent classification.
type Classifier interface {
	Classify(ctx context.Context, prompt string) ClassificationResult
}

// ClassifierConfig configures the hybrid classification engine.
type ClassifierConfig struct {
	UpstreamURL  string
	APIKey       string
	Model        string        // Fast classifier model (e.g., llama-3.3-70b-versatile)
	Timeout      time.Duration // Maximum LLM classification timeout (default 40ms)
	HTTPClient   *http.Client
	ForceTier0   bool // If true, skips LLM Tier 1 (useful for testing and zero-overhead)
}

// HybridClassifier evaluates lexical heuristics in <1ms (Tier 0) and falls back to LLM 1-token prompt (Tier 1).
type HybridClassifier struct {
	cfg ClassifierConfig

	// Compiled heuristic patterns for Tier 0
	codeFenceRe   *regexp.Regexp
	codeKeywordRe *regexp.Regexp
	mathRe        *regexp.Regexp
	searchRe      *regexp.Regexp
	docRe         *regexp.Regexp
}

// NewHybridClassifier initializes compiled regexes and default timeouts.
func NewHybridClassifier(cfg ClassifierConfig) *HybridClassifier {
	if cfg.Timeout == 0 {
		cfg.Timeout = 40 * time.Millisecond
	}
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{Timeout: cfg.Timeout}
	}
	if cfg.Model == "" {
		cfg.Model = "llama-3.3-70b-versatile"
	}

	return &HybridClassifier{
		cfg: cfg,
		codeFenceRe:   regexp.MustCompile("(?s)```[a-zA-Z0-9_-]*\\n.*"),
		codeKeywordRe: regexp.MustCompile(`(?i)\b(func|def|class|import|package|struct|interface|impl|fn|pub fn|return|console\.log|printf|SELECT\s+.*\s+FROM|INSERT\s+INTO|CREATE\s+TABLE|ALTER\s+TABLE|refactor|debug|syntax\s+tree|binary\s+search|algorithm|compile|git\s+commit|pull\s+request|npm\s+run|go\s+build|pytest|rustc)\b`),
		mathRe:        regexp.MustCompile(`(?i)\b(prove\s+that|proof|theorem|axiom|integral|differential|derivative|\bdy/dx\b|eigenvalue|matrix\s+multiplication|calculate\s+the\s+limit|solve\s+for\s+x|irrational\s+number|primes?\s+between|fibonacci|modular\s+arithmetic)\b`),
		searchRe:      regexp.MustCompile(`(?i)\b(latest\s+news|today'?s\s+weather|current\s+price\s+of|stock\s+price\s+of|who\s+is\s+the\s+current|what\s+happened\s+in|search\s+(the\s+web\s+for|for)|https?://)\b`),
		docRe:         regexp.MustCompile(`(?i)\b(analyze\s+this\s+document|attached\s+pdf|attached\s+file|summarize\s+this\s+file|review\s+the\s+attached|transcript\s+below)\b`),
	}
}

// Classify executes the hybrid classification pipeline within the latency budget.
func (c *HybridClassifier) Classify(ctx context.Context, prompt string) ClassificationResult {
	start := time.Now()

	// 1. Tier 0: Fast Heuristic Pass (<1ms)
	intent, confidence, ok := c.evaluateHeuristics(prompt)
	if ok && confidence >= 0.85 {
		pool, exists := IntentPoolMap[intent]
		if !exists {
			pool = "pool/general"
		}
		return ClassificationResult{
			Intent:     intent,
			TargetPool: pool,
			Tier:       0,
			Latency:    time.Since(start),
			Confidence: confidence,
			Reasoning:  "Heuristic pattern match",
		}
	}

	// If ForceTier0 or UpstreamURL not configured, return best heuristic guess
	if c.cfg.ForceTier0 || c.cfg.UpstreamURL == "" {
		if !ok {
			intent = IntentGeneral
			confidence = 0.70
		}
		pool, exists := IntentPoolMap[intent]
		if !exists {
			pool = "pool/general"
		}
		return ClassificationResult{
			Intent:     intent,
			TargetPool: pool,
			Tier:       0,
			Latency:    time.Since(start),
			Confidence: confidence,
			Reasoning:  "Fallback heuristic guess",
		}
	}

	// 2. Tier 1: LLM 1-Token Edge Classifier (<40ms)
	llmIntent, err := c.queryLLMClassifier(ctx, prompt)
	if err == nil && llmIntent != "" {
		pool, exists := IntentPoolMap[llmIntent]
		if !exists {
			pool = "pool/general"
			llmIntent = IntentGeneral
		}
		return ClassificationResult{
			Intent:     llmIntent,
			TargetPool: pool,
			Tier:       1,
			Latency:    time.Since(start),
			Confidence: 0.95,
			Reasoning:  "Tier 1 LLM single-token arbitration",
		}
	}

	// Fallback if LLM classification times out or fails
	if !ok {
		intent = IntentGeneral
		confidence = 0.60
	}
	pool, exists := IntentPoolMap[intent]
	if !exists {
		pool = "pool/general"
	}
	return ClassificationResult{
		Intent:     intent,
		TargetPool: pool,
		Tier:       0,
		Latency:    time.Since(start),
		Confidence: confidence,
		Reasoning:  "LLM timeout/error fallback to heuristic",
	}
}

// evaluateHeuristics examines prompt characteristics and patterns.
func (c *HybridClassifier) evaluateHeuristics(prompt string) (string, float64, bool) {
	trimmed := strings.TrimSpace(prompt)
	if len(trimmed) == 0 {
		return IntentGeneral, 1.0, true
	}

	// Large context / document input (>32k characters)
	if len(trimmed) > 32000 || c.docRe.MatchString(trimmed) {
		return IntentDoc, 0.95, true
	}

	// Explicit code markdown block
	if c.codeFenceRe.MatchString(trimmed) {
		return IntentCode, 0.98, true
	}

	// Code syntax and keywords
	if c.codeKeywordRe.MatchString(trimmed) {
		return IntentCode, 0.90, true
	}

	// Math proofs and complex logic
	if c.mathRe.MatchString(trimmed) {
		return IntentReason, 0.92, true
	}

	// Search queries and real-time indicators
	if c.searchRe.MatchString(trimmed) {
		return IntentSearch, 0.90, true
	}

	// Short greetings and casual smalltalk
	lower := strings.ToLower(trimmed)
	if lower == "hi" || lower == "hello" || lower == "hey" || strings.HasPrefix(lower, "how are you") ||
		strings.HasPrefix(lower, "tell me a joke") || strings.HasPrefix(lower, "summarize") {
		return IntentGeneral, 0.95, true
	}

	return IntentGeneral, 0.50, false
}

// queryLLMClassifier sends the 1-token classification prompt with strict deadline.
func (c *HybridClassifier) queryLLMClassifier(ctx context.Context, prompt string) (string, error) {
	deadlineCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
	defer cancel()

	const classifierSystemPrompt = `Analyze the user query. Output EXACTLY ONE token representing the routing pool:
GENERAL (default chat, conversation, summary)
REASON (math, complex proofs, multi-step logic)
CODE (programming, refactoring, bug fixing, SQL)
SEARCH (real-time facts, current news, URLs)
DOC (analyzing attached documents, large inputs > 32k chars)

Output:`

	truncatedPrompt := prompt
	if len(truncatedPrompt) > 500 {
		truncatedPrompt = truncatedPrompt[:500] + "..."
	}

	payload := map[string]interface{}{
		"model": c.cfg.Model,
		"messages": []map[string]string{
			{"role": "system", "content": classifierSystemPrompt},
			{"role": "user", "content": fmt.Sprintf("Query: %s\nOutput:", truncatedPrompt)},
		},
		"max_tokens":  2,
		"temperature": 0.0,
		"stream":      false,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(deadlineCtx, http.MethodPost, c.cfg.UpstreamURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.cfg.APIKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.cfg.APIKey)
	}

	resp, err := c.cfg.HTTPClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("classifier upstream status: %d", resp.StatusCode)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var completion struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.Unmarshal(respBody, &completion); err != nil {
		return "", err
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("empty choices from classifier")
	}

	token := strings.ToUpper(strings.TrimSpace(completion.Choices[0].Message.Content))
	// Match token against supported intents
	switch {
	case strings.Contains(token, "CODE"):
		return IntentCode, nil
	case strings.Contains(token, "REASON"):
		return IntentReason, nil
	case strings.Contains(token, "SEARCH"):
		return IntentSearch, nil
	case strings.Contains(token, "DOC"):
		return IntentDoc, nil
	case strings.Contains(token, "GENERAL"):
		return IntentGeneral, nil
	default:
		return IntentGeneral, nil
	}
}
