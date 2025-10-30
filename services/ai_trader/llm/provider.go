package llm

import (
	"context"
)

// Provider defines the interface for Large Language Model backends (e.g., Groq).
// Implementations should be stateless and safe for concurrent use.
type Provider interface {
	// GenerateDecision asks the LLM to produce a trading decision based on structured inputs.
	GenerateDecision(ctx context.Context, in PromptInput) (DecisionOut, error)

	// Name returns a short provider identifier (e.g., "groq").
	Name() string
}
