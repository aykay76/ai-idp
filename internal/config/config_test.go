package config

import (
	"os"
	"testing"
	"time"
)

func TestLoad(t *testing.T) {
	// Save original environment
	originalEnv := saveEnvironment()
	defer restoreEnvironment(originalEnv)

	// Clear environment
	clearEnvironment()

	// Test default values
	config := Load()

	if config.ServiceName != "ai-idp" {
		t.Errorf("Expected service name 'ai-idp', got '%s'", config.ServiceName)
	}

	if config.Environment != "development" {
		t.Errorf("Expected environment 'development', got '%s'", config.Environment)
	}

	if config.Server.Port != "8080" {
		t.Errorf("Expected port '8080', got '%s'", config.Server.Port)
	}

	if config.Server.Host != "0.0.0.0" {
		t.Errorf("Expected host '0.0.0.0', got '%s'", config.Server.Host)
	}

	if config.Logging.Level != "info" {
		t.Errorf("Expected log level 'info', got '%s'", config.Logging.Level)
	}
}

func TestLoadWithDefaults(t *testing.T) {
	// Save original environment
	originalEnv := saveEnvironment()
	defer restoreEnvironment(originalEnv)

	// Clear environment
	clearEnvironment()

	// Test custom defaults
	config := LoadWithDefaults("test-service", "9000")

	if config.ServiceName != "test-service" {
		t.Errorf("Expected service name 'test-service', got '%s'", config.ServiceName)
	}

	if config.Server.Port != "9000" {
		t.Errorf("Expected port '9000', got '%s'", config.Server.Port)
	}
}

func TestEnvironmentVariables(t *testing.T) {
	// Save original environment
	originalEnv := saveEnvironment()
	defer restoreEnvironment(originalEnv)

	// Set test environment variables
	testEnvVars := map[string]string{
		"SERVICE_NAME":       "custom-service",
		"PORT":               "3000",
		"HOST":               "127.0.0.1",
		"ENVIRONMENT":        "production",
		"DEBUG":              "true",
		"DATABASE_URL":       "postgres://test:test@localhost/test",
		"DB_MAX_CONNECTIONS": "50",
		"DB_MIN_CONNECTIONS": "10",
		"DB_CONNECT_TIMEOUT": "30s",
		"DB_MAX_IDLE_TIME":   "1h",
		"REDIS_URL":          "redis://localhost:6380",
		"REDIS_PASSWORD":     "secret",
		"REDIS_DB":           "2",
		"LOG_LEVEL":          "debug",
		"LOG_FORMAT":         "text",
		"JWT_SECRET":         "super-secret",
		"GITHUB_APP_ID":      "12345",
		"GITHUB_PRIVATE_KEY": "private-key-content",
		"SHUTDOWN_TIMEOUT":   "60s",
	}

	for key, value := range testEnvVars {
		os.Setenv(key, value)
	}

	config := Load()

	// Test all environment variables are loaded correctly
	if config.ServiceName != "custom-service" {
		t.Errorf("Expected service name 'custom-service', got '%s'", config.ServiceName)
	}

	if config.Server.Port != "3000" {
		t.Errorf("Expected port '3000', got '%s'", config.Server.Port)
	}

	if config.Server.Host != "127.0.0.1" {
		t.Errorf("Expected host '127.0.0.1', got '%s'", config.Server.Host)
	}

	if config.Environment != "production" {
		t.Errorf("Expected environment 'production', got '%s'", config.Environment)
	}

	if !config.Server.Debug {
		t.Error("Expected debug to be true")
	}

	if config.Database.URL != "postgres://test:test@localhost/test" {
		t.Errorf("Expected database URL 'postgres://test:test@localhost/test', got '%s'", config.Database.URL)
	}

	if config.Database.MaxConnections != 50 {
		t.Errorf("Expected max connections 50, got %d", config.Database.MaxConnections)
	}

	if config.Database.MinConnections != 10 {
		t.Errorf("Expected min connections 10, got %d", config.Database.MinConnections)
	}

	if config.Database.ConnectTimeout != 30*time.Second {
		t.Errorf("Expected connect timeout 30s, got %v", config.Database.ConnectTimeout)
	}

	if config.Database.MaxIdleTime != 1*time.Hour {
		t.Errorf("Expected max idle time 1h, got %v", config.Database.MaxIdleTime)
	}

	if config.Redis.URL != "redis://localhost:6380" {
		t.Errorf("Expected Redis URL 'redis://localhost:6380', got '%s'", config.Redis.URL)
	}

	if config.Redis.Password != "secret" {
		t.Errorf("Expected Redis password 'secret', got '%s'", config.Redis.Password)
	}

	if config.Redis.DB != 2 {
		t.Errorf("Expected Redis DB 2, got %d", config.Redis.DB)
	}

	if config.Logging.Level != "debug" {
		t.Errorf("Expected log level 'debug', got '%s'", config.Logging.Level)
	}

	if config.Logging.Format != "text" {
		t.Errorf("Expected log format 'text', got '%s'", config.Logging.Format)
	}

	if config.Security.JWTSecret != "super-secret" {
		t.Errorf("Expected JWT secret 'super-secret', got '%s'", config.Security.JWTSecret)
	}

	if config.GitHub.AppID != "12345" {
		t.Errorf("Expected GitHub app ID '12345', got '%s'", config.GitHub.AppID)
	}

	if config.GitHub.PrivateKey != "private-key-content" {
		t.Errorf("Expected GitHub private key 'private-key-content', got '%s'", config.GitHub.PrivateKey)
	}

	if config.Server.ShutdownTimeout != 60*time.Second {
		t.Errorf("Expected shutdown timeout 60s, got %v", config.Server.ShutdownTimeout)
	}
}

