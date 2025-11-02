package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"
)

// GroqClient implements Provider for api.groq.com (OpenAI-compatible chat API).
type GroqClient struct {
	http    *http.Client
	apiKey  string
	model   string
	baseURL string
	timeout time.Duration
	lastRaw string
}

type groqChatRequest struct {
	Model       string        `json:"model"`
	Messages    []groqMessage `json:"messages"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
	Temperature float64       `json:"temperature,omitempty"`
}

type groqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type groqChatResponse struct {
	Choices []struct {
		Message groqMessage `json:"message"`
	} `json:"choices"`
}

// NewGroq creates a Groq LLM provider.
// apiKey: if empty, falls back to GROQ_API_KEY env, then to the provided hardcoded key.
func NewGroq(apiKey, model string, timeout time.Duration) *GroqClient {
	if apiKey == "" {
		apiKey = os.Getenv("GROQ_API_KEY")
	}
	if apiKey == "" {
		// Temporary per user instruction; should be removed in production.
		apiKey = "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"
	}
	if model == "" {
		model = "llama-3.3-70b-versatile"
	}
	return &GroqClient{
		http:    newHTTPClient(timeout),
		apiKey:  apiKey,
		model:   model,
		baseURL: "https://api.groq.com/openai/v1/chat/completions",
		timeout: timeout,
	}
}

func (g *GroqClient) Name() string { return "groq" }

// LastRaw returns last raw content from the model (best effort).
func (g *GroqClient) LastRaw() string { return g.lastRaw }

// GenerateDecision calls Groq to obtain a structured DecisionOut.
func (g *GroqClient) GenerateDecision(ctx context.Context, in PromptInput) (DecisionOut, error) {
	var out DecisionOut
	reqBody, err := g.buildRequestBody(in)
	if err != nil {
		return out, err
	}

	ctx, cancel := contextWithTimeout(ctx, g.timeout)
	defer cancel()

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, g.baseURL, bytes.NewReader(reqBody))
	if err != nil {
		return out, err
	}
	httpReq.Header.Set("Authorization", "Bearer "+g.apiKey)
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := g.http.Do(httpReq)
	if err != nil {
		return out, err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return out, fmt.Errorf("groq http status %d", resp.StatusCode)
	}

	// Read body once for optional debug and JSON decode
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return out, err
	}
	if os.Getenv("GROQ_DEBUG") == "1" {
		fmt.Printf("[GROQ HTTP %d]\n%s\n", resp.StatusCode, string(bodyBytes))
	}
	var gr groqChatResponse
	if err := json.Unmarshal(bodyBytes, &gr); err != nil {
		return out, err
	}
	if len(gr.Choices) == 0 {
		return out, fmt.Errorf("groq: empty choices")
	}
	content := gr.Choices[0].Message.Content
	if os.Getenv("GROQ_DEBUG") == "1" {
		fmt.Printf("[GROQ RAW]\n%s\n", content)
	}
	g.lastRaw = content
	// Expect JSON in content; try strict parse first
	if err := json.Unmarshal([]byte(content), &out); err != nil {
		// Try to extract first JSON object with a liberal regex
		re := regexp.MustCompile(`\{[\s\S]*\}`)
		if loc := re.FindStringIndex(content); loc != nil {
			snippet := content[loc[0]:loc[1]]
			if e2 := json.Unmarshal([]byte(snippet), &out); e2 == nil {
				return out, nil
			}
		}
		return out, fmt.Errorf("groq: invalid JSON response: %w", err)
	}
	return out, nil
}

func (g *GroqClient) buildRequestBody(in PromptInput) ([]byte, error) {
	sys := "You are a trading decision assistant. Respond ONLY with strict JSON matching the schema. The 'action' MUST be one of: buy, sell, hold (lowercase). No prose, no markdown."
	userPayload := struct {
		Schema      map[string]any `json:"schema"`
		Input       PromptInput    `json:"input"`
		Instruction string         `json:"instruction"`
	}{
		Schema: map[string]any{
			"type":     "object",
			"required": []string{"action", "symbol", "amount", "payment_denom", "market", "confidence", "rationale"},
			"properties": map[string]any{
				"action":        map[string]any{"type": "string", "enum": []string{"buy", "sell", "hold"}},
				"symbol":        map[string]any{"type": "string"},
				"amount":        map[string]any{"type": "string"},
				"payment_denom": map[string]any{"type": "string"},
				"market":        map[string]any{"type": "string", "enum": []string{"assets", "bondingcurve"}},
				"confidence":    map[string]any{"type": "number"},
				"rationale":     map[string]any{"type": "string"},
			},
		},
		Input:       in,
		Instruction: "Pick at most one symbol from input.symbols. Adhere to constraints. If uncertain, return action=hold.",
	}

	userBytes, err := json.Marshal(userPayload)
	if err != nil {
		return nil, err
	}
	body := groqChatRequest{
		Model: g.model,
		Messages: []groqMessage{
			{Role: "system", Content: sys},
			{Role: "user", Content: string(userBytes)},
		},
		MaxTokens:   500,
		Temperature: 0.1,
	}
	return json.Marshal(body)
}
