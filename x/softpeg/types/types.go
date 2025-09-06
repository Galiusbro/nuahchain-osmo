package types

import (
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/osmosis-labs/osmosis/osmomath"
)

// PriceData represents price information for NUAH token
type PriceData struct {
	Timestamp    int64           `json:"timestamp"`
	Price        osmomath.Dec    `json:"price"`
	TargetPrice  osmomath.Dec    `json:"target_price"`
	Deviation    osmomath.Dec    `json:"deviation"`
	Volume24h    osmomath.Int    `json:"volume_24h"`
	Liquidity    osmomath.Int    `json:"liquidity"`
	Source       string          `json:"source"`
	Confidence   osmomath.Dec    `json:"confidence"`
}

// CommunityMetrics represents community engagement and sentiment data
type CommunityMetrics struct {
	Timestamp         int64        `json:"timestamp"`
	TrustIndex        osmomath.Dec `json:"trust_index"`
	CommunitySize     int64        `json:"community_size"`
	ActiveUsers24h    int64        `json:"active_users_24h"`
	GovernanceVotes   int64        `json:"governance_votes"`
	SentimentScore    osmomath.Dec `json:"sentiment_score"`
	EngagementRate    osmomath.Dec `json:"engagement_rate"`
	FeedbackCount     int64        `json:"feedback_count"`
}

// PegConfig represents the configuration for the soft peg mechanism
type PegConfig struct {
	TargetPrice       osmomath.Dec `json:"target_price"`
	DeviationThreshold osmomath.Dec `json:"deviation_threshold"`
	AlertThreshold    osmomath.Dec `json:"alert_threshold"`
	MonitoringEnabled bool         `json:"monitoring_enabled"`
	UpdateInterval    time.Duration `json:"update_interval"`
	MinLiquidity      osmomath.Int `json:"min_liquidity"`
	MaxDeviation      osmomath.Dec `json:"max_deviation"`
}

// Alert represents a price deviation or system alert
type Alert struct {
	ID          string       `json:"id"`
	Timestamp   int64        `json:"timestamp"`
	Type        AlertType    `json:"type"`
	Severity    AlertSeverity `json:"severity"`
	Message     string       `json:"message"`
	Price       osmomath.Dec `json:"price"`
	Deviation   osmomath.Dec `json:"deviation"`
	Resolved    bool         `json:"resolved"`
	ResolvedAt  int64        `json:"resolved_at"`
}

// AlertType defines the type of alert
type AlertType string

const (
	AlertTypePriceDeviation AlertType = "price_deviation"
	AlertTypeLowLiquidity   AlertType = "low_liquidity"
	AlertTypeLowConfidence  AlertType = "low_confidence"
	AlertTypeSystemError    AlertType = "system_error"
	AlertTypeCommunity      AlertType = "community"
)

// AlertSeverity defines the severity level of an alert
type AlertSeverity string

const (
	AlertSeverityLow      AlertSeverity = "low"
	AlertSeverityMedium   AlertSeverity = "medium"
	AlertSeverityHigh     AlertSeverity = "high"
	AlertSeverityCritical AlertSeverity = "critical"
)

// PegStatus represents the current status of the soft peg
type PegStatus struct {
	IsHealthy         bool         `json:"is_healthy"`
	CurrentPrice      osmomath.Dec `json:"current_price"`
	TargetPrice       osmomath.Dec `json:"target_price"`
	Deviation         osmomath.Dec `json:"deviation"`
	LastUpdate        int64        `json:"last_update"`
	ActiveAlerts      int64        `json:"active_alerts"`
	TrustIndex        osmomath.Dec `json:"trust_index"`
	LiquidityHealth   osmomath.Dec `json:"liquidity_health"`
}

// FeedbackSubmission represents community feedback about the peg
type FeedbackSubmission struct {
	Submitter   sdk.AccAddress `json:"submitter"`
	Timestamp   int64          `json:"timestamp"`
	FeedbackType FeedbackType  `json:"feedback_type"`
	Rating      int32          `json:"rating"` // 1-5 scale
	Comment     string         `json:"comment"`
	PriceView   osmomath.Dec   `json:"price_view"`
}

// FeedbackType defines the type of community feedback
type FeedbackType string

const (
	FeedbackTypePrice      FeedbackType = "price"
	FeedbackTypeLiquidity  FeedbackType = "liquidity"
	FeedbackTypeStability  FeedbackType = "stability"
	FeedbackTypeGeneral    FeedbackType = "general"
)

// Validate validates the PriceData
func (pd PriceData) Validate() error {
	if pd.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	if pd.Price.IsNegative() {
		return ErrInvalidPrice
	}
	if pd.TargetPrice.IsNegative() {
		return ErrInvalidTargetPrice
	}
	return nil
}

// Validate validates the CommunityMetrics
func (cm CommunityMetrics) Validate() error {
	if cm.Timestamp <= 0 {
		return ErrInvalidTimestamp
	}
	if cm.TrustIndex.IsNegative() || cm.TrustIndex.GT(osmomath.OneDec()) {
		return ErrInvalidTrustIndex
	}
	return nil
}

// Validate validates the PegConfig
func (pc PegConfig) Validate() error {
	if pc.TargetPrice.IsNegative() || pc.TargetPrice.IsZero() {
		return ErrInvalidTargetPrice
	}
	if pc.DeviationThreshold.IsNegative() || pc.DeviationThreshold.GT(osmomath.OneDec()) {
		return ErrInvalidDeviationThreshold
	}
	if pc.UpdateInterval <= 0 {
		return ErrInvalidUpdateInterval
	}
	return nil
}