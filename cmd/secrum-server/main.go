package main

import (
	"github.com/hidden-life/secrum-server/internal/config"
	"github.com/hidden-life/secrum-server/internal/logger"
	"github.com/hidden-life/secrum-server/internal/server"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.New(cfg.ApplicationEnv, cfg.LogLevel)
	defer log.Sync()

	// Start a HTTP server
	srv := server.NewHTTPServer(log, cfg.HTTPPort)
	if err := srv.Start(); err != nil {
		log.Fatal("Server start failed", zap.Error(err))
	}
}
