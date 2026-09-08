package supervisor

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHybridClassifier_Heuristics(t *testing.T) {
	c := NewHybridClassifier(ClassifierConfig{ForceTier0: true})
	ctx := context.Background()

	cases := []struct {
		name           string
		prompt         string
		expectedIntent string
		expectedPool   string
	}{
		{
			name:           "Code markdown block",
			prompt:         "Refactor this:\n```go\nfunc add(a, b int) int { return a + b }\n```",
			expectedIntent: IntentCode,
			expectedPool:   "pool/agent-coding",
		},
		{
			name:           "Code keyword prompt",
			prompt:         "Write a binary search algorithm and refactor the syntax tree",
			expectedIntent: IntentCode,
			expectedPool:   "pool/agent-coding",
		},
		{
			name:           "SQL Code Prompt",
			prompt:         "SELECT id, name FROM users WHERE active = 1",
			expectedIntent: IntentCode,
			expectedPool:   "pool/agent-coding",
		},
		{
			name:           "Math and reasoning prompt",
			prompt:         "Prove that the square root of 2 is an irrational number",
			expectedIntent: IntentReason,
			expectedPool:   "pool/deep-reasoning",
		},
		{
			name:           "Differential calculus reasoning prompt",
			prompt:         "Calculate the integral of e^(-x^2) dx from 0 to infinity",
			expectedIntent: IntentReason,
			expectedPool:   "pool/deep-reasoning",
		},
		{
			name:           "Search prompt with URL",
			prompt:         "Check https://news.ycombinator.com for latest tech updates",
			expectedIntent: IntentSearch,
			expectedPool:   "pool/web-research",
		},
		{
			name:           "Search prompt current events",
			prompt:         "What is today's weather in Tokyo?",
			expectedIntent: IntentSearch,
			expectedPool:   "pool/web-research",
		},
		{
			name:           "Document prompt attachment",
			prompt:         "Please analyze this document and review the attached pdf file",
			expectedIntent: IntentDoc,
			expectedPool:   "pool/document-analysis",
		},
		{
			name:           "General conversational greeting",
			prompt:         "Hello! How are you today?",
			expectedIntent: IntentGeneral,
			expectedPool:   "pool/general",
		},
		{
			name:           "General joke request",
			prompt:         "Tell me a joke about distributed systems",
			expectedIntent: IntentGeneral,
			expectedPool:   "pool/general",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			res := c.Classify(ctx, tc.prompt)
			if res.Intent != tc.expectedIntent {
				t.Errorf("expected intent %s, got %s (reason: %s)", tc.expectedIntent, res.Intent, res.Reasoning)
			}
			if res.TargetPool != tc.expectedPool {
				t.Errorf("expected pool %s, got %s", tc.expectedPool, res.TargetPool)
			}
			if res.Latency > 10*time.Millisecond {
				t.Errorf("heuristic latency too slow: %v (budget: <1ms)", res.Latency)
			}
		})
	}
}

func TestHybridClassifier_LLMFallback(t *testing.T) {
	// Mock LLM server returning single token "CODE"
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
			"id": "chatcmpl-test",
			"object": "chat.completion",
			"choices": [{"message": {"role": "assistant", "content": "CODE"}}]
		}`))
	}))
	defer ts.Close()

	c := NewHybridClassifier(ClassifierConfig{
		UpstreamURL: ts.URL,
		Timeout:     50 * time.Millisecond,
	})

	ctx := context.Background()
	res := c.Classify(ctx, "Unusual prompt without obvious heuristic markers")
	if res.Intent != IntentCode {
		t.Errorf("expected LLM classified intent CODE, got %s", res.Intent)
	}
	if res.TargetPool != "pool/agent-coding" {
		t.Errorf("expected pool/agent-coding, got %s", res.TargetPool)
	}
	if res.Tier != 1 {
		t.Errorf("expected Tier 1 LLM classification, got Tier %d", res.Tier)
	}
}

func TestHybridClassifier_LLMTimeoutFallback(t *testing.T) {
	// Slow mock server exceeding 20ms timeout
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer ts.Close()

	c := NewHybridClassifier(ClassifierConfig{
		UpstreamURL: ts.URL,
		Timeout:     20 * time.Millisecond,
	})

	ctx := context.Background()
	start := time.Now()
	res := c.Classify(ctx, "Arbitrary query")
	duration := time.Since(start)

	// Must finish quickly and fallback to general
	if duration > 80*time.Millisecond {
		t.Errorf("classifier took too long: %v", duration)
	}
	if res.TargetPool != "pool/general" {
		t.Errorf("expected fallback to pool/general, got %s", res.TargetPool)
	}
}

func BenchmarkHybridClassifier_Tier0(b *testing.B) {
	c := NewHybridClassifier(ClassifierConfig{ForceTier0: true})
	ctx := context.Background()
	prompt := "Write a quicksort implementation in Rust with syntax tree analysis"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Classify(ctx, prompt)
	}
}
