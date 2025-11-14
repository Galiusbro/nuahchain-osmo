package monitor

import (
	"time"
)

// BlockchainTransaction represents a full blockchain transaction in the database
type BlockchainTransaction struct {
	TxHash        string                 `json:"tx_hash" db:"tx_hash"`
	Height        int64                  `json:"height" db:"height"`
	Code          int                    `json:"code" db:"code"`
	Codespace     *string                `json:"codespace,omitempty" db:"codespace"`
	Success       bool                   `json:"success" db:"success"`
	Log           *string                `json:"log,omitempty" db:"log"`
	RawLog        *string                `json:"raw_log,omitempty" db:"raw_log"`
	Info          *string                `json:"info,omitempty" db:"info"`
	GasWanted     *int64                 `json:"gas_wanted,omitempty" db:"gas_wanted"`
	GasUsed       *int64                 `json:"gas_used,omitempty" db:"gas_used"`
	FeeAmount     map[string]interface{} `json:"fee_amount,omitempty" db:"fee_amount"`
	Messages      []map[string]interface{} `json:"messages,omitempty" db:"messages"`
	Signers       []string               `json:"signers,omitempty" db:"signers"`
	Memo          *string                `json:"memo,omitempty" db:"memo"`
	Events        map[string]interface{} `json:"events" db:"events"`
	MsgResponses  []map[string]interface{} `json:"msg_responses,omitempty" db:"msg_responses"`
	OperationType *string                `json:"operation_type,omitempty" db:"operation_type"`
	Sender        *string                `json:"sender,omitempty" db:"sender"`
	ModuleName    *string                `json:"module_name,omitempty" db:"module_name"`
	BlockTimestamp *time.Time            `json:"block_timestamp,omitempty" db:"block_timestamp"`
	CreatedAt     time.Time              `json:"created_at" db:"created_at"`
	Enriched      bool                   `json:"enriched" db:"enriched"`
	EnrichedAt    *time.Time             `json:"enriched_at,omitempty" db:"enriched_at"`
	Data          *string                `json:"data,omitempty" db:"data"`
	TxBytes       *string                `json:"tx_bytes,omitempty" db:"tx_bytes"`
}

// CreateBlockchainTransactionRequest represents data for creating a blockchain transaction record
type CreateBlockchainTransactionRequest struct {
	TxHash        string
	Height        int64
	Code          int
	Codespace     *string
	Success       bool
	Log           *string
	RawLog        *string
	Info          *string
	GasWanted     *int64
	GasUsed       *int64
	FeeAmount     map[string]interface{}
	Messages      []map[string]interface{}
	Signers       []string
	Memo          *string
	Events        map[string]interface{}
	MsgResponses  []map[string]interface{}
	OperationType *string
	Sender        *string
	ModuleName    *string
	BlockTimestamp *time.Time
	TxBytes       *string
	Data          *string
	EnrichedAt    *time.Time
}

// UpdateBlockchainTransactionRequest represents data for updating a blockchain transaction
type UpdateBlockchainTransactionRequest struct {
	Codespace     *string
	RawLog        *string
	Info          *string
	GasWanted     *int64
	GasUsed       *int64
	FeeAmount     map[string]interface{}
	Messages      []map[string]interface{}
	Signers       []string
	Memo          *string
	MsgResponses  []map[string]interface{}
	BlockTimestamp *time.Time
	Enriched      bool
	EnrichedAt    *time.Time
	Data          *string
}

