package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// Config represents the server configuration
type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Logger     LoggerConfig     `yaml:"logger"`
	Auth       AuthConfig       `yaml:"auth"`
	Blockchain BlockchainConfig `yaml:"blockchain"`
}

// ServerConfig contains HTTP server settings
type ServerConfig struct {
	// Server address (host:port)
	Address string `yaml:"address"`

	// Read timeout
	ReadTimeout time.Duration `yaml:"read_timeout"`

	// Write timeout
	WriteTimeout time.Duration `yaml:"write_timeout"`

	// Idle timeout
	IdleTimeout time.Duration `yaml:"idle_timeout"`
}

// DatabaseConfig contains PostgreSQL database settings
type DatabaseConfig struct {
	// Database host
	Host string `yaml:"host"`

	// Database port
	Port string `yaml:"port"`

	// Database user
	User string `yaml:"user"`

	// Database password
	Password string `yaml:"password"`

	// Database name
	Database string `yaml:"database"`

	// SSL mode (disable, require, verify-ca, verify-full)
	SSLMode string `yaml:"sslmode"`

	// Connection timeout
	ConnectTimeout time.Duration `yaml:"connect_timeout"`

	// Maximum number of open connections
	MaxOpenConns int `yaml:"max_open_conns"`

	// Maximum number of idle connections
	MaxIdleConns int `yaml:"max_idle_conns"`

	// Connection max lifetime
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`

	// Connection max idle time
	ConnMaxIdleTime time.Duration `yaml:"conn_max_idle_time"`
}

// LoggerConfig contains logging settings
type LoggerConfig struct {
	// Enabled enables/disables logging (useful for production)
	Enabled bool `yaml:"enabled"`

	// Level is the logging level (debug, info, warn, error)
	Level string `yaml:"level"`

	// Format is the output format (json, text)
	Format string `yaml:"format"`

	// Environment (dev, prod)
	Environment string `yaml:"environment"`
}

// AuthConfig contains authentication settings
type AuthConfig struct {
	// JWT secret key
	JWTSecret string `yaml:"jwt_secret"`

	// Token expiry duration
	TokenExpiry time.Duration `yaml:"token_expiry"`

	// Refresh token expiry duration
	RefreshExpiry time.Duration `yaml:"refresh_expiry"`
}

// BlockchainConfig contains blockchain connection settings
type BlockchainConfig struct {
	// Node URL for gRPC connection
	NodeURL string `yaml:"node_url"`

	// Chain ID
	ChainID string `yaml:"chain_id"`

	// RPC URL for REST/WebSocket (default: localhost:26657)
	RPCURL string `yaml:"rpc_url"`

	// WebSocket URL (default: ws://localhost:26657/websocket)
	WebSocketURL string `yaml:"websocket_url"`

	// Enable WebSocket for transaction tracking
	WebSocketEnabled bool `yaml:"websocket_enabled"`

	// Reconnect interval for WebSocket
	ReconnectInterval time.Duration `yaml:"reconnect_interval"`

	// Timeout for WebSocket operations
	WebSocketTimeout time.Duration `yaml:"websocket_timeout"`
}

// Load loads and returns the server configuration
func Load() (*Config, error) {
	cfg := &Config{
		Server: ServerConfig{
			Address:      getEnv("SERVER_ADDRESS", "0.0.0.0:8080"),
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
		},
		Database: DatabaseConfig{
			Host:            getEnv("DB_HOST", "localhost"),
			Port:            getEnv("DB_PORT", "5432"),
			User:            getEnv("DB_USER", "postgres"),
			Password:        getEnv("DB_PASSWORD", "postgres"),
			Database:        getEnv("DB_NAME", "serverdb"),
			SSLMode:         getEnv("DB_SSLMODE", "disable"),
			ConnectTimeout:  10 * time.Second,
			MaxOpenConns:    25,
			MaxIdleConns:    5,
			ConnMaxLifetime: 5 * time.Minute,
			ConnMaxIdleTime: 10 * time.Minute,
		},
		Logger: LoggerConfig{
			Enabled:     getEnvBool("LOG_ENABLED", true),
			Level:       getEnv("LOG_LEVEL", "debug"),
			Format:      getEnv("LOG_FORMAT", "text"),
			Environment: getEnv("ENVIRONMENT", "dev"),
		},
		Auth: AuthConfig{
			JWTSecret:     getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
			TokenExpiry:   24 * time.Hour,
			RefreshExpiry: 7 * 24 * time.Hour,
		},
		Blockchain: BlockchainConfig{
			NodeURL:           getEnv("BLOCKCHAIN_NODE_URL", "localhost:9090"),
			ChainID:           getEnv("BLOCKCHAIN_CHAIN_ID", "nuahchain"),
			RPCURL:            getEnv("BLOCKCHAIN_RPC_URL", "localhost:26657"),
			WebSocketURL:      getEnv("BLOCKCHAIN_WEBSOCKET_URL", "ws://localhost:26657/websocket"),
			WebSocketEnabled:  getEnvBool("BLOCKCHAIN_WEBSOCKET_ENABLED", true),
			ReconnectInterval: parseDuration(getEnv("BLOCKCHAIN_WEBSOCKET_RECONNECT_INTERVAL", "5s"), 5*time.Second),
			WebSocketTimeout:  parseDuration(getEnv("BLOCKCHAIN_WEBSOCKET_TIMEOUT", "30s"), 30*time.Second),
		},
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Address == "" {
		return fmt.Errorf("server address is required")
	}

	if c.Server.ReadTimeout <= 0 {
		return fmt.Errorf("read timeout must be positive")
	}

	if c.Server.WriteTimeout <= 0 {
		return fmt.Errorf("write timeout must be positive")
	}

	if c.Server.IdleTimeout <= 0 {
		return fmt.Errorf("idle timeout must be positive")
	}

	if err := c.Database.Validate(); err != nil {
		return fmt.Errorf("database config validation failed: %w", err)
	}

	if err := c.Logger.Validate(); err != nil {
		return fmt.Errorf("logger config validation failed: %w", err)
	}

	if err := c.Blockchain.Validate(); err != nil {
		return fmt.Errorf("blockchain config validation failed: %w", err)
	}

	return nil
}

// Validate validates database configuration
func (d *DatabaseConfig) Validate() error {
	if d.Host == "" {
		return fmt.Errorf("database host is required")
	}

	if d.Port == "" {
		return fmt.Errorf("database port is required")
	}

	if d.User == "" {
		return fmt.Errorf("database user is required")
	}

	if d.Database == "" {
		return fmt.Errorf("database name is required")
	}

	validSSLModes := map[string]bool{
		"disable":     true,
		"require":     true,
		"verify-ca":   true,
		"verify-full": true,
	}

	if d.SSLMode != "" && !validSSLModes[d.SSLMode] {
		return fmt.Errorf("invalid SSL mode: %s", d.SSLMode)
	}

	if d.MaxOpenConns <= 0 {
		return fmt.Errorf("max open connections must be positive")
	}

	if d.MaxIdleConns < 0 {
		return fmt.Errorf("max idle connections must be non-negative")
	}

	if d.MaxIdleConns > d.MaxOpenConns {
		return fmt.Errorf("max idle connections cannot exceed max open connections")
	}

	return nil
}

// DSN returns the PostgreSQL connection string
func (d *DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		d.Host,
		d.Port,
		d.User,
		d.Password,
		d.Database,
		d.SSLMode,
		int(d.ConnectTimeout.Seconds()),
	)
}

// getEnv gets an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvBool gets an environment variable as boolean or returns a default value
func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		return value == "true" || value == "1" || value == "yes"
	}
	return defaultValue
}

// parseDuration parses a duration string or returns a default value
func parseDuration(value string, defaultValue time.Duration) time.Duration {
	if value == "" {
		return defaultValue
	}
	d, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return d
}

// Validate validates blockchain configuration
func (b *BlockchainConfig) Validate() error {
	if b.NodeURL == "" {
		return fmt.Errorf("blockchain node URL is required")
	}

	if b.ChainID == "" {
		return fmt.Errorf("blockchain chain ID is required")
	}

	return nil
}

// Validate validates logger configuration
func (l *LoggerConfig) Validate() error {
	validLevels := map[string]bool{
		"debug": true,
		"info":  true,
		"warn":  true,
		"error": true,
		"fatal": true,
		"panic": true,
	}

	if l.Level != "" && !validLevels[strings.ToLower(l.Level)] {
		return fmt.Errorf("invalid log level: %s", l.Level)
	}

	validFormats := map[string]bool{
		"json": true,
		"text": true,
	}

	if l.Format != "" && !validFormats[strings.ToLower(l.Format)] {
		return fmt.Errorf("invalid log format: %s", l.Format)
	}

	validEnvironments := map[string]bool{
		"dev":  true,
		"prod": true,
	}

	if l.Environment != "" && !validEnvironments[strings.ToLower(l.Environment)] {
		return fmt.Errorf("invalid environment: %s", l.Environment)
	}

	return nil
}
