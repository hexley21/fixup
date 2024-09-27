package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
	chi_middleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	v1 "github.com/hexley21/fixup/internal/user/delivery/http/v1"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/internal/user/service/verifier"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/http/binder"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/writer"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/hexley21/fixup/pkg/validator"
	plaground_validator "github.com/hexley21/fixup/pkg/validator/playground_validator"
)

type services struct {
	authService service.AuthService
	userService service.UserService
}

type jWTManagers struct {
	accessJWTManager       auth_jwt.JWTManager
	refreshJWTManager      auth_jwt.JWTManager
	verificationJWTManager verifier.JWTManager
}

type requestComponents struct {
	logger     logger.Logger
	binder     binder.FullBinder
	validator  validator.Validator
	httpWriter writer.HTTPWriter
}

type server struct {
	router            chi.Router
	mux               *http.Server
	cfg               *config.Config
	dbPool            *pgxpool.Pool
	requestComponents *requestComponents
	jWTManagers       *jWTManagers
	services          *services
}

func NewServer(
	cfg *config.Config,
	dbPool *pgxpool.Pool,
	logger logger.Logger,
	snowflakeNode *snowflake.Node,
	validator validator.Validator,
	s3Bucket s3.Bucket,
	cdnFileInvalidator cdn.FileInvalidator,
	hasher hasher.Hasher,
	encryptor encryption.Encryptor,
	mailer mailer.Mailer,
) *server {
	cdnURLSigner := cdn.NewCloudFrontURLSigner(cfg.AWS.CDN)

	userRepository := repository.NewUserRepository(dbPool, snowflakeNode)
	providerRepository := repository.NewProviderRepository(dbPool)

	authService := service.NewAuthService(
		userRepository,
		providerRepository,
		dbPool,
		hasher,
		encryptor,
		mailer,
		cfg.Server.Email,
		cdnURLSigner,
	)
	if err := authService.ParseTemplates(); err != nil {
		logger.Fatalf("error starting server %w", err)
	}

	userService := service.NewUserService(
		userRepository,
		s3Bucket,
		cdnFileInvalidator,
		cdnURLSigner,
		hasher,
	)

	services := &services{
		authService: authService,
		userService: userService,
	}

	jWTManagers := &jWTManagers{
		accessJWTManager:       auth_jwt.NewJWTManager(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL),
		refreshJWTManager:      auth_jwt.NewJWTManager(cfg.JWT.RefreshSecret, cfg.JWT.RefreshTTL),
		verificationJWTManager: verifier.NewJWTManager(cfg.JWT.VerificationSecret, cfg.JWT.VerificationTTL),
	}

	jsonManager := std_json.New()

	requestComponents := &requestComponents{
		logger:     logger,
		binder:     std_binder.New(jsonManager),
		validator:  plaground_validator.New(),
		httpWriter: json_writer.New(logger, jsonManager),
	}

	router := chi.NewMux()
	mux := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      router,
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	return &server{
		router:            router,
		mux:               mux,
		cfg:               cfg,
		dbPool:            dbPool,
		requestComponents: requestComponents,
		jWTManagers:       jWTManagers,
		services:          services,
	}
}

func (s *server) Run() error {
	middlewareFactory := middleware.NewMiddlewareFactory(s.requestComponents.binder, s.requestComponents.httpWriter)
	chiLogger := &chi_middleware.DefaultLogFormatter{
		Logger:  s.requestComponents.logger,
		NoColor: false,
	}

	s.router.Use(cors.AllowAll().Handler)
	s.router.Use(chi_middleware.Recoverer)
	s.router.Use(chi_middleware.RequestLogger(chiLogger))

	v1.MapV1Routes(v1.RouterArgs{
		AuthService:            s.services.authService,
		UserService:            s.services.userService,
		MiddlewareFactory:      middlewareFactory,
		Logger:                 s.requestComponents.logger,
		Binder:                 s.requestComponents.binder,
		Validator:              s.requestComponents.validator,
		Writer:                 s.requestComponents.httpWriter,
		AccessJWTManager:       s.jWTManagers.accessJWTManager,
		RefreshJWTManager:      s.jWTManagers.refreshJWTManager,
		VerificationJWTManager: s.jWTManagers.verificationJWTManager,
	}, s.router)

	return s.mux.ListenAndServe()
}

func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	err := s.mux.Shutdown(ctx)
	if err != nil {
		return err
	}

	err = postgres.Close(s.dbPool)
	if err != nil {
		return err
	}

	return nil
}
