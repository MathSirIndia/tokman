package shaper

import (
	"context"
	"math/rand"
	"net/url"
	"strings"
	"sync"
	"time"
)

// DomainJitterEngine enforces an inter-request delay floor and randomized jitter per upstream domain.
type DomainJitterEngine struct {
	mu           sync.Mutex
	lastDispatch map[string]time.Time
	baseFloor    time.Duration
	maxJitterMs  int
}

// NewDomainJitterEngine initializes a jitter engine with default 150ms floor and 200ms jitter.
func NewDomainJitterEngine(baseFloor time.Duration, maxJitterMs int) *DomainJitterEngine {
	if baseFloor <= 0 {
		baseFloor = 150 * time.Millisecond
	}
	if maxJitterMs <= 0 {
		maxJitterMs = 200
	}
	return &DomainJitterEngine{
		lastDispatch: make(map[string]time.Time),
		baseFloor:    baseFloor,
		maxJitterMs:  maxJitterMs,
	}
}

// ExtractDomain parses the host from a raw URL or domain string.
func ExtractDomain(rawURL string) string {
	if !strings.HasPrefix(rawURL, "http://") && !strings.HasPrefix(rawURL, "https://") {
		rawURL = "https://" + rawURL
	}
	parsed, err := url.Parse(rawURL)
	if err != nil || parsed.Host == "" {
		return "default"
	}
	return strings.ToLower(parsed.Host)
}

// WaitBeforeDispatch reserves a dispatch time slot for the domain and sleeps until that slot arrives.
// If the context expires before the delay completes, it returns ctx.Err().
func (e *DomainJitterEngine) WaitBeforeDispatch(ctx context.Context, domainOrURL string) error {
	domain := ExtractDomain(domainOrURL)

	e.mu.Lock()
	now := time.Now()
	last := e.lastDispatch[domain]

	// Compute randomized jitter for this slot
	jitter := time.Duration(rand.Intn(e.maxJitterMs)) * time.Millisecond
	requiredGap := e.baseFloor + jitter

	var sleepDuration time.Duration
	if last.IsZero() || now.Sub(last) >= requiredGap {
		// First request or sufficient time has elapsed
		e.lastDispatch[domain] = now
		sleepDuration = 0
	} else {
		// Needs delay to respect domain floor + jitter
		sleepDuration = requiredGap - now.Sub(last)
		e.lastDispatch[domain] = now.Add(sleepDuration)
	}
	e.mu.Unlock()

	if sleepDuration > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleepDuration):
			return nil
		}
	}

	return nil
}
