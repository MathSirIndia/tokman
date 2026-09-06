package filter

import (
	"testing"
)

func TestPruneMessages_NoPruneNeeded(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "system", Content: "You are a helpful assistant."},
		{Role: "user", Content: "Hello!"},
		{Role: "assistant", Content: "Hi there!"},
		{Role: "user", Content: "How is the weather?"},
	}

	pruned, dropped := PruneMessages(msgs, 500)
	if dropped != 0 {
		t.Errorf("expected 0 dropped turns, got %d", dropped)
	}
	if len(pruned) != len(msgs) {
		t.Errorf("expected %d messages, got %d", len(msgs), len(pruned))
	}
}

func TestPruneMessages_PreserveSystemAndFinalTurn(t *testing.T) {
	systemMsg := "You are a system prompt that must NEVER be pruned."
	finalUserMsg := "This is the newest active user query that must be preserved."

	msgs := []ChatMessage{
		{Role: "system", Content: systemMsg},
		{Role: "user", Content: "Old question 1 that is very long: " + string(make([]byte, 200))},
		{Role: "assistant", Content: "Old answer 1 that is very long: " + string(make([]byte, 200))},
		{Role: "user", Content: "Old question 2 that is also long: " + string(make([]byte, 200))},
		{Role: "assistant", Content: "Old answer 2 that is also long: " + string(make([]byte, 200))},
		{Role: "user", Content: finalUserMsg},
	}

	// Set a tight budget that only fits system + final message + maybe 1 intermediate turn
	budget := EstimateTokens(systemMsg) + EstimateTokens(finalUserMsg) + 30

	pruned, dropped := PruneMessages(msgs, budget)

	if dropped == 0 {
		t.Errorf("expected turns to be dropped under tight budget, got 0")
	}

	// System prompt must be Turn 0
	if len(pruned) < 2 {
		t.Fatalf("pruned slice too short: %d", len(pruned))
	}

	if pruned[0].Role != "system" || pruned[0].Content != systemMsg {
		t.Errorf("system prompt was not preserved or modified: got %v", pruned[0])
	}

	// Final prompt must be preserved
	last := pruned[len(pruned)-1]
	if last.Role != "user" || last.Content != finalUserMsg {
		t.Errorf("final user prompt was not preserved: got %v", last)
	}
}

func TestPruneMessages_OldestTurnsDroppedFirst(t *testing.T) {
	msgs := []ChatMessage{
		{Role: "system", Content: "SysPrompt"},
		{Role: "user", Content: "Turn 1 - oldest user"},
		{Role: "assistant", Content: "Turn 2 - oldest assistant"},
		{Role: "user", Content: "Turn 3 - newer user"},
		{Role: "assistant", Content: "Turn 4 - newer assistant"},
		{Role: "user", Content: "Turn 5 - newest question"},
	}

	// Budget that only allows 4 messages total (System, Turn 3, Turn 4, Turn 5)
	budget := EstimateTokens("SysPrompt") +
		EstimateTokens("Turn 3 - newer user") +
		EstimateTokens("Turn 4 - newer assistant") +
		EstimateTokens("Turn 5 - newest question") + 25

	pruned, dropped := PruneMessages(msgs, budget)

	if dropped != 2 {
		t.Errorf("expected 2 oldest intermediate turns dropped, got %d", dropped)
	}

	expectedContents := []string{
		"SysPrompt",
		"Turn 3 - newer user",
		"Turn 4 - newer assistant",
		"Turn 5 - newest question",
	}

	if len(pruned) != len(expectedContents) {
		t.Fatalf("expected %d messages, got %d", len(expectedContents), len(pruned))
	}

	for i, exp := range expectedContents {
		if pruned[i].Content != exp {
			t.Errorf("index %d: expected %q, got %q", i, exp, pruned[i].Content)
		}
	}
}
