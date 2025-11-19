package config

import (
	"log"
	"os"
	"strconv"
)

// Config holds global application configuration
type Config struct {
	ApplicationName string
	ApplicationEnv  string
	HTTPPort        string
	LogLevel        string

	DatabaseURL  string
	RedisAddress string

	FileStorageDir string

	JWTAccessSecret      string
	JWTRefreshSecret     string
	JWTAccessTTLMinutes  int
	JWTRefreshTTLMinutes int
}

// getEnv retrieves the value of the environment variable named by the key.
func getEnv(key, fallbackStr string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallbackStr
}

// getEnvInt retrieves the integer value of the environment variable named by the key.
func getEnvInt(key string, fallbackInt int) int {
	if value, exists := os.LookupEnv(key); exists {
		if n, err := strconv.Atoi(value); err == nil {
			return n
		}
	}
	return fallbackInt
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	cfg := &Config{
		ApplicationName: getEnv("APP_NAME", "SecrumServer"),
		ApplicationEnv:  getEnv("APP_ENV", "dev"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),

		DatabaseURL:  getEnv("DATABASE_URL", "postgres://secrum_user:secrum_password@localhost:5432/secrum_db?sslmode=disable"),
		RedisAddress: getEnv("REDIS_ADDR", "localhost:6379"),

		FileStorageDir: getEnv("FILE_STORAGE_DIR", "attachments"),

		JWTAccessSecret:      getEnv("JWT_ACCESS_SECRET", "dev_access_secret"),
		JWTRefreshSecret:     getEnv("JWT_REFRESH_SECRET", "dev_refresh_secret"),
		JWTAccessTTLMinutes:  getEnvInt("JWT_ACCESS_TTL_MINUTES", 15),     // 15 minutes
		JWTRefreshTTLMinutes: getEnvInt("JWT_REFRESH_TTL_MINUTES", 60*24), // 24 hours
	}

	log.Printf("[CONFIG] Loaded configuration for %s (%s)", cfg.ApplicationName, cfg.ApplicationEnv)

	return cfg
}
