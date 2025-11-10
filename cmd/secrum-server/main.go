package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hidden-life/secrum-server/internal/adapters/http"
	"github.com/hidden-life/secrum-server/internal/adapters/otp"
	internalRedis "github.com/hidden-life/secrum-server/internal/adapters/redis"
	"github.com/hidden-life/secrum-server/internal/app/auth"
	"github.com/hidden-life/secrum-server/internal/config"
	"github.com/hidden-life/secrum-server/internal/logger"
	"github.com/hidden-life/secrum-server/internal/server"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	log := logger.New(cfg.ApplicationEnv, cfg.LogLevel)
	defer log.Sync()

	// Init redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Warn("Failed to connect to Redis", zap.Error(err))
	}

	// OTP store + provider
	otpStore := internalRedis.New(rdb)
	otpProvider := otp.NewMockProvider(log, cfg.ApplicationEnv)

	// Auth service
	authSvc := auth.NewService(log, otpStore, otpProvider)

	// Start a HTTP server
	srv := server.NewHTTPServer(log, cfg.HTTPPort)
	http.RegisterAuthRoutes(srv.Router(), authSvc)

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Warn("HTTP server shutdown error", zap.Error(err))
	}

	log.Info("Shutting down server...")

	_ = rdb.Close()
}
