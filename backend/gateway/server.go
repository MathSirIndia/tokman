package gateway

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"tokman/backend/filter"
	"tokman/backend/interceptor"
	"tokman/backend/storage"
)

// CanonicalPools defines the 14 capability pools from docs/report.md
var CanonicalPools = []string{
	"pool/auto",
	"pool/general",
	"pool/deep-reasoning",
	"pool/agent-coding",
	"pool/document-analysis",
	"pool/web-research",
	"pool/presentation",
	"pool/image-gen",
	"pool/architect",
	"pool/security-tester",
	"pool/stack-optimizer",
	"pool/document-gen",
	"pool/audio-gen",
	"pool/video-conductor",
}

// ChatMessage represents a single turn in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatCompletionRequest represents an OpenAI-compatible request payload.
type ChatCompletionRequest struct {
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Temperature *float64      `json:"temperature,omitempty"`
	MaxTokens   *int          `json:"max_tokens,omitempty"`
	Stream      bool          `json:"stream,omitempty"`
}

// Choice represents a completion choice.
type Choice struct {
	Index        int         `json:"index"`
	Message      ChatMessage `json:"message"`
	FinishReason string      `json:"finish_reason"`
}

// Usage reports token metrics.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ChatCompletionResponse represents an OpenAI-compatible completion response.
type ChatCompletionResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// ModelEntry represents an available model in /v1/models.
type ModelEntry struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	OwnedBy string `json:"owned_by"`
}

// ModelListResponse represents the /v1/models response.
type ModelListResponse struct {
	Object string       `json:"object"`
	Data   []ModelEntry `json:"data"`
}

// Config configures the Go HTTP Gateway server.
type Config struct {
	Port         int
	MasterKey    string
	GroqAPIKey   string
	UpstreamURL  string // Optional override for testing (defaults to Groq API)
	Cache        *storage.LRUCache
	DB           *storage.DB
	Interceptor  *interceptor.Interceptor
}

// Server is the native Gateway HTTP server.
type Server struct {
	Config      Config
	httpServer  *http.Server
	listener    net.Listener
	startTime   time.Time
	mu          sync.RWMutex
	groqKey     string
	Interceptor *interceptor.Interceptor
}

// NewServer initializes a new Gateway HTTP server instance.
func NewServer(cfg Config) *Server {
	if cfg.Port <= 0 {
		cfg.Port = 8000
	}
	if cfg.Cache == nil {
		cfg.Cache = storage.NewLRUCache(1000, 24*time.Hour)
	}
	if cfg.UpstreamURL == "" {
		cfg.UpstreamURL = "https://api.groq.com/openai/v1/chat/completions"
	}
	ic := cfg.Interceptor
	if ic == nil {
		ic = interceptor.NewInterceptor(nil, nil)
	}

	s := &Server{
		Config:      cfg,
		startTime:   time.Now(),
		groqKey:     cfg.GroqAPIKey,
		Interceptor: ic,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/health/readiness", s.handleReadiness)
	mux.HandleFunc("/health/liveness", s.handleLiveness)
	mux.HandleFunc("/v1/models", s.handleModels)
	mux.HandleFunc("/v1/chat/completions", s.handleChatCompletions)
	mux.HandleFunc("/admin", s.handleAdmin)

	s.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: s.corsMiddleware(mux),
	}

	return s
}

// SetGroqAPIKey updates the active Groq key dynamically.
func (s *Server) SetGroqAPIKey(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.groqKey = key
}

// GetGroqAPIKey gets the current Groq API key safely.
func (s *Server) GetGroqAPIKey() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.groqKey != "" {
		return s.groqKey
	}
	return os.Getenv("GROQ_API_KEY")
}

// Start runs the HTTP server. If listener is not set, it listens on configured port.
func (s *Server) Start() error {
	var err error
	if s.listener == nil {
		s.listener, err = net.Listen("tcp", s.httpServer.Addr)
		if err != nil {
			return err
		}
	}
	return s.httpServer.Serve(s.listener)
}

// SetListener allows injecting a custom net.Listener (useful for testing).
func (s *Server) SetListener(l net.Listener) {
	s.listener = l
}

