package monitoring

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRESTAPI_Integration(t *testing.T) {
	// Create logger
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	// Create audit store
	auditStore := NewAuditStore(logger, 1000, 100)

	// Create REST API
	restAPI := NewRESTAPI(auditStore, logger)

	// Create test server
	server := httptest.NewServer(restAPI.GetRouter())
	defer server.Close()

	// Test data
	trader := "cosmos1test1234567890abcdefghijklmnopqrstuvwxyz"
	symbol := "BTC"

	// Store some test events
	testEvents := []AuditEvent{
		{
			ID:        "event_1",
			Timestamp: time.Now().Add(-2 * time.Hour),
			EventType: EventTypeTradeRequest,
			Trader:    trader,
			Symbol:    symbol,
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     "50000.00",
			Success:   true,
			RiskLevel: RiskLevelLow,
		},
		{
			ID:        "event_2",
			Timestamp: time.Now().Add(-1 * time.Hour),
			EventType: EventTypeTradeExecuted,
			Trader:    trader,
			Symbol:    symbol,
			Action:    "buy",
			Amount:    sdk.NewCoin("factory/test/ndollar", math.NewInt(1000)),
			Price:     "50000.00",
			TxHash:    "0x1234567890abcdef",
			Success:   true,
			RiskLevel: RiskLevelLow,
		},
		{
			ID:         "event_3",
			Timestamp:  time.Now().Add(-30 * time.Minute),
			EventType:  EventTypePolicyViolation,
			Trader:     trader,
			Symbol:     symbol,
			Reason:     "Volume limit exceeded",
			Violations: []string{"daily_volume_limits"},
			RiskLevel:  RiskLevelHigh,
		},
	}

	for _, event := range testEvents {
		auditStore.StoreEvent(event)
	}

	// Store some test alerts
	testAlerts := []Alert{
		{
			ID:        "alert_1",
			Timestamp: time.Now().Add(-1 * time.Hour),
			AlertType: AlertTypeLimitExceeded,
			Severity:  AlertSeverityWarning,
			Title:     "Volume Limit Exceeded",
			Message:   "Trading volume has exceeded the daily limit",
			Trader:    trader,
			Symbol:    symbol,
			Metadata: map[string]interface{}{
				"limit":  1000000,
				"actual": 1500000,
			},
		},
		{
			ID:        "alert_2",
			Timestamp: time.Now().Add(-30 * time.Minute),
			AlertType: AlertTypePolicyViolation,
			Severity:  AlertSeverityError,
			Title:     "Policy Violation",
			Message:   "Policy violation detected",
			Trader:    trader,
			Symbol:    symbol,
		},
	}

	for _, alert := range testAlerts {
		auditStore.StoreAlert(alert)
	}

	t.Run("HealthCheck", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/health")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]interface{}
		err = json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		assert.Equal(t, "healthy", status["status"])
		assert.Contains(t, status, "timestamp")
		assert.Contains(t, status, "version")
	})

	t.Run("GetEvents", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/events")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response EventQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 3, response.Total)
		assert.Len(t, response.Events, 3)
		assert.Equal(t, 0, response.Offset)
		assert.Equal(t, 100, response.Limit)
	})

	t.Run("GetEventsWithFilters", func(t *testing.T) {
		// Test filtering by trader
		resp, err := http.Get(server.URL + "/api/v1/events?trader=" + trader)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response EventQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 3, response.Total)
		assert.Len(t, response.Events, 3)

		// Test filtering by event type
		resp, err = http.Get(server.URL + "/api/v1/events?event_type=trade_request")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 1, response.Total)
		assert.Len(t, response.Events, 1)
		assert.Equal(t, EventTypeTradeRequest, response.Events[0].EventType)
	})

	t.Run("GetAlerts", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/alerts")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response AlertQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Total)
		assert.Len(t, response.Alerts, 2)
	})

	t.Run("GetAlertsWithFilters", func(t *testing.T) {
		// Test filtering by severity
		resp, err := http.Get(server.URL + "/api/v1/alerts?severity=warning")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response AlertQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 1, response.Total)
		assert.Len(t, response.Alerts, 1)
		assert.Equal(t, AlertSeverityWarning, response.Alerts[0].Severity)
	})

	t.Run("AcknowledgeAlert", func(t *testing.T) {
		alertID := "alert_1"

		resp, err := http.Post(server.URL+"/api/v1/alerts/"+alertID+"/acknowledge", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var status map[string]string
		err = json.NewDecoder(resp.Body).Decode(&status)
		require.NoError(t, err)

		assert.Equal(t, "acknowledged", status["status"])

		// Verify alert was acknowledged
		alerts, err := auditStore.GetAlerts(context.Background(), AlertQueryRequest{
			Trader: trader,
			Limit:  10,
		})
		require.NoError(t, err)

		var acknowledgedAlert *Alert
		for _, alert := range alerts {
			if alert.ID == alertID {
				acknowledgedAlert = &alert
				break
			}
		}

		require.NotNil(t, acknowledgedAlert)
		assert.True(t, acknowledgedAlert.Acknowledged)
		assert.NotNil(t, acknowledgedAlert.AcknowledgedAt)
	})

	t.Run("GetTraderHistory", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/history/" + trader)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var history TradeHistory
		err = json.NewDecoder(resp.Body).Decode(&history)
		require.NoError(t, err)

		assert.Equal(t, trader, history.Trader)
		assert.Equal(t, 2, history.TotalTrades) // Both trade_request and trade_executed
		assert.Equal(t, 1, history.SuccessfulTrades)
		assert.Equal(t, 0, history.FailedTrades)
	})

	t.Run("GetSymbolHistory", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/history/" + trader + "/" + symbol)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var history TradeHistory
		err = json.NewDecoder(resp.Body).Decode(&history)
		require.NoError(t, err)

		assert.Equal(t, trader, history.Trader)
		assert.Equal(t, symbol, history.Symbol)
		assert.Equal(t, 2, history.TotalTrades)
	})

	t.Run("GetSystemMetrics", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/metrics")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var metrics SystemMetrics
		err = json.NewDecoder(resp.Body).Decode(&metrics)
		require.NoError(t, err)

		assert.Equal(t, 1, metrics.ActiveTraders)
		assert.Equal(t, 1, metrics.TotalTrades)
		assert.Equal(t, 1, metrics.PolicyViolations)
		assert.Equal(t, 2, metrics.ActiveAlerts)
		assert.Contains(t, metrics.Metrics, "events_per_hour")
		assert.Contains(t, metrics.Metrics, "average_trade_size")
	})

	t.Run("GetTraderMetrics", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/metrics/trader/" + trader)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var history TradeHistory
		err = json.NewDecoder(resp.Body).Decode(&history)
		require.NoError(t, err)

		assert.Equal(t, trader, history.Trader)
		assert.Equal(t, 2, history.TotalTrades)
	})

	t.Run("ErrorHandling", func(t *testing.T) {
		// Test invalid endpoint
		resp, err := http.Get(server.URL + "/api/v1/invalid")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)

		// Test invalid alert ID for acknowledgment
		resp, err = http.Post(server.URL+"/api/v1/alerts/invalid_id/acknowledge", "application/json", nil)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})

	t.Run("Pagination", func(t *testing.T) {
		// Test pagination with limit
		resp, err := http.Get(server.URL + "/api/v1/events?limit=2")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		var response EventQueryResponse
		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Total)
		assert.Len(t, response.Events, 2)
		assert.Equal(t, 0, response.Offset)
		assert.Equal(t, 2, response.Limit)

		// Test pagination with offset
		resp, err = http.Get(server.URL + "/api/v1/events?offset=1&limit=2")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)

		err = json.NewDecoder(resp.Body).Decode(&response)
		require.NoError(t, err)

		assert.Equal(t, 2, response.Total)
		assert.Len(t, response.Events, 2)
		assert.Equal(t, 1, response.Offset)
		assert.Equal(t, 2, response.Limit)
	})
}

