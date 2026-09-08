package types

// ChatMessage represents a single turn in a chat conversation.
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// PoolInfo contains capability pool metadata and target configurations.
type PoolInfo struct {
	Name        string `json:"name"`
	Tier        string `json:"tier"`
	TargetModel string `json:"target_model"`
	Provider    string `json:"provider"`
	ContextMax  string `json:"context_max"`
	Status      string `json:"status"`
	Description string `json:"description"`
}
