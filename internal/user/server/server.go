package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/redis/go-redis/v9"

	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/internal/user/delivery/http/v1"
	"github.com/hexley21/fixup/internal/user/jwt/refresh_jwt"
	"github.com/hexley21/fixup/internal/user/jwt/verify_jwt"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/hexley21/fixup/pkg/validator"
)

type services struct {
	authService service.AuthService
	userService service.UserService
}

type jWTManagers struct {
	accessJWTManager       auth_jwt.Manager
	refreshJWTManager      refresh_jwt.Manager
	verificationJWTManager verify_jwt.Manager
}
type server struct {
	router            chi.Router
	metricsRouter     chi.Router
	mux               *http.Server
	metricsMux        *http.Server
	cfg               *config.Config
	dbPool            *pgxpool.Pool
	redisCluster      *redis.ClusterClient
	handlerComponents *handler.Components
	jWTManagers       *jWTManagers
	services          *services
	cdnUrlSigner      cdn.URLSigner
}

// NewServer initializes and returns a new server instance with the provided configuration and dependencies.
// It sets up repositories, services, JWT managers, handler components, and HTTP servers for both main and metrics endpoints.
func NewServer(
	cfg *config.Config,
	dbPool *pgxpool.Pool,
	redisCluster *redis.ClusterClient,
	logger logger.Logger,
	snowflakeNode *snowflake.Node,
	validator validator.Validator,
	s3Bucket s3.Bucket,
	cdnFileInvalidator cdn.FileInvalidator,
	hasher hasher.Hasher,
	encryptor encryption.Encryptor,
	mailer mailer.Mailer,
) *server {
	userRepository := repository.NewUserRepository(dbPool, snowflakeNode)
	providerRepository := repository.NewProviderRepository(dbPool)
	verificationRepository := repository.NewVerificationRepository(redisCluster)

	authService := service.NewAuthService(
		userRepository,
		providerRepository,
		verificationRepository,
		cfg.JWT.VerificationTTL,
		dbPool,
		hasher,
		encryptor,
		mailer,
		cfg.Server.Email,
	)
	if err := authService.ParseTemplates(cfg.Templates); err != nil {
		logger.Fatalf("error starting server %v", err)
	}

	userService := service.NewUserService(
		userRepository,
		s3Bucket,
		cdnFileInvalidator,
		hasher,
	)

	services := &services{
		authService: authService,
		userService: userService,
	}

	jWTManagers := &jWTManagers{
		accessJWTManager:       auth_jwt.NewManager(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL),
		refreshJWTManager:      refresh_jwt.NewManager(cfg.JWT.RefreshSecret, cfg.JWT.RefreshTTL),
		verificationJWTManager: verify_jwt.NewManager(cfg.JWT.VerificationSecret, cfg.JWT.VerificationTTL),
	}

	jsonManager := std_json.New()

	handlerComponents := &handler.Components{
		Logger:    logger,
		Binder:    std_binder.New(jsonManager),
		Validator: validator,
		Writer:    json_writer.New(logger, jsonManager),
	}

	router := chi.NewMux()
	mux := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      router,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	metricsRouter := chi.NewMux()
	metricsMux := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Metrics.Port),
		Handler:      metricsRouter,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	return &server{
		router:            router,
		metricsRouter:     metricsRouter,
		mux:               mux,
		metricsMux:        metricsMux,
		cfg:               cfg,
		dbPool:            dbPool,
		handlerComponents: handlerComponents,
		jWTManagers:       jWTManagers,
		services:          services,
		cdnUrlSigner:      cdn.NewCloudFrontURLSigner(cfg.AWS.CDN),
	}
}

// Run starts the server and its associated components, including the main HTTP server and metrics server.
// It returns an error if either the main server or the metrics server fails to start or run.
func (s *server) Run() error {
	// Initialize middleware with binder and writer components
	Middleware := middleware.NewMiddleware(s.handlerComponents.Binder, s.handlerComponents.Writer)

	// Set up logging middleware for chi router
	chiLogger := &chi_middleware.DefaultLogFormatter{
		Logger:  s.handlerComponents.Logger,
		NoColor: false,
	}

	s.router.Use(chi_middleware.Recoverer)
	s.router.Use(chi_middleware.RequestLogger(chiLogger))
	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(s.cfg.HTTP.CorsOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	v1.MapV1Routes(v1.RouterArgs{
		AuthService:            s.services.authService,
		UserService:            s.services.userService,
		Middleware:             Middleware,
		HandlerComponents:      s.handlerComponents,
		AccessJWTManager:       s.jWTManagers.accessJWTManager,
		RefreshJWTManager:      s.jWTManagers.refreshJWTManager,
		VerificationJWTManager: s.jWTManagers.verificationJWTManager,
		CdnUrlSigner:           s.cdnUrlSigner,
	}, s.router)

	// Setup metrics endpoint
	s.metricsRouter.Use(chi_middleware.Recoverer)
	s.metricsRouter.Handle("/metrics", promhttp.Handler())

	mainErrChan := make(chan error, 1)
	metricsErrChan := make(chan error, 1)

	go func() {
		mainErrChan <- s.mux.ListenAndServe()
	}()

	go func() {
		metricsErrChan <- s.metricsMux.ListenAndServe()
	}()

	select {
	case mainErr := <-mainErrChan:
		return mainErr
	case metricsErr := <-metricsErrChan:
		return metricsErr
	}
}

// Close gracefully shuts down the server, including its HTTP mux, metrics mux, database pool, and Redis cluster.
// Errors during shutdown are logged, but the function returns nil to ensure all components attempt to close.
// Complies to io.Closer interface.
func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	err := s.mux.Shutdown(ctx)
	if err != nil {
		s.handlerComponents.Logger.Error(err)
		err = nil
	}

	err = s.metricsMux.Shutdown(ctx)
	if err != nil {
		s.handlerComponents.Logger.Error(err)
		err = nil
	}

	err = postgres.Close(s.dbPool)
	if err != nil {
		s.handlerComponents.Logger.Error(err)
		err = nil
	}

	err = s.redisCluster.Close()
	if err != nil {
		s.handlerComponents.Logger.Error(err)
	}

	return nil
}
