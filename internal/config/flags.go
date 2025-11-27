package config

import "flag"

type Flags struct {
	ConfigPath string
	Env        string
	HTTPort    string
	DBUrl      string
	RedisAddr  string
}

func ParseFlags() *Flags {
	f := &Flags{}

	flag.StringVar(&f.ConfigPath, "config", "", "Path to config file")
	flag.StringVar(&f.Env, "env", "", "Application environment (dev | prod)")
	flag.StringVar(&f.HTTPort, "port", "8080", "Application port to listen on")
	flag.StringVar(&f.DBUrl, "db", "", "Database URL")
	flag.StringVar(&f.RedisAddr, "redis", "", "Redis address")

	flag.Parse()
	return f
}
