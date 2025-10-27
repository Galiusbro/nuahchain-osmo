package monitoring

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
)

// RESTAPI provides REST endpoints for monitoring and audit
type RESTAPI struct {
	auditStore *AuditStore
	logger     *Logger
	router     *mux.Router
}

// NewRESTAPI creates a new REST API instance
func NewRESTAPI(auditStore *AuditStore, logger *Logger) *RESTAPI {
	api := &RESTAPI{
		auditStore: auditStore,
		logger:     logger,
		router:     mux.NewRouter(),
	}

	api.setupRoutes()
	return api
}

// setupRoutes sets up the API routes
func (api *RESTAPI) setupRoutes() {
	// Events endpoints
	api.router.HandleFunc("/api/v1/events", api.getEvents).Methods("GET")
	api.router.HandleFunc("/api/v1/events/{id}", api.getEvent).Methods("GET")

	// Alerts endpoints
	api.router.HandleFunc("/api/v1/alerts", api.getAlerts).Methods("GET")
	api.router.HandleFunc("/api/v1/alerts/{id}/acknowledge", api.acknowledgeAlert).Methods("POST")

	// History endpoints
	api.router.HandleFunc("/api/v1/history/{trader}", api.getTraderHistory).Methods("GET")
	api.router.HandleFunc("/api/v1/history/{trader}/{symbol}", api.getSymbolHistory).Methods("GET")

	// Metrics endpoints
	api.router.HandleFunc("/api/v1/metrics", api.getSystemMetrics).Methods("GET")
	api.router.HandleFunc("/api/v1/metrics/trader/{trader}", api.getTraderMetrics).Methods("GET")

	// Health check
	api.router.HandleFunc("/health", api.healthCheck).Methods("GET")
}

// GetRouter returns the HTTP router
func (api *RESTAPI) GetRouter() *mux.Router {
	return api.router
}

// getEvents handles GET /api/v1/events
func (api *RESTAPI) getEvents(w http.ResponseWriter, r *http.Request) {
	req := api.parseEventQueryRequest(r)

	events, err := api.auditStore.GetEvents(r.Context(), req)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to get events", err)
		return
	}

	response := EventQueryResponse{
		Events: events,
		Total:  len(events),
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	api.writeJSON(w, http.StatusOK, response)
}

// getEvent handles GET /api/v1/events/{id}
func (api *RESTAPI) getEvent(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	_ = vars["id"] // eventID

	// This would require a GetEventByID method in AuditStore
	// For now, return not implemented
	api.writeError(w, http.StatusNotImplemented, "Get event by ID not implemented", nil)
}

// getAlerts handles GET /api/v1/alerts
func (api *RESTAPI) getAlerts(w http.ResponseWriter, r *http.Request) {
	req := api.parseAlertQueryRequest(r)

	alerts, err := api.auditStore.GetAlerts(r.Context(), req)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to get alerts", err)
		return
	}

	response := AlertQueryResponse{
		Alerts: alerts,
		Total:  len(alerts),
		Offset: req.Offset,
		Limit:  req.Limit,
	}

	api.writeJSON(w, http.StatusOK, response)
}

// acknowledgeAlert handles POST /api/v1/alerts/{id}/acknowledge
func (api *RESTAPI) acknowledgeAlert(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	alertID := vars["id"]

	// Get acknowledgedBy from request body or header
	acknowledgedBy := r.Header.Get("X-User-ID")
	if acknowledgedBy == "" {
		acknowledgedBy = "system"
	}

	err := api.auditStore.AcknowledgeAlert(r.Context(), alertID, acknowledgedBy)
	if err != nil {
		api.writeError(w, http.StatusNotFound, "Failed to acknowledge alert", err)
		return
	}

	api.writeJSON(w, http.StatusOK, map[string]string{"status": "acknowledged"})
}

// getTraderHistory handles GET /api/v1/history/{trader}
func (api *RESTAPI) getTraderHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trader := vars["trader"]

	history, err := api.auditStore.GetTradeHistory(r.Context(), trader, "")
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to get trader history", err)
		return
	}

	api.writeJSON(w, http.StatusOK, history)
}

