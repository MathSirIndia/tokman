package registry

import (
	"fmt"
	"tokman/backend/types"
)

// CanonicalPools contains the definitive 14 capability pools from docs/report.md.
// This is the single source of truth across the gateway router, GUI, and supervisor.
var CanonicalPools = []types.PoolInfo{
	{
		Name:        "pool/auto",
		Tier:        "Meta-Orchestrator",
		TargetModel: "Hierarchical Supervisor",
		Provider:    "Two-Tier Arbitrator",
		ContextMax:  "<40ms DAG",
		Status:      "Active",
		Description: "Frontline 5 intent classifier & mid-query tool handoff DAG",
	},
	{
		Name:        "pool/general",
		Tier:        "Frontline Core",
		TargetModel: "groq/llama-3.3-70b",
		Provider:    "Groq / GitHub / CF",
		ContextMax:  "8,192 tkns",
		Status:      "Active",
		Description: "Balanced daily tasks, summarization & flagship routing",
	},
	{
		Name:        "pool/deep-reasoning",
		Tier:        "Frontline Core",
		TargetModel: "sambanova/deepseek-r1",
		Provider:    "SambaNova / Cloudflare",
		ContextMax:  "8,192 tkns",
		Status:      "Active",
		Description: "Complex logic, mathematical deduction & step-by-step reasoning",
	},
	{
		Name:        "pool/agent-coding",
		Tier:        "Frontline Core",
		TargetModel: "sambanova/qwen-2.5-coder",
		Provider:    "SambaNova / Mistral / Groq",
		ContextMax:  "8,192 tkns",
		Status:      "Active",
		Description: "Autonomous developer assistants, refactoring & code synthesis",
	},
	{
		Name:        "pool/document-analysis",
		Tier:        "Frontline Core",
		TargetModel: "gemini/gemini-2.0-flash",
		Provider:    "Google AI Studio / Groq",
		ContextMax:  "1,000,000 tkns",
		Status:      "Active",
		Description: "Massive context window ingestion & long-document reasoning",
	},
	{
		Name:        "pool/web-research",
		Tier:        "Frontline Core",
		TargetModel: "cerebras/llama-3.3-70b",
		Provider:    "Cerebras / SearXNG",
		ContextMax:  "8,192 tkns",
		Status:      "Staged (Module 4)",
		Description: "Live internet grounding & multi-source web synthesis",
	},
	{
		Name:        "pool/presentation",
		Tier:        "Artifact Engine",
		TargetModel: "groq/llama-3.3-70b",
		Provider:    "Groq / Headless Marp",
		ContextMax:  "8,192 tkns",
		Status:      "Staged (Module 6)",
		Description: "Marp presentation synthesis & Reveal.js visual decks",
	},
	{
		Name:        "pool/image-gen",
		Tier:        "Artifact Engine",
		TargetModel: "cf/flux-1-schnell",
		Provider:    "Cloudflare Diffusion Engine",
		ContextMax:  "Keyframe / 1:1",
		Status:      "Staged (Module 6)",
		Description: "Cloudflare Flux.1 Schnell / SDXL text-to-image synthesis",
	},
	{
		Name:        "pool/architect",
		Tier:        "Artifact Engine",
		TargetModel: "gemini/gemini-2.0-flash",
		Provider:    "Google AI Studio / Groq",
		ContextMax:  "Scene Graph",
		Status:      "Staged (Module 6)",
		Description: "Video Conductor Scene Graph JSON manifest generator",
	},
	{
		Name:        "pool/security-tester",
		Tier:        "Frontline Core",
		TargetModel: "sambanova/qwen-coder",
		Provider:    "SambaNova / Cloudflare",
		ContextMax:  "8,192 tkns",
		Status:      "Active",
		Description: "Automated static analysis & vulnerability critic pass",
	},
	{
		Name:        "pool/stack-optimizer",
		Tier:        "System Utility",
		TargetModel: "gemini/gemini-2.0-flash",
		Provider:    "Google AI Studio",
		ContextMax:  "Telemetry Diff",
		Status:      "Staged (Module 7)",
		Description: "7-day token burn analyzer & zero-touch configuration optimizer",
	},
	{
		Name:        "pool/document-gen",
		Tier:        "Artifact Engine",
		TargetModel: "gemini/gemini-2.0-flash",
		Provider:    "Google AI Studio / Typst",
		ContextMax:  "1,000,000 tkns",
		Status:      "Staged (Module 6)",
		Description: "Headless Typst / Pandoc mathematical PDF compiler",
	},
	{
		Name:        "pool/audio-gen",
		Tier:        "Artifact Engine",
		TargetModel: "cf/openai-whisper",
		Provider:    "Cloudflare / HuggingFace",
		ContextMax:  "Kokoro-82M TTS",
		Status:      "Staged (Module 6)",
		Description: "Local Kokoro-82M speech synthesis & Cloudflare Whisper STT",
	},
	{
		Name:        "pool/video-conductor",
		Tier:        "Meta-Orchestrator",
		TargetModel: "FFmpeg Ken Burns Engine",
		Provider:    "Multi-Modal Compositor",
		ContextMax:  "1080p MP4 FIFO",
		Status:      "Staged (Module 6)",
		Description: "End-to-end automated documentary video rendering pipeline",
	},
}

// CanonicalPoolNames returns a slice of all 14 capability pool identifiers.
func CanonicalPoolNames() []string {
	names := make([]string, len(CanonicalPools))
	for i, p := range CanonicalPools {
		names[i] = p.Name
	}
	return names
}

// ActivePoolNames returns a slice of currently active/routed capability pool identifiers.
func ActivePoolNames() []string {
	var names []string
	for _, p := range CanonicalPools {
		if p.Status == "Active" {
			names = append(names, p.Name)
		}
	}
	return names
}

// FindPool searches CanonicalPools for a pool matching name.
func FindPool(name string) (types.PoolInfo, bool) {
	for _, p := range CanonicalPools {
		if p.Name == name {
			return p, true
		}
	}
	return types.PoolInfo{}, false
}

// GetTargetModelNote returns a formatted target string for the given pool name.
func GetTargetModelNote(name string) string {
	if p, found := FindPool(name); found {
		return fmt.Sprintf("Target: %s (%s)", p.TargetModel, p.Provider)
	}
	return "Target: " + name
}