// Addr returns the bound address of the server.
func (s *Server) Addr() string {
	if s.listener != nil {
		return s.listener.Addr().String()
	}
	return s.httpServer.Addr
}

// Shutdown gracefully terminates the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Master-Key")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleReadiness(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"status":         "healthy",
		"version":        "1.0.0",
		"pools":          CanonicalPools,
		"uptime_seconds": int(time.Since(s.startTime).Seconds()),
		"cache_entries":  s.Config.Cache.Len(),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleLiveness(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"alive"}`))
}

//go:embed admin.html
var adminHTML []byte

func (s *Server) handleAdmin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(adminHTML)
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	data := make([]ModelEntry, 0, len(CanonicalPools))
	for _, id := range CanonicalPools {
		data = append(data, ModelEntry{
			ID:      id,
			Object:  "model",
			Created: s.startTime.Unix(),
			OwnedBy: "tokman-mesh",
		})
	}

	models := ModelListResponse{
		Object: "list",
		Data:   data,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(models)
}

func (s *Server) handleChatCompletions(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Verify Authorization header if MasterKey is set
	if s.Config.MasterKey != "" {
		auth := r.Header.Get("Authorization")
		expected := "Bearer " + s.Config.MasterKey
		if auth != expected && auth != s.Config.MasterKey {
			http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
			return
		}
	}

	// Rate limiting & identity evaluation
	if s.Interceptor != nil {
		_, _, allowed := s.Interceptor.InterceptRequest(w, r)
		if !allowed {
			return
		}
	}

	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, `{"error": "Failed to read request body"}`, http.StatusBadRequest)
		return
	}

	// Dynamic context pruning
	if s.Interceptor != nil {
		if prunedBytes, dropped, err := s.Interceptor.PruneRequestBody(bodyBytes, 8192); err == nil && dropped > 0 {
			bodyBytes = prunedBytes
		}
	}

	var req ChatCompletionRequest
	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		http.Error(w, `{"error": "Invalid JSON schema"}`, http.StatusBadRequest)
		return
	}

	// Resolve capability pool to upstream model
	requestedPool := req.Model
	upstreamModel := req.Model
	switch req.Model {
	case "pool/general", "pool/auto":
		upstreamModel = "openai/gpt-oss-120b"
	case "pool/deep-reasoning":
		upstreamModel = "openai/gpt-oss-120b"
	case "pool/agent-coding", "pool/security-tester":
		upstreamModel = "openai/gpt-oss-120b"
	case "pool/document-analysis", "pool/architect", "pool/stack-optimizer", "pool/document-gen":
		upstreamModel = "openai/gpt-oss-120b"
	case "pool/web-research", "pool/presentation":
		upstreamModel = "openai/gpt-oss-120b"
	case "pool/image-gen", "pool/audio-gen", "pool/video-conductor":
		upstreamModel = "openai/gpt-oss-120b"
	}

	cacheKey := s.computeCacheKey(requestedPool, req.Messages)

	// Check in-memory LRU cache for non-streaming requests
	if !req.Stream {
		if cachedResp, found := s.Config.Cache.Get(cacheKey); found {
			if respBytes, ok := cachedResp.([]byte); ok {
				w.Header().Set("Content-Type", "application/json")
				w.Header().Set("X-Cache", "HIT")
				w.WriteHeader(http.StatusOK)
				w.Write(respBytes)

				// Log cached completion asynchronously
				if s.Config.DB != nil {
					go s.Config.DB.LogSpend(storage.SpendRecord{
						Model:     req.Model,
						LatencyMs: 0.5,
						Cached:    true,
					})
				}
				return
			}
		}
	}

	// Dispatch to upstream provider
	startTime := time.Now()
	req.Model = upstreamModel
	modifiedBody, err := json.Marshal(req)
	if err != nil {
		http.Error(w, `{"error": "Failed to prepare upstream payload"}`, http.StatusInternalServerError)
		return
	}

	upstreamReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost, s.Config.UpstreamURL, bytes.NewReader(modifiedBody))
	if err != nil {
		http.Error(w, `{"error": "Failed to create upstream request"}`, http.StatusInternalServerError)
		return
	}

	upstreamReq.Header.Set("Content-Type", "application/json")
	apiKey := s.GetGroqAPIKey()
	if apiKey != "" {
		upstreamReq.Header.Set("Authorization", "Bearer "+apiKey)
	}

	// Inter-request jitter and domain pacing
	if s.Interceptor != nil && s.Interceptor.JitterEngine != nil {
		_ = s.Interceptor.JitterEngine.WaitBeforeDispatch(r.Context(), s.Config.UpstreamURL)
	}

	client := &http.Client{Timeout: 60 * time.Second}
	upstreamResp, err := client.Do(upstreamReq)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Upstream error: %s"}`, maskSensitiveCredentials(err.Error())), http.StatusBadGateway)
		return
	}
	defer upstreamResp.Body.Close()

	if req.Stream {
		// Flush SSE chunks directly to client with stateful <think> filter
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("X-Cache", "MISS")
		w.WriteHeader(upstreamResp.StatusCode)

		flusher, ok := w.(http.Flusher)
		streamFilter := filter.NewStreamFilter()
		reader := bufio.NewReader(upstreamResp.Body)

		for {
			line, err := reader.ReadBytes('\n')
			if len(line) > 0 {
				filtered := streamFilter.FilterSSEChunk(line)
				if len(filtered) > 0 {
					w.Write(filtered)
					if ok {
						flusher.Flush()
					}
				}
			}
			if err != nil {
				break
			}
		}
		return
	}

	// Read full non-streaming response
	respBytes, err := io.ReadAll(upstreamResp.Body)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"error": "Failed to read upstream response: %s"}`, maskSensitiveCredentials(err.Error())), http.StatusBadGateway)
		return
	}

	// Rewrite model identifier back to requested pool name in output and sanitize thinking tokens
	if upstreamResp.StatusCode == http.StatusOK {
		var completionResp ChatCompletionResponse
		if err := json.Unmarshal(respBytes, &completionResp); err == nil {
			completionResp.Model = requestedPool
			for i := range completionResp.Choices {
				completionResp.Choices[i].Message.Content = filter.SanitizeNonStreamingContent(completionResp.Choices[i].Message.Content)
			}
			if rewritten, err := json.Marshal(completionResp); err == nil {
				respBytes = rewritten
			}

			// Store in LRU cache
			s.Config.Cache.Set(cacheKey, respBytes)

			// Record spend log in SQLite
			durationMs := float64(time.Since(startTime).Microseconds()) / 1000.0
			if s.Config.DB != nil {
				go s.Config.DB.LogSpend(storage.SpendRecord{
					Model:            requestedPool,
					PromptTokens:     completionResp.Usage.PromptTokens,
					CompletionTokens: completionResp.Usage.CompletionTokens,
					TotalTokens:      completionResp.Usage.TotalTokens,
					LatencyMs:        durationMs,
					Cached:           false,
				})
			}
		}
	}

	// Set exact sanitized headers for client
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(respBytes)))
	w.Header().Set("X-Cache", "MISS")
	w.WriteHeader(upstreamResp.StatusCode)
	w.Write(respBytes)
}

func (s *Server) computeCacheKey(model string, messages []ChatMessage) string {
	h := sha256.New()
	h.Write([]byte(model))
	for _, m := range messages {
		h.Write([]byte(m.Role))
		h.Write([]byte(m.Content))
	}
	return hex.EncodeToString(h.Sum(nil))
}

// maskSensitiveCredentials masks any API keys or bearer tokens (G-1.4 resolution)
func maskSensitiveCredentials(input string) string {
	patterns := []string{"gsk_", "sk-", "AIzaSy", "Bearer "}
	result := input
	for _, p := range patterns {
		for {
			idx := strings.Index(result, p)
			if idx == -1 {
				break
			}
			start := idx
			if p == "Bearer " {
				start = idx + len("Bearer ")
			}
			end := start + len(p)
			if p == "Bearer " {
				end = start
			}
			for end < len(result) && (result[end] >= 'a' && result[end] <= 'z' ||
				result[end] >= 'A' && result[end] <= 'Z' ||
				result[end] >= '0' && result[end] <= '9' ||
				result[end] == '-' || result[end] == '_') {
				end++
			}
			if end == start {
				break
			}
			result = result[:start] + "***" + result[end:]
		}
	}
	return result
}
