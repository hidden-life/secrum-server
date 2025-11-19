package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/hidden-life/secrum-server/internal/adapters/http"
	"github.com/hidden-life/secrum-server/internal/adapters/jwt"
	"github.com/hidden-life/secrum-server/internal/adapters/otp"
	"github.com/hidden-life/secrum-server/internal/adapters/postgres"
	internalRedis "github.com/hidden-life/secrum-server/internal/adapters/redis"
	"github.com/hidden-life/secrum-server/internal/adapters/storage"
	"github.com/hidden-life/secrum-server/internal/app/attachments"
	"github.com/hidden-life/secrum-server/internal/app/auth"
	"github.com/hidden-life/secrum-server/internal/app/chats"
	"github.com/hidden-life/secrum-server/internal/app/contact"
	"github.com/hidden-life/secrum-server/internal/app/devices"
	"github.com/hidden-life/secrum-server/internal/app/groups"
	"github.com/hidden-life/secrum-server/internal/app/keys"
	"github.com/hidden-life/secrum-server/internal/app/messages"
	"github.com/hidden-life/secrum-server/internal/app/profile"
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

	ctx := context.Background()

	// Init redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Warn("Failed to connect to Redis", zap.Error(err))
	}

	// Init postgresql
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatal("Failed to connect to Postgres", zap.Error(err))
	}
	defer pool.Close()

	keyRepository := postgres.NewKeyRepository(pool)
	otpkRepo := postgres.NewOTPKRepository(pool)
	keyService := keys.NewService(log, keyRepository, otpkRepo)

	// Repositories
	userRepo := postgres.NewUserRepository(pool)
	deviceRepo := postgres.NewDeviceRepository(pool)

	// OTP store + provider
	otpStore := internalRedis.New(rdb)
	otpProvider := otp.NewMockProvider(log, cfg.ApplicationEnv)

	// Token manager (JWT)
	tokenManager := jwt.NewManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.ApplicationName,
		cfg.JWTAccessTTLMinutes,
		cfg.JWTRefreshTTLMinutes,
	)

	// Auth service
	authSvc := auth.NewService(log, cfg.ApplicationEnv, otpStore, otpProvider, userRepo, deviceRepo, tokenManager)

	// Start a HTTP server
	srv := server.NewHTTPServer(log, cfg.HTTPPort)

	// message repository
	msgRepo := postgres.NewMessageRepository(pool)
	msgSvc := messages.NewService(log, msgRepo, userRepo, deviceRepo)

	// auth middleware
	authMW := http.AuthMiddleware(tokenManager)

	// user profile
	profileSvc := profile.NewService(log, userRepo)
	// Register new routes
	http.RegisterAuthRoutes(srv.Router(), authSvc)
	http.RegisterKeyRoutes(srv.Router(), keyService, tokenManager)

	// devices
	deviceSvc := devices.NewService(log, deviceRepo)

	// Contacts
	contactRepo := postgres.NewContactRepository(pool)
	contactSvc := contact.NewService(userRepo, contactRepo)
	// Chats
	chatSvc := chats.NewService(log, msgRepo, userRepo)

	groupRepo := postgres.NewGroupRepository(pool)
	groupMemberRepo := postgres.NewGroupMemberRepository(pool)
	groupSvc := groups.NewService(log, groupRepo, groupMemberRepo, userRepo, deviceRepo, msgRepo)

	attachmentsRepo := postgres.NewAttachmentRepository(pool)
	localStorage, err := storage.NewLocalStorage(cfg.FileStorageDir) // @todo: Change to using from configuration
	attachmentSvc := attachments.NewService(log, attachmentsRepo, localStorage, "attachments", 1024*1024*50)

	srv.Router().Group(func(r chi.Router) {
		r.Use(authMW)
		http.RegisterMessagesRoutes(r, msgSvc)
		http.RegisterProfileRoutes(r, profileSvc)
		http.RegisterContactRoutes(r, contactSvc)
		http.RegisterChatRoutes(r, chatSvc)
		http.RegisterDevicesRoutes(r, deviceSvc)
		http.RegisterGroupsRoutes(r, groupSvc, msgSvc)
		http.RegisterAttachmentsRoutes(r, attachmentSvc)
	})
	// Start server async
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatal("Failed to start HTTP server", zap.Error(err))
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Stop(shutdownCtx); err != nil {
		log.Warn("HTTP server shutdown error", zap.Error(err))
	}
	if err := rdb.Close(); err != nil {
		log.Warn("Redis client close error", zap.Error(err))
	}

	log.Info("SecRum server stopped")
}
