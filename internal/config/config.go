package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// ServerConfig holds server-specific configuration
type ServerConfig struct {
	Port            string        `json:"port" mapstructure:"port"`
	Host            string        `json:"host" mapstructure:"host"`
	ShutdownTimeout time.Duration `json:"shutdown_timeout" mapstructure:"shutdown_timeout"`
	Debug           bool          `json:"debug" mapstructure:"debug"`
}

// DatabaseConfig holds database configuration
type DatabaseConfig struct {
	URL            string        `json:"url" mapstructure:"url" validate:"required"`
	MaxConnections int32         `json:"max_connections" mapstructure:"max_connections"`
	MinConnections int32         `json:"min_connections" mapstructure:"min_connections"`
	ConnectTimeout time.Duration `json:"connect_timeout" mapstructure:"connect_timeout"`
	MaxIdleTime    time.Duration `json:"max_idle_time" mapstructure:"max_idle_time"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL      string `json:"url" mapstructure:"url"`
	Password string `json:"password" mapstructure:"password"`
	DB       int    `json:"db" mapstructure:"db"`
}

// LoggingConfig holds logging configuration
type LoggingConfig struct {
	Level  string `json:"level" mapstructure:"level"`
	Format string `json:"format" mapstructure:"format"`
}

// SecurityConfig holds security-related configuration
type SecurityConfig struct {
	JWTSecret string `json:"jwt_secret" mapstructure:"jwt_secret"`
}

// GitHubConfig holds GitHub integration configuration
type GitHubConfig struct {
	AppID      string `json:"app_id" mapstructure:"app_id"`
	PrivateKey string `json:"private_key" mapstructure:"private_key"`
}

// Config holds the complete application configuration
type Config struct {
	// Environment and service info
	Environment string `json:"environment" mapstructure:"environment"`
	ServiceName string `json:"service_name" mapstructure:"service_name"`

	// Component configurations
	Server   ServerConfig   `json:"server" mapstructure:"server"`
	Database DatabaseConfig `json:"database" mapstructure:"database"`
	Redis    RedisConfig    `json:"redis" mapstructure:"redis"`
	Logging  LoggingConfig  `json:"logging" mapstructure:"logging"`
	Security SecurityConfig `json:"security" mapstructure:"security"`
	GitHub   GitHubConfig   `json:"github" mapstructure:"github"`
}

// Load loads configuration from environment variables with defaults
func Load() *Config {
	return LoadWithDefaults("ai-idp", "8080")
}

// LoadWithDefaults loads configuration with custom service name and port defaults
func LoadWithDefaults(serviceName, defaultPort string) *Config {
	config := &Config{
		Environment: getEnv("ENVIRONMENT", "development"),
		ServiceName: getEnv("SERVICE_NAME", serviceName),

		Server: ServerConfig{
			Port:            getEnv("PORT", defaultPort),
			Host:            getEnv("HOST", "0.0.0.0"),
			ShutdownTimeout: getDurationEnv("SHUTDOWN_TIMEOUT", 30*time.Second),
			Debug:           getBoolEnv("DEBUG", false),
		},

		Database: DatabaseConfig{
			URL:            getEnv("DATABASE_URL", "postgres://platform:platform_dev_password@localhost:5432/platform?sslmode=disable"),
			MaxConnections: getIntEnv("DB_MAX_CONNECTIONS", 25),
			MinConnections: getIntEnv("DB_MIN_CONNECTIONS", 5),
			ConnectTimeout: getDurationEnv("DB_CONNECT_TIMEOUT", 10*time.Second),
			MaxIdleTime:    getDurationEnv("DB_MAX_IDLE_TIME", 30*time.Minute),
		},

		Redis: RedisConfig{
			URL:      getEnv("REDIS_URL", "redis://:redis_dev_password@localhost:6379/0"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       int(getIntEnv("REDIS_DB", 0)),
		},

		Logging: LoggingConfig{
			Level:  getEnv("LOG_LEVEL", "info"),
			Format: getEnv("LOG_FORMAT", "json"),
		},

		Security: SecurityConfig{
			JWTSecret: getEnv("JWT_SECRET", "dev_jwt_secret_change_in_production"),
		},

		GitHub: GitHubConfig{
			AppID:      getEnv("GITHUB_APP_ID", ""),
			PrivateKey: getEnv("GITHUB_PRIVATE_KEY", ""),
		},
	}

	return config
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Database.URL == "" {
		return fmt.Errorf("DATABASE_URL is required")
	}

	if c.Environment == "production" && c.Security.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required in production")
	}

	// Validate server configuration
	if c.Server.Port == "" {
		return fmt.Errorf("server port is required")
	}

	// Validate database connection limits
	if c.Database.MaxConnections < c.Database.MinConnections {
		return fmt.Errorf("database max connections cannot be less than min connections")
	}

	// Validate environment
	validEnvs := []string{"development", "staging", "production"}
	validEnv := false
	for _, env := range validEnvs {
		if c.Environment == env {
			validEnv = true
			break
		}
	}
	if !validEnv {
		return fmt.Errorf("invalid environment '%s', must be one of: development, staging, production", c.Environment)
	}

	// Validate log level
	validLogLevels := []string{"debug", "info", "warn", "error", "fatal", "panic"}
	validLogLevel := false
	for _, level := range validLogLevels {
		if c.Logging.Level == level {
			validLogLevel = true
			break
		}
	}
	if !validLogLevel {
		return fmt.Errorf("invalid log level '%s', must be one of: debug, info, warn, error, fatal, panic", c.Logging.Level)
	}

	return nil
}

// IsDevelopment returns true if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// getEnv gets an environment variable with a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv gets a boolean environment variable with a default value
func getBoolEnv(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		switch value {
		case "true", "1", "yes", "on":
			return true
		case "false", "0", "no", "off":
			return false
		}
	}
	return defaultValue
}

// getIntEnv gets an integer environment variable with a default value
func getIntEnv(key string, defaultValue int32) int32 {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.ParseInt(value, 10, 32); err == nil {
			return int32(intValue)
		}
	}
	return defaultValue
}

// getDurationEnv gets a duration environment variable with a default value
func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		// Try parsing as seconds if duration parsing fails
		if seconds, err := strconv.Atoi(value); err == nil {
			return time.Duration(seconds) * time.Second
		}
	}
	return defaultValue
}

// DatabaseURL returns the database URL for backward compatibility
func (c *Config) DatabaseURL() string {
	return c.Database.URL
}

// RedisURL returns the Redis URL for backward compatibility
func (c *Config) RedisURL() string {
	return c.Redis.URL
}

// Port returns the server port for backward compatibility
func (c *Config) Port() string {
	return c.Server.Port
}

// JWTSecret returns the JWT secret for backward compatibility
func (c *Config) JWTSecret() string {
	return c.Security.JWTSecret
}
