package marketdata

import (
	"testing"
	"time"
)

func TestRateLimiter(t *testing.T) {
	l := NewRateLimiter(3, 50*time.Millisecond)
	key := "k1"
	for i := 0; i < 3; i++ {
		if !l.Allow(key) {
			t.Fatalf("expected allow at %d", i)
		}
	}
	if l.Allow(key) {
		t.Fatalf("expected limit after capacity")
	}
	time.Sleep(60 * time.Millisecond)
	if !l.Allow(key) {
		t.Fatalf("expected allow after refill")
	}
}

