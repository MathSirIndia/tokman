package e2e

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"tokman/backend/gateway"
	"tokman/backend/interceptor"
	"tokman/backend/shaper"
	"tokman/backend/storage"
)

func TestModule2_E2E_TrafficShaperAndFilter(t *testing.T) {
	// 1. Mock Upstream that returns SSE streaming with <think> tags
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gateway.ChatCompletionRequest
		json.Unmarshal(body, &req)

		if req.Stream {
			w.Header().Set("Content-Type", "text/event-stream")
			w.WriteHeader(http.StatusOK)

			flusher, _ := w.(http.Flusher)

			sendChunk := func(content string) {
				payload := map[string]interface{}{
					"id":      "chatcmpl-test",
					"choices": []map[string]interface{}{{"delta": map[string]string{"content": content}}},
				}
				data, _ := json.Marshal(payload)
				fmt.Fprintf(w, "data: %s\n\n", string(data))
				if flusher != nil {
					flusher.Flush()
				}
			}

			sendChunk("Thinking start: ")
			sendChunk("<think>secret internal reasoning chain</think>")
			sendChunk("The final answer is 42.")
			fmt.Fprintf(w, "data: [DONE]\n\n")
			if flusher != nil {
				flusher.Flush()
			}
			return
		}

		// Non-streaming response
		resp := gateway.ChatCompletionResponse{
			ID:      "chatcmpl-nonstream",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []gateway.Choice{
				{
					Index: 0,
					Message: gateway.ChatMessage{
						Role:    "assistant",
						Content: "Direct response: <think>hide me</think>Clean Output.",
					},
					FinishReason: "stop",
				},
			},
			Usage: gateway.Usage{
				PromptTokens:     10,
				CompletionTokens: 10,
				TotalTokens:      20,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstreamServer.Close()

	// 2. Setup SQLite & Cache
	tmpDir, err := os.MkdirTemp("", "tokman_e2e_m2_*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := storage.InitDB(filepath.Join(tmpDir, "test.db"))
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	defer db.Close()

	cache := storage.NewLRUCache(100, 5*time.Minute)

	// Configure interceptor with small rate limits for testing
	customQuotas := map[shaper.PriorityClass]shaper.RateLimitQuota{
		shaper.PriorityP0: {RPM: 5, TPM: 10000},
	}
	limiter := shaper.NewLeakyBucketLimiter(customQuotas)
	jitter := shaper.NewDomainJitterEngine(10*time.Millisecond, 5)
	ic := interceptor.NewInterceptor(limiter, jitter)

	// 3. Start Gateway Server on loopback
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind listener: %v", err)
	}

	gw := gateway.NewServer(gateway.Config{
		Port:        listener.Addr().(*net.TCPAddr).Port,
		MasterKey:   "sk-master-internal-network-key",
		UpstreamURL: upstreamServer.URL,
		GroqAPIKey:  "gsk_valid_test_key",
		Cache:       cache,
		DB:          db,
		Interceptor: ic,
	})
	gw.SetListener(listener)

	go func() {
		_ = gw.Start()
	}()
	time.Sleep(30 * time.Millisecond)

	baseURL := fmt.Sprintf("http://%s", gw.Addr())

	// Test 1: SSE Streaming with <think> tag suppression
	t.Run("Streaming_Think_Suppression", func(t *testing.T) {
		reqPayload := `{"model": "pool/general", "stream": true, "messages": [{"role": "user", "content": "solve problem"}]}`
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader([]byte(reqPayload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-master-internal-network-key")
		req.Header.Set("X-User-Id", "user-stream-test")

		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("streaming request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}

		var collectedContent strings.Builder
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "data: ") && !strings.Contains(line, "[DONE]") {
				dataStr := strings.TrimPrefix(line, "data: ")
				var chunk struct {
					Choices []struct {
						Delta struct {
							Content string `json:"content"`
						} `json:"delta"`
					} `json:"choices"`
				}
				if err := json.Unmarshal([]byte(dataStr), &chunk); err == nil && len(chunk.Choices) > 0 {
					collectedContent.WriteString(chunk.Choices[0].Delta.Content)
				}
			}
		}

		streamedText := collectedContent.String()
		if strings.Contains(streamedText, "<think>") || strings.Contains(streamedText, "secret internal reasoning") {
			t.Errorf("stream contained unredacted <think> tokens: %q", streamedText)
		}
		if !strings.Contains(streamedText, "The final answer is 42.") {
			t.Errorf("stream missing final answer: %q", streamedText)
		}
	})

	// Test 2: Non-Streaming <think> Sanitization
	t.Run("NonStreaming_Think_Sanitization", func(t *testing.T) {
		reqPayload := `{"model": "pool/general", "stream": false, "messages": [{"role": "user", "content": "hello"}]}`
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader([]byte(reqPayload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-master-internal-network-key")
		req.Header.Set("X-User-Id", "user-nonstream-test")

		client := &http.Client{Timeout: 5 * time.Second}
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		var chatResp gateway.ChatCompletionResponse
		json.NewDecoder(res.Body).Decode(&chatResp)

		content := chatResp.Choices[0].Message.Content
		if strings.Contains(content, "<think>") || strings.Contains(content, "hide me") {
			t.Errorf("content contained unredacted <think> tags: %q", content)
		}
		if !strings.Contains(content, "Clean Output.") {
			t.Errorf("content missing clean output: %q", content)
		}
	})

	// Test 3: Rate Limiting 429 Exhaustion
	t.Run("RateLimiter_429_Enforcement", func(t *testing.T) {
		client := &http.Client{Timeout: 5 * time.Second}
		rateUser := "user-spam-429"

		// Send 5 allowed requests (quota is 5 RPM for P0)
		for i := 1; i <= 5; i++ {
			reqPayload := fmt.Sprintf(`{"model": "pool/general", "stream": false, "messages": [{"role": "user", "content": "req %d"}]}`, i)
			req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader([]byte(reqPayload)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer sk-master-internal-network-key")
			req.Header.Set("X-User-Id", rateUser)

			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("request %d failed: %v", i, err)
			}
			res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Fatalf("request %d expected 200, got %d", i, res.StatusCode)
			}
		}

		// 6th request must trigger HTTP 429
		reqPayload := `{"model": "pool/general", "stream": false, "messages": [{"role": "user", "content": "req 6 - overflow"}]}`
		req, _ := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader([]byte(reqPayload)))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-master-internal-network-key")
		req.Header.Set("X-User-Id", rateUser)

		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("overflow request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusTooManyRequests {
			t.Errorf("expected 429 Too Many Requests, got %d", res.StatusCode)
		}
		if res.Header.Get("Retry-After") == "" {
			t.Errorf("expected Retry-After header to be set on 429 response")
		}
	})
}
