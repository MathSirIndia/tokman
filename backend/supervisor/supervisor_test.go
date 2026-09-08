package supervisor

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestSupervisor_RouteRequest_PoolAuto(t *testing.T) {
	sup := NewSupervisor(SupervisorConfig{
		ClassifierConfig: ClassifierConfig{ForceTier0: true},
		DefaultModel:     "llama-3.3-70b-versatile",
	})
	ctx := context.Background()

	// Capture telemetry events
	var events []TelemetryEvent
	var mu sync.Mutex
	sup.Telemetry.Subscribe(func(ev TelemetryEvent) {
		mu.Lock()
		defer mu.Unlock()
		events = append(events, ev)
	})

	// 1. Coding prompt routes to pool/agent-coding
	orchCode := sup.RouteRequest(ctx, "req-1", "pool/auto", "Implement a red-black tree in Go:\n```go\ntype Node struct{}\n```")
	if orchCode.TargetPool != "pool/agent-coding" {
		t.Errorf("expected pool/agent-coding, got %s", orchCode.TargetPool)
	}
	if orchCode.Scratchpad == nil {
		t.Errorf("expected initialized scratchpad")
	}

	// 2. Math prompt routes to pool/deep-reasoning
	orchMath := sup.RouteRequest(ctx, "req-2", "pool/auto", "Prove that there are infinitely many prime numbers")
	if orchMath.TargetPool != "pool/deep-reasoning" {
		t.Errorf("expected pool/deep-reasoning, got %s", orchMath.TargetPool)
	}

	// 3. Conversational prompt routes to pool/general
	orchGeneral := sup.RouteRequest(ctx, "req-3", "pool/auto", "Hi there, how is your day going?")
	if orchGeneral.TargetPool != "pool/general" {
		t.Errorf("expected pool/general, got %s", orchGeneral.TargetPool)
	}

	// Verify telemetry events arrived
	time.Sleep(10 * time.Millisecond)
	mu.Lock()
	count := len(events)
	mu.Unlock()
	if count < 3 {
		t.Errorf("expected at least 3 telemetry events, got %d", count)
	}
}

func TestSupervisor_PerformSecurityCritic(t *testing.T) {
	sup := NewSupervisor(SupervisorConfig{
		ClassifierConfig: ClassifierConfig{ForceTier0: true},
		CriticConfig:     CriticConfig{MaxCriticLoops: 2, Enabled: true},
	})
	ctx := context.Background()

	envelope := NewScratchpadEnvelope("req-critic-1", "Write code")
	envelope.SelectedPool = "pool/agent-coding"

	vulnerableCode := "```go\nquery := \"SELECT * FROM users WHERE id = '\" + id + \"'\"\n```"
	finalCode, verdict := sup.PerformSecurityCritic(ctx, envelope, vulnerableCode)

	if verdict.IsClean {
		t.Errorf("expected vulnerable code to be flagged")
	}
	if !strings.Contains(finalCode, "TokMan Security Critic (`pool/security-tester`) Advisory") {
		t.Errorf("expected security advisory note appended to code")
	}
	if len(envelope.SecurityWarnings) == 0 {
		t.Errorf("expected security warning recorded in scratchpad")
	}
}
