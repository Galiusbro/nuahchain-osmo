package balances

import (
	"database/sql"
	"errors"
	"time"
)

// Repository handles database operations for user balances
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new balances repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// UserBalance represents a user's balance for a specific denom
type UserBalance struct {
	ID              int64     `json:"id"`
	UserID          int64     `json:"user_id"`
	Address         string    `json:"address"`
	Denom           string    `json:"denom"`
	Amount          string    `json:"amount"` // NUMERIC as string for precision
	UpdatedAt       time.Time `json:"updated_at"`
	UpdatedByTxHash *string   `json:"updated_by_tx_hash,omitempty"`
	UpdatedByHeight *int64    `json:"updated_by_height,omitempty"`
}

// BalanceHistory represents a balance change in history
type BalanceHistory struct {
	ID           int64     `json:"id"`
	UserID       int64     `json:"user_id"`
	Address      string    `json:"address"`
	Denom        string    `json:"denom"`
	AmountBefore *string   `json:"amount_before,omitempty"` // NULL for first entry
	AmountAfter  string    `json:"amount_after"`
	AmountDelta  string    `json:"amount_delta"` // positive = increase, negative = decrease
	TxHash       string    `json:"tx_hash"`
	Height       int64     `json:"height"`
	EventType    *string   `json:"event_type,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// UpdateBalanceRequest represents a request to update a balance
type UpdateBalanceRequest struct {
	UserID       int64
	Address      string
	Denom        string
	Amount       string
	TxHash       string
	Height       int64
	EventType    string
	AmountBefore *string // For history
}

// UpsertBalance updates or inserts a balance atomically with history
func (r *Repository) UpsertBalance(req UpdateBalanceRequest) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Get current balance for history
	var currentAmount sql.NullString
	err = tx.QueryRow(`
		SELECT amount FROM user_balances
		WHERE address = $1 AND denom = $2
	`, req.Address, req.Denom).Scan(&currentAmount)

	var amountBefore *string
	if err == nil && currentAmount.Valid {
		amountBefore = &currentAmount.String
	} else if !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// Upsert balance
	_, err = tx.Exec(`
		INSERT INTO user_balances (
			user_id, address, denom, amount,
			updated_at, updated_by_tx_hash, updated_by_height
		)
		VALUES ($1, $2, $3, $4, NOW(), $5, $6)
		ON CONFLICT (address, denom) DO UPDATE SET
			amount = EXCLUDED.amount,
			updated_at = NOW(),
			updated_by_tx_hash = EXCLUDED.updated_by_tx_hash,
			updated_by_height = EXCLUDED.updated_by_height
	`, req.UserID, req.Address, req.Denom, req.Amount, req.TxHash, req.Height)
	if err != nil {
		return err
	}

	// Calculate delta properly using SQL
	var calculatedDelta string
	if amountBefore != nil {
		// Calculate delta: amount_after - amount_before
		err = tx.QueryRow(`
			SELECT ($1::NUMERIC - $2::NUMERIC)::TEXT
		`, req.Amount, *amountBefore).Scan(&calculatedDelta)
		if err != nil {
			return err
		}
	} else {
		calculatedDelta = req.Amount // First entry, delta equals amount
	}

	// Insert history
	_, err = tx.Exec(`
		INSERT INTO balance_history (
			user_id, address, denom,
			amount_before, amount_after, amount_delta,
			tx_hash, height, event_type
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`, req.UserID, req.Address, req.Denom,
		amountBefore, req.Amount, calculatedDelta,
		req.TxHash, req.Height, req.EventType)
	if err != nil {
		return err
	}

	return tx.Commit()
}

// GetUserBalances gets all balances for a user
func (r *Repository) GetUserBalances(userID int64, denomFilter string) ([]UserBalance, error) {
	var query string
	var args []interface{}

	if denomFilter != "" {
		query = `
			SELECT id, user_id, address, denom, amount,
				updated_at, updated_by_tx_hash, updated_by_height
			FROM user_balances
			WHERE user_id = $1 AND denom = $2
			ORDER BY updated_at DESC
		`
		args = []interface{}{userID, denomFilter}
	} else {
		query = `
			SELECT id, user_id, address, denom, amount,
				updated_at, updated_by_tx_hash, updated_by_height
			FROM user_balances
			WHERE user_id = $1
			ORDER BY updated_at DESC
		`
		args = []interface{}{userID}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var balances []UserBalance
	for rows.Next() {
		var balance UserBalance
		var txHash sql.NullString
		var height sql.NullInt64

		err := rows.Scan(
			&balance.ID,
			&balance.UserID,
			&balance.Address,
			&balance.Denom,
			&balance.Amount,
			&balance.UpdatedAt,
			&txHash,
			&height,
		)
		if err != nil {
			return nil, err
		}

		if txHash.Valid {
			balance.UpdatedByTxHash = &txHash.String
		}
		if height.Valid {
			balance.UpdatedByHeight = &height.Int64
		}

		balances = append(balances, balance)
	}

	return balances, rows.Err()
}

// GetBalanceByAddress gets balance for a specific address and denom
func (r *Repository) GetBalanceByAddress(address, denom string) (*UserBalance, error) {
	var balance UserBalance
	var txHash sql.NullString
	var height sql.NullInt64

	err := r.db.QueryRow(`
		SELECT id, user_id, address, denom, amount,
			updated_at, updated_by_tx_hash, updated_by_height
		FROM user_balances
		WHERE address = $1 AND denom = $2
	`, address, denom).Scan(
		&balance.ID,
		&balance.UserID,
		&balance.Address,
		&balance.Denom,
		&balance.Amount,
		&balance.UpdatedAt,
		&txHash,
		&height,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if txHash.Valid {
		balance.UpdatedByTxHash = &txHash.String
	}
	if height.Valid {
		balance.UpdatedByHeight = &height.Int64
	}

	return &balance, nil
}

// GetUserIDByAddress gets user ID by wallet address
func (r *Repository) GetUserIDByAddress(address string) (*int64, error) {
	var userID int64
	err := r.db.QueryRow(`
		SELECT user_id FROM wallets WHERE address = $1
	`, address).Scan(&userID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &userID, nil
}

// UpdateIndexerState updates the last processed height
func (r *Repository) UpdateIndexerState(height int64, errMsg *string) error {
	_, err := r.db.Exec(`
		UPDATE balance_indexer_state
		SET last_processed_height = $1,
			last_processed_at = NOW(),
			last_error = $2,
			updated_at = NOW()
		WHERE id = 1
	`, height, errMsg)
	return err
}

// GetIndexerState gets the last processed height
func (r *Repository) GetIndexerState() (int64, time.Time, error) {
	var height int64
	var lastProcessedAt time.Time
	err := r.db.QueryRow(`
		SELECT last_processed_height, last_processed_at
		FROM balance_indexer_state
		WHERE id = 1
	`).Scan(&height, &lastProcessedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			// Initialize if not exists
			_, err := r.db.Exec(`
				INSERT INTO balance_indexer_state (id, last_processed_height, last_processed_at)
				VALUES (1, 0, NOW())
			`)
			if err != nil {
				return 0, time.Time{}, err
			}
			return 0, time.Now(), nil
		}
		return 0, time.Time{}, err
	}

	return height, lastProcessedAt, nil
}

// SaveFailedUpdate saves a failed balance update to DLQ
func (r *Repository) SaveFailedUpdate(txHash string, height int64, address, denom, amountDelta, errorMsg string) error {
	_, err := r.db.Exec(`
		INSERT INTO failed_balance_updates (
			tx_hash, height, address, denom, amount_delta, error_message
		)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, txHash, height, address, denom, amountDelta, errorMsg)
	return err
}

