package transactions

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// Repository обрабатывает операции с базой данных для транзакций
type Repository struct {
	db *sql.DB
}

// NewRepository создает новый репозиторий для транзакций
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateTransaction создает новую запись о транзакции
func (r *Repository) CreateTransaction(req CreateTransactionRequest) (*Transaction, error) {
	// Преобразуем operation_data в JSON
	operationDataJSON, err := json.Marshal(req.OperationData)
	if err != nil {
		return nil, err
	}

	var transactionID int64
	err = r.db.QueryRow(`
		INSERT INTO transactions (
			user_id,
			operation_type,
			tx_hash,
			status,
			operation_data,
			error_message,
			created_at,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
		RETURNING id
	`, req.UserID, req.OperationType, req.TxHash, req.Status, operationDataJSON, req.ErrorMessage).Scan(&transactionID)

	if err != nil {
		return nil, err
	}

	return r.GetTransactionByID(transactionID)
}

// GetTransactionByID получает транзакцию по ID
func (r *Repository) GetTransactionByID(id int64) (*Transaction, error) {
	transaction := &Transaction{ID: id}
	var operationDataJSON []byte
	var errorMessage sql.NullString

	err := r.db.QueryRow(`
		SELECT
			user_id,
			operation_type,
			tx_hash,
			status,
			operation_data,
			error_message,
			created_at,
			updated_at
		FROM transactions
		WHERE id = $1
	`, id).Scan(
		&transaction.UserID,
		&transaction.OperationType,
		&transaction.TxHash,
		&transaction.Status,
		&operationDataJSON,
		&errorMessage,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Парсим JSON данные
	if err := json.Unmarshal(operationDataJSON, &transaction.OperationData); err != nil {
		return nil, err
	}

	// Обрабатываем error_message
	if errorMessage.Valid {
		transaction.ErrorMessage = &errorMessage.String
	}

	return transaction, nil
}

// GetTransactionByTxHash получает транзакцию по хешу
func (r *Repository) GetTransactionByTxHash(txHash string) (*Transaction, error) {
	transaction := &Transaction{TxHash: txHash}
	var operationDataJSON []byte
	var errorMessage sql.NullString

	err := r.db.QueryRow(`
		SELECT
			id,
			user_id,
			operation_type,
			status,
			operation_data,
			error_message,
			created_at,
			updated_at
		FROM transactions
		WHERE tx_hash = $1
		ORDER BY created_at DESC
		LIMIT 1
	`, txHash).Scan(
		&transaction.ID,
		&transaction.UserID,
		&transaction.OperationType,
		&transaction.Status,
		&operationDataJSON,
		&errorMessage,
		&transaction.CreatedAt,
		&transaction.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("transaction not found")
		}
		return nil, err
	}

	// Парсим JSON данные
	if err := json.Unmarshal(operationDataJSON, &transaction.OperationData); err != nil {
		return nil, err
	}

	// Обрабатываем error_message
	if errorMessage.Valid {
		transaction.ErrorMessage = &errorMessage.String
	}

	return transaction, nil
}

// UpdateTransaction обновляет статус и данные транзакции
func (r *Repository) UpdateTransaction(req UpdateTransactionRequest) error {
	var operationDataJSON []byte
	var err error

	// Если есть обновленные данные, преобразуем в JSON
	if req.OperationData != nil {
		operationDataJSON, err = json.Marshal(req.OperationData)
		if err != nil {
			return err
		}
	}

	// Если есть обновленные данные, обновляем их
	if req.OperationData != nil {
		_, err = r.db.Exec(`
			UPDATE transactions
			SET
				status = $1,
				operation_data = $2,
				error_message = $3,
				updated_at = NOW()
			WHERE id = $4
		`, req.Status, operationDataJSON, req.ErrorMessage, req.ID)
	} else {
		// Если данных нет, обновляем только статус и ошибку
		_, err = r.db.Exec(`
			UPDATE transactions
			SET
				status = $1,
				error_message = $2,
				updated_at = NOW()
			WHERE id = $3
		`, req.Status, req.ErrorMessage, req.ID)
	}

	return err
}

