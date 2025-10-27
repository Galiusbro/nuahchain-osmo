package monitoring

import "time"

// EventQueryRequest represents a request to query audit events
type EventQueryRequest struct {
	Trader    string    `json:"trader,omitempty"`
	Symbol    string    `json:"symbol,omitempty"`
	EventType string    `json:"event_type,omitempty"`
	RiskLevel string    `json:"risk_level,omitempty"`
	Success   *bool     `json:"success,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
	Offset    int       `json:"offset"`
	Limit     int       `json:"limit"`
}

// AlertQueryRequest represents a request to query alerts
type AlertQueryRequest struct {
	Trader       string    `json:"trader,omitempty"`
	Symbol       string    `json:"symbol,omitempty"`
	AlertType    string    `json:"alert_type,omitempty"`
	Severity     string    `json:"severity,omitempty"`
	Acknowledged *bool     `json:"acknowledged,omitempty"`
	StartTime    time.Time `json:"start_time,omitempty"`
	EndTime      time.Time `json:"end_time,omitempty"`
	Offset       int       `json:"offset"`
	Limit        int       `json:"limit"`
}

// EventQueryResponse represents the response to an event query
type EventQueryResponse struct {
	Events []AuditEvent `json:"events"`
	Total  int          `json:"total"`
	Offset int          `json:"offset"`
	Limit  int          `json:"limit"`
}

// AlertQueryResponse represents the response to an alert query
type AlertQueryResponse struct {
	Alerts []Alert `json:"alerts"`
	Total  int     `json:"total"`
	Offset int     `json:"offset"`
	Limit  int     `json:"limit"`
}

// DefaultQueryRequest returns a default query request
func DefaultEventQueryRequest() EventQueryRequest {
	return EventQueryRequest{
		Offset: 0,
		Limit:  100,
	}
}

// DefaultAlertQueryRequest returns a default alert query request
func DefaultAlertQueryRequest() AlertQueryRequest {
	return AlertQueryRequest{
		Offset: 0,
		Limit:  100,
	}
}
