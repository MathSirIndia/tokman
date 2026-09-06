package shaper

import (
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

// PriorityClass defines the QoS tier for a client request.
type PriorityClass string

const (
	PriorityP0 PriorityClass = "P0" // Interactive Chat: 12 RPM
	PriorityP1 PriorityClass = "P1" // IDE / Coding Agents: 60 RPM
	PriorityP2 PriorityClass = "P2" // Deep Research: 5 RPM
	PriorityP3 PriorityClass = "P3" // Media & Bulk Jobs: 2 RPM
)

// RateLimitQuota defines maximum requests and tokens per minute for a tier.
type RateLimitQuota struct {
	RPM int
	TPM int
}

// DefaultQuotas maps priority tiers to default fair-share limits.
var DefaultQuotas = map[PriorityClass]RateLimitQuota{
	PriorityP0: {RPM: 12, TPM: 40000},
	PriorityP1: {RPM: 60, TPM: 150000},
	PriorityP2: {RPM: 5, TPM: 30000},
	PriorityP3: {RPM: 2, TPM: 20000},
}

// ClientBucket tracks timestamps of requests made within the current sliding window.
type ClientBucket struct {
	Timestamps []time.Time
}

// LeakyBucketLimiter enforces P0–P3 fair-share rate limits per client identity.
type LeakyBucketLimiter struct {
	mu      sync.Mutex
	quotas  map[PriorityClass]RateLimitQuota
	buckets map[string]*ClientBucket
	window  time.Duration
}

// NewLeakyBucketLimiter initializes a rate limiter with a sliding 60-second window.
func NewLeakyBucketLimiter(customQuotas map[PriorityClass]RateLimitQuota) *LeakyBucketLimiter {
	quotas := make(map[PriorityClass]RateLimitQuota)
	for k, v := range DefaultQuotas {
		quotas[k] = v
	}
	for k, v := range customQuotas {
		quotas[k] = v
	}

	return &LeakyBucketLimiter{
		quotas:  quotas,
		buckets: make(map[string]*ClientBucket),
		window:  60 * time.Second,
	}
}

// ResolveIdentity extracts the canonical client identifier and priority class from an HTTP request.
func ResolveIdentity(r *http.Request) (string, PriorityClass) {
	priority := PriorityP0
	if pHeader := r.Header.Get("X-Priority"); pHeader != "" {
		pUpper := PriorityClass(strings.ToUpper(strings.TrimSpace(pHeader)))
		switch pUpper {
		case PriorityP0, PriorityP1, PriorityP2, PriorityP3:
			priority = pUpper
		}
	}

	var identity string
	if uid := r.Header.Get("X-User-Id"); uid != "" {
		identity = "user:" + uid
	} else if tgID := r.Header.Get("X-Telegram-Chat-Id"); tgID != "" {
		identity = "telegram:" + tgID
	} else if discID := r.Header.Get("X-Discord-User-Id"); discID != "" {
		identity = "discord:" + discID
	} else if auth := r.Header.Get("Authorization"); auth != "" {
		token := strings.TrimPrefix(auth, "Bearer ")
		if len(token) > 12 {
			identity = "token:" + token[:12]
		} else {
			identity = "token:" + token
		}
	} else {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			host = r.RemoteAddr
		}
		identity = "ip:" + host
	}

	return identity, priority
}

// RateLimitResult contains the outcome of an Allow check.
type RateLimitResult struct {
	Allowed    bool
	RetryAfter int
	LimitRPM   int
	Remaining  int
	Priority   PriorityClass
	Identity   string
}

// Allow evaluates whether a client is permitted to make a request under their priority quota.
func (l *LeakyBucketLimiter) Allow(identity string, priority PriorityClass) RateLimitResult {
	l.mu.Lock()
	defer l.mu.Unlock()

	quota, exists := l.quotas[priority]
	if !exists {
		quota = DefaultQuotas[PriorityP0]
	}

	key := fmt.Sprintf("%s:%s", identity, priority)
	bucket, exists := l.buckets[key]
	if !exists {
		bucket = &ClientBucket{Timestamps: make([]time.Time, 0, quota.RPM)}
		l.buckets[key] = bucket
	}

	now := time.Now()
	cutoff := now.Add(-l.window)

	// Filter out expired timestamps
	valid := bucket.Timestamps[:0]
	for _, t := range bucket.Timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}
	bucket.Timestamps = valid

	if len(bucket.Timestamps) >= quota.RPM {
		oldest := bucket.Timestamps[0]
		retryAfterSec := int(oldest.Add(l.window).Sub(now).Seconds())
		if retryAfterSec < 1 {
			retryAfterSec = 1
		}

		return RateLimitResult{
			Allowed:    false,
			RetryAfter: retryAfterSec,
			LimitRPM:   quota.RPM,
			Remaining:  0,
			Priority:   priority,
			Identity:   identity,
		}
	}

	// Request allowed: record timestamp
	bucket.Timestamps = append(bucket.Timestamps, now)
	remaining := quota.RPM - len(bucket.Timestamps)

	return RateLimitResult{
		Allowed:    true,
		RetryAfter: 0,
		LimitRPM:   quota.RPM,
		Remaining:  remaining,
		Priority:   priority,
		Identity:   identity,
	}
}

// Reset clears recorded request buckets (useful for tests).
func (l *LeakyBucketLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]*ClientBucket)
}
