package types

const (
	EventTypePoolCreated   = "treasury_pool_created"
	EventTypePoolUpdated   = "treasury_pool_updated"
	EventTypeDeposit       = "treasury_deposit"
	EventTypeWithdrawal    = "treasury_withdrawal"
	EventTypeReserveUpdate = "treasury_reserve_update"
	EventTypePayout        = "treasury_payout"

	AttributeKeyPoolID   = "pool_id"
	AttributeKeyDenom    = "denom"
	AttributeKeyAmount   = "amount"
	AttributeKeySender   = "sender"
	AttributeKeyReceiver = "receiver"
)