func TestValidation(t *testing.T) {
	tests := []struct {
		name        string
		config      *Config
		expectError bool
		errorMsg    string
	}{
		{
			name: "valid config",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					URL:            "postgres://localhost/test",
					MaxConnections: 25,
					MinConnections: 5,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				Security: SecurityConfig{
					JWTSecret: "secret",
				},
			},
			expectError: false,
		},
		{
			name: "missing database URL",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					MaxConnections: 25,
					MinConnections: 5,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			expectError: true,
			errorMsg:    "DATABASE_URL is required",
		},
		{
			name: "production without JWT secret",
			config: &Config{
				Environment: "production",
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					URL:            "postgres://localhost/test",
					MaxConnections: 25,
					MinConnections: 5,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
				Security: SecurityConfig{},
			},
			expectError: true,
			errorMsg:    "JWT_SECRET is required in production",
		},
		{
			name: "invalid environment",
			config: &Config{
				Environment: "invalid",
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					URL:            "postgres://localhost/test",
					MaxConnections: 25,
					MinConnections: 5,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			expectError: true,
			errorMsg:    "invalid environment 'invalid', must be one of: development, staging, production",
		},
		{
			name: "invalid log level",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					URL:            "postgres://localhost/test",
					MaxConnections: 25,
					MinConnections: 5,
				},
				Logging: LoggingConfig{
					Level: "invalid",
				},
			},
			expectError: true,
			errorMsg:    "invalid log level 'invalid', must be one of: debug, info, warn, error, fatal, panic",
		},
		{
			name: "max connections less than min",
			config: &Config{
				Environment: "development",
				Server: ServerConfig{
					Port: "8080",
				},
				Database: DatabaseConfig{
					URL:            "postgres://localhost/test",
					MaxConnections: 5,
					MinConnections: 10,
				},
				Logging: LoggingConfig{
					Level: "info",
				},
			},
			expectError: true,
			errorMsg:    "database max connections cannot be less than min connections",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.config.Validate()
			if tt.expectError && err == nil {
				t.Errorf("Expected error but got none")
			}
			if !tt.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if tt.expectError && err != nil && tt.errorMsg != "" {
				if err.Error() != tt.errorMsg {
					t.Errorf("Expected error message '%s', got '%s'", tt.errorMsg, err.Error())
				}
			}
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	config := &Config{Environment: "development"}
	if !config.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return true for development environment")
	}

	config.Environment = "production"
	if config.IsDevelopment() {
		t.Error("Expected IsDevelopment() to return false for production environment")
	}
}

func TestIsProduction(t *testing.T) {
	config := &Config{Environment: "production"}
	if !config.IsProduction() {
		t.Error("Expected IsProduction() to return true for production environment")
	}

	config.Environment = "development"
	if config.IsProduction() {
		t.Error("Expected IsProduction() to return false for development environment")
	}
}

func TestBackwardCompatibilityMethods(t *testing.T) {
	config := &Config{
		Database: DatabaseConfig{
			URL: "postgres://test",
		},
		Redis: RedisConfig{
			URL: "redis://test",
		},
		Server: ServerConfig{
			Port: "9000",
		},
		Security: SecurityConfig{
			JWTSecret: "secret",
		},
	}

	if config.DatabaseURL() != "postgres://test" {
		t.Errorf("Expected DatabaseURL() to return 'postgres://test', got '%s'", config.DatabaseURL())
	}

	if config.RedisURL() != "redis://test" {
		t.Errorf("Expected RedisURL() to return 'redis://test', got '%s'", config.RedisURL())
	}

	if config.Port() != "9000" {
		t.Errorf("Expected Port() to return '9000', got '%s'", config.Port())
	}

	if config.JWTSecret() != "secret" {
		t.Errorf("Expected JWTSecret() to return 'secret', got '%s'", config.JWTSecret())
	}
}