// getSymbolHistory handles GET /api/v1/history/{trader}/{symbol}
func (api *RESTAPI) getSymbolHistory(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trader := vars["trader"]
	symbol := vars["symbol"]

	history, err := api.auditStore.GetTradeHistory(r.Context(), trader, symbol)
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to get symbol history", err)
		return
	}

	api.writeJSON(w, http.StatusOK, history)
}

// getSystemMetrics handles GET /api/v1/metrics
func (api *RESTAPI) getSystemMetrics(w http.ResponseWriter, r *http.Request) {
	metrics, err := api.auditStore.GetSystemMetrics(r.Context())
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to get system metrics", err)
		return
	}

	api.writeJSON(w, http.StatusOK, metrics)
}

// getTraderMetrics handles GET /api/v1/metrics/trader/{trader}
func (api *RESTAPI) getTraderMetrics(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	trader := vars["trader"]

	history, err := api.auditStore.GetTradeHistory(r.Context(), trader, "")
	if err != nil {
		api.writeError(w, http.StatusInternalServerError, "Failed to get trader metrics", err)
		return
	}

	api.writeJSON(w, http.StatusOK, history)
}

// healthCheck handles GET /health
func (api *RESTAPI) healthCheck(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"status":    "healthy",
		"timestamp": time.Now(),
		"version":   "1.0.0",
	}

	api.writeJSON(w, http.StatusOK, status)
}

// Helper methods

func (api *RESTAPI) parseEventQueryRequest(r *http.Request) EventQueryRequest {
	req := DefaultEventQueryRequest()

	if trader := r.URL.Query().Get("trader"); trader != "" {
		req.Trader = trader
	}
	if symbol := r.URL.Query().Get("symbol"); symbol != "" {
		req.Symbol = symbol
	}
	if eventType := r.URL.Query().Get("event_type"); eventType != "" {
		req.EventType = eventType
	}
	if riskLevel := r.URL.Query().Get("risk_level"); riskLevel != "" {
		req.RiskLevel = riskLevel
	}
	if success := r.URL.Query().Get("success"); success != "" {
		if successBool, err := strconv.ParseBool(success); err == nil {
			req.Success = &successBool
		}
	}
	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			req.StartTime = t
		}
	}
	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			req.EndTime = t
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	return req
}

func (api *RESTAPI) parseAlertQueryRequest(r *http.Request) AlertQueryRequest {
	req := DefaultAlertQueryRequest()

	if trader := r.URL.Query().Get("trader"); trader != "" {
		req.Trader = trader
	}
	if symbol := r.URL.Query().Get("symbol"); symbol != "" {
		req.Symbol = symbol
	}
	if alertType := r.URL.Query().Get("alert_type"); alertType != "" {
		req.AlertType = alertType
	}
	if severity := r.URL.Query().Get("severity"); severity != "" {
		req.Severity = severity
	}
	if acknowledged := r.URL.Query().Get("acknowledged"); acknowledged != "" {
		if ackBool, err := strconv.ParseBool(acknowledged); err == nil {
			req.Acknowledged = &ackBool
		}
	}
	if startTime := r.URL.Query().Get("start_time"); startTime != "" {
		if t, err := time.Parse(time.RFC3339, startTime); err == nil {
			req.StartTime = t
		}
	}
	if endTime := r.URL.Query().Get("end_time"); endTime != "" {
		if t, err := time.Parse(time.RFC3339, endTime); err == nil {
			req.EndTime = t
		}
	}
	if offset := r.URL.Query().Get("offset"); offset != "" {
		if o, err := strconv.Atoi(offset); err == nil {
			req.Offset = o
		}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		if l, err := strconv.Atoi(limit); err == nil {
			req.Limit = l
		}
	}

	return req
}

func (api *RESTAPI) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		api.logger.logger.WithError(err).Error("Failed to encode JSON response")
	}
}

func (api *RESTAPI) writeError(w http.ResponseWriter, status int, message string, err error) {
	errorResponse := map[string]interface{}{
		"error":   message,
		"status":  status,
		"details": err.Error(),
	}

	api.writeJSON(w, status, errorResponse)
}
