package shaper

import (
	"net/http/httptest"
	"testing"
	"time"
)

func TestResolveIdentity(t *testing.T) {
	// 1. Channel 'ide' maps to P1
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req1.Header.Set("X-User-Id", "usr-abc-123")
	req1.Header.Set("X-Channel", "ide")
	id1, p1 := ResolveIdentity(req1)
	if id1 != "user:usr-abc-123" || p1 != PriorityP1 {
		t.Errorf("expected user:usr-abc-123 / P1, got %s / %s", id1, p1)
	}

	// 2. Untrusted client attempting to self-assign X-Priority: P1 without X-Tokman-Internal remains P0 (SEC-5)
	reqUnauthPrio := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	reqUnauthPrio.Header.Set("X-User-Id", "usr-attacker")
	reqUnauthPrio.Header.Set("X-Priority", "P1")
	idUnauth, pUnauth := ResolveIdentity(reqUnauthPrio)
	if idUnauth != "user:usr-attacker" || pUnauth != PriorityP0 {
		t.Errorf("expected unauthenticated priority to default to P0, got %s / %s", idUnauth, pUnauth)
	}

	// 3. Trusted internal caller with X-Tokman-Internal: true can set X-Priority: P1
	reqInternal := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	reqInternal.Header.Set("X-Tokman-Internal", "true")
	reqInternal.Header.Set("X-Priority", "P1")
	_, pInternal := ResolveIdentity(reqInternal)
	if pInternal != PriorityP1 {
		t.Errorf("expected internal priority override to be P1, got %s", pInternal)
	}

	// 4. Telegram ID maps to P0
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req2.Header.Set("X-Telegram-Chat-Id", "987654321")
	id2, p2 := ResolveIdentity(req2)
	if id2 != "telegram:987654321" || p2 != PriorityP0 {
		t.Errorf("expected telegram:987654321 / P0, got %s / %s", id2, p2)
	}

	// 5. Fallback IP maps to P0
	req3 := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req3.RemoteAddr = "192.168.1.50:54321"
	id3, p3 := ResolveIdentity(req3)
	if id3 != "ip:192.168.1.50" || p3 != PriorityP0 {
		t.Errorf("expected ip:192.168.1.50 / P0, got %s / %s", id3, p3)
	}
}

func TestLeakyBucketLimiter_EvictStale(t *testing.T) {
	limiter := NewLeakyBucketLimiter(nil)
	limiter.Allow("user:stale1", PriorityP0)
	limiter.Allow("user:stale2", PriorityP0)

	if limiter.BucketCount() != 2 {
		t.Fatalf("expected 2 active buckets, got %d", limiter.BucketCount())
	}

	// Fast-forward 70s into future
	futureTime := time.Now().Add(70 * time.Second)
	evicted := limiter.EvictStaleBuckets(futureTime)
	if evicted != 2 {
		t.Errorf("expected 2 evicted buckets, got %d", evicted)
	}
	if limiter.BucketCount() != 0 {
		t.Errorf("expected 0 remaining buckets, got %d", limiter.BucketCount())
	}
}

func TestLeakyBucketLimiter_P0Quota(t *testing.T) {
	limiter := NewLeakyBucketLimiter(nil)
	identity := "user:test-interactive"

	// First 12 requests should be allowed (P0 RPM = 12)
	for i := 1; i <= 12; i++ {
		res := limiter.Allow(identity, PriorityP0)
		if !res.Allowed {
			t.Fatalf("request %d was unexpectedly blocked", i)
		}
		expectedRemaining := 12 - i
		if res.Remaining != expectedRemaining {
			t.Errorf("request %d: expected remaining %d, got %d", i, expectedRemaining, res.Remaining)
		}
	}

	// 13th request must be blocked
	blockedRes := limiter.Allow(identity, PriorityP0)
	if blockedRes.Allowed {
		t.Fatalf("13th request should be blocked under P0 quota")
	}
	if blockedRes.RetryAfter <= 0 {
		t.Errorf("expected positive RetryAfter, got %d", blockedRes.RetryAfter)
	}
	if blockedRes.Remaining != 0 {
		t.Errorf("expected remaining 0, got %d", blockedRes.Remaining)
	}
}

func TestLeakyBucketLimiter_IsolationBetweenUsers(t *testing.T) {
	limiter := NewLeakyBucketLimiter(nil)

	userA := "user:alice"
	userB := "user:bob"

	// Exhaust userA's quota (12 requests)
	for i := 0; i < 12; i++ {
		limiter.Allow(userA, PriorityP0)
	}
	if limiter.Allow(userA, PriorityP0).Allowed {
		t.Errorf("userA should be blocked")
	}

	// userB should still be allowed
	resB := limiter.Allow(userB, PriorityP0)
	if !resB.Allowed {
		t.Errorf("userB should not be affected by userA's rate limit")
	}
}
