package bootstrap

import (
	"context"

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
	"github.com/hidden-life/secrum-server/internal/app/sync"
	"github.com/hidden-life/secrum-server/internal/config"
	"github.com/hidden-life/secrum-server/internal/logger"
	"github.com/hidden-life/secrum-server/internal/presence"
	"github.com/hidden-life/secrum-server/internal/real_time"
	"github.com/hidden-life/secrum-server/internal/server"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

func InitApp(ctx context.Context, cfg *config.Config) (*App, error) {
	log := logger.New(cfg.ApplicationEnv, cfg.LogLevel)

	// redis
	rdb := redis.NewClient(&redis.Options{
		Addr: cfg.RedisAddress,
	})

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Warn("Failed to connect to redis", zap.Error(err))
	}

	// postgres
	pool, err := postgres.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, err
	}

	// repositories
	keyRepo := postgres.NewKeyRepository(pool)
	otkpRepo := postgres.NewOTPKRepository(pool)
	userRepo := postgres.NewUserRepository(pool)
	deviceRepo := postgres.NewDeviceRepository(pool)
	msgRepo := postgres.NewMessageRepository(pool)
	groupRepo := postgres.NewGroupRepository(pool)
	groupMemberRepo := postgres.NewGroupMemberRepository(pool)
	attachmentsRepo := postgres.NewAttachmentRepository(pool)
	contactRepo := postgres.NewContactRepository(pool)
	chatStateRepo := postgres.NewChatStateRepository(pool)
	syncRepo := postgres.NewSyncEventRepository(pool)

	// services
	keySvc := keys.NewService(log, keyRepo, otkpRepo)

	// redis storage
	otpStore := internalRedis.New(rdb)
	otpProvider := otp.NewMockProvider(log, cfg.ApplicationEnv)

	tokenManager := jwt.NewManager(
		cfg.JWTAccessSecret,
		cfg.JWTRefreshSecret,
		cfg.ApplicationName,
		cfg.JWTRefreshTTLMinutes,
		cfg.JWTRefreshTTLMinutes,
	)

	// session store
	sessionStore := internalRedis.NewSessionStore(rdb)

	authSvc := auth.NewService(log, cfg.ApplicationEnv, otpStore, otpProvider, userRepo, deviceRepo, tokenManager)
	deviceSvc := devices.NewService(log, deviceRepo, sessionStore)
	contactSvc := contact.NewService(userRepo, contactRepo)
	profileSvc := profile.NewService(log, userRepo)
	chatSvc := chats.NewService(log, msgRepo, userRepo, chatStateRepo)
	syncSvc := sync.NewService(log, chatSvc, syncRepo)

	realtimeHub := real_time.NewDeliveryHub(log)

	groupSvc := groups.NewService(log, groupRepo, groupMemberRepo, userRepo, deviceRepo, msgRepo, realtimeHub, syncRepo)
	msgSvc := messages.NewService(log, msgRepo, userRepo, deviceRepo, realtimeHub, syncRepo, groupRepo, groupMemberRepo)

	st, _ := storage.NewLocalStorage(cfg.FileStorageDir)
	attachmentsSvc := attachments.NewService(log, attachmentsRepo, msgRepo, groupRepo, groupMemberRepo, st, "attachments", 50<<20)
	presenceSvc := presence.New(rdb)

	// HTTP
	srv := server.NewHTTPServer(log, cfg.HTTPPort)
	authMW := http.AuthMiddleware(tokenManager, sessionStore, deviceRepo)

	router := srv.Router()

	http.RegisterAuthRoutes(router, authSvc)
	http.RegisterKeyRoutes(router, keySvc, tokenManager, sessionStore, deviceRepo)

	router.Group(func(r chi.Router) {
		r.Use(authMW)

		http.RegisterMessagesRoutes(r, msgSvc)
		http.RegisterProfileRoutes(r, profileSvc)
		http.RegisterChatRoutes(r, chatSvc)
		http.RegisterContactRoutes(r, contactSvc)
		http.RegisterDevicesRoutes(r, deviceSvc)
		http.RegisterGroupsRoutes(r, groupSvc, msgSvc)
		http.RegisterAttachmentsRoutes(r, attachmentsSvc)
		http.RegisterSyncEventRoutes(r, syncSvc)
		http.RegisterWSRoutes(r, log, realtimeHub, msgSvc, presenceSvc)
	})

	return &App{
		Logger: log,
		Server: srv,
		Redis:  rdb,
	}, nil
}
