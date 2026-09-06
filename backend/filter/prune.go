package filter

import (
	"unicode/utf8"
)

// ChatMessage represents a single turn in a conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// EstimateTokens calculates an approximate token count for a string using the standard 4-char heuristic + envelope overhead.
func EstimateTokens(text string) int {
	charCount := utf8.RuneCountInString(text)
	tokens := (charCount + 3) / 4
	if tokens < 1 && charCount > 0 {
		return 1
	}
	return tokens
}

// EstimateMessagesTokens computes total tokens for an entire array of ChatMessage.
func EstimateMessagesTokens(messages []ChatMessage) int {
	total := 0
	for _, m := range messages {
		// ~4 tokens per message for formatting overhead (<|im_start|>role\n...<|im_end|>)
		total += EstimateTokens(m.Content) + 4
	}
	return total + 3 // Priming tokens
}

// PruneMessages prunes intermediate conversation turns strictly at message boundaries
// to fit within maxTokens, while strictly protecting:
// 1. Turn 0 if it is a system prompt (role: "system")
// 2. The final turn (active user prompt)
// Returns the pruned message slice and the number of turns dropped.
func PruneMessages(messages []ChatMessage, maxTokens int) ([]ChatMessage, int) {
	if len(messages) <= 2 || maxTokens <= 0 {
		return messages, 0
	}

	totalTokens := EstimateMessagesTokens(messages)
	if totalTokens <= maxTokens {
		return messages, 0
	}

	hasSystemPrompt := messages[0].Role == "system"
	var systemPrompt *ChatMessage
	var turnsToPrune []ChatMessage
	finalTurn := messages[len(messages)-1]

	if hasSystemPrompt {
		systemPrompt = &messages[0]
		turnsToPrune = append(turnsToPrune, messages[1:len(messages)-1]...)
	} else {
		turnsToPrune = append(turnsToPrune, messages[0:len(messages)-1]...)
	}

	droppedCount := 0
	// Prune from oldest to newest
	for len(turnsToPrune) > 0 {
		// Calculate current total
		currentTokens := EstimateTokens(finalTurn.Content) + 4 + 3
		if systemPrompt != nil {
			currentTokens += EstimateTokens(systemPrompt.Content) + 4
		}
		for _, m := range turnsToPrune {
			currentTokens += EstimateTokens(m.Content) + 4
		}

		if currentTokens <= maxTokens {
			break
		}

		// Drop oldest intermediate turn
		turnsToPrune = turnsToPrune[1:]
		droppedCount++
	}

	// Reassemble pruned slice
	result := make([]ChatMessage, 0, len(turnsToPrune)+2)
	if systemPrompt != nil {
		result = append(result, *systemPrompt)
	}
	result = append(result, turnsToPrune...)
	result = append(result, finalTurn)

	return result, droppedCount
}
