package interceptor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"tokman/backend/filter"
	"tokman/backend/shaper"
)

// Interceptor manages the rate limiting, context pruning, and domain pacing pipeline.
type Interceptor struct {
	Limiter      *shaper.LeakyBucketLimiter
	JitterEngine *shaper.DomainJitterEngine
	MaxContext   int
}

// NewInterceptor initializes a new Interceptor instance.
func NewInterceptor(limiter *shaper.LeakyBucketLimiter, jitter *shaper.DomainJitterEngine) *Interceptor {
	if limiter == nil {
		limiter = shaper.NewLeakyBucketLimiter(nil)
	}
	if jitter == nil {
		jitter = shaper.NewDomainJitterEngine(0, 0)
	}
	return &Interceptor{
		Limiter:      limiter,
		JitterEngine: jitter,
		MaxContext:   8192,
	}
}

// InterceptRequest processes an incoming request: resolves identity, checks rate limits,
// and prunes the context if necessary.
// Returns a boolean indicating if the request is permitted, and an optional error response.
func (ic *Interceptor) InterceptRequest(w http.ResponseWriter, r *http.Request) (*filter.ChatMessage, []filter.ChatMessage, bool) {
	identity, priority := shaper.ResolveIdentity(r)
	rateRes := ic.Limiter.Allow(identity, priority)

	w.Header().Set("X-RateLimit-Limit-RPM", strconv.Itoa(rateRes.LimitRPM))
	w.Header().Set("X-RateLimit-Remaining-RPM", strconv.Itoa(rateRes.Remaining))

	if !rateRes.Allowed {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Retry-After", strconv.Itoa(rateRes.RetryAfter))
		w.WriteHeader(http.StatusTooManyRequests)

		errResp := fmt.Sprintf(`{
  "error": {
    "message": "Rate limit exceeded for priority class %s. Quota: %d RPM. Retry in %d seconds.",
    "type": "requests_per_minute_exceeded",
    "code": 429
  }
}`, rateRes.Priority, rateRes.LimitRPM, rateRes.RetryAfter)
		w.Write([]byte(errResp))
		return nil, nil, false
	}

	return nil, nil, true
}

// PruneRequestBody parses the request body, executes dynamic context pruning on messages,
// and returns the modified body bytes and the pruned message slice.
func (ic *Interceptor) PruneRequestBody(bodyBytes []byte, maxTokens int) ([]byte, int, error) {
	if maxTokens <= 0 {
		maxTokens = ic.MaxContext
	}

	var req struct {
		Model       string               `json:"model"`
		Messages    []filter.ChatMessage `json:"messages"`
		Temperature *float64             `json:"temperature,omitempty"`
		MaxTokens   *int                 `json:"max_tokens,omitempty"`
		Stream      bool                 `json:"stream,omitempty"`
	}

	if err := json.Unmarshal(bodyBytes, &req); err != nil {
		return bodyBytes, 0, err
	}

	if len(req.Messages) == 0 {
		return bodyBytes, 0, nil
	}

	prunedMessages, dropped := filter.PruneMessages(req.Messages, maxTokens)
	if dropped > 0 {
		req.Messages = prunedMessages
		newBody, err := json.Marshal(req)
		if err != nil {
			return bodyBytes, 0, err
		}
		return newBody, dropped, nil
	}

	return bodyBytes, 0, nil
}

// ReplaceRequestBody resets r.Body with a new byte buffer.
func ReplaceRequestBody(r *http.Request, body []byte) {
	r.Body = io.NopCloser(bytes.NewReader(body))
	r.ContentLength = int64(len(body))
	r.Header.Set("Content-Length", strconv.Itoa(len(body)))
}
