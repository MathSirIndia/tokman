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
	TPM int // Reserved for token-budget enforcement (Module 7)
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
	mu           sync.Mutex
	quotas       map[PriorityClass]RateLimitQuota
	buckets      map[string]*ClientBucket
	window       time.Duration
	lastEviction time.Time
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
		quotas:       quotas,
		buckets:      make(map[string]*ClientBucket),
		window:       60 * time.Second,
		lastEviction: time.Now(),
	}
}

// ResolveIdentity extracts the canonical client identifier and priority class from an HTTP request.
// Security: Priority is resolved server-side from X-Channel or trusted internal mesh contexts (SEC-5).
func ResolveIdentity(r *http.Request) (string, PriorityClass) {
	priority := PriorityP0

	channel := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Channel")))
	switch channel {
	case "ide", "cursor", "aider", "vscode", "agent":
		priority = PriorityP1
	case "research", "deep-reasoning":
		priority = PriorityP2
	case "media", "conductor", "bulk":
		priority = PriorityP3
	default:
		// Direct X-Priority override is restricted strictly to internal mesh callers
		if r.Header.Get("X-Tokman-Internal") == "true" {
			if pHeader := r.Header.Get("X-Priority"); pHeader != "" {
				pUpper := PriorityClass(strings.ToUpper(strings.TrimSpace(pHeader)))
				switch pUpper {
				case PriorityP0, PriorityP1, PriorityP2, PriorityP3:
					priority = pUpper
				}
			}
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

	now := time.Now()

	// Periodic stale bucket cleanup (BP-1: prevents unbounded memory growth)
	if now.Sub(l.lastEviction) >= l.window || len(l.buckets) >= 500 {
		l.evictStaleLocked(now)
		l.lastEviction = now
	}

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

// evictStaleLocked removes buckets that have no active timestamps within the sliding window.
// Caller must hold l.mu.
func (l *LeakyBucketLimiter) evictStaleLocked(now time.Time) int {
	cutoff := now.Add(-l.window)
	evicted := 0
	for key, b := range l.buckets {
		valid := b.Timestamps[:0]
		for _, t := range b.Timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		b.Timestamps = valid
		if len(b.Timestamps) == 0 {
			delete(l.buckets, key)
			evicted++
		}
	}
	return evicted
}

// EvictStaleBuckets triggers eviction of expired buckets and returns the count of purged entries.
func (l *LeakyBucketLimiter) EvictStaleBuckets(now time.Time) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.evictStaleLocked(now)
}

// BucketCount returns the current count of allocated buckets.
func (l *LeakyBucketLimiter) BucketCount() int {
	l.mu.Lock()
	defer l.mu.Unlock()
	return len(l.buckets)
}

// Reset clears recorded request buckets (useful for tests).
func (l *LeakyBucketLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string]*ClientBucket)
}
