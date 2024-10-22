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

	"github.com/hexley21/fixup/internal/catalog/delivery/http/v1"
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
)

type services struct {
	categoryTypes service.CategoryTypeService
	category      service.CategoryService
	subcategory   service.SubcategoryService
}

type jWTManagers struct {
	accessJWTManager auth_jwt.Manager
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

// NewServer initializes and returns a new server instance with the provided configuration and dependencies.
// It sets up repositories, services, JWT managers, handler components, and HTTP servers for both main and metrics endpoints.
func NewServer(
	cfg *config.Config,
	dbPool *pgxpool.Pool,
	logger logger.Logger,
	_ *snowflake.Node,
	validator validator.Validator,
) *server {
	categoryTypeRepository := repository.NewCategoryTypeRepository(dbPool)
	categoryRepository := repository.NewCategoryRepository(dbPool)
	subcategoryRepository := repository.NewSubcategoryRepository(dbPool)

	services := &services{
		categoryTypes: service.NewCategoryTypeService(categoryTypeRepository),
		category:      service.NewCategoryService(categoryRepository),
		subcategory:   service.NewSubcategoryService(subcategoryRepository),
	}

	jWTManagers := &jWTManagers{
		accessJWTManager: auth_jwt.NewManager(cfg.JWT.AccessSecret, cfg.JWT.AccessTTL),
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
	}
}

func (s *server) Run() error {
	// Initialize middleware with binder and writer components
	Middleware := middleware.NewMiddleware(s.handlerComponents.Binder, s.handlerComponents.Writer)

	// Set up logging middleware for chi router
	chiLogger := &chi_middleware.DefaultLogFormatter{
		Logger:  s.handlerComponents.Logger,
		NoColor: false,
	}

	// TODO: Add CSRF middleware
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
		CategoryService:     s.services.category,
		SubcategoryService:  s.services.subcategory,
		Middleware:          Middleware,
		HandlerComponents:   s.handlerComponents,
		AccessJWTManager:    s.jWTManagers.accessJWTManager,
		PaginationConfig:    &s.cfg.Pagination,
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

// Close gracefully shuts down the server, including its HTTP mux, metrics mux and database pool.
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
	}

	return nil
}
