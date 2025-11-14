package monitor

import (
	"database/sql"
	"encoding/json"
	"fmt"

	"github.com/lib/pq"
)

// Repository handles database operations for blockchain transactions
type Repository struct {
	db *sql.DB
}

// NewRepository creates a new repository
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// CreateTransaction creates a new blockchain transaction record
func (r *Repository) CreateTransaction(req CreateBlockchainTransactionRequest) (*BlockchainTransaction, error) {
	eventsJSON, err := json.Marshal(req.Events)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal events: %w", err)
	}

	var messagesJSON sql.NullString
	if req.Messages != nil && len(req.Messages) > 0 {
		marshaled, err := json.Marshal(req.Messages)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal messages: %w", err)
		}
		messagesJSON = sql.NullString{String: string(marshaled), Valid: true}
	}

	var feeAmountJSON sql.NullString
	if req.FeeAmount != nil && len(req.FeeAmount) > 0 {
		marshaled, err := json.Marshal(req.FeeAmount)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal fee_amount: %w", err)
		}
		feeAmountJSON = sql.NullString{String: string(marshaled), Valid: true}
	}

	var msgResponsesJSON sql.NullString
	if req.MsgResponses != nil && len(req.MsgResponses) > 0 {
		marshaled, err := json.Marshal(req.MsgResponses)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal msg_responses: %w", err)
		}
		msgResponsesJSON = sql.NullString{String: string(marshaled), Valid: true}
	}

	var signersJSON sql.NullString
	if req.Signers != nil && len(req.Signers) > 0 {
		marshaled, err := json.Marshal(req.Signers)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal signers: %w", err)
		}
		signersJSON = sql.NullString{String: string(marshaled), Valid: true}
	}

	query := `
		INSERT INTO blockchain_transactions (
			tx_hash, height, code, codespace, success, log, raw_log, info,
			gas_wanted, gas_used, fee_amount, messages, signers, memo,
			events, msg_responses, operation_type, sender, module_name,
			block_timestamp, tx_bytes, data, enriched, enriched_at, created_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24, NOW())
		ON CONFLICT (tx_hash) DO UPDATE SET
			codespace = COALESCE(EXCLUDED.codespace, blockchain_transactions.codespace),
			raw_log = COALESCE(EXCLUDED.raw_log, blockchain_transactions.raw_log),
			info = COALESCE(EXCLUDED.info, blockchain_transactions.info),
			gas_wanted = COALESCE(EXCLUDED.gas_wanted, blockchain_transactions.gas_wanted),
			gas_used = COALESCE(EXCLUDED.gas_used, blockchain_transactions.gas_used),
			fee_amount = COALESCE(EXCLUDED.fee_amount, blockchain_transactions.fee_amount),
			messages = COALESCE(EXCLUDED.messages, blockchain_transactions.messages),
			signers = COALESCE(EXCLUDED.signers, blockchain_transactions.signers),
			memo = COALESCE(EXCLUDED.memo, blockchain_transactions.memo),
			msg_responses = COALESCE(EXCLUDED.msg_responses, blockchain_transactions.msg_responses),
			block_timestamp = COALESCE(EXCLUDED.block_timestamp, blockchain_transactions.block_timestamp),
			data = COALESCE(EXCLUDED.data, blockchain_transactions.data),
			enriched = EXCLUDED.enriched,
			enriched_at = COALESCE(EXCLUDED.enriched_at, blockchain_transactions.enriched_at)
		RETURNING tx_hash, height, code, codespace, success, log, raw_log, info,
			gas_wanted, gas_used, fee_amount, messages, signers, memo,
			events, msg_responses, operation_type, sender, module_name,
			block_timestamp, created_at, enriched, enriched_at, data, tx_bytes
	`

	var tx BlockchainTransaction
	var eventsJSONStr, messagesJSONStr, feeAmountJSONStr, msgResponsesJSONStr sql.NullString
	var signersStr sql.NullString

	enriched := false
	if req.EnrichedAt != nil {
		enriched = true
	}

	err = r.db.QueryRow(query,
		req.TxHash, req.Height, req.Code, req.Codespace, req.Success,
		req.Log, req.RawLog, req.Info,
		req.GasWanted, req.GasUsed,
		feeAmountJSON, messagesJSON,
		signersJSON, req.Memo,
		eventsJSON, msgResponsesJSON,
		req.OperationType, req.Sender, req.ModuleName,
		req.BlockTimestamp, req.TxBytes, req.Data,
		enriched, req.EnrichedAt,
	).Scan(
		&tx.TxHash, &tx.Height, &tx.Code, &tx.Codespace, &tx.Success,
		&tx.Log, &tx.RawLog, &tx.Info,
		&tx.GasWanted, &tx.GasUsed,
		&feeAmountJSONStr, &messagesJSONStr,
		&signersStr, &tx.Memo,
		&eventsJSONStr, &msgResponsesJSONStr,
		&tx.OperationType, &tx.Sender, &tx.ModuleName,
		&tx.BlockTimestamp, &tx.CreatedAt,
		&tx.Enriched, &tx.EnrichedAt,
		&tx.Data, &tx.TxBytes,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create transaction: %w", err)
	}

	// Unmarshal JSON fields
	if eventsJSONStr.Valid {
		json.Unmarshal([]byte(eventsJSONStr.String), &tx.Events)
	}
	if messagesJSONStr.Valid {
		json.Unmarshal([]byte(messagesJSONStr.String), &tx.Messages)
	}
	if feeAmountJSONStr.Valid {
		json.Unmarshal([]byte(feeAmountJSONStr.String), &tx.FeeAmount)
	}
	if msgResponsesJSONStr.Valid {
		json.Unmarshal([]byte(msgResponsesJSONStr.String), &tx.MsgResponses)
	}
	if signersStr.Valid {
		json.Unmarshal([]byte(signersStr.String), &tx.Signers)
	}

	return &tx, nil
}

