package config

import (
	"fmt"
)

// Config represents the AI trader configuration
type Config struct {
	// Bot settings
	Bot BotConfig `toml:"bot" yaml:"bot"`

	// Trading limits and policies
	Limits TradingLimits `toml:"limits" yaml:"limits"`

	// Risk management
	RiskManagement RiskManagement `toml:"risk_management" yaml:"risk_management"`

	// Data sources
	DataSources DataSources `toml:"data_sources" yaml:"data_sources"`

	// Monitoring and logging
	Monitoring Monitoring `toml:"monitoring" yaml:"monitoring"`
}

// BotConfig contains basic bot settings
type BotConfig struct {
	// Bot name for identification
	Name string `toml:"name" yaml:"name"`

	// Granter address (user who granted permissions)
	GranterAddress string `toml:"granter_address" yaml:"granter_address"`

	// Grantee address (bot's address)
	GranteeAddress string `toml:"grantee_address" yaml:"grantee_address"`

	// Chain configuration
	ChainID string `toml:"chain_id" yaml:"chain_id"`
	NodeURL string `toml:"node_url" yaml:"node_url"`

	// Trading interval
	TradingInterval DurationString `toml:"trading_interval" yaml:"trading_interval"`

	// Enable/disable trading
	Enabled bool `toml:"enabled" yaml:"enabled"`
}

// TradingLimits defines trading limits and constraints
type TradingLimits struct {
	// Maximum daily trading volume in NDOLLAR
	MaxDailyVolume CoinString `toml:"max_daily_volume" yaml:"max_daily_volume"`

	// Maximum single trade amount
	MaxSingleTrade CoinString `toml:"max_single_trade" yaml:"max_single_trade"`

	// Minimum trade amount
	MinSingleTrade CoinString `toml:"min_single_trade" yaml:"min_single_trade"`

	// Allowed trading symbols (whitelist)
	AllowedSymbols []string `toml:"allowed_symbols" yaml:"allowed_symbols"`

	// Maximum number of trades per day
	MaxTradesPerDay int `toml:"max_trades_per_day" yaml:"max_trades_per_day"`

	// Maximum number of trades per hour
	MaxTradesPerHour int `toml:"max_trades_per_hour" yaml:"max_trades_per_hour"`
}

// RiskManagement contains risk management settings
type RiskManagement struct {
	// Stop loss percentage (0.1 = 10%)
	StopLossPercentage float64 `toml:"stop_loss_percentage" yaml:"stop_loss_percentage"`

	// Take profit percentage (0.2 = 20%)
	TakeProfitPercentage float64 `toml:"take_profit_percentage" yaml:"take_profit_percentage"`

	// Maximum price deviation from oracle (0.05 = 5%)
	MaxPriceDeviation float64 `toml:"max_price_deviation" yaml:"max_price_deviation"`

	// Cool-down period between trades
	CoolDownPeriod DurationString `toml:"cool_down_period" yaml:"cool_down_period"`

	// Emergency stop conditions
	EmergencyStop EmergencyStop `toml:"emergency_stop" yaml:"emergency_stop"`
}

// EmergencyStop defines emergency stop conditions
type EmergencyStop struct {
	// Stop trading if daily loss exceeds this amount
	MaxDailyLoss CoinString `toml:"max_daily_loss" yaml:"max_daily_loss"`

	// Stop trading if consecutive losses exceed this number
	MaxConsecutiveLosses int `toml:"max_consecutive_losses" yaml:"max_consecutive_losses"`

	// Stop trading if oracle price is unavailable for this duration
	MaxOracleDowntime DurationString `toml:"max_oracle_downtime" yaml:"max_oracle_downtime"`
}

// DataSources defines data source configurations
type DataSources struct {
	// Oracle configuration
	Oracle OracleConfig `toml:"oracle" yaml:"oracle"`

	// Price update interval
	PriceUpdateInterval DurationString `toml:"price_update_interval" yaml:"price_update_interval"`

	// Maximum price age before considering stale
	MaxPriceAge DurationString `toml:"max_price_age" yaml:"max_price_age"`
}

// OracleConfig contains oracle-specific settings
type OracleConfig struct {
	// Oracle query timeout
	QueryTimeout DurationString `toml:"query_timeout" yaml:"query_timeout"`

	// Retry configuration
	MaxRetries int `toml:"max_retries" yaml:"max_retries"`

	// Retry delay
	RetryDelay DurationString `toml:"retry_delay" yaml:"retry_delay"`
}

// Monitoring contains monitoring and logging settings
type Monitoring struct {
	// Log level (debug, info, warn, error)
	LogLevel string `toml:"log_level" yaml:"log_level"`

	// Log format (json, text)
	LogFormat string `toml:"log_format" yaml:"log_format"`

	// Enable metrics collection
	EnableMetrics bool `toml:"enable_metrics" yaml:"enable_metrics"`

	// Metrics port
	MetricsPort int `toml:"metrics_port" yaml:"metrics_port"`

	// Enable audit logging
	EnableAuditLog bool `toml:"enable_audit_log" yaml:"enable_audit_log"`

	// Audit log file path
	AuditLogPath string `toml:"audit_log_path" yaml:"audit_log_path"`
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if err := c.Bot.Validate(); err != nil {
		return fmt.Errorf("bot config validation failed: %w", err)
	}

	if err := c.Limits.Validate(); err != nil {
		return fmt.Errorf("trading limits validation failed: %w", err)
	}

	if err := c.RiskManagement.Validate(); err != nil {
		return fmt.Errorf("risk management validation failed: %w", err)
	}

	if err := c.DataSources.Validate(); err != nil {
		return fmt.Errorf("data sources validation failed: %w", err)
	}

	if err := c.Monitoring.Validate(); err != nil {
		return fmt.Errorf("monitoring validation failed: %w", err)
	}

	return nil
}

