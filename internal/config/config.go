package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
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

	MaxUploadMB int
}

// LoadConfig loads configuration from environment variables with defaults
func LoadConfig() *Config {
	flags := ParseFlags()
	// load .env (optional)
	_ = godotenv.Load()

	env := pick(flags.Env, os.Getenv("APP_ENV"), "dev")

	v := viper.New()
	v.SetConfigType("yaml")
	if flags.ConfigPath != "" {
		v.SetConfigFile(flags.ConfigPath)
	} else {
		v.SetConfigName(fmt.Sprintf("config.%s", env))
		v.AddConfigPath("config")
		v.AddConfigPath(".")
	}

	v.AutomaticEnv()

	// env mapping
	_ = v.BindEnv("database.url", "DB_URL")
	_ = v.BindEnv("jwt.access_secret", "JWT_ACCESS_SECRET")
	_ = v.BindEnv("jwt.refresh_secret", "JWT_REFRESH_SECRET")
	_ = v.BindEnv("redis.addr", "REDIS_ADDR")
	_ = v.BindEnv("http.port", "HTTP_PORT")

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Errorf("fatal error config: %w", err))
	}

	cfg := &Config{
		ApplicationName:      v.GetString("app.name"),
		ApplicationEnv:       v.GetString("app.env"),
		HTTPPort:             v.GetString("http.port"),
		DatabaseURL:          v.GetString("database.url"),
		RedisAddress:         v.GetString("redis.addr"),
		JWTAccessSecret:      v.GetString("jwt.access_secret"),
		JWTRefreshSecret:     v.GetString("jwt.refresh_secret"),
		JWTAccessTTLMinutes:  v.GetInt("jwt.access_ttl_minutes"),
		JWTRefreshTTLMinutes: v.GetInt("jwt.refresh_ttl_minutes"),
		FileStorageDir:       v.GetString("storage.local_dir"),
		MaxUploadMB:          v.GetInt("security.max_upload_mb"),
	}

	overrideByFlags(cfg, flags)
	validate(cfg)

	return cfg
}

func validate(cfg *Config) {
	assert(cfg.DatabaseURL, "database.url")
	assert(cfg.JWTAccessSecret, "jwt.access_secret")
	assert(cfg.JWTRefreshSecret, "jwt.refresh_secret")
}

func assert(val, name string) {
	if val == "" {
		panic("config field is required: " + name)
	}
}

func pick(values ...string) string {
	for _, v := range values {
		if v != "" {
			return v
		}
	}

	return ""
}