// GetBalanceHistory gets balance history for a user
func (r *Repository) GetBalanceHistory(userID int64, denom string, limit int) ([]BalanceHistory, error) {
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}

	var query string
	var args []interface{}

	if denom != "" {
		query = `
			SELECT id, user_id, address, denom,
				amount_before, amount_after, amount_delta,
				tx_hash, height, event_type, created_at
			FROM balance_history
			WHERE user_id = $1 AND denom = $2
			ORDER BY created_at DESC LIMIT $3
		`
		args = []interface{}{userID, denom, limit}
	} else {
		query = `
			SELECT id, user_id, address, denom,
				amount_before, amount_after, amount_delta,
				tx_hash, height, event_type, created_at
			FROM balance_history
			WHERE user_id = $1
			ORDER BY created_at DESC LIMIT $2
		`
		args = []interface{}{userID, limit}
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var history []BalanceHistory
	for rows.Next() {
		var h BalanceHistory
		var amountBefore sql.NullString
		var eventType sql.NullString

		err := rows.Scan(
			&h.ID,
			&h.UserID,
			&h.Address,
			&h.Denom,
			&amountBefore,
			&h.AmountAfter,
			&h.AmountDelta,
			&h.TxHash,
			&h.Height,
			&eventType,
			&h.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		if amountBefore.Valid {
			h.AmountBefore = &amountBefore.String
		}
		if eventType.Valid {
			h.EventType = &eventType.String
		}

		history = append(history, h)
	}

	return history, rows.Err()
}
