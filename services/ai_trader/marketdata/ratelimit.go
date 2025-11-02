package marketdata

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RateLimiter enforces token bucket per key with optional global fallback.
type RateLimiter struct {
	capacity int
	refill   time.Duration
	mu       sync.Mutex
	buckets  map[string]*bucket
}

type bucket struct {
	tokens    int
	updatedAt time.Time
}

// NewRateLimiter builds a limiter with given capacity and refill interval.
func NewRateLimiter(capacity int, refill time.Duration) *RateLimiter {
	if capacity <= 0 {
		capacity = 100
	}
	if refill <= 0 {
		refill = time.Minute
	}
	return &RateLimiter{
		capacity: capacity,
		refill:   refill,
		buckets:  make(map[string]*bucket),
	}
}

func newLimiterFromEnv() *RateLimiter {
	cap := 60
	if v := strings.TrimSpace(os.Getenv("AI_TRADER_RATE_LIMIT")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			cap = parsed
		}
	}
	refill := time.Minute
	if v := strings.TrimSpace(os.Getenv("AI_TRADER_RATE_INTERVAL")); v != "" {
		if d, err := time.ParseDuration(v); err == nil && d > 0 {
			refill = d
		}
	}
	return NewRateLimiter(cap, refill)
}

// Allow consumes one token for the provided key; returns false when limited.
func (l *RateLimiter) Allow(key string) bool {
	if key == "" {
		key = "anonymous"
	}
	now := time.Now()
	l.mu.Lock()
	defer l.mu.Unlock()
	bkt := l.buckets[key]
	if bkt == nil {
		bkt = &bucket{tokens: l.capacity - 1, updatedAt: now}
		l.buckets[key] = bkt
		return true
	}
	// Refill tokens based on elapsed intervals.
	delta := now.Sub(bkt.updatedAt)
	if delta >= l.refill {
		steps := int(delta / l.refill)
		bkt.tokens += steps
		if bkt.tokens > l.capacity {
			bkt.tokens = l.capacity
		}
		bkt.updatedAt = now
	}
	if bkt.tokens <= 0 {
		return false
	}
	bkt.tokens--
	return true
}