// UpdateTransaction updates a blockchain transaction with enriched data
func (r *Repository) UpdateTransaction(txHash string, req UpdateBlockchainTransactionRequest) error {
	updates := []string{}
	args := []interface{}{}
	argIdx := 1

	// ЛОГИРУЕМ ВХОДЯЩИЕ ДАННЫЕ для отладки
	fmt.Printf("[UpdateTransaction] DEBUG: Codespace=%v, RawLog=%v, GasWanted=%v, GasUsed=%v, Data=%v\n",
		req.Codespace != nil, req.RawLog != nil, req.GasWanted != nil, req.GasUsed != nil, req.Data != nil)
	fmt.Printf("[UpdateTransaction] DEBUG: Messages=%v (len=%d), Signers=%v (len=%d), FeeAmount=%v, Memo=%v, BlockTimestamp=%v\n",
		req.Messages != nil, len(req.Messages), req.Signers != nil, len(req.Signers), req.FeeAmount != nil, req.Memo != nil, req.BlockTimestamp != nil)

	// ФИКСИРУЕМ ПОРЯДОК: всегда добавляем параметры в одном порядке
	// 1. Codespace
	if req.Codespace != nil {
		updates = append(updates, fmt.Sprintf("codespace = $%d::varchar", argIdx))
		args = append(args, *req.Codespace)
		fmt.Printf("[UpdateTransaction] Added codespace at $%d\n", argIdx)
		argIdx++
	}
	// 2. RawLog
	if req.RawLog != nil {
		updates = append(updates, fmt.Sprintf("raw_log = $%d::text", argIdx))
		args = append(args, *req.RawLog)
		fmt.Printf("[UpdateTransaction] Added raw_log at $%d\n", argIdx)
		argIdx++
	}
	// 3. Info
	if req.Info != nil {
		updates = append(updates, fmt.Sprintf("info = $%d::text", argIdx))
		args = append(args, *req.Info)
		fmt.Printf("[UpdateTransaction] Added info at $%d\n", argIdx)
		argIdx++
	}
	// 4. GasWanted
	if req.GasWanted != nil {
		updates = append(updates, fmt.Sprintf("gas_wanted = $%d::bigint", argIdx))
		args = append(args, *req.GasWanted)
		fmt.Printf("[UpdateTransaction] Added gas_wanted at $%d\n", argIdx)
		argIdx++
	}
	// 5. GasUsed
	if req.GasUsed != nil {
		updates = append(updates, fmt.Sprintf("gas_used = $%d::bigint", argIdx))
		args = append(args, *req.GasUsed)
		fmt.Printf("[UpdateTransaction] Added gas_used at $%d\n", argIdx)
		argIdx++
	}
	// 6. FeeAmount
	if req.FeeAmount != nil {
		feeJSON, err := json.Marshal(req.FeeAmount)
		if err != nil {
			return fmt.Errorf("failed to marshal fee_amount: %w", err)
		}
		updates = append(updates, fmt.Sprintf("fee_amount = $%d::jsonb", argIdx))
		args = append(args, string(feeJSON))
		fmt.Printf("[UpdateTransaction] Added fee_amount at $%d\n", argIdx)
		argIdx++
	}
	// 7. Messages
	if req.Messages != nil {
		messagesJSON, err := json.Marshal(req.Messages)
		if err != nil {
			return fmt.Errorf("failed to marshal messages: %w", err)
		}
		updates = append(updates, fmt.Sprintf("messages = $%d::jsonb", argIdx))
		args = append(args, string(messagesJSON))
		fmt.Printf("[UpdateTransaction] Added messages at $%d\n", argIdx)
		argIdx++
	}
	// 8. Signers (ВАЖНО: даже если пустой массив, добавляем как пустой массив!)
	if req.Signers != nil {
		// Use pq.Array for PostgreSQL array type
		updates = append(updates, fmt.Sprintf("signers = $%d::text[]", argIdx))
		args = append(args, pq.Array(req.Signers))
		fmt.Printf("[UpdateTransaction] Added signers at $%d (len=%d)\n", argIdx, len(req.Signers))
		argIdx++
	}
	// 9. Memo
	if req.Memo != nil {
		updates = append(updates, fmt.Sprintf("memo = $%d::text", argIdx))
		args = append(args, *req.Memo)
		fmt.Printf("[UpdateTransaction] Added memo at $%d\n", argIdx)
		argIdx++
	}
	// 10. MsgResponses
	if req.MsgResponses != nil && len(req.MsgResponses) > 0 {
		msgResponsesJSON, err := json.Marshal(req.MsgResponses)
		if err != nil {
			return fmt.Errorf("failed to marshal msg_responses: %w", err)
		}
		updates = append(updates, fmt.Sprintf("msg_responses = $%d::jsonb", argIdx))
		args = append(args, string(msgResponsesJSON))
		fmt.Printf("[UpdateTransaction] Added msg_responses at $%d\n", argIdx)
		argIdx++
	}
	// 11. BlockTimestamp
	if req.BlockTimestamp != nil {
		updates = append(updates, fmt.Sprintf("block_timestamp = $%d::timestamptz", argIdx))
		args = append(args, *req.BlockTimestamp)
		fmt.Printf("[UpdateTransaction] Added block_timestamp at $%d\n", argIdx)
		argIdx++
	}
	// 12. Data
	if req.Data != nil {
		updates = append(updates, fmt.Sprintf("data = $%d::text", argIdx))
		args = append(args, *req.Data)
		fmt.Printf("[UpdateTransaction] Added data at $%d\n", argIdx)
		argIdx++
	}
	// 13. Enriched (всегда)
	updates = append(updates, fmt.Sprintf("enriched = $%d::boolean", argIdx))
	args = append(args, req.Enriched)
	fmt.Printf("[UpdateTransaction] Added enriched at $%d\n", argIdx)
	argIdx++
	// 14. EnrichedAt
	if req.EnrichedAt != nil {
		updates = append(updates, fmt.Sprintf("enriched_at = $%d::timestamptz", argIdx))
		args = append(args, *req.EnrichedAt)
		fmt.Printf("[UpdateTransaction] Added enriched_at at $%d\n", argIdx)
		argIdx++
	} else if req.Enriched {
		// Only set enriched_at if enriched is true and EnrichedAt is not provided
		updates = append(updates, "enriched_at = NOW()")
		fmt.Printf("[UpdateTransaction] Added enriched_at = NOW()\n")
	}

	fmt.Printf("[UpdateTransaction] Total updates: %d, Total args: %d\n", len(updates), len(args))

	if len(updates) == 0 {
		return nil // Nothing to update
	}

	// Формируем SET clause
	setClause := updates[0]
	for i := 1; i < len(updates); i++ {
		setClause += ", " + updates[i]
	}

	// WHERE clause использует следующий номер параметра (после всех SET параметров)
	// Вычисляем ДО добавления txHash в args
	whereParamNum := len(args) + 1
	args = append(args, txHash)

	query := fmt.Sprintf(`
		UPDATE blockchain_transactions
		SET %s
		WHERE tx_hash = $%d
	`, setClause, whereParamNum)

	fmt.Printf("[UpdateTransaction] SQL: %s\n", query)
	fmt.Printf("[UpdateTransaction] Args count: %d, WHERE param: $%d\n", len(args), whereParamNum)

	_, err := r.db.Exec(query, args...)
	if err != nil {
		fmt.Printf("[UpdateTransaction] SQL ERROR: %v\n", err)
	}
	return err
}

