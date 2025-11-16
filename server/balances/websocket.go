package balances

import (
	"net/http"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/osmosis-labs/osmosis/v30/server/auth"
)

var globalBalancesIndexer *Indexer
var globalAuthServiceForWS interface {
	ValidateToken(token string) (*auth.User, error)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins in development
	},
}

// SetIndexerForWS sets the global indexer for WebSocket
func SetIndexerForWS(indexer *Indexer) {
	globalBalancesIndexer = indexer
}

// SetAuthServiceForWS sets the global auth service for WebSocket
func SetAuthServiceForWS(s interface {
	ValidateToken(token string) (*auth.User, error)
}) {
	globalAuthServiceForWS = s
}

// HandleBalanceWebSocket handles WebSocket connection for real-time balance updates
// (GET /api/users/balances/ws)
func HandleBalanceWebSocket(w http.ResponseWriter, r *http.Request) {
	// Authenticate user
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Authorization header required", http.StatusUnauthorized)
		return
	}

	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
		return
	}

	token := parts[1]
	user, err := globalAuthServiceForWS.ValidateToken(token)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	// Subscribe to balance updates
	if globalBalancesIndexer == nil {
		conn.WriteJSON(map[string]interface{}{
			"type":    "error",
			"message": "Balance indexer not configured",
		})
		return
	}

	updateChannel := globalBalancesIndexer.GetUpdateChannel()

	// Send initial message
	conn.WriteJSON(map[string]interface{}{
		"type":    "connected",
		"user_id": user.ID,
		"message": "Connected to balance updates",
	})

	// Handle incoming messages (for subscribe/unsubscribe)
	var subscribedDenoms map[string]bool
	subscribedDenoms = make(map[string]bool) // Empty = all denoms

	go func() {
		for {
			var msg map[string]interface{}
			if err := conn.ReadJSON(&msg); err != nil {
				break
			}

			action, ok := msg["action"].(string)
			if !ok {
				continue
			}

			switch action {
			case "subscribe":
				// Subscribe to specific denoms
				if denoms, ok := msg["denoms"].([]interface{}); ok {
					subscribedDenoms = make(map[string]bool)
					for _, denom := range denoms {
						if d, ok := denom.(string); ok {
							subscribedDenoms[d] = true
						}
					}
				} else {
					// Subscribe to all
					subscribedDenoms = nil
				}

				conn.WriteJSON(map[string]interface{}{
					"type":    "subscribed",
					"denoms":  msg["denoms"],
					"message": "Subscribed to balance updates",
				})

			case "unsubscribe":
				subscribedDenoms = make(map[string]bool)
				conn.WriteJSON(map[string]interface{}{
					"type":    "unsubscribed",
					"message": "Unsubscribed from balance updates",
				})
			}
		}
	}()

	// Send balance updates
	for {
		select {
		case update := <-updateChannel:
			if update == nil {
				return
			}

			// Only send updates for this user
			if update.UserID != user.ID {
				continue
			}

			// Filter by denom if subscribed to specific denoms
			if subscribedDenoms != nil && len(subscribedDenoms) > 0 {
				if !subscribedDenoms[update.Denom] {
					continue
				}
			}

			// Send update
			if err := conn.WriteJSON(map[string]interface{}{
				"type":      "balance_update",
				"user_id":   update.UserID,
				"address":   update.Address,
				"denom":     update.Denom,
				"amount":    update.Amount,
				"tx_hash":   update.TxHash,
				"height":    update.Height,
				"timestamp": update.Timestamp,
			}); err != nil {
				return
			}
		}
	}
}

// BalanceWebSocketManager manages WebSocket connections for balance updates
type BalanceWebSocketManager struct {
	connections map[int64]map[*websocket.Conn]bool // user_id -> connections
	mu          sync.RWMutex
	updateChan  <-chan *BalanceUpdate
}

// NewBalanceWebSocketManager creates a new WebSocket manager
func NewBalanceWebSocketManager(updateChan <-chan *BalanceUpdate) *BalanceWebSocketManager {
	return &BalanceWebSocketManager{
		connections: make(map[int64]map[*websocket.Conn]bool),
		updateChan:  updateChan,
	}
}

// AddConnection adds a WebSocket connection for a user
func (m *BalanceWebSocketManager) AddConnection(userID int64, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[userID] == nil {
		m.connections[userID] = make(map[*websocket.Conn]bool)
	}
	m.connections[userID][conn] = true
}

// RemoveConnection removes a WebSocket connection
func (m *BalanceWebSocketManager) RemoveConnection(userID int64, conn *websocket.Conn) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.connections[userID] != nil {
		delete(m.connections[userID], conn)
		if len(m.connections[userID]) == 0 {
			delete(m.connections, userID)
		}
	}
}

// BroadcastUpdate broadcasts a balance update to all connected clients for that user
func (m *BalanceWebSocketManager) BroadcastUpdate(update *BalanceUpdate) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	conns, exists := m.connections[update.UserID]
	if !exists {
		return
	}

	msg := map[string]interface{}{
		"type":      "balance_update",
		"user_id":   update.UserID,
		"address":   update.Address,
		"denom":     update.Denom,
		"amount":    update.Amount,
		"tx_hash":   update.TxHash,
		"height":    update.Height,
		"timestamp": update.Timestamp,
	}

	for conn := range conns {
		if err := conn.WriteJSON(msg); err != nil {
			// Remove failed connection
			delete(conns, conn)
			conn.Close()
		}
	}
}

