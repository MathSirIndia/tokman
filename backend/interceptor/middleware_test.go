package interceptor

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"tokman/backend/filter"
)

func TestInterceptor_AllowAndHeaders(t *testing.T) {
	ic := NewInterceptor(nil, nil)

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("X-User-Id", "user-test-1")
	req.Header.Set("X-Priority", "P0")
	rec := httptest.NewRecorder()

	_, _, allowed := ic.InterceptRequest(rec, req)
	if !allowed {
		t.Fatalf("first request should be allowed")
	}

	if rec.Header().Get("X-RateLimit-Limit-RPM") != "12" {
		t.Errorf("expected Limit-RPM 12, got %s", rec.Header().Get("X-RateLimit-Limit-RPM"))
	}
	if rec.Header().Get("X-RateLimit-Remaining-RPM") != "11" {
		t.Errorf("expected Remaining-RPM 11, got %s", rec.Header().Get("X-RateLimit-Remaining-RPM"))
	}
}

func TestInterceptor_Block429(t *testing.T) {
	ic := NewInterceptor(nil, nil)

	req := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req.Header.Set("X-User-Id", "user-flooder")
	req.Header.Set("X-Priority", "P0")

	// Consume all 12 requests
	for i := 0; i < 12; i++ {
		rec := httptest.NewRecorder()
		ic.InterceptRequest(rec, req)
	}

	// 13th request
	rec429 := httptest.NewRecorder()
	_, _, allowed := ic.InterceptRequest(rec429, req)
	if allowed {
		t.Fatalf("13th request should have been blocked")
	}

	if rec429.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", rec429.Code)
	}
	if rec429.Header().Get("Retry-After") == "" {
		t.Errorf("expected Retry-After header to be set")
	}
}

func TestInterceptor_PruneRequestBody(t *testing.T) {
	ic := NewInterceptor(nil, nil)

	longStr := string(make([]byte, 1000))
	msgs := []filter.ChatMessage{
		{Role: "system", Content: "System prompt"},
		{Role: "user", Content: "Intermediate 1 " + longStr},
		{Role: "assistant", Content: "Intermediate 2 " + longStr},
		{Role: "user", Content: "Final question"},
	}

	reqPayload := map[string]interface{}{
		"model":    "pool/general",
		"messages": msgs,
	}
	bodyBytes, _ := json.Marshal(reqPayload)

	// Set small budget of 200 tokens
	prunedBytes, dropped, err := ic.PruneRequestBody(bodyBytes, 200)
	if err != nil {
		t.Fatalf("pruning failed: %v", err)
	}
	if dropped == 0 {
		t.Errorf("expected turns to be dropped, got 0")
	}

	var parsed struct {
		Messages []filter.ChatMessage `json:"messages"`
	}
	json.Unmarshal(prunedBytes, &parsed)

	if len(parsed.Messages) >= len(msgs) {
		t.Errorf("expected messages length to decrease from %d, got %d", len(msgs), len(parsed.Messages))
	}
	if parsed.Messages[0].Role != "system" {
		t.Errorf("system prompt was not preserved")
	}
	if parsed.Messages[len(parsed.Messages)-1].Content != "Final question" {
		t.Errorf("final question was not preserved")
	}
}
