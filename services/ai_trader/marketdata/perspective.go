package marketdata

import (
	"encoding/json"
	"fmt"
	"strings"
)

// PerspectiveConfig captures decider perspective settings per bot.
type PerspectiveConfig struct {
	PreTF       Timeframe `json:"pre_tf"`
	PreLimit    int       `json:"pre_limit"`
	TargetTF    Timeframe `json:"target_tf"`
	TargetLimit int       `json:"target_limit"`
	PostTF      Timeframe `json:"post_tf"`
	PostLimit   int       `json:"post_limit"`
}

// DefaultPerspectiveConfig returns sensible defaults matching AIDecider defaults.
func DefaultPerspectiveConfig() PerspectiveConfig {
	return PerspectiveConfig{
		PreTF:       TF1h,
		PreLimit:    48,
		TargetTF:    TF5m,
		TargetLimit: 96,
		PostTF:      TF1m,
		PostLimit:   60,
	}
}

// Validate ensures all fields are set and timeframes supported.
func (p PerspectiveConfig) Validate() error {
	if err := validateTimeframe(p.PreTF); err != nil {
		return fmt.Errorf("pre_tf: %w", err)
	}
	if err := validateTimeframe(p.TargetTF); err != nil {
		return fmt.Errorf("target_tf: %w", err)
	}
	if err := validateTimeframe(p.PostTF); err != nil {
		return fmt.Errorf("post_tf: %w", err)
	}
	if p.PreLimit <= 0 {
		return fmt.Errorf("pre_limit must be > 0")
	}
	if p.TargetLimit <= 0 {
		return fmt.Errorf("target_limit must be > 0")
	}
	if p.PostLimit <= 0 {
		return fmt.Errorf("post_limit must be > 0")
	}
	return nil
}

func validateTimeframe(tf Timeframe) error {
	switch tf {
	case TF1m, TF5m, TF1h, TF1d:
		return nil
	default:
		return fmt.Errorf("invalid timeframe %q", tf)
	}
}

// Marshal returns JSON string.
func (p PerspectiveConfig) Marshal() (string, error) {
	b, err := json.Marshal(p)
	return string(b), err
}

// ParsePerspective parses perspective config from JSON string.
func ParsePerspective(data string) (PerspectiveConfig, error) {
	if strings.TrimSpace(data) == "" {
		return DefaultPerspectiveConfig(), nil
	}
	var cfg PerspectiveConfig
	if err := json.Unmarshal([]byte(data), &cfg); err != nil {
		return PerspectiveConfig{}, err
	}
	if err := cfg.Validate(); err != nil {
		return PerspectiveConfig{}, err
	}
	return cfg, nil
}
