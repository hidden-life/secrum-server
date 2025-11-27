package config

// overrideByFlags overrides given CLI flags by Config
func overrideByFlags(c *Config, f *Flags) {
	if f.HTTPort != "" {
		c.HTTPPort = f.HTTPort
	}

	if f.DBUrl != "" {
		c.DatabaseURL = f.DBUrl
	}

	if f.RedisAddr != "" {
		c.RedisAddress = f.RedisAddr
	}

	if f.Env != "" {
		c.ApplicationEnv = f.Env
	}
}
