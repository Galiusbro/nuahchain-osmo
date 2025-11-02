package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNormalizeGRPCAddress(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"plain", "localhost:9090", "localhost:9090"},
		{"http prefix", "http://localhost:9090", "localhost:9090"},
		{"https prefix", "https://example.com:9090", "example.com:9090"},
		{"extra spaces", "  localhost:9090  ", "localhost:9090"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := normalizeGRPCAddress(tc.in)
			require.Equal(t, tc.want, got)
		})
	}
}

func TestParseExpirationWithDefault(t *testing.T) {
	t.Run("duration input", func(t *testing.T) {
		exp, err := parseExpirationWithDefault("24h", time.Time{}, 0)
		require.NoError(t, err)
		diff := time.Until(exp)
		require.InDelta(t, 24*time.Hour.Seconds(), diff.Seconds(), 5)
	})

	t.Run("rfc3339 input", func(t *testing.T) {
		target := time.Now().Add(48 * time.Hour).UTC().Truncate(time.Second)
		exp, err := parseExpirationWithDefault(target.Format(time.RFC3339), time.Time{}, 0)
		require.NoError(t, err)
		require.WithinDuration(t, target, exp, time.Second)
	})

	t.Run("default time fallback", func(t *testing.T) {
		fallback := time.Now().Add(72 * time.Hour).UTC()
		exp, err := parseExpirationWithDefault("", fallback, 0)
		require.NoError(t, err)
		require.WithinDuration(t, fallback, exp, time.Second)
	})

	t.Run("default duration fallback", func(t *testing.T) {
		exp, err := parseExpirationWithDefault("", time.Time{}, 6*time.Hour)
		require.NoError(t, err)
		diff := time.Until(exp)
		require.InDelta(t, (6 * time.Hour).Seconds(), diff.Seconds(), 5)
	})

	t.Run("invalid input", func(t *testing.T) {
		_, err := parseExpirationWithDefault("not-a-duration", time.Time{}, time.Hour)
		require.Error(t, err)
	})
}

func TestSanitizeSymbols(t *testing.T) {
	input := []string{" osmo ", "OSMO", "eth", "", "btc", "Eth"}
	out := sanitizeSymbols(input)
	require.ElementsMatch(t, []string{"OSMO", "ETH", "BTC"}, out)
}
