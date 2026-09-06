package mocks

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"testing"
	"time"
)

func TestMockUpstream_BasicFlow(t *testing.T) {
	mock := NewMockUpstream()
	defer mock.Close()

	// 1. Success response
	resp, err := http.Post(mock.URL()+"/v1/chat/completions", "application/json", bytes.NewBufferString(`{"model":"llama-3.3-70b"}`))
	if err != nil {
		t.Fatalf("failed to post to mock upstream: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	var completion MockChatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		t.Fatalf("failed to parse mock json: %v", err)
	}

	if completion.Choices[0].Message.Content != "Mock response from zero-credential upstream engine." {
		t.Fatalf("unexpected content: %s", completion.Choices[0].Message.Content)
	}
	if mock.CallCount() != 1 {
		t.Fatalf("expected call count 1, got %d", mock.CallCount())
	}
}

func TestMockUpstream_AuthAndErrors(t *testing.T) {
	mock := NewMockUpstream()
	defer mock.Close()

	mock.SetRequireAuth(true, "test-secret-key")

	// Missing auth -> 401
	res, err := http.Post(mock.URL()+"/v1/chat/completions", "application/json", nil)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if res.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	res.Body.Close()

	// With valid auth
	req, _ := http.NewRequest("POST", mock.URL()+"/v1/chat/completions", bytes.NewBufferString(`{}`))
	req.Header.Set("Authorization", "Bearer test-secret-key")
	client := &http.Client{Timeout: 5 * time.Second}
	res2, err := client.Do(req)
	if err != nil {
		t.Fatalf("authenticated request failed: %v", err)
	}
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", res2.StatusCode)
	}
	res2.Body.Close()

	// Simulate 429 Rate Limit
	mock.SetStatusCode(http.StatusTooManyRequests)
	res3, _ := client.Do(req)
	if res3.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", res3.StatusCode)
	}
	res3.Body.Close()
}
