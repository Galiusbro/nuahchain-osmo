package types

const (
	EventTypePremiumPlanCreated   = "premium_plan_created"
	EventTypePremiumPaymentMade   = "premium_payment_made"
	EventTypePremiumMarkedOverdue = "premium_marked_overdue"

	AttributeKeyPlanID    = "plan_id"
	AttributeKeyPaymentID = "payment_id"
	AttributeKeyPayer     = "payer"
)
