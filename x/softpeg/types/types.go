package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	paramtypes "github.com/cosmos/cosmos-sdk/x/params/types"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// Additional types not covered by protobuf definitions

// Alert represents a price deviation or system alert
type Alert struct {
	ID         string        `json:"id"`
	Timestamp  int64         `json:"timestamp"`
	Type       AlertType     `json:"type"`
	Severity   AlertSeverity `json:"severity"`
	Message    string        `json:"message"`
	Price      osmomath.Dec  `json:"price"`
	Deviation  osmomath.Dec  `json:"deviation"`
	Resolved   bool          `json:"resolved"`
	ResolvedAt int64         `json:"resolved_at"`
}

// PegStatus represents the current status of the soft peg
type PegStatus struct {
	IsHealthy       bool         `json:"is_healthy"`
	CurrentPrice    osmomath.Dec `json:"current_price"`
	TargetPrice     osmomath.Dec `json:"target_price"`
	Deviation       osmomath.Dec `json:"deviation"`
	LastUpdate      int64        `json:"last_update"`
	ActiveAlerts    int64        `json:"active_alerts"`
	TrustIndex      osmomath.Dec `json:"trust_index"`
	LiquidityHealth osmomath.Dec `json:"liquidity_health"`
}

// FeedbackSubmission represents community feedback about the peg
type FeedbackSubmission struct {
	Submitter    sdk.AccAddress `json:"submitter"`
	Timestamp    int64          `json:"timestamp"`
	FeedbackType FeedbackType   `json:"feedback_type"`
	Rating       int32          `json:"rating"` // 1-5 scale
	Comment      string         `json:"comment"`
	PriceView    osmomath.Dec   `json:"price_view"`
}

// FeedbackType defines the type of community feedback
type FeedbackType string

const (
	FeedbackTypePrice     FeedbackType = "price"
	FeedbackTypeLiquidity FeedbackType = "liquidity"
	FeedbackTypeStability FeedbackType = "stability"
	FeedbackTypeGeneral   FeedbackType = "general"
)

// Validation methods for protobuf types

// Validate validates PriceData
func (pd PriceData) Validate() error {
	if pd.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	if pd.Price.IsNil() || pd.Price.IsNegative() {
		return ErrInvalidPrice
	}
	if pd.Confidence.IsNil() || pd.Confidence.IsNegative() || pd.Confidence.GT(osmomath.OneDec()) {
		return ErrInvalidConfidence
	}
	if pd.Source == "" {
		return ErrInvalidSource
	}
	return nil
}

// Validate validates CommunityMetrics
func (cm CommunityMetrics) Validate() error {
	if cm.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	if cm.SentimentScore.IsNil() || cm.SentimentScore.IsNegative() || cm.SentimentScore.GT(osmomath.OneDec()) {
		return ErrInvalidSentimentScore
	}
	if cm.TrustScore.IsNil() || cm.TrustScore.IsNegative() || cm.TrustScore.GT(osmomath.OneDec()) {
		return ErrInvalidTrustScore
	}
	if cm.ParticipationRate.IsNil() || cm.ParticipationRate.IsNegative() || cm.ParticipationRate.GT(osmomath.OneDec()) {
		return ErrInvalidParticipationRate
	}
	if cm.GovernanceActivity < 0 {
		return ErrInvalidGovernanceActivity
	}
	if cm.SocialMentions < 0 {
		return ErrInvalidSocialMentions
	}
	if cm.CommunitySize < 0 {
		return ErrInvalidCommunitySize
	}
	return nil
}

// Validate validates PegConfig
func (pc PegConfig) Validate() error {
	if pc.TargetAsset == "" {
		return ErrInvalidTargetAsset
	}
	if pc.AdjustmentFactor.IsNil() || pc.AdjustmentFactor.IsNegative() || pc.AdjustmentFactor.GT(osmomath.OneDec()) {
		return ErrInvalidAdjustmentFactor
	}
	if pc.LastAdjustmentTime < 0 {
		return ErrInvalidLastAdjustmentTime
	}
	return nil
}

// ParamKeyTable returns the parameter key table for the softpeg module.
func ParamKeyTable() paramtypes.KeyTable {
	return paramtypes.NewKeyTable().RegisterParamSet(&Params{})
}