// Validate validates bot configuration
func (b *BotConfig) Validate() error {
	if b.Name == "" {
		return fmt.Errorf("bot name is required")
	}

	if b.GranterAddress == "" {
		return fmt.Errorf("granter address is required")
	}

	if b.GranteeAddress == "" {
		return fmt.Errorf("grantee address is required")
	}

	if b.ChainID == "" {
		return fmt.Errorf("chain ID is required")
	}

	if b.NodeURL == "" {
		return fmt.Errorf("node URL is required")
	}

	if b.TradingInterval.duration == "" {
		return fmt.Errorf("trading interval is required")
	}

	_, err := b.TradingInterval.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid trading interval: %w", err)
	}

	return nil
}

// Validate validates trading limits
func (l *TradingLimits) Validate() error {
	if l.MaxDailyVolume.IsZero() {
		return fmt.Errorf("max daily volume is required")
	}

	if l.MaxSingleTrade.IsZero() {
		return fmt.Errorf("max single trade is required")
	}

	if l.MinSingleTrade.IsZero() {
		return fmt.Errorf("min single trade is required")
	}

	if l.MaxSingleTrade.Coin.IsLT(l.MinSingleTrade.Coin) {
		return fmt.Errorf("max single trade must be greater than min single trade")
	}

	if len(l.AllowedSymbols) == 0 {
		return fmt.Errorf("at least one allowed symbol is required")
	}

	if l.MaxTradesPerDay <= 0 {
		return fmt.Errorf("max trades per day must be positive")
	}

	if l.MaxTradesPerHour <= 0 {
		return fmt.Errorf("max trades per hour must be positive")
	}

	return nil
}

// Validate validates risk management settings
func (r *RiskManagement) Validate() error {
	if r.StopLossPercentage < 0 || r.StopLossPercentage > 1 {
		return fmt.Errorf("stop loss percentage must be between 0 and 1")
	}

	if r.TakeProfitPercentage < 0 || r.TakeProfitPercentage > 1 {
		return fmt.Errorf("take profit percentage must be between 0 and 1")
	}

	if r.MaxPriceDeviation < 0 || r.MaxPriceDeviation > 1 {
		return fmt.Errorf("max price deviation must be between 0 and 1")
	}

	if r.CoolDownPeriod.duration == "" {
		return fmt.Errorf("cool down period is required")
	}

	_, err := r.CoolDownPeriod.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid cool down period: %w", err)
	}

	return r.EmergencyStop.Validate()
}

// Validate validates emergency stop settings
func (e *EmergencyStop) Validate() error {
	if e.MaxDailyLoss.IsZero() {
		return fmt.Errorf("max daily loss is required")
	}

	if e.MaxConsecutiveLosses < 0 {
		return fmt.Errorf("max consecutive losses must be non-negative")
	}

	if e.MaxOracleDowntime.duration == "" {
		return fmt.Errorf("max oracle downtime is required")
	}

	_, err := e.MaxOracleDowntime.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid max oracle downtime: %w", err)
	}

	return nil
}

// Validate validates data sources configuration
func (d *DataSources) Validate() error {
	if d.PriceUpdateInterval.duration == "" {
		return fmt.Errorf("price update interval is required")
	}

	_, err := d.PriceUpdateInterval.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid price update interval: %w", err)
	}

	if d.MaxPriceAge.duration == "" {
		return fmt.Errorf("max price age is required")
	}

	_, err = d.MaxPriceAge.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid max price age: %w", err)
	}

	return d.Oracle.Validate()
}

// Validate validates oracle configuration
func (o *OracleConfig) Validate() error {
	if o.QueryTimeout.duration == "" {
		return fmt.Errorf("query timeout is required")
	}

	_, err := o.QueryTimeout.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid query timeout: %w", err)
	}

	if o.MaxRetries < 0 {
		return fmt.Errorf("max retries must be non-negative")
	}

	if o.RetryDelay.duration == "" {
		return fmt.Errorf("retry delay is required")
	}

	_, err = o.RetryDelay.ParseDuration()
	if err != nil {
		return fmt.Errorf("invalid retry delay: %w", err)
	}

	return nil
}

// Validate validates monitoring configuration
func (m *Monitoring) Validate() error {
	validLogLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
	}

	if !validLogLevels[m.LogLevel] {
		return fmt.Errorf("invalid log level: %s", m.LogLevel)
	}

	validLogFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if !validLogFormats[m.LogFormat] {
		return fmt.Errorf("invalid log format: %s", m.LogFormat)
	}

	if m.EnableMetrics && m.MetricsPort <= 0 {
		return fmt.Errorf("metrics port must be positive when metrics are enabled")
	}

	if m.EnableAuditLog && m.AuditLogPath == "" {
		return fmt.Errorf("audit log path is required when audit logging is enabled")
	}

	return nil
}
