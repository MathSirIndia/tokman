package e2e

import (
	"bytes"
	"context"
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
	"tokman/backend/storage"
	"tokman/backend/supervisor"
)

func TestModule3_E2E_AutonomousSupervisorAndFederation(t *testing.T) {
	// 1. Mock Upstream returning completions based on request
	upstreamServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req gateway.ChatCompletionRequest
		json.Unmarshal(body, &req)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)

		content := "Standard response from upstream model."
		if len(req.Messages) > 0 {
			lastMsg := req.Messages[len(req.Messages)-1].Content
			if strings.Contains(lastMsg, "vulnerable") {
				// Inject vulnerable code to test critic pass
				content = "Here is the code:\n```go\nquery := \"SELECT * FROM users WHERE id = '\" + id + \"'\"\n```"
			} else if strings.Contains(lastMsg, "binary search") {
				content = "Here is the clean implementation:\n```go\nfunc binarySearch(arr []int, target int) int { return 0 }\n```"
			}
		}

		resp := gateway.ChatCompletionResponse{
			ID:      "chatcmpl-mod3",
			Object:  "chat.completion",
			Created: time.Now().Unix(),
			Model:   req.Model,
			Choices: []gateway.Choice{
				{
					Index: 0,
					Message: gateway.ChatMessage{
						Role:    "assistant",
						Content: content,
					},
					FinishReason: "stop",
				},
			},
			Usage: gateway.Usage{
				PromptTokens:     15,
				CompletionTokens: 25,
				TotalTokens:      40,
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer upstreamServer.Close()

	// 2. Initialize storage, cache, interceptor, and supervisor
	tmpDir, err := os.MkdirTemp("", "tokman-e2e-mod3-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	db, err := storage.InitDB(filepath.Join(tmpDir, "mod3.db"))
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}
	defer db.Close()

	cache := storage.NewLRUCache(100, 1*time.Hour)
	ic := interceptor.NewInterceptor(nil, nil)

	sup := supervisor.NewSupervisor(supervisor.SupervisorConfig{
		ClassifierConfig: supervisor.ClassifierConfig{
			ForceTier0: true, // test sub-1ms deterministic classification
		},
		CriticConfig: supervisor.CriticConfig{
			MaxCriticLoops: 2,
			Enabled:        true,
		},
		DefaultModel: "llama-3.3-70b-versatile",
	})

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to bind test listener: %v", err)
	}
	gwAddr := listener.Addr().String()

	gw := gateway.NewServer(gateway.Config{
		Port:        listener.Addr().(*net.TCPAddr).Port,
		MasterKey:   "sk-test-master-key",
		UpstreamURL: upstreamServer.URL,
		Cache:       cache,
		DB:          db,
		Interceptor: ic,
		Supervisor:  sup,
	})
	gw.SetListener(listener)

	go func() {
		_ = gw.Start()
	}()
	defer gw.Shutdown(context.Background())

	client := &http.Client{Timeout: 5 * time.Second}

	// Sub-test 1: pool/auto with code prompt -> routes to pool/agent-coding
	t.Run("PoolAuto_CodingIntent", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "pool/auto",
			"messages": []map[string]string{
				{"role": "user", "content": "Write a binary search algorithm in Go:\n```go\nfunc bs() {}\n```"},
			},
		}
		data, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/chat/completions", gwAddr), bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-test-master-key")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		poolHeader := resp.Header.Get("X-Tokman-Pool")
		if poolHeader != "pool/agent-coding" {
			t.Errorf("expected X-Tokman-Pool: pool/agent-coding, got %s", poolHeader)
		}

		intentHeader := resp.Header.Get("X-Tokman-Intent")
		if intentHeader != "CODE" {
			t.Errorf("expected X-Tokman-Intent: CODE, got %s", intentHeader)
		}

		if resp.Header.Get("X-Tokman-Classification-Ms") == "" {
			t.Errorf("expected X-Tokman-Classification-Ms header to be present")
		}
	})

	// Sub-test 2: pool/auto with reasoning prompt -> routes to pool/deep-reasoning
	t.Run("PoolAuto_ReasoningIntent", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "pool/auto",
			"messages": []map[string]string{
				{"role": "user", "content": "Prove that the square root of 2 is irrational step-by-step"},
			},
		}
		data, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/chat/completions", gwAddr), bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-test-master-key")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		poolHeader := resp.Header.Get("X-Tokman-Pool")
		if poolHeader != "pool/deep-reasoning" {
			t.Errorf("expected X-Tokman-Pool: pool/deep-reasoning, got %s", poolHeader)
		}

		intentHeader := resp.Header.Get("X-Tokman-Intent")
		if intentHeader != "REASON" {
			t.Errorf("expected X-Tokman-Intent: REASON, got %s", intentHeader)
		}
	})

	// Sub-test 3: pool/auto with general greeting -> routes to pool/general
	t.Run("PoolAuto_GeneralIntent", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "pool/auto",
			"messages": []map[string]string{
				{"role": "user", "content": "Hello! Tell me a fun fact about computer science."},
			},
		}
		data, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/chat/completions", gwAddr), bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-test-master-key")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		poolHeader := resp.Header.Get("X-Tokman-Pool")
		if poolHeader != "pool/general" {
			t.Errorf("expected X-Tokman-Pool: pool/general, got %s", poolHeader)
		}

		intentHeader := resp.Header.Get("X-Tokman-Intent")
		if intentHeader != "GENERAL" {
			t.Errorf("expected X-Tokman-Intent: GENERAL, got %s", intentHeader)
		}
	})

	// Sub-test 4: Security Critic flags vulnerable generated code
	t.Run("SecurityCritic_Flagging", func(t *testing.T) {
		payload := map[string]interface{}{
			"model": "pool/agent-coding",
			"messages": []map[string]string{
				{"role": "user", "content": "Generate a vulnerable database query function"},
			},
		}
		data, _ := json.Marshal(payload)

		req, _ := http.NewRequest(http.MethodPost, fmt.Sprintf("http://%s/v1/chat/completions", gwAddr), bytes.NewReader(data))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer sk-test-master-key")

		resp, err := client.Do(req)
		if err != nil {
			t.Fatalf("request failed: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("expected 200 OK, got %d", resp.StatusCode)
		}

		criticVerdict := resp.Header.Get("X-Tokman-Critic-Verdict")
		if criticVerdict != "FLAGGED" {
			t.Errorf("expected X-Tokman-Critic-Verdict: FLAGGED, got %s", criticVerdict)
		}

		respBody, _ := io.ReadAll(resp.Body)
		var compResp gateway.ChatCompletionResponse
		json.Unmarshal(respBody, &compResp)

		if len(compResp.Choices) == 0 {
			t.Fatalf("expected at least 1 choice")
		}
		if !strings.Contains(compResp.Choices[0].Message.Content, "TokMan Security Critic (`pool/security-tester`) Advisory") {
			t.Errorf("expected security advisory appended to completion output: %s", compResp.Choices[0].Message.Content)
		}

		t.Logf("\n=======================================================\nCRITIC HEADER: X-Tokman-Critic-Verdict: %s\n=======================================================\nRESPONSE DELIVERED TO USER WITH ADVISORY NOTICE:\n%s\n=======================================================", criticVerdict, compResp.Choices[0].Message.Content)
	})
}
