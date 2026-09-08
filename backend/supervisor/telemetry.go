package supervisor

import (
	"fmt"
	"sync"
	"time"
)

// TelemetryEventType defines the type of event being emitted.
type TelemetryEventType string

const (
	EventIntentClassifying TelemetryEventType = "INTENT_CLASSIFYING"
	EventIntentClassified  TelemetryEventType = "INTENT_CLASSIFIED"
	EventUpstreamDispatch  TelemetryEventType = "UPSTREAM_DISPATCH"
	EventCriticStarted     TelemetryEventType = "CRITIC_STARTED"
	EventCriticComplete    TelemetryEventType = "CRITIC_COMPLETE"
	EventCompleted         TelemetryEventType = "COMPLETED"
	EventFailed            TelemetryEventType = "FAILED"
)

// TelemetryEvent represents a live execution milestone.
type TelemetryEvent struct {
	RequestID string             `json:"request_id"`
	Type      TelemetryEventType `json:"type"`
	Message   string             `json:"message"`
	Timestamp time.Time          `json:"timestamp"`
	Metadata  map[string]string  `json:"metadata,omitempty"`
}

// TelemetryListener is a callback function for receiving real-time events.
type TelemetryListener func(event TelemetryEvent)

// TelemetryDispatcher manages event subscribers and broadcasts execution events.
type TelemetryDispatcher struct {
	mu        sync.RWMutex
	listeners []TelemetryListener
}

// NewTelemetryDispatcher initializes a dispatcher.
func NewTelemetryDispatcher() *TelemetryDispatcher {
	return &TelemetryDispatcher{
		listeners: make([]TelemetryListener, 0),
	}
}

// Subscribe registers a new listener callback.
func (d *TelemetryDispatcher) Subscribe(listener TelemetryListener) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.listeners = append(d.listeners, listener)
}

// Emit broadcasts an event to all subscribers.
func (d *TelemetryDispatcher) Emit(event TelemetryEvent) {
	d.mu.RLock()
	defer d.mu.RUnlock()

	for _, l := range d.listeners {
		go l(event)
	}
}

// FormatStatusMessage produces human-friendly status strings for bastions/UIs.
func FormatStatusMessage(eventType TelemetryEventType, detail string) string {
	switch eventType {
	case EventIntentClassifying:
		return "🔍 Analyzing prompt intent..."
	case EventIntentClassified:
		return fmt.Sprintf("🎯 Routed to %s", detail)
	case EventUpstreamDispatch:
		return fmt.Sprintf("⚡ Generating completion via %s...", detail)
	case EventCriticStarted:
		return "🛡️ Executing automated AST security critic pass..."
	case EventCriticComplete:
		return fmt.Sprintf("✅ Security review completed: %s", detail)
	case EventCompleted:
		return "🏁 Generation finished."
	case EventFailed:
		return fmt.Sprintf("❌ Error during orchestration: %s", detail)
	default:
		return detail
	}
}

// FormatWebUIAccordion formats status for Open WebUI collapsible containers.
func FormatWebUIAccordion(status string) string {
	return fmt.Sprintf(":::status %s\n:::\n", status)
}
