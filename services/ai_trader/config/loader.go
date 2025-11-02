package config

import (
	"fmt"
	"os"
	"path/filepath"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v3"
)

// LoadConfig loads configuration from file
func LoadConfig(configPath string) (*Config, error) {
	if configPath == "" {
		return nil, fmt.Errorf("config path is required")
	}

	// Check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", configPath)
	}

	// Determine file format based on extension
	ext := filepath.Ext(configPath)

	var config Config

	switch ext {
	case ".toml":
		if _, err := toml.DecodeFile(configPath, &config); err != nil {
			return nil, fmt.Errorf("failed to decode TOML config: %w", err)
		}
	case ".yaml", ".yml":
		data, err := os.ReadFile(configPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read YAML config: %w", err)
		}

		if err := yaml.Unmarshal(data, &config); err != nil {
			return nil, fmt.Errorf("failed to unmarshal YAML config: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported config file format: %s", ext)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &config, nil
}

// SaveConfig saves configuration to file
func SaveConfig(config *Config, configPath string) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	if configPath == "" {
		return fmt.Errorf("config path is required")
	}

	// Validate configuration before saving
	if err := config.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Determine file format based on extension
	ext := filepath.Ext(configPath)

	switch ext {
	case ".toml":
		file, err := os.Create(configPath)
		if err != nil {
			return fmt.Errorf("failed to create TOML config file: %w", err)
		}
		defer file.Close()

		// Create a map structure for TOML encoding
		tomlData := map[string]interface{}{
			"bot": map[string]interface{}{
				"name":             config.Bot.Name,
				"granter_address":  config.Bot.GranterAddress,
				"grantee_address":  config.Bot.GranteeAddress,
				"chain_id":         config.Bot.ChainID,
				"node_url":         config.Bot.NodeURL,
				"trading_interval": config.Bot.TradingInterval.String(),
				"enabled":          config.Bot.Enabled,
			},
			"api": map[string]interface{}{
				"rate_limit":    config.API.RateLimit,
				"rate_interval": config.API.RateInterval.String(),
				"bind":          config.API.Bind,
				"tls_cert":      config.API.TLSCertPath,
				"tls_key":       config.API.TLSKeyPath,
				"cors_origins":  config.API.CORSOrigins,
			},
			"limits": map[string]interface{}{
				"max_daily_volume":    config.Limits.MaxDailyVolume.String(),
				"max_single_trade":    config.Limits.MaxSingleTrade.String(),
				"min_single_trade":    config.Limits.MinSingleTrade.String(),
				"allowed_symbols":     config.Limits.AllowedSymbols,
				"max_trades_per_day":  config.Limits.MaxTradesPerDay,
				"max_trades_per_hour": config.Limits.MaxTradesPerHour,
			},
			"risk_management": map[string]interface{}{
				"stop_loss_percentage":   config.RiskManagement.StopLossPercentage,
				"take_profit_percentage": config.RiskManagement.TakeProfitPercentage,
				"max_price_deviation":    config.RiskManagement.MaxPriceDeviation,
				"cool_down_period":       config.RiskManagement.CoolDownPeriod.String(),
				"emergency_stop": map[string]interface{}{
					"max_daily_loss":         config.RiskManagement.EmergencyStop.MaxDailyLoss.String(),
					"max_consecutive_losses": config.RiskManagement.EmergencyStop.MaxConsecutiveLosses,
					"max_oracle_downtime":    config.RiskManagement.EmergencyStop.MaxOracleDowntime.String(),
				},
			},
			"data_sources": map[string]interface{}{
				"price_update_interval": config.DataSources.PriceUpdateInterval.String(),
				"max_price_age":         config.DataSources.MaxPriceAge.String(),
				"oracle": map[string]interface{}{
					"query_timeout": config.DataSources.Oracle.QueryTimeout.String(),
					"max_retries":   config.DataSources.Oracle.MaxRetries,
					"retry_delay":   config.DataSources.Oracle.RetryDelay.String(),
				},
			},
			"monitoring": map[string]interface{}{
				"log_level":        config.Monitoring.LogLevel,
				"log_format":       config.Monitoring.LogFormat,
				"enable_metrics":   config.Monitoring.EnableMetrics,
				"metrics_port":     config.Monitoring.MetricsPort,
				"enable_audit_log": config.Monitoring.EnableAuditLog,
				"audit_log_path":   config.Monitoring.AuditLogPath,
			},
		}

		encoder := toml.NewEncoder(file)
		if err := encoder.Encode(tomlData); err != nil {
			return fmt.Errorf("failed to encode TOML config: %w", err)
		}
	case ".yaml", ".yml":
		data, err := yaml.Marshal(config)
		if err != nil {
			return fmt.Errorf("failed to marshal YAML config: %w", err)
		}

		if err := os.WriteFile(configPath, data, 0644); err != nil {
			return fmt.Errorf("failed to write YAML config: %w", err)
		}
	default:
		return fmt.Errorf("unsupported config file format: %s", ext)
	}

	return nil
}

// DefaultConfig returns a default configuration
func DefaultConfig() *Config {
	return &Config{
		Bot: BotConfig{
			Name:            "ai-trader",
			GranterAddress:  "nuah1granter1234567890abcdefghijklmnopqrstuvwxyz",
			GranteeAddress:  "nuah1grantee1234567890abcdefghijklmnopqrstuvwxyz",
			ChainID:         "nuahchain-1",
			NodeURL:         "http://localhost:26657",
			TradingInterval: DurationString{duration: "30s"},
			Enabled:         true,
		},
	API: APIConfig{
		RateLimit:    60,
		RateInterval: DurationString{duration: "1m"},
		Bind:         "127.0.0.1:8080",
		CORSOrigins:  []string{"http://localhost"},
	},
		Limits: TradingLimits{
			MaxDailyVolume:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000000))},
			MaxSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(10000))},
			MinSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(100))},
			AllowedSymbols:   []string{"BTC", "ETH", "OSMO"},
			MaxTradesPerDay:  100,
			MaxTradesPerHour: 10,
		},
		RiskManagement: RiskManagement{
			StopLossPercentage:   0.1,  // 10%
			TakeProfitPercentage: 0.2,  // 20%
			MaxPriceDeviation:    0.05, // 5%
			CoolDownPeriod:       DurationString{duration: "5m"},
			EmergencyStop: EmergencyStop{
				MaxDailyLoss:         CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(100000))},
				MaxConsecutiveLosses: 5,
				MaxOracleDowntime:    DurationString{duration: "10m"},
			},
		},
		DataSources: DataSources{
			PriceUpdateInterval: DurationString{duration: "10s"},
			MaxPriceAge:         DurationString{duration: "30s"},
			Oracle: OracleConfig{
				QueryTimeout: DurationString{duration: "5s"},
				MaxRetries:   3,
				RetryDelay:   DurationString{duration: "1s"},
			},
		},
		Monitoring: Monitoring{
			LogLevel:       "info",
			LogFormat:      "json",
			EnableMetrics:  true,
			MetricsPort:    8080,
			EnableAuditLog: true,
			AuditLogPath:   "/var/log/ai-trader/audit.log",
		},
	}
}

// CreateDefaultConfigFile creates a default configuration file
func CreateDefaultConfigFile(configPath string) error {
	config := DefaultConfig()
	return SaveConfig(config, configPath)
}
