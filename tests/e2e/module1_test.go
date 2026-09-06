package e2e_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"tokman/backend/gateway"
	"tokman/backend/storage"
)

func TestModule1_E2E_FullFlow(t *testing.T) {
	// 1. Upstream mock server (simulating Groq LLM provider)
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"message":"Invalid API Key","type":"invalid_request_error"}}`))
			return
		}

		resp := gateway.ChatCompletionResponse{
			ID:      "chatcmpl-e2e-test",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   "llama-3.3-70b-versatile",
			Choices: []gateway.Choice{
				{
					Index: 0,
					Message: gateway.ChatMessage{
						Role:    "assistant",
						Content: "Hello from mock Groq provider!",
					},
					FinishReason: "stop",
				},
			},
			Usage: gateway.Usage{
				PromptTokens:     10,
				CompletionTokens: 8,
				TotalTokens:      18,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstreamServer.Close()

	// 2. Setup SQLite & Cache
	tmpDir, err := os.MkdirTemp("", "tokman_e2e_*")
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
	})
	gw.SetListener(listener)

	go func() {
		_ = gw.Start()
	}()
	time.Sleep(30 * time.Millisecond)

	baseURL := fmt.Sprintf("http://%s", gw.Addr())

	// Test 1: Health Readiness
	t.Run("Health_Readiness", func(t *testing.T) {
		res, err := http.Get(baseURL + "/health/readiness")
		if err != nil {
			t.Fatalf("GET /health/readiness failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var body map[string]interface{}
		if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode response: %v", err)
		}
		if body["status"] != "healthy" {
			t.Errorf("expected status healthy, got %v", body["status"])
		}
	})

	// Test 2: Models listing
	t.Run("Models_Listing", func(t *testing.T) {
		res, err := http.Get(baseURL + "/v1/models")
		if err != nil {
			t.Fatalf("GET /v1/models failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusOK {
			t.Fatalf("expected 200, got %d", res.StatusCode)
		}
		var models gateway.ModelListResponse
		if err := json.NewDecoder(res.Body).Decode(&models); err != nil {
			t.Fatalf("failed to decode models: %v", err)
		}
		if len(models.Data) < 2 {
			t.Errorf("expected at least 2 pools, got %d", len(models.Data))
		}
	})

	// Test 3: Chat Completions - Unauthorized
	t.Run("Chat_Unauthorized", func(t *testing.T) {
		reqBody := `{"model": "pool/general", "messages": [{"role": "user", "content": "hello"}]}`
		res, err := http.Post(baseURL+"/v1/chat/completions", "application/json", bytes.NewReader([]byte(reqBody)))
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != http.StatusUnauthorized {
			t.Errorf("expected 401 Unauthorized, got %d", res.StatusCode)
		}
	})

	// Test 4: Chat Completions - Valid Request & Caching
	t.Run("Chat_Completion_And_Cache", func(t *testing.T) {
		payload := `{"model": "pool/general", "messages": [{"role": "user", "content": "Hello tokman"}]}`

		req, err := http.NewRequest(http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader([]byte(payload)))
		if err != nil {
			t.Fatalf("failed to create request: %v", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-master-internal-network-key")

		client := &http.Client{Timeout: 5 * time.Second}

		// First call: MISS
		res1, err := client.Do(req)
		if err != nil {
			t.Fatalf("request 1 failed: %v", err)
		}
		defer res1.Body.Close()

		if res1.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", res1.StatusCode)
		}
		if res1.Header.Get("X-Cache") != "MISS" {
			t.Errorf("expected X-Cache MISS, got %s", res1.Header.Get("X-Cache"))
		}

		body1, _ := io.ReadAll(res1.Body)
		var respObj gateway.ChatCompletionResponse
		if err := json.Unmarshal(body1, &respObj); err != nil {
			t.Fatalf("failed to parse completion: %v", err)
		}
		if respObj.Model != "pool/general" {
			t.Errorf("expected model pool/general (rewritten), got %s", respObj.Model)
		}

		// Second call: HIT
		req2, _ := http.NewRequest(http.MethodPost, baseURL+"/v1/chat/completions", bytes.NewReader([]byte(payload)))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("Authorization", "Bearer sk-master-internal-network-key")

		res2, err := client.Do(req2)
		if err != nil {
			t.Fatalf("request 2 failed: %v", err)
		}
		defer res2.Body.Close()

		if res2.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK on cached call, got %d", res2.StatusCode)
		}
		if res2.Header.Get("X-Cache") != "HIT" {
			t.Errorf("expected X-Cache HIT, got %s", res2.Header.Get("X-Cache"))
		}
	})
}
