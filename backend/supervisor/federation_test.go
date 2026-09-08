package supervisor

import (
	"testing"
)

func TestFederationEngine_MasterCatalog(t *testing.T) {
	fed := NewFederationEngine("llama-3.3-70b-versatile")

	// Ensure all 12 core pools have complete 3-tier definitions (4x4x4)
	corePools := []string{
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
	}

	for _, pool := range corePools {
		tiers, err := fed.GetPoolTiers(pool)
		if err != nil {
			t.Fatalf("failed to get tiers for %s: %v", pool, err)
		}
		if len(tiers) != 3 {
			t.Errorf("%s must have exactly 3 tiers, got %d", pool, len(tiers))
		}
		for i, tier := range tiers {
			if len(tier.Models) != 4 {
				t.Errorf("%s tier %d must have 4 models, got %d", pool, i, len(tier.Models))
			}
		}
	}
}

func TestFederationEngine_ResolvePoolModel(t *testing.T) {
	fed := NewFederationEngine("llama-3.3-70b-versatile")

	cases := []struct {
		pool          string
		expectedModel string
	}{
		{"pool/general", "groq/llama-3.3-70b"},
		{"pool/deep-reasoning", "sambanova/deepseek-r1"},
		{"pool/agent-coding", "sambanova/qwen-2.5-coder"},
		{"pool/document-analysis", "gemini/gemini-2.0-flash"},
		{"pool/security-tester", "sambanova/qwen-coder"},
	}

	for _, tc := range cases {
		model, err := fed.ResolvePoolModel(tc.pool)
		if err != nil {
			t.Fatalf("unexpected error resolving %s: %v", tc.pool, err)
		}
		if model != tc.expectedModel {
			t.Errorf("for %s: expected %s, got %s", tc.pool, tc.expectedModel, model)
		}
	}
}

func TestFederationEngine_IntraPoolFailover(t *testing.T) {
	fed := NewFederationEngine("llama-3.3-70b-versatile")
	pool := "pool/general"

	// From Primary model 0 -> Primary model 1
	nextModel, nextTier, nextIdx, hasNext := fed.GetNextFallbackModel(pool, 0, 0)
	if !hasNext || nextTier != 0 || nextIdx != 1 || nextModel != "cerebras/llama-3.3-70b" {
		t.Errorf("expected primary model 1, got %s (tier %d, idx %d)", nextModel, nextTier, nextIdx)
	}

	// From Primary model 3 -> Backup Tier 1 model 0
	nextModel, nextTier, nextIdx, hasNext = fed.GetNextFallbackModel(pool, 0, 3)
	if !hasNext || nextTier != 1 || nextIdx != 0 || nextModel != "github/gpt-4o-mini" {
		t.Errorf("expected backup tier 1 model 0, got %s (tier %d, idx %d)", nextModel, nextTier, nextIdx)
	}

	// From Backup Tier 2 model 3 -> Exhausted (no more models)
	_, _, _, hasNext = fed.GetNextFallbackModel(pool, 2, 3)
	if hasNext {
		t.Errorf("expected exhausted fallback array, got hasNext=true")
	}
}

func TestMapToGroqUpstream(t *testing.T) {
	cases := []struct {
		catalogModel string
		expectedGroq string
	}{
		{"sambanova/deepseek-r1", "openai/gpt-oss-120b"},
		{"groq/deepseek-r1-70b", "openai/gpt-oss-120b"},
		{"sambanova/qwen-2.5-coder", "openai/gpt-oss-120b"},
		{"cf/openai-whisper", "whisper-large-v3"},
		{"groq/llama-3.3-70b", "openai/gpt-oss-120b"},
	}

	for _, tc := range cases {
		mapped := MapToGroqUpstream(tc.catalogModel)
		if mapped != tc.expectedGroq {
			t.Errorf("for %s: expected %s, got %s", tc.catalogModel, tc.expectedGroq, mapped)
		}
	}
}
