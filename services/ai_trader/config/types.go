package config

import (
	"fmt"
	"strings"
	"time"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// CoinString is a custom type for parsing Coin from string
type CoinString struct {
	sdk.Coin
}

// UnmarshalTOML implements toml.Unmarshaler interface
func (c *CoinString) UnmarshalTOML(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", data)
	}

	coin, err := sdk.ParseCoinNormalized(str)
	if err != nil {
		return fmt.Errorf("failed to parse coin: %w", err)
	}

	c.Coin = coin
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (c *CoinString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	coin, err := sdk.ParseCoinNormalized(str)
	if err != nil {
		return fmt.Errorf("failed to parse coin: %w", err)
	}

	c.Coin = coin
	return nil
}

// MarshalTOML implements toml.Marshaler interface
func (c CoinString) MarshalTOML() (interface{}, error) {
	return c.Coin.String(), nil
}

// MarshalYAML implements yaml.Marshaler interface
func (c CoinString) MarshalYAML() (interface{}, error) {
	return c.Coin.String(), nil
}

// CoinsString is a custom type for parsing Coins from string
type CoinsString struct {
	sdk.Coins
}

// UnmarshalTOML implements toml.Unmarshaler interface
func (c *CoinsString) UnmarshalTOML(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", data)
	}

	coins, err := sdk.ParseCoinsNormalized(str)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	c.Coins = coins
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (c *CoinsString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	coins, err := sdk.ParseCoinsNormalized(str)
	if err != nil {
		return fmt.Errorf("failed to parse coins: %w", err)
	}

	c.Coins = coins
	return nil
}

// MarshalTOML implements toml.Marshaler interface
func (c CoinsString) MarshalTOML() (interface{}, error) {
	return c.Coins.String(), nil
}

// MarshalYAML implements yaml.Marshaler interface
func (c CoinsString) MarshalYAML() (interface{}, error) {
	return c.Coins.String(), nil
}

// DurationString is a custom type for parsing duration from string
type DurationString struct {
	duration string
}

// NewDurationString creates a new DurationString from a time.Duration
func NewDurationString(d time.Duration) DurationString {
	return DurationString{duration: d.String()}
}

// UnmarshalTOML implements toml.Unmarshaler interface
func (d *DurationString) UnmarshalTOML(data interface{}) error {
	str, ok := data.(string)
	if !ok {
		return fmt.Errorf("expected string, got %T", data)
	}

	d.duration = str
	return nil
}

// UnmarshalYAML implements yaml.Unmarshaler interface
func (d *DurationString) UnmarshalYAML(unmarshal func(interface{}) error) error {
	var str string
	if err := unmarshal(&str); err != nil {
		return err
	}

	d.duration = str
	return nil
}

// MarshalTOML implements toml.Marshaler interface
func (d DurationString) MarshalTOML() (interface{}, error) {
	return d.duration, nil
}

// MarshalYAML implements yaml.Marshaler interface
func (d DurationString) MarshalYAML() (interface{}, error) {
	return d.duration, nil
}

// String returns the duration string
func (d DurationString) String() string {
	return d.duration
}

// ParseDuration parses the duration string into time.Duration
func (d DurationString) ParseDuration() (time.Duration, error) {
	// Handle common duration formats
	str := strings.TrimSpace(d.duration)
	if str == "" {
		return 0, fmt.Errorf("empty duration string")
	}

	// Try to parse as Go duration format
	duration, err := time.ParseDuration(str)
	if err != nil {
		return 0, fmt.Errorf("failed to parse duration '%s': %w", str, err)
	}

	return duration, nil
}

// LLMConfig defines LLM provider settings (kept separate to avoid breaking existing Config).
type LLMConfig struct {
	Provider    string        `json:"provider" yaml:"provider"`
	APIKey      string        `json:"api_key" yaml:"api_key"`
	Model       string        `json:"model" yaml:"model"`
	Temperature float64       `json:"temperature" yaml:"temperature"`
	MaxTokens   int           `json:"max_tokens" yaml:"max_tokens"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
}

// MarketDataConfig defines market data source settings.
type MarketDataConfig struct {
	Source          string        `json:"source" yaml:"source"`
	Timeframes      []string      `json:"timeframes" yaml:"timeframes"`
	RefreshInterval time.Duration `json:"refresh_interval" yaml:"refresh_interval"`
}
