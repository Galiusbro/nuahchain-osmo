package trading_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osmosis-labs/osmosis/v30/services/ai_trader/client/trading"
)

func TestTradingClient_ValidateTradeRequest(t *testing.T) {
	tests := []struct {
		name          string
		request       *trading.TradeRequest
		expectedError bool
	}{
		{
			name: "valid buy request",
			request: &trading.TradeRequest{
				Symbol: "BTC",
				Amount: "1000000",
				Type:   "buy",
			},
			expectedError: false,
		},
		{
			name: "valid sell request",
			request: &trading.TradeRequest{
				Symbol: "BTC",
				Amount: "0.02",
				Type:   "sell",
			},
			expectedError: false,
		},
		{
			name:          "nil request",
			request:       nil,
			expectedError: true,
		},
		{
			name: "empty symbol",
			request: &trading.TradeRequest{
				Symbol: "",
				Amount: "1000000",
				Type:   "buy",
			},
			expectedError: true,
		},
		{
			name: "empty amount",
			request: &trading.TradeRequest{
				Symbol: "BTC",
				Amount: "",
				Type:   "buy",
			},
			expectedError: true,
		},
		{
			name: "invalid type",
			request: &trading.TradeRequest{
				Symbol: "BTC",
				Amount: "1000000",
				Type:   "invalid",
			},
			expectedError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &trading.Client{}
			err := client.ValidateTradeRequest(tt.request)

			if tt.expectedError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