// GetTransactionByHash gets a transaction by hash
func (r *Repository) GetTransactionByHash(txHash string) (*BlockchainTransaction, error) {
	query := `
		SELECT tx_hash, height, code, codespace, success, log, raw_log, info,
			gas_wanted, gas_used, fee_amount, messages, signers, memo,
			events, msg_responses, operation_type, sender, module_name,
			block_timestamp, created_at, enriched, enriched_at, data, tx_bytes
		FROM blockchain_transactions
		WHERE tx_hash = $1
	`

	var tx BlockchainTransaction
	var eventsJSON, messagesJSON, feeAmountJSON, msgResponsesJSON, signersJSON sql.NullString

	err := r.db.QueryRow(query, txHash).Scan(
		&tx.TxHash, &tx.Height, &tx.Code, &tx.Codespace, &tx.Success,
		&tx.Log, &tx.RawLog, &tx.Info,
		&tx.GasWanted, &tx.GasUsed,
		&feeAmountJSON, &messagesJSON,
		&signersJSON, &tx.Memo,
		&eventsJSON, &msgResponsesJSON,
		&tx.OperationType, &tx.Sender, &tx.ModuleName,
		&tx.BlockTimestamp, &tx.CreatedAt,
		&tx.Enriched, &tx.EnrichedAt,
		&tx.Data, &tx.TxBytes,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get transaction: %w", err)
	}

	// Unmarshal JSON fields
	if eventsJSON.Valid {
		json.Unmarshal([]byte(eventsJSON.String), &tx.Events)
	}
	if messagesJSON.Valid {
		json.Unmarshal([]byte(messagesJSON.String), &tx.Messages)
	}
	if feeAmountJSON.Valid {
		json.Unmarshal([]byte(feeAmountJSON.String), &tx.FeeAmount)
	}
	if msgResponsesJSON.Valid {
		json.Unmarshal([]byte(msgResponsesJSON.String), &tx.MsgResponses)
	}
	if signersJSON.Valid {
		json.Unmarshal([]byte(signersJSON.String), &tx.Signers)
	}

	return &tx, nil
}

