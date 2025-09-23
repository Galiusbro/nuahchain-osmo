package types

const (
	EventTypeClaimSubmitted = "claim_submitted"
	EventTypeClaimReviewed  = "claim_reviewed"
	EventTypeClaimEvidence  = "claim_evidence_added"
	EventTypeClaimPaid      = "claim_paid"

	AttributeKeyClaimID  = "claim_id"
	AttributeKeyPolicyID = "policy_id"
	AttributeKeyStatus   = "status"
	AttributeKeyReviewer = "reviewer"
)
