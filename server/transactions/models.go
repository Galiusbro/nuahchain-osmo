package transactions

import (
	"time"
)

// Операции с токенами
const (
	OperationTypeTokenCreate = "TOKEN_CREATE" // Создание токена
	OperationTypeTokenBuy    = "TOKEN_BUY"    // Покупка токена
	OperationTypeTokenSell   = "TOKEN_SELL"   // Продажа токена
)

// Операции с активами
const (
	OperationTypeAssetEnsure = "ASSET_ENSURE" // Создание/обеспечение актива
	OperationTypeAssetBuy    = "ASSET_BUY"    // Покупка актива
	OperationTypeAssetSell   = "ASSET_SELL"   // Продажа актива
	OperationTypeAssetMarginOpen  = "ASSET_MARGIN_OPEN"  // Открытие маржинальной позиции
	OperationTypeAssetMarginClose = "ASSET_MARGIN_CLOSE" // Закрытие маржинальной позиции
)

// Статусы транзакций
const (
	StatusPending = "PENDING" // Транзакция отправлена, ожидает подтверждения
	StatusSuccess = "SUCCESS" // Транзакция успешно выполнена
	StatusFailed  = "FAILED"  // Транзакция не удалась
)

// Transaction представляет запись о транзакции в БД
type Transaction struct {
	ID            int64                  `json:"id"`
	UserID        int64                  `json:"user_id"`
	OperationType string                 `json:"operation_type"`          // Тип операции (TOKEN_CREATE, TOKEN_BUY и т.д.)
	TxHash        string                 `json:"tx_hash"`                 // Хеш транзакции в блокчейне
	Status        string                 `json:"status"`                  // Статус (PENDING, SUCCESS, FAILED)
	OperationData map[string]interface{} `json:"operation_data"`          // Данные операции (JSON)
	ErrorMessage  *string                `json:"error_message,omitempty"` // Сообщение об ошибке
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// CreateTransactionRequest представляет данные для создания записи о транзакции
type CreateTransactionRequest struct {
	UserID        int64                  `json:"user_id"`
	OperationType string                 `json:"operation_type"`
	TxHash        string                 `json:"tx_hash"`
	Status        string                 `json:"status"`
	OperationData map[string]interface{} `json:"operation_data"`
	ErrorMessage  *string                `json:"error_message,omitempty"`
}

// UpdateTransactionRequest представляет данные для обновления статуса транзакции
type UpdateTransactionRequest struct {
	ID            int64                  `json:"id"`
	Status        string                 `json:"status"`
	OperationData map[string]interface{} `json:"operation_data,omitempty"` // Обновленные данные (например, полученные из блокчейна)
	ErrorMessage  *string                `json:"error_message,omitempty"`
}

// Вспомогательные функции для создания данных операций

// TokenCreateData создает данные для операции создания токена
func TokenCreateData(denom, name, symbol, image, description string) map[string]interface{} {
	return map[string]interface{}{
		"denom":       denom,
		"name":        name,
		"symbol":      symbol,
		"image":       image,
		"description": description,
	}
}

// TokenBuyData создает данные для операции покупки токена
func TokenBuyData(denom, paymentDenom, paymentAmount, tokensOut, pricePaid string) map[string]interface{} {
	return map[string]interface{}{
		"denom":          denom,
		"payment_denom":  paymentDenom,
		"payment_amount": paymentAmount,
		"tokens_out":     tokensOut,
		"price_paid":     pricePaid,
	}
}

// TokenSellData создает данные для операции продажи токена
func TokenSellData(denom, tokenAmount, paymentDenom, paymentOut, priceReceived string) map[string]interface{} {
	return map[string]interface{}{
		"denom":          denom,
		"token_amount":   tokenAmount,
		"payment_denom":  paymentDenom,
		"payment_out":    paymentOut,
		"price_received": priceReceived,
	}
}

// AssetEnsureData создает данные для операции создания/обеспечения актива
func AssetEnsureData(symbol string) map[string]interface{} {
	return map[string]interface{}{
		"symbol": symbol,
	}
}

// AssetBuyData создает данные для операции покупки актива
func AssetBuyData(symbol, denom, amount, baseAmount string) map[string]interface{} {
	return map[string]interface{}{
		"symbol":      symbol,
		"denom":       denom,
		"amount":      amount,
		"base_amount": baseAmount,
	}
}

// AssetSellData создает данные для операции продажи актива
func AssetSellData(symbol, baseAmount, payoutNDOLLAR string) map[string]interface{} {
	return map[string]interface{}{
		"symbol":         symbol,
		"base_amount":    baseAmount,
		"payout_ndollar": payoutNDOLLAR,
	}
}

// AssetMarginOpenData создает данные для операции открытия маржинальной позиции
func AssetMarginOpenData(symbol, side, quoteAmount, leverage string, positionID uint64, baseQuantity, entryPrice string) map[string]interface{} {
	data := map[string]interface{}{
		"symbol":        symbol,
		"side":          side,
		"quote_amount":  quoteAmount,
		"leverage":      leverage,
		"base_quantity": baseQuantity,
		"entry_price":   entryPrice,
	}
	if positionID != 0 {
		data["position_id"] = positionID
	}
	return data
}

// AssetMarginCloseData создает данные для операции закрытия маржинальной позиции
func AssetMarginCloseData(positionID uint64, pnl string) map[string]interface{} {
	data := map[string]interface{}{}
	if positionID != 0 {
		data["position_id"] = positionID
	}
	if pnl != "" {
		data["pnl"] = pnl
	}
	return data
}
