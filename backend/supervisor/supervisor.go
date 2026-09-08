package supervisor

import (
	"context"
	"strings"
	"time"
)

// SupervisorConfig defines configuration options for the supervisor engine.
type SupervisorConfig struct {
	ClassifierConfig ClassifierConfig
	CriticConfig     CriticConfig
	DefaultModel     string
}

// Supervisor coordinates autonomous classification, 14-pool federation, and security critic passes.
type Supervisor struct {
	Classifier *HybridClassifier
	Federation *FederationEngine
	Critic     *SecurityCritic
	Telemetry  *TelemetryDispatcher
}

// OrchestrationResult contains the routing decision and execution metadata.
type OrchestrationResult struct {
	RequestedModel       string               `json:"requested_model"`
	TargetPool           string               `json:"target_pool"`
	UpstreamModel        string               `json:"upstream_model"`
	Classification       ClassificationResult `json:"classification"`
	Scratchpad           *ScratchpadEnvelope  `json:"scratchpad"`
	CriticVerdict        *CriticVerdict       `json:"critic_verdict,omitempty"`
	SecurityReviewPassed bool                 `json:"security_review_passed"`
}

// NewSupervisor initializes the supervisor with all sub-engines.
func NewSupervisor(cfg SupervisorConfig) *Supervisor {
	if cfg.DefaultModel == "" {
		cfg.DefaultModel = "llama-3.3-70b-versatile"
	}

	return &Supervisor{
		Classifier: NewHybridClassifier(cfg.ClassifierConfig),
		Federation: NewFederationEngine(cfg.DefaultModel),
		Critic:     NewSecurityCritic(cfg.CriticConfig),
		Telemetry:  NewTelemetryDispatcher(),
	}
}

// RouteRequest evaluates an incoming request and determines the optimal capability pool and upstream model.
func (s *Supervisor) RouteRequest(ctx context.Context, requestID string, model string, prompt string) OrchestrationResult {
	envelope := NewScratchpadEnvelope(requestID, prompt)

	var classification ClassificationResult
	targetPool := model

	// If model is "pool/auto", perform Tier 1 sub-40ms intent classification
	if model == "pool/auto" {
		s.Telemetry.Emit(TelemetryEvent{
			RequestID: requestID,
			Type:      EventIntentClassifying,
			Message:   FormatStatusMessage(EventIntentClassifying, ""),
			Timestamp: time.Now(),
		})

		classification = s.Classifier.Classify(ctx, prompt)
		targetPool = classification.TargetPool

		envelope.Classification = classification
		envelope.RecordStep("IntentClassification", "completed", classification.Intent, classification.Latency)

		s.Telemetry.Emit(TelemetryEvent{
			RequestID: requestID,
			Type:      EventIntentClassified,
			Message:   FormatStatusMessage(EventIntentClassified, targetPool),
			Timestamp: time.Now(),
			Metadata: map[string]string{
				"intent":     classification.Intent,
				"pool":       targetPool,
				"latency_ms": classification.Latency.String(),
			},
		})
	} else {
		// Explicit pool request
		classification = ClassificationResult{
			Intent:     "EXPLICIT",
			TargetPool: model,
			Tier:       0,
			Confidence: 1.0,
			Reasoning:  "Explicit pool requested by client",
		}
		envelope.Classification = classification
	}

	// Resolve capability pool to active primary model in the 3-tier federation
	upstreamModel, err := s.Federation.ResolvePoolModel(targetPool)
	if err != nil {
		upstreamModel = "llama-3.3-70b-versatile"
	}

	envelope.SelectedPool = targetPool
	envelope.ResolvedModel = upstreamModel

	s.Telemetry.Emit(TelemetryEvent{
		RequestID: requestID,
		Type:      EventUpstreamDispatch,
		Message:   FormatStatusMessage(EventUpstreamDispatch, upstreamModel),
		Timestamp: time.Now(),
	})

	return OrchestrationResult{
		RequestedModel:       model,
		TargetPool:           targetPool,
		UpstreamModel:        upstreamModel,
		Classification:       classification,
		Scratchpad:           envelope,
		SecurityReviewPassed: true,
	}
}

// PerformSecurityCritic executes an automated AST security review pass on generated code.
func (s *Supervisor) PerformSecurityCritic(
	ctx context.Context,
	envelope *ScratchpadEnvelope,
	generatedContent string,
) (finalContent string, verdict CriticVerdict) {
	// Only run critic pass if content appears to contain code or comes from agent-coding
	hasCodeFence := strings.Contains(generatedContent, "```")
	isCodingPool := envelope.SelectedPool == "pool/agent-coding" || envelope.Classification.Intent == IntentCode

	if !hasCodeFence && !isCodingPool {
		return generatedContent, CriticVerdict{IsClean: true}
	}

	s.Telemetry.Emit(TelemetryEvent{
		RequestID: envelope.RequestID,
		Type:      EventCriticStarted,
		Message:   FormatStatusMessage(EventCriticStarted, ""),
		Timestamp: time.Now(),
	})

	start := time.Now()
	finalContent, verdict = s.Critic.ExecuteCriticLoop(ctx, generatedContent, nil)
	criticDuration := time.Since(start)

	envelope.CriticPassesExecuted = verdict.Iterations
	envelope.RecordStep("SecurityCriticPass", "completed", verdictSummary(verdict), criticDuration)

	for _, issue := range verdict.Issues {
		envelope.AddSecurityWarning(issue.Description)
	}

	s.Telemetry.Emit(TelemetryEvent{
		RequestID: envelope.RequestID,
		Type:      EventCriticComplete,
		Message:   FormatStatusMessage(EventCriticComplete, verdictSummary(verdict)),
		Timestamp: time.Now(),
	})

	return finalContent, verdict
}

func verdictSummary(v CriticVerdict) string {
	if v.IsClean {
		return "Clean AST pass"
	}
	return "Vulnerabilities flagged"
}
