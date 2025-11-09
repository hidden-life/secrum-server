package config

import (
	"log"
	"os"
)

// Config holds global application configuration
type Config struct {
	ApplicationName string
	ApplicationEnv  string
	HTTPPort        string
	LogLevel        string
	DatabaseURL     string
	RedisAddress    string
}

// getEnv retrieves the value of the environment variable named by the key.
func getEnv(key, fallbackStr string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}

	return fallbackStr
}

func LoadConfig() *Config {
	cfg := &Config{
		ApplicationName: getEnv("APP_NAME", "SecrumServer"),
		ApplicationEnv:  getEnv("APP_ENV", "dev"),
		HTTPPort:        getEnv("HTTP_PORT", "8080"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://secrum_user:secrum_password@localhost:5432/secrum_db?sslmode=disable"),
		RedisAddress:    getEnv("REDIS_ADDR", "localhost:6379"),
	}

	log.Printf("[CONFIG] Loaded configuration for %s (%s)", cfg.ApplicationName, cfg.ApplicationEnv)

	return cfg
}