// GetUserTransactions получает список транзакций пользователя
func (r *Repository) GetUserTransactions(userID int64, limit, offset int) ([]*Transaction, error) {
	rows, err := r.db.Query(`
		SELECT
			id,
			user_id,
			operation_type,
			tx_hash,
			status,
			operation_data,
			error_message,
			created_at,
			updated_at
		FROM transactions
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, userID, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		transaction := &Transaction{}
		var operationDataJSON []byte
		var errorMessage sql.NullString

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.OperationType,
			&transaction.TxHash,
			&transaction.Status,
			&operationDataJSON,
			&errorMessage,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Парсим JSON данные
		if err := json.Unmarshal(operationDataJSON, &transaction.OperationData); err != nil {
			return nil, err
		}

		// Обрабатываем error_message
		if errorMessage.Valid {
			transaction.ErrorMessage = &errorMessage.String
		}

		transactions = append(transactions, transaction)
	}

	return transactions, rows.Err()
}

// GetUserTransactionsByType получает транзакции пользователя определенного типа
func (r *Repository) GetUserTransactionsByType(userID int64, operationType string, limit, offset int) ([]*Transaction, error) {
	rows, err := r.db.Query(`
		SELECT
			id,
			user_id,
			operation_type,
			tx_hash,
			status,
			operation_data,
			error_message,
			created_at,
			updated_at
		FROM transactions
		WHERE user_id = $1 AND operation_type = $2
		ORDER BY created_at DESC
		LIMIT $3 OFFSET $4
	`, userID, operationType, limit, offset)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []*Transaction
	for rows.Next() {
		transaction := &Transaction{}
		var operationDataJSON []byte
		var errorMessage sql.NullString

		err := rows.Scan(
			&transaction.ID,
			&transaction.UserID,
			&transaction.OperationType,
			&transaction.TxHash,
			&transaction.Status,
			&operationDataJSON,
			&errorMessage,
			&transaction.CreatedAt,
			&transaction.UpdatedAt,
		)

		if err != nil {
			return nil, err
		}

		// Парсим JSON данные
		if err := json.Unmarshal(operationDataJSON, &transaction.OperationData); err != nil {
			return nil, err
		}

		// Обрабатываем error_message
		if errorMessage.Valid {
			transaction.ErrorMessage = &errorMessage.String
		}

		transactions = append(transactions, transaction)
	}

	return transactions, rows.Err()
}

// UpdateTransactionByTxHash обновляет транзакцию по хешу (удобно для обновления статуса после проверки в блокчейне)
func (r *Repository) UpdateTransactionByTxHash(txHash string, status TransactionStatus, operationData map[string]interface{}, errorMessage *string) error {
	var operationDataJSON []byte
	var err error

	if operationData != nil {
		operationDataJSON, err = json.Marshal(operationData)
		if err != nil {
			return err
		}
	}

	if operationData != nil {
		_, err = r.db.Exec(`
			UPDATE transactions
			SET
				status = $1,
				operation_data = $2,
				error_message = $3,
				updated_at = NOW()
			WHERE tx_hash = $4
		`, status, operationDataJSON, errorMessage, txHash)
	} else {
		_, err = r.db.Exec(`
			UPDATE transactions
			SET
				status = $1,
				error_message = $2,
				updated_at = NOW()
			WHERE tx_hash = $3
		`, status, errorMessage, txHash)
	}

	return err
}

// ListPendingTransactions возвращает все транзакции со статусом PENDING (можно ограничить количеством)
func (r *Repository) ListPendingTransactions(limit int) ([]*Transaction, error) {
	query := `
		SELECT
			id,
			user_id,
			operation_type,
			tx_hash,
			status,
			operation_data,
			error_message,
			created_at,
			updated_at
		FROM transactions
		WHERE status = $1
		ORDER BY created_at ASC`

	args := []interface{}{StatusPending}
	if limit > 0 {
		query += ` LIMIT $2`
		args = append(args, limit)
	}

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []*Transaction
	for rows.Next() {
		var (
			tx                = &Transaction{}
			operationDataJSON []byte
			errorMessage      sql.NullString
		)

		if err := rows.Scan(
			&tx.ID,
			&tx.UserID,
			&tx.OperationType,
			&tx.TxHash,
			&tx.Status,
			&operationDataJSON,
			&errorMessage,
			&tx.CreatedAt,
			&tx.UpdatedAt,
		); err != nil {
			return nil, err
		}

		if err := json.Unmarshal(operationDataJSON, &tx.OperationData); err != nil {
			return nil, err
		}

		if errorMessage.Valid {
			tx.ErrorMessage = &errorMessage.String
		}

		result = append(result, tx)
	}

	return result, rows.Err()
}

// GetUserTransactionStats получает статистику транзакций пользователя
func (r *Repository) GetUserTransactionStats(userID int64) (*UserTransactionStats, error) {
	stats := &UserTransactionStats{}

	// Общее количество транзакций
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM transactions WHERE user_id = $1
	`, userID).Scan(&stats.TotalTransactions)
	if err != nil {
		return nil, err
	}

	// Количество по статусам
	err = r.db.QueryRow(`
		SELECT
			COUNT(*) FILTER (WHERE status = 'SUCCESS') as successful,
			COUNT(*) FILTER (WHERE status = 'FAILED') as failed,
			COUNT(*) FILTER (WHERE status = 'PENDING') as pending
		FROM transactions WHERE user_id = $1
	`, userID).Scan(&stats.SuccessfulTransactions, &stats.FailedTransactions, &stats.PendingTransactions)
	if err != nil {
		return nil, err
	}

	// Количество по типам операций
	rows, err := r.db.Query(`
		SELECT operation_type, COUNT(*) as count
		FROM transactions
		WHERE user_id = $1
		GROUP BY operation_type
		ORDER BY count DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	stats.TransactionsByType = make(map[string]int)
	for rows.Next() {
		var opType string
		var count int
		if err := rows.Scan(&opType, &count); err != nil {
			return nil, err
		}
		stats.TransactionsByType[opType] = count
	}

	// Дата последней транзакции
	var lastTxAt sql.NullTime
	err = r.db.QueryRow(`
		SELECT MAX(created_at) FROM transactions WHERE user_id = $1
	`, userID).Scan(&lastTxAt)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	if lastTxAt.Valid {
		stats.LastTransactionAt = &lastTxAt.Time
	}

	// Количество созданных токенов
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM transactions
		WHERE user_id = $1 AND operation_type = 'TOKEN_CREATE' AND status = 'SUCCESS'
	`, userID).Scan(&stats.TokensCreated)
	if err != nil {
		return nil, err
	}

	// Количество открытых margin позиций (позиции, которые были открыты, но не закрыты)
	// Считаем открытые позиции как те, у которых есть ASSET_MARGIN_OPEN, но нет соответствующего ASSET_MARGIN_CLOSE
	err = r.db.QueryRow(`
		SELECT COUNT(DISTINCT operation_data->>'position_id')
		FROM transactions
		WHERE user_id = $1
			AND operation_type = 'ASSET_MARGIN_OPEN'
			AND status = 'SUCCESS'
			AND operation_data->>'position_id' IS NOT NULL
			AND operation_data->>'position_id' != ''
			AND NOT EXISTS (
				SELECT 1 FROM transactions t2
				WHERE t2.user_id = $1
					AND t2.operation_type = 'ASSET_MARGIN_CLOSE'
					AND t2.status = 'SUCCESS'
					AND t2.operation_data->>'position_id' = transactions.operation_data->>'position_id'
			)
	`, userID).Scan(&stats.MarginPositionsOpen)
	if err != nil {
		// Если ошибка, просто ставим 0 (может быть проблема с JSON запросом)
		stats.MarginPositionsOpen = 0
	}

	return stats, nil
}

// UserTransactionStats представляет статистику транзакций пользователя
type UserTransactionStats struct {
	TotalTransactions      int            `json:"total_transactions"`
	SuccessfulTransactions int            `json:"successful_transactions"`
	FailedTransactions     int            `json:"failed_transactions"`
	PendingTransactions    int            `json:"pending_transactions"`
	TransactionsByType     map[string]int `json:"transactions_by_type"`
	LastTransactionAt      *time.Time     `json:"last_transaction_at,omitempty"`
	TokensCreated          int            `json:"tokens_created"`
	MarginPositionsOpen    int            `json:"margin_positions_open"`
}