func TestGetBoolEnv(t *testing.T) {
	tests := []struct {
		value    string
		expected bool
	}{
		{"true", true},
		{"1", true},
		{"yes", true},
		{"on", true},
		{"false", false},
		{"0", false},
		{"no", false},
		{"off", false},
		{"invalid", false}, // should use default
	}

	for _, tt := range tests {
		t.Run(tt.value, func(t *testing.T) {
			os.Setenv("TEST_BOOL", tt.value)
			result := getBoolEnv("TEST_BOOL", false)
			if result != tt.expected {
				t.Errorf("Expected %v for value '%s', got %v", tt.expected, tt.value, result)
			}
			os.Unsetenv("TEST_BOOL")
		})
	}

	// Test default value when env var is not set
	result := getBoolEnv("NONEXISTENT_BOOL", true)
	if result != true {
		t.Errorf("Expected default value true, got %v", result)
	}
}

func TestGetIntEnv(t *testing.T) {
	os.Setenv("TEST_INT", "42")
	result := getIntEnv("TEST_INT", 10)
	if result != 42 {
		t.Errorf("Expected 42, got %d", result)
	}
	os.Unsetenv("TEST_INT")

	// Test invalid value falls back to default
	os.Setenv("TEST_INT", "invalid")
	result = getIntEnv("TEST_INT", 10)
	if result != 10 {
		t.Errorf("Expected default value 10, got %d", result)
	}
	os.Unsetenv("TEST_INT")

	// Test default value when env var is not set
	result = getIntEnv("NONEXISTENT_INT", 99)
	if result != 99 {
		t.Errorf("Expected default value 99, got %d", result)
	}
}

func TestGetDurationEnv(t *testing.T) {
	// Test duration format
	os.Setenv("TEST_DURATION", "30s")
	result := getDurationEnv("TEST_DURATION", 10*time.Second)
	if result != 30*time.Second {
		t.Errorf("Expected 30s, got %v", result)
	}
	os.Unsetenv("TEST_DURATION")

	// Test seconds format
	os.Setenv("TEST_DURATION", "60")
	result = getDurationEnv("TEST_DURATION", 10*time.Second)
	if result != 60*time.Second {
		t.Errorf("Expected 60s, got %v", result)
	}
	os.Unsetenv("TEST_DURATION")

	// Test invalid value falls back to default
	os.Setenv("TEST_DURATION", "invalid")
	result = getDurationEnv("TEST_DURATION", 5*time.Second)
	if result != 5*time.Second {
		t.Errorf("Expected default value 5s, got %v", result)
	}
	os.Unsetenv("TEST_DURATION")
}

// Helper functions for environment management in tests
func saveEnvironment() map[string]string {
	env := make(map[string]string)
	envVars := []string{
		"SERVICE_NAME", "PORT", "HOST", "ENVIRONMENT", "DEBUG",
		"DATABASE_URL", "DB_MAX_CONNECTIONS", "DB_MIN_CONNECTIONS",
		"DB_CONNECT_TIMEOUT", "DB_MAX_IDLE_TIME",
		"REDIS_URL", "REDIS_PASSWORD", "REDIS_DB",
		"LOG_LEVEL", "LOG_FORMAT", "JWT_SECRET",
		"GITHUB_APP_ID", "GITHUB_PRIVATE_KEY", "SHUTDOWN_TIMEOUT",
	}

	for _, key := range envVars {
		if value := os.Getenv(key); value != "" {
			env[key] = value
		}
	}
	return env
}

func restoreEnvironment(env map[string]string) {
	// Clear all test env vars first
	envVars := []string{
		"SERVICE_NAME", "PORT", "HOST", "ENVIRONMENT", "DEBUG",
		"DATABASE_URL", "DB_MAX_CONNECTIONS", "DB_MIN_CONNECTIONS",
		"DB_CONNECT_TIMEOUT", "DB_MAX_IDLE_TIME",
		"REDIS_URL", "REDIS_PASSWORD", "REDIS_DB",
		"LOG_LEVEL", "LOG_FORMAT", "JWT_SECRET",
		"GITHUB_APP_ID", "GITHUB_PRIVATE_KEY", "SHUTDOWN_TIMEOUT",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}

	// Restore original values
	for key, value := range env {
		os.Setenv(key, value)
	}
}

func clearEnvironment() {
	envVars := []string{
		"SERVICE_NAME", "PORT", "HOST", "ENVIRONMENT", "DEBUG",
		"DATABASE_URL", "DB_MAX_CONNECTIONS", "DB_MIN_CONNECTIONS",
		"DB_CONNECT_TIMEOUT", "DB_MAX_IDLE_TIME",
		"REDIS_URL", "REDIS_PASSWORD", "REDIS_DB",
		"LOG_LEVEL", "LOG_FORMAT", "JWT_SECRET",
		"GITHUB_APP_ID", "GITHUB_PRIVATE_KEY", "SHUTDOWN_TIMEOUT",
	}

	for _, key := range envVars {
		os.Unsetenv(key)
	}
}
