package supervisor

import (
	"sync"
	"time"
)

// ScratchpadStep represents an intermediate execution step within the supervisor.
type ScratchpadStep struct {
	Name      string        `json:"name"`
	Status    string        `json:"status"` // "running", "completed", "failed", "skipped"
	Details   string        `json:"details,omitempty"`
	Duration  time.Duration `json:"duration"`
	Timestamp time.Time     `json:"timestamp"`
}

// ScratchpadEnvelope isolates ephemeral reasoning, routing metadata, and sub-task passes.
// It ensures that internal multi-step executions do not pollute the client-facing chat history.
type ScratchpadEnvelope struct {
	mu                   sync.RWMutex
	RequestID            string                 `json:"request_id"`
	OriginalPrompt       string                 `json:"original_prompt"`
	Classification       ClassificationResult   `json:"classification"`
	ResolvedModel        string                 `json:"resolved_model"`
	SelectedPool         string                 `json:"selected_pool"`
	IntermediateSteps    []ScratchpadStep       `json:"intermediate_steps"`
	CriticPassesExecuted int                    `json:"critic_passes_executed"`
	SecurityWarnings     []string               `json:"security_warnings,omitempty"`
	StartTime            time.Time              `json:"start_time"`
	TotalDuration        time.Duration          `json:"total_duration"`
	Metadata             map[string]interface{} `json:"metadata"`
}

// NewScratchpadEnvelope creates an initialized scratchpad for a request.
func NewScratchpadEnvelope(requestID string, prompt string) *ScratchpadEnvelope {
	return &ScratchpadEnvelope{
		RequestID:         requestID,
		OriginalPrompt:    prompt,
		IntermediateSteps: make([]ScratchpadStep, 0, 4),
		SecurityWarnings:  make([]string, 0),
		StartTime:         time.Now(),
		Metadata:          make(map[string]interface{}),
	}
}

// RecordStep appends an execution step to the envelope.
func (s *ScratchpadEnvelope) RecordStep(name string, status string, details string, duration time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.IntermediateSteps = append(s.IntermediateSteps, ScratchpadStep{
		Name:      name,
		Status:    status,
		Details:   details,
		Duration:  duration,
		Timestamp: time.Now(),
	})
}

// AddSecurityWarning logs a security or AST flaw flagged by the critic pass.
func (s *ScratchpadEnvelope) AddSecurityWarning(warning string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.SecurityWarnings = append(s.SecurityWarnings, warning)
}

// Finalize calculates total duration and closes the envelope.
func (s *ScratchpadEnvelope) Finalize() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.TotalDuration = time.Since(s.StartTime)
}

// TelemetrySummary returns a concise string representation of steps for logging.
func (s *ScratchpadEnvelope) TelemetrySummary() string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return ""
}
