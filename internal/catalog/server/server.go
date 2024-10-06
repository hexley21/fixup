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

	v1 "github.com/hexley21/fixup/internal/catalog/delivery/http/v1"
	"github.com/hexley21/fixup/internal/catalog/repository"
	"github.com/hexley21/fixup/internal/catalog/service"
	"github.com/hexley21/fixup/internal/common/auth_jwt"
	"github.com/hexley21/fixup/internal/common/middleware"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/http/binder/std_binder"
	"github.com/hexley21/fixup/pkg/http/handler"
	"github.com/hexley21/fixup/pkg/http/json/std_json"
	"github.com/hexley21/fixup/pkg/http/writer/json_writer"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/validator"
	"github.com/hexley21/fixup/pkg/validator/playground_validator"
)

type services struct {
	categoryTypes service.CategoryTypeService
}

type jWTManagers struct {
	accessJWTManager auth_jwt.JWTManager
}

type server struct {
	router            chi.Router
	metricsRouter     chi.Router
	mux               *http.Server
	metricsMux        *http.Server
	cfg               *config.Config
	dbPool            *pgxpool.Pool
	handlerComponents *handler.Components
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
	categoryTypeRepository := repository.NewCategoryTypeRepository(dbPool)

	services := &services{
		categoryTypes: service.NewCategoryTypeService(categoryTypeRepository),
	}

	jWTManagers := &jWTManagers{
		accessJWTManager: auth_jwt.NewJWTManager(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL),
	}

	jsonManager := std_json.New()
	handlerComponents := &handler.Components{
		Logger:    logger,
		Binder:    std_binder.New(jsonManager),
		Validator: playground_validator.New(),
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
	}
}

func (s *server) Run() error {
	middlewareFactory := middleware.NewMiddlewareFactory(s.handlerComponents.Binder, s.handlerComponents.Writer)
	chiLogger := &chi_middleware.DefaultLogFormatter{
		Logger:  s.handlerComponents.Logger,
		NoColor: false,
	}

	s.router.Use(cors.Handler(cors.Options{
		AllowedOrigins:   strings.Split(s.cfg.HTTP.CorsOrigins, ","),
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Idempotency-Key", "X-CSRF-Token"},
		AllowCredentials: true,
		MaxAge:           300,
	}))
	s.router.Use(chi_middleware.Recoverer)
	s.router.Use(chi_middleware.RequestLogger(chiLogger))

	v1.MapV1Routes(v1.RouterArgs{
		CategoryTypeService: s.services.categoryTypes,
		MiddlewareFactory:   middlewareFactory,
		HandlerComponents:   s.handlerComponents,
		AccessJWTManager:    s.jWTManagers.accessJWTManager,
	}, s.router)

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
	}

	return nil
}
