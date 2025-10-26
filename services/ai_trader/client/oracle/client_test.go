package oracle_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/oracle"
)

func TestOracleClient_ValidatePrice(t *testing.T) {
	tests := []struct {
		name          string
		price         *oracle.PriceData
		expectedError bool
	}{
		{
			name: "valid price",
			price: &oracle.PriceData{
				Symbol:     "BTC",
				Value:      "50000.00",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 0.95,
			},
			expectedError: false,
		},
		{
			name:          "nil price",
			price:         nil,
			expectedError: true,
		},
		{
			name: "empty symbol",
			price: &oracle.PriceData{
				Symbol:     "",
				Value:      "50000.00",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 0.95,
			},
			expectedError: true,
		},
		{
			name: "empty value",
			price: &oracle.PriceData{
				Symbol:     "BTC",
				Value:      "",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 0.95,
			},
			expectedError: true,
		},
		{
			name: "invalid confidence",
			price: &oracle.PriceData{
				Symbol:     "BTC",
				Value:      "50000.00",
				Source:     "coinbase",
				Timestamp:  time.Now(),
				Confidence: 1.5,
			},
			expectedError: true,
		},
		{
			name: "stale price",
			price: &oracle.PriceData{
				Symbol:     "BTC",
				Value:      "50000.00",
				Source:     "coinbase",
				Timestamp:  time.Now().Add(-2 * time.Hour),
				Confidence: 0.95,
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &oracle.Client{}
			err := client.ValidatePrice(tt.price)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestOracleClient_IsPriceStale(t *testing.T) {
	tests := []struct {
		name     string
		price    *oracle.PriceData
		maxAge   time.Duration
		expected bool
	}{
		{
			name: "fresh price",
			price: &oracle.PriceData{
				Timestamp: time.Now().Add(-5 * time.Minute),
			},
			maxAge:   time.Hour,
			expected: false,
		},
		{
			name: "stale price",
			price: &oracle.PriceData{
				Timestamp: time.Now().Add(-2 * time.Hour),
			},
			maxAge:   time.Hour,
			expected: true,
		},
		{
			name:     "nil price",
			price:    nil,
			maxAge:   time.Hour,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &oracle.Client{}
			result := client.IsPriceStale(tt.price, tt.maxAge)
			assert.Equal(t, tt.expected, result)
		})
	}
}
