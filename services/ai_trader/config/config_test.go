package config

import (
	"os"
	"path/filepath"
	"testing"

	"cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name:        "valid config",
			config:      DefaultConfig(),
			expectError: false,
		},
		{
			name: "invalid bot config - empty name",
			config: &Config{
				Bot: BotConfig{
					Name: "",
				},
			},
			expectError: true,
			errorMsg:    "bot name is required",
		},
		{
			name: "invalid bot config - empty granter address",
			config: &Config{
				Bot: BotConfig{
					Name:            "test-bot",
					GranterAddress:  "",
					GranteeAddress:  "nuah1test123",
					ChainID:         "test-chain",
					NodeURL:         "http://localhost:26657",
					TradingInterval: DurationString{duration: "30s"},
				},
			},
			expectError: true,
			errorMsg:    "granter address is required",
		},
		{
			name: "invalid trading limits - zero max daily volume",
			config: &Config{
				Bot: BotConfig{
					Name:            "test-bot",
					GranterAddress:  "nuah1granter",
					GranteeAddress:  "nuah1grantee",
					ChainID:         "test-chain",
					NodeURL:         "http://localhost:26657",
					TradingInterval: DurationString{duration: "30s"},
				},
				Limits: TradingLimits{
					MaxDailyVolume: CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(0))},
				},
			},
			expectError: true,
			errorMsg:    "max daily volume is required",
		},
		{
			name: "invalid trading limits - max single trade less than min",
			config: &Config{
				Bot: BotConfig{
					Name:            "test-bot",
					GranterAddress:  "nuah1granter",
					GranteeAddress:  "nuah1grantee",
					ChainID:         "test-chain",
					NodeURL:         "http://localhost:26657",
					TradingInterval: DurationString{duration: "30s"},
				},
				Limits: TradingLimits{
					MaxDailyVolume:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000000))},
					MaxSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(100))},
					MinSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(200))},
					AllowedSymbols:   []string{"BTC"},
					MaxTradesPerDay:  100,
					MaxTradesPerHour: 10,
				},
			},
			expectError: true,
			errorMsg:    "max single trade must be greater than min single trade",
		},
		{
			name: "invalid risk management - stop loss percentage out of range",
			config: &Config{
				Bot: BotConfig{
					Name:            "test-bot",
					GranterAddress:  "nuah1granter",
					GranteeAddress:  "nuah1grantee",
					ChainID:         "test-chain",
					NodeURL:         "http://localhost:26657",
					TradingInterval: DurationString{duration: "30s"},
				},
				Limits: TradingLimits{
					MaxDailyVolume:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000000))},
					MaxSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(10000))},
					MinSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(100))},
					AllowedSymbols:   []string{"BTC"},
					MaxTradesPerDay:  100,
					MaxTradesPerHour: 10,
				},
				RiskManagement: RiskManagement{
					StopLossPercentage: 1.5, // Invalid: > 1
				},
			},
			expectError: true,
			errorMsg:    "stop loss percentage must be between 0 and 1",
		},
		{
			name: "invalid monitoring - invalid log level",
			config: &Config{
				Bot: BotConfig{
					Name:            "test-bot",
					GranterAddress:  "nuah1granter",
					GranteeAddress:  "nuah1grantee",
					ChainID:         "test-chain",
					NodeURL:         "http://localhost:26657",
					TradingInterval: DurationString{duration: "30s"},
				},
				Limits: TradingLimits{
					MaxDailyVolume:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(1000000))},
					MaxSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(10000))},
					MinSingleTrade:   CoinString{Coin: sdk.NewCoin("factory/test/ndollar", math.NewInt(100))},
					AllowedSymbols:   []string{"BTC"},
					MaxTradesPerDay:  100,
					MaxTradesPerHour: 10,
				},
				RiskManagement: RiskManagement{
					StopLossPercentage:   0.1,
					TakeProfitPercentage: 0.2,
					MaxPriceDeviation:    0.05,
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
					LogLevel: "invalid-level",
				},
			},
			expectError: true,
			errorMsg:    "invalid log level: invalid-level",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorMsg)
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func TestLoadConfig(t *testing.T) {
	// Create temporary directory for test files
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "load TOML config",
			format:      "toml",
			expectError: false,
		},
		{
			name:        "load YAML config",
			format:      "yaml",
			expectError: false,
		},
		{
			name:        "load YML config",
			format:      "yml",
			expectError: false,
		},
		{
			name:        "unsupported format",
			format:      "json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "config."+tt.format)

			if tt.format == "toml" || tt.format == "yaml" || tt.format == "yml" {
				// Create a valid config file
				err := CreateDefaultConfigFile(configPath)
				require.NoError(t, err)

				// Load the config
				config, err := LoadConfig(configPath)
				if tt.expectError {
					require.Error(t, err)
				} else {
					require.NoError(t, err)
					require.NotNil(t, config)
					assert.Equal(t, "ai-trader", config.Bot.Name)
				}
			} else {
				// Test unsupported format
				_, err := LoadConfig(configPath)
				require.Error(t, err)
			}
		})
	}
}

func TestSaveConfig(t *testing.T) {
	tempDir := t.TempDir()

	tests := []struct {
		name        string
		format      string
		expectError bool
	}{
		{
			name:        "save TOML config",
			format:      "toml",
			expectError: false,
		},
		{
			name:        "save YAML config",
			format:      "yaml",
			expectError: false,
		},
		{
			name:        "save YML config",
			format:      "yml",
			expectError: false,
		},
		{
			name:        "unsupported format",
			format:      "json",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			configPath := filepath.Join(tempDir, "config."+tt.format)
			config := DefaultConfig()

			err := SaveConfig(config, configPath)
			if tt.expectError {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify file was created
				_, err := os.Stat(configPath)
				require.NoError(t, err)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	// Validate default config
	err := config.Validate()
	require.NoError(t, err)

	// Check some default values
	assert.Equal(t, "ai-trader", config.Bot.Name)
	assert.Equal(t, DurationString{duration: "30s"}, config.Bot.TradingInterval)
	assert.True(t, config.Bot.Enabled)

	assert.Equal(t, "factory/test/ndollar", config.Limits.MaxDailyVolume.Denom)
	assert.Equal(t, int64(1000000), config.Limits.MaxDailyVolume.Amount.Int64())

	assert.Equal(t, 0.1, config.RiskManagement.StopLossPercentage)
	assert.Equal(t, 0.2, config.RiskManagement.TakeProfitPercentage)

	assert.Equal(t, "info", config.Monitoring.LogLevel)
	assert.Equal(t, "json", config.Monitoring.LogFormat)
	assert.True(t, config.Monitoring.EnableMetrics)
}

func TestCreateDefaultConfigFile(t *testing.T) {
	tempDir := t.TempDir()
	configPath := filepath.Join(tempDir, "default.toml")

	err := CreateDefaultConfigFile(configPath)
	require.NoError(t, err)

	// Verify file was created
	_, err = os.Stat(configPath)
	require.NoError(t, err)

	// Load and validate the created config
	config, err := LoadConfig(configPath)
	require.NoError(t, err)
	require.NotNil(t, config)

	// Validate the loaded config
	err = config.Validate()
	require.NoError(t, err)
}
