package mocks

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"time"
)

// MockChatMessage represents a chat message in the mock completion.
type MockChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// MockChoice represents a completion choice in the mock response.
type MockChoice struct {
	Index        int             `json:"index"`
	Message      MockChatMessage `json:"message"`
	FinishReason string          `json:"finish_reason"`
}

// MockUsage reports token metrics for mock responses.
type MockUsage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// MockChatCompletionResponse mimics an OpenAI/Groq/Gemini response structure.
type MockChatCompletionResponse struct {
	ID      string       `json:"id"`
	Object  string       `json:"object"`
	Created int64        `json:"created"`
	Model   string       `json:"model"`
	Choices []MockChoice `json:"choices"`
	Usage   MockUsage    `json:"usage"`
}

// MockServerConfig defines runtime behavior of the mock upstream server.
type MockServerConfig struct {
	StatusCode       int
	SimulatedLatency time.Duration
	DefaultModel     string
	DefaultResponse  string
	RequireAuth      bool
	ExpectedBearer   string
}

// MockUpstream provides an offline, zero-credential upstream test server.
type MockUpstream struct {
	server    *httptest.Server
	mu        sync.RWMutex
	config    MockServerConfig
	callCount int64
}

// NewMockUpstream starts an offline mock upstream HTTP server with default configuration.
func NewMockUpstream() *MockUpstream {
	m := &MockUpstream{
		config: MockServerConfig{
			StatusCode:       http.StatusOK,
			SimulatedLatency: 0,
			DefaultModel:     "llama-3.3-70b-versatile",
			DefaultResponse:  "Mock response from zero-credential upstream engine.",
			RequireAuth:      false,
		},
	}

	m.server = httptest.NewServer(http.HandlerFunc(m.handleRequest))
	return m
}

// SetConfig updates the mock server configuration dynamically.
func (m *MockUpstream) SetConfig(cfg MockServerConfig) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config = cfg
}

// SetStatusCode changes the HTTP status code returned by the mock server.
func (m *MockUpstream) SetStatusCode(code int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.StatusCode = code
}

// SetLatency sets simulated processing latency before replying.
func (m *MockUpstream) SetLatency(d time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.SimulatedLatency = d
}

// SetRequireAuth enforces Authorization Bearer check.
func (m *MockUpstream) SetRequireAuth(require bool, expectedToken string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.config.RequireAuth = require
	m.config.ExpectedBearer = expectedToken
}

// CallCount returns the total number of requests received by the mock server.
func (m *MockUpstream) CallCount() int64 {
	return atomic.LoadInt64(&m.callCount)
}

// ResetCallCount resets the counter to 0.
func (m *MockUpstream) ResetCallCount() {
	atomic.StoreInt64(&m.callCount, 0)
}

// URL returns the loopback URL of the mock upstream server.
func (m *MockUpstream) URL() string {
	return m.server.URL
}

// Close gracefully terminates the mock HTTP server.
func (m *MockUpstream) Close() {
	m.server.Close()
}

func (m *MockUpstream) handleRequest(w http.ResponseWriter, r *http.Request) {
	atomic.AddInt64(&m.callCount, 1)

	m.mu.RLock()
	cfg := m.config
	m.mu.RUnlock()

	if cfg.SimulatedLatency > 0 {
		time.Sleep(cfg.SimulatedLatency)
	}

	// Verify Authorization if configured
	if cfg.RequireAuth {
		authHeader := r.Header.Get("Authorization")
		expected := "Bearer " + cfg.ExpectedBearer
		if authHeader != expected {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			w.Write([]byte(`{"error":{"message":"Invalid API Key","type":"invalid_request_error","code":"invalid_api_key"}}`))
			return
		}
	}

	// Return configured status code if non-200
	if cfg.StatusCode != http.StatusOK {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(cfg.StatusCode)
		if cfg.StatusCode == http.StatusTooManyRequests {
			w.Write([]byte(`{"error":{"message":"Rate limit reached for requests","type":"tokens","code":"rate_limit_exceeded"}}`))
		} else {
			w.Write([]byte(fmt.Sprintf(`{"error":{"message":"Upstream simulated error %d","code":%d}}`, cfg.StatusCode, cfg.StatusCode)))
		}
		return
	}

	// Standard Chat Completion Response
	resp := MockChatCompletionResponse{
		ID:      fmt.Sprintf("chatcmpl-mock-%d", atomic.LoadInt64(&m.callCount)),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   cfg.DefaultModel,
		Choices: []MockChoice{
			{
				Index: 0,
				Message: MockChatMessage{
					Role:    "assistant",
					Content: cfg.DefaultResponse,
				},
				FinishReason: "stop",
			},
		},
		Usage: MockUsage{
			PromptTokens:     12,
			CompletionTokens: 10,
			TotalTokens:      22,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
