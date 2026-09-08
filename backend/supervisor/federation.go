package supervisor

import (
	"fmt"
	"strings"
	"sync"
)

// ModelTier represents a tier containing 4 candidate models.
type ModelTier struct {
	TierNumber int      `json:"tier_number"` // 0: Primary, 1: Backup 1, 2: Backup 2
	Name       string   `json:"name"`        // e.g. "Primary Tier", "Backup Tier 1"
	Models     []string `json:"models"`      // 4 model identifiers
}

// PoolDefinition encapsulates the 3-Tier Multi-Model Fallback Array for a capability pool.
type PoolDefinition struct {
	Name        string      `json:"name"`
	Purpose     string      `json:"purpose"`
	PrimaryTier ModelTier   `json:"primary_tier"`
	BackupTier1 ModelTier   `json:"backup_tier_1"`
	BackupTier2 ModelTier   `json:"backup_tier_2"`
}

// MasterPoolCatalog contains the full 144-model capability catalog across all 14 pools.
var MasterPoolCatalog = map[string]PoolDefinition{
	"pool/general": {
		Name:    "pool/general",
		Purpose: "Frontline Chat, Summarization & Balanced Reasoning",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"groq/llama-3.3-70b", "cerebras/llama-3.3-70b", "sambanova/deepseek-v3", "gemini/gemini-2.0-flash"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"github/gpt-4o-mini", "sambanova/llama-3.3-70b", "mistral/mistral-small", "openrouter/gemini-flash"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"cloudflare/llama-3.3-70b", "openrouter/free-models", "huggingface/llama-3.3", "github/phi-4"},
		},
	},
	"pool/deep-reasoning": {
		Name:    "pool/deep-reasoning",
		Purpose: "Math, Formal Logic, Multi-turn Chain-of-Thought",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"sambanova/deepseek-r1", "groq/deepseek-r1-70b", "github/deepseek-r1", "gemini/gemini-2.0-think"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"cloudflare/deepseek-r1", "huggingface/deepseek-r1", "sambanova/qwen-72b-inst", "openrouter/deepseek-r1"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"github/phi-4-reasoning", "groq/llama-3.3-spec", "mistral/mistral-large", "openrouter/qwen-72b"},
		},
	},
	"pool/agent-coding": {
		Name:    "pool/agent-coding",
		Purpose: "Autonomous Coding, Refactoring, Debugging & Syntax Analysis",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"sambanova/qwen-2.5-coder", "mistral/codestral", "github/gpt-4o-mini", "gemini/gemini-2.0-flash"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"cloudflare/qwen-2.5-coder", "huggingface/qwen-coder", "sambanova/llama-3.3-70b", "openrouter/qwen-coder"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"groq/llama-3.3-70b", "cerebras/llama-3.3-70b", "github/llama-3.3-70b", "openrouter/deepseek-v3"},
		},
	},
	"pool/document-analysis": {
		Name:    "pool/document-analysis",
		Purpose: "Massive Context Ingestion & Deep Document Reasoning (1M tokens)",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"gemini/gemini-2.0-flash", "gemini/gemini-flash-lite", "gemini/gemini-1.5-flash", "openrouter/gemini-flash"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"groq/llama-3.3-70b", "mistral/mistral-small", "github/llama-3.3-70b", "github/gpt-4o-mini"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"sambanova/llama-3.3-70b", "sambanova/deepseek-v3", "cerebras/llama-3.3-70b", "openrouter/llama-70b"},
		},
	},
	"pool/web-research": {
		Name:    "pool/web-research",
		Purpose: "SearXNG Grounded Search & Live Web Synthesis",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"cerebras/llama-3.3-70b", "groq/llama-3.3-70b", "gemini/gemini-2.0-flash", "groq/llama-3.1-8b"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"github/phi-4", "mistral/mistral-small", "sambanova/llama-3.3-70b", "github/gpt-4o-mini"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"cloudflare/llama-3.3", "openrouter/free-models", "huggingface/llama-3.3", "openrouter/gemini-flash"},
		},
	},
	"pool/presentation": {
		Name:    "pool/presentation",
		Purpose: "Marp & Reveal.js Presentation Deck Synthesis",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"groq/llama-3.3-70b", "cerebras/llama-3.3-70b", "github/gpt-4o-mini", "gemini/gemini-2.0-flash"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"sambanova/llama-3.3-70b", "mistral/mistral-small", "github/llama-3.3-70b", "sambanova/qwen-coder"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"openrouter/llama-70b", "cloudflare/llama-3.3", "huggingface/llama-3.3", "openrouter/deepseek-v3"},
		},
	},
	"pool/image-gen": {
		Name:    "pool/image-gen",
		Purpose: "Cloudflare Flux.1 Schnell / SDXL Visual Art & Keyframe Generation",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"cf/flux-1-schnell", "hf/flux-1-schnell", "hf/sdxl-base-1.0", "cf/sdxl-base-1.0"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"hf/stable-diffusion-3.5", "hf/sdxl-lightning", "cf/sdxl-lightning", "pollinations/flux-model"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"pollinations/turbo-gen", "hf/stable-diffusion-1.5", "cf/stable-diffusion-xl", "pollinations/midjourney"},
		},
	},
	"pool/architect": {
		Name:    "pool/architect",
		Purpose: "Video Conductor Scene Graph & Manifest Planning",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"gemini/gemini-2.0-flash", "gemini/gemini-2.0-think", "groq/llama-3.2-90b-vis", "sambanova/deepseek-r1"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"github/gpt-4o-mini", "mistral/pixtral-12b", "cf/llama-3.2-11b-vision", "openrouter/gemini-flash"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"huggingface/llama-11b-v", "gemini/gemini-1.5-flash", "sambanova/llama-3.3-70b", "cerebras/llama-3.3-70b"},
		},
	},
	"pool/security-tester": {
		Name:    "pool/security-tester",
		Purpose: "Automated Code Critic Loop & AST Vulnerability Review",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"sambanova/qwen-coder", "gemini/gemini-2.0-flash", "mistral/codestral", "github/gpt-4o-mini"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"cf/qwen-2.5-coder", "groq/deepseek-r1-70b", "mistral/pixtral-12b", "openrouter/qwen-coder"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"sambanova/deepseek-r1", "github/llama-3.3-70b", "huggingface/qwen-coder", "openrouter/deepseek-r1"},
		},
	},
	"pool/stack-optimizer": {
		Name:    "pool/stack-optimizer",
		Purpose: "Token Burn Telemetry & Zero-Touch Tuning Evaluator",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"gemini/gemini-2.0-flash", "sambanova/deepseek-r1", "groq/llama-3.3-70b", "github/gpt-4o-mini"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"cerebras/llama-3.3-70b", "mistral/mistral-small", "cf/llama-3.3-70b", "openrouter/gemini-flash"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"github/phi-4", "sambanova/qwen-72b-inst", "huggingface/llama-3.3", "openrouter/deepseek-v3"},
		},
	},
	"pool/document-gen": {
		Name:    "pool/document-gen",
		Purpose: "Headless Typst & Pandoc Mathematical Document Compilation",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"gemini/gemini-2.0-flash", "cerebras/llama-3.3-70b", "github/gpt-4o-mini", "gemini/gemini-flash-lite"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"groq/llama-3.3-70b", "sambanova/llama-3.3-70b", "mistral/codestral", "github/llama-3.3-70b"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"mistral/mistral-large", "openrouter/free-models", "cf/llama-3.3-70b", "huggingface/llama-3.3"},
		},
	},
	"pool/audio-gen": {
		Name:    "pool/audio-gen",
		Purpose: "Kokoro-82M Speech Synthesis & Cloudflare Whisper STT",
		PrimaryTier: ModelTier{
			TierNumber: 0,
			Name:       "Primary Tier",
			Models:     []string{"cf/openai-whisper", "hf/kokoro-82m-tts", "groq/whisper-large-v3", "cf/melo-tts-speech"},
		},
		BackupTier1: ModelTier{
			TierNumber: 1,
			Name:       "Backup Tier 1",
			Models:     []string{"groq/whisper-large-v3", "hf/xtts-v2-speech", "cf/fastspeech2-tts", "huggingface/speecht5"},
		},
		BackupTier2: ModelTier{
			TierNumber: 2,
			Name:       "Backup Tier 2",
			Models:     []string{"hf/mms-tts-speech", "pollinations/audio-synth", "deepgram/aura-free", "huggingface/piper-tts"},
		},
	},
}

