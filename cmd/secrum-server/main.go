package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hidden-life/secrum-server/internal/bootstrap"
	"github.com/hidden-life/secrum-server/internal/config"
	"go.uber.org/zap"
)

func main() {
	cfg := config.LoadConfig()
	ctx := context.Background()

	app, err := bootstrap.InitApp(ctx, cfg)
	if err != nil {
		panic(err)
	}

	log := app.Logger
	defer log.Sync()

	go func() {
		if err := app.Server.Start(); err != nil {
			log.Fatal("Server crashed", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Info("Shutdown signal received")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	_ = app.Server.Stop(shutdownCtx)
	_ = app.Redis.Close()

	log.Info("Server stopped successfully")
}
