package bootstrap

import (
	"github.com/hidden-life/secrum-server/internal/server"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

type App struct {
	Logger *zap.Logger
	Server *server.HTTPServer
	Redis  *redis.Client
}
