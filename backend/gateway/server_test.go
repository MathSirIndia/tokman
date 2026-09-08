package gateway

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
	"time"

	"tokman/backend/storage"
)

func TestHealthEndpoints(t *testing.T) {
	srv := NewServer(Config{Port: 0})

	// Test Readiness
	req := httptest.NewRequest(http.MethodGet, "/health/readiness", nil)
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var data map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &data); err != nil {
		t.Fatalf("failed to decode readiness json: %v", err)
	}
	if data["status"] != "healthy" {
		t.Fatalf("expected status healthy, got %v", data["status"])
	}

	// Test Liveness
	reqLive := httptest.NewRequest(http.MethodGet, "/health/liveness", nil)
	wLive := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wLive, reqLive)

	if wLive.Code != http.StatusOK {
		t.Fatalf("expected status 200 for liveness, got %d", wLive.Code)
	}
}

func TestModelsEndpoint(t *testing.T) {
	srv := NewServer(Config{Port: 0})

	req := httptest.NewRequest(http.MethodGet, "/v1/models", nil)
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp ModelListResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to parse models: %v", err)
	}
	if len(resp.Data) < 2 {
		t.Fatalf("expected at least 2 models, got %d", len(resp.Data))
	}
}

func TestChatCompletion_UpstreamAndCaching(t *testing.T) {
	var upstreamCallCount int32

	// Create synthetic upstream server (mocking Groq)
	mockUpstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&upstreamCallCount, 1)

		resp := ChatCompletionResponse{
			ID:      "chatcmpl-test-123",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "llama-3.3-70b-versatile",
			Choices: []Choice{
				{
					Index: 0,
					Message: ChatMessage{
						Role:    "assistant",
						Content: "mocked online",
					},
					FinishReason: "stop",
				},
			},
			Usage: Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer mockUpstream.Close()

	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_gateway.db")
	defer os.Remove(dbPath)

	db, err := storage.InitDB(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	cache := storage.NewLRUCache(100, time.Hour)

	srv := NewServer(Config{
		Port:        0,
		MasterKey:   "test-key",
		UpstreamURL: mockUpstream.URL,
		Cache:       cache,
		DB:          db,
	})

	chatReq := ChatCompletionRequest{
		Model: "pool/general",
		Messages: []ChatMessage{
			{Role: "user", Content: "hello test"},
		},
	}
	body, _ := json.Marshal(chatReq)

	// First Request: Cache MISS -> hits mock upstream
	req1 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req1.Header.Set("Authorization", "Bearer test-key")
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first call failed with status %d: %s", w1.Code, w1.Body.String())
	}
	if w1.Header().Get("X-Cache") != "MISS" {
		t.Fatalf("expected X-Cache MISS on first call, got %s", w1.Header().Get("X-Cache"))
	}
	if atomic.LoadInt32(&upstreamCallCount) != 1 {
		t.Fatalf("expected 1 upstream call, got %d", atomic.LoadInt32(&upstreamCallCount))
	}

	// Second Request: Cache HIT -> resolved directly from memory in < 1ms
	start := time.Now()
	req2 := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req2.Header.Set("Authorization", "Bearer test-key")
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w2, req2)
	latency := time.Since(start)

	if w2.Code != http.StatusOK {
		t.Fatalf("second call failed with status %d: %s", w2.Code, w2.Body.String())
	}
	if w2.Header().Get("X-Cache") != "HIT" {
		t.Fatalf("expected X-Cache HIT on second call, got %s", w2.Header().Get("X-Cache"))
	}
	if atomic.LoadInt32(&upstreamCallCount) != 1 {
		t.Fatalf("expected upstreamCallCount to remain 1, got %d", atomic.LoadInt32(&upstreamCallCount))
	}
	if latency > 10*time.Millisecond {
		t.Fatalf("cache hit latency should be < 10ms, took %v", latency)
	}

	// Test Unauthorized
	reqUnauth := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	wUnauth := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wUnauth, reqUnauth)
	if wUnauth.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized without master key, got %d", wUnauth.Code)
	}
}

func TestMaskSensitiveCredentials(t *testing.T) {
	cases := []struct {
		input    string
		expected string
	}{
		{
			input:    "Error connecting with gsk_abcdef1234567890 to upstream",
			expected: "Error connecting with *** to upstream",
		},
		{
			input:    "Authorization: Bearer sk-master-internal-key failed",
			expected: "Authorization: Bearer *** failed",
		},
		{
			input:    "Gemini error with AIzaSy_testkey123456",
			expected: "Gemini error with ***",
		},
	}

	for _, c := range cases {
		out := maskSensitiveCredentials(c.input)
		if out != c.expected {
			t.Errorf("maskSensitiveCredentials(%q) = %q, expected %q", c.input, out, c.expected)
		}
	}
}

func TestCORSOriginRestriction(t *testing.T) {
	srv := NewServer(Config{Port: 0})

	// Localhost origin should be allowed
	reqLocal := httptest.NewRequest(http.MethodOptions, "/v1/models", nil)
	reqLocal.Header.Set("Origin", "http://localhost:3000")
	wLocal := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wLocal, reqLocal)
	if wLocal.Header().Get("Access-Control-Allow-Origin") != "http://localhost:3000" {
		t.Errorf("expected Access-Control-Allow-Origin for localhost, got %q", wLocal.Header().Get("Access-Control-Allow-Origin"))
	}

	// 127.0.0.1 origin should be allowed
	reqIP := httptest.NewRequest(http.MethodOptions, "/v1/models", nil)
	reqIP.Header.Set("Origin", "http://127.0.0.1:8080")
	wIP := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wIP, reqIP)
	if wIP.Header().Get("Access-Control-Allow-Origin") != "http://127.0.0.1:8080" {
		t.Errorf("expected Access-Control-Allow-Origin for 127.0.0.1, got %q", wIP.Header().Get("Access-Control-Allow-Origin"))
	}

	// Malicious/unknown origin should NOT receive allow-origin header (SEC-1)
	reqEvil := httptest.NewRequest(http.MethodOptions, "/v1/models", nil)
	reqEvil.Header.Set("Origin", "https://evil-site.com")
	wEvil := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(wEvil, reqEvil)
	if wEvil.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Errorf("expected empty Access-Control-Allow-Origin for evil-site.com, got %q", wEvil.Header().Get("Access-Control-Allow-Origin"))
	}
}

func TestRequestBodySizeLimit(t *testing.T) {
	srv := NewServer(Config{Port: 0})

	// Create a payload larger than 10MB
	oversizedBody := make([]byte, 11*1024*1024)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(oversizedBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("expected 413 StatusRequestEntityTooLarge for >10MB body, got %d", w.Code)
	}
}

func TestUpstreamErrorSanitization(t *testing.T) {
	// Point upstream URL to a non-existent port to simulate upstream outage
	srv := NewServer(Config{
		Port:        0,
		UpstreamURL: "http://127.0.0.1:59999/v1/chat/completions",
	})

	chatReq := ChatCompletionRequest{
		Model: "pool/general",
		Messages: []ChatMessage{
			{Role: "user", Content: "hello"},
		},
	}
	body, _ := json.Marshal(chatReq)
	req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(w, req)

	if w.Code != http.StatusBadGateway {
		t.Fatalf("expected 502 Bad Gateway, got %d", w.Code)
	}
	// Verify raw network/DNS errors are NOT leaked to client (SEC-4)
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "Upstream service error" || resp["code"] != "upstream_unavailable" {
		t.Errorf("expected sanitized error response, got %v", resp)
	}
}
