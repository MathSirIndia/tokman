package shaper

import (
	"context"
	"testing"
	"time"
)

func TestExtractDomain(t *testing.T) {
	cases := map[string]string{
		"https://api.groq.com/openai/v1/chat/completions": "api.groq.com",
		"http://api.cerebras.ai/v1":                       "api.cerebras.ai",
		"api.openai.com":                                  "api.openai.com",
		"":                                                "default",
	}

	for input, expected := range cases {
		actual := ExtractDomain(input)
		if actual != expected {
			t.Errorf("input %q: expected %q, got %q", input, expected, actual)
		}
	}
}

func TestDomainJitterEngine_PacesConsecutiveRequests(t *testing.T) {
	// Set 50ms floor and 10ms jitter for fast unit testing
	floor := 50 * time.Millisecond
	jitterMs := 10
	engine := NewDomainJitterEngine(floor, jitterMs)

	ctx := context.Background()
	domain := "api.groq.com"

	// First call should be immediate
	t0 := time.Now()
	if err := engine.WaitBeforeDispatch(ctx, domain); err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	d0 := time.Since(t0)
	if d0 > 15*time.Millisecond {
		t.Errorf("first dispatch took unexpectedly long: %v", d0)
	}

	// Second consecutive call should wait at least the 50ms floor
	t1 := time.Now()
	if err := engine.WaitBeforeDispatch(ctx, domain); err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	d1 := time.Since(t1)
	if d1 < floor {
		t.Errorf("second dispatch did not observe floor: took %v, expected >= %v", d1, floor)
	}
}

func TestDomainJitterEngine_IndependentDomains(t *testing.T) {
	engine := NewDomainJitterEngine(100*time.Millisecond, 10)
	ctx := context.Background()

	// Dispatch to Domain A
	_ = engine.WaitBeforeDispatch(ctx, "api.groq.com")

	// Immediate dispatch to Domain B should NOT wait for Domain A
	t0 := time.Now()
	if err := engine.WaitBeforeDispatch(ctx, "api.cerebras.ai"); err != nil {
		t.Fatalf("call to independent domain failed: %v", err)
	}
	elapsed := time.Since(t0)
	if elapsed > 25*time.Millisecond {
		t.Errorf("call to independent domain was delayed: %v", elapsed)
	}
}