// ListTransactions lists transactions with pagination
func (r *Repository) ListTransactions(limit, offset int) ([]*BlockchainTransaction, error) {
	query := `
		SELECT tx_hash, height, code, codespace, success, log, raw_log, info,
			gas_wanted, gas_used, fee_amount, messages, signers, memo,
			events, msg_responses, operation_type, sender, module_name,
			block_timestamp, created_at, enriched, enriched_at, data, tx_bytes
		FROM blockchain_transactions
		ORDER BY height DESC, created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list transactions: %w", err)
	}
	defer rows.Close()

	var transactions []*BlockchainTransaction
	for rows.Next() {
		var tx BlockchainTransaction
		var eventsJSON, messagesJSON, feeAmountJSON, msgResponsesJSON, signersJSON sql.NullString

		err := rows.Scan(
			&tx.TxHash, &tx.Height, &tx.Code, &tx.Codespace, &tx.Success,
			&tx.Log, &tx.RawLog, &tx.Info,
			&tx.GasWanted, &tx.GasUsed,
			&feeAmountJSON, &messagesJSON,
			&signersJSON, &tx.Memo,
			&eventsJSON, &msgResponsesJSON,
			&tx.OperationType, &tx.Sender, &tx.ModuleName,
			&tx.BlockTimestamp, &tx.CreatedAt,
			&tx.Enriched, &tx.EnrichedAt,
			&tx.Data, &tx.TxBytes,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan transaction: %w", err)
		}

		// Unmarshal JSON fields
		if eventsJSON.Valid {
			json.Unmarshal([]byte(eventsJSON.String), &tx.Events)
		}
		if messagesJSON.Valid {
			json.Unmarshal([]byte(messagesJSON.String), &tx.Messages)
		}
		if feeAmountJSON.Valid {
			json.Unmarshal([]byte(feeAmountJSON.String), &tx.FeeAmount)
		}
		if msgResponsesJSON.Valid {
			json.Unmarshal([]byte(msgResponsesJSON.String), &tx.MsgResponses)
		}
		if signersJSON.Valid {
			json.Unmarshal([]byte(signersJSON.String), &tx.Signers)
		}

		transactions = append(transactions, &tx)
	}

	return transactions, rows.Err()
}
