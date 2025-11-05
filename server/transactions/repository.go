package transactions

import (
	"database/sql"
	"encoding/json"
	"errors"
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
func (r *Repository) UpdateTransactionByTxHash(txHash string, status string, operationData map[string]interface{}, errorMessage *string) error {
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