// FederationEngine provides model resolution and intra-pool 3-tier failover.
type FederationEngine struct {
	mu           sync.RWMutex
	catalog      map[string]PoolDefinition
	defaultModel string
}

// NewFederationEngine initializes the federation engine with the Master 144-Model Catalog.
func NewFederationEngine(defaultModel string) *FederationEngine {
	if defaultModel == "" {
		defaultModel = "openai/gpt-oss-120b"
	}
	return &FederationEngine{
		catalog:      MasterPoolCatalog,
		defaultModel: defaultModel,
	}
}

// ResolvePoolModel resolves a capability pool to its current primary model.
func (f *FederationEngine) ResolvePoolModel(poolName string) (string, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	def, exists := f.catalog[poolName]
	if !exists {
		// Fallback for custom or direct model requests
		if strings.HasPrefix(poolName, "pool/") {
			return f.defaultModel, nil
		}
		return poolName, nil
	}

	if len(def.PrimaryTier.Models) > 0 {
		return def.PrimaryTier.Models[0], nil
	}

	return f.defaultModel, nil
}

// GetPoolTiers returns the ordered 3-tier model array for the specified pool.
func (f *FederationEngine) GetPoolTiers(poolName string) ([]ModelTier, error) {
	f.mu.RLock()
	defer f.mu.RUnlock()

	def, exists := f.catalog[poolName]
	if !exists {
		return nil, fmt.Errorf("unknown pool: %s", poolName)
	}

	return []ModelTier{
		def.PrimaryTier,
		def.BackupTier1,
		def.BackupTier2,
	}, nil
}

// GetNextFallbackModel finds the next model within the pool's 3-tier array.
// Returns (nextModel, nextTier, nextIndex, hasNext).
func (f *FederationEngine) GetNextFallbackModel(poolName string, currentTier int, currentIndex int) (string, int, int, bool) {
	tiers, err := f.GetPoolTiers(poolName)
	if err != nil || len(tiers) == 0 {
		return "", -1, -1, false
	}

	tier := currentTier
	idx := currentIndex + 1

	for tier < len(tiers) {
		if idx < len(tiers[tier].Models) {
			return tiers[tier].Models[idx], tier, idx, true
		}
		tier++
		idx = 0
	}

	return "", -1, -1, false
}

// MapToGroqUpstream translates abstract catalog models to concrete Groq API targets
// for the current active provider environment.
func MapToGroqUpstream(catalogModel string) string {
	lower := strings.ToLower(catalogModel)
	if strings.Contains(lower, "whisper") {
		return "whisper-large-v3"
	}
	return "openai/gpt-oss-120b"
}