func TestRESTAPI_QueryParameters(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	auditStore := NewAuditStore(logger, 1000, 100)
	restAPI := NewRESTAPI(auditStore, logger)

	server := httptest.NewServer(restAPI.GetRouter())
	defer server.Close()

	t.Run("EventQueryParameters", func(t *testing.T) {
		// Test all query parameters
		url := server.URL + "/api/v1/events?" +
			"trader=test_trader&" +
			"symbol=BTC&" +
			"event_type=trade_request&" +
			"risk_level=high&" +
			"success=true&" +
			"start_time=2024-01-01T00:00:00Z&" +
			"end_time=2024-01-02T00:00:00Z&" +
			"offset=0&" +
			"limit=10"

		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("AlertQueryParameters", func(t *testing.T) {
		// Test all query parameters
		url := server.URL + "/api/v1/alerts?" +
			"trader=test_trader&" +
			"symbol=BTC&" +
			"alert_type=limit_exceeded&" +
			"severity=warning&" +
			"acknowledged=false&" +
			"start_time=2024-01-01T00:00:00Z&" +
			"end_time=2024-01-02T00:00:00Z&" +
			"offset=0&" +
			"limit=10"

		resp, err := http.Get(url)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}

func TestRESTAPI_ContentType(t *testing.T) {
	logger, err := NewLogger(DefaultLoggerConfig())
	require.NoError(t, err)
	defer logger.Close()

	auditStore := NewAuditStore(logger, 1000, 100)
	restAPI := NewRESTAPI(auditStore, logger)

	server := httptest.NewServer(restAPI.GetRouter())
	defer server.Close()

	t.Run("JSONContentType", func(t *testing.T) {
		resp, err := http.Get(server.URL + "/api/v1/events")
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	})
}
