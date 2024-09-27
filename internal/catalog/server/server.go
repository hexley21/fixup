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
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/http/binder"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/writer"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/validator"
	"github.com/hexley21/fixup/pkg/validator/playground_validator"
)

type services struct {
}

type jWTManagers struct {
	accessJWTManager       auth_jwt.JWTManager
	refreshJWTManager      auth_jwt.JWTManager
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
) *server {
	// cdnURLSigner := cdn.NewCloudFrontURLSigner(cfg.AWS.CDN)

	services := &services{}

	jWTManagers := &jWTManagers{
		accessJWTManager:       auth_jwt.NewJWTManager(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL),
		refreshJWTManager:      auth_jwt.NewJWTManager(cfg.JWT.RefreshSecret, cfg.JWT.RefreshTTL),
	}

	jsonManager := std_json.New()

	requestComponents := &requestComponents{
		logger:     logger,
		binder:     std_binder.New(jsonManager),
		validator:  playground_validator.New(),
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
	// middlewareFactory := middleware.NewMiddlewareFactory(s.requestComponents.binder, s.requestComponents.httpWriter)
	chiLogger := &chi_middleware.DefaultLogFormatter{
		Logger:  s.requestComponents.logger,
		NoColor: false,
	}

	s.router.Use(cors.AllowAll().Handler)
	s.router.Use(chi_middleware.Recoverer)
	s.router.Use(chi_middleware.RequestLogger(chiLogger))

	return nil
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
