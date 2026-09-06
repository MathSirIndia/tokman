package shaper

import (
	"net/http/httptest"
	"testing"
)

func TestResolveIdentity(t *testing.T) {
	// 1. X-User-Id
	req1 := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req1.Header.Set("X-User-Id", "usr-abc-123")
	req1.Header.Set("X-Priority", "P1")
	id1, p1 := ResolveIdentity(req1)
	if id1 != "user:usr-abc-123" || p1 != PriorityP1 {
		t.Errorf("expected user:usr-abc-123 / P1, got %s / %s", id1, p1)
	}

	// 2. Telegram ID
	req2 := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req2.Header.Set("X-Telegram-Chat-Id", "987654321")
	id2, p2 := ResolveIdentity(req2)
	if id2 != "telegram:987654321" || p2 != PriorityP0 {
		t.Errorf("expected telegram:987654321 / P0, got %s / %s", id2, p2)
	}

	// 3. Fallback IP
	req3 := httptest.NewRequest("POST", "/v1/chat/completions", nil)
	req3.RemoteAddr = "192.168.1.50:54321"
	id3, p3 := ResolveIdentity(req3)
	if id3 != "ip:192.168.1.50" || p3 != PriorityP0 {
		t.Errorf("expected ip:192.168.1.50 / P0, got %s / %s", id3, p3)
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
