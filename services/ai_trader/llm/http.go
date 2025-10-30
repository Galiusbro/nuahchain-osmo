package llm

import (
	"context"
	"net/http"
	"time"
)

// newHTTPClient creates a tuned HTTP client for LLM calls.
func newHTTPClient(timeout time.Duration) *http.Client {
	if timeout <= 0 {
		timeout = 12 * time.Second
	}
	transport := &http.Transport{
		MaxIdleConns:        16,
		MaxIdleConnsPerHost: 8,
		IdleConnTimeout:     30 * time.Second,
	}
	return &http.Client{Timeout: timeout, Transport: transport}
}

// contextWithTimeout returns a context with timeout if parent has none.
func contextWithTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if _, hasDeadline := ctx.Deadline(); hasDeadline {
		return ctx, func() {}
	}
	if d <= 0 {
		d = 12 * time.Second
	}
	return context.WithTimeout(ctx, d)
}
