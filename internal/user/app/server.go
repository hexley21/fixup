package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/go-chi/chi/v5"
	v1_http "github.com/hexley21/handy/internal/user/delivery/http/v1"
	"github.com/hexley21/handy/internal/user/repository"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/internal/user/util"
	"github.com/hexley21/handy/pkg/config"
	"github.com/hexley21/handy/pkg/infra/postgres"
	"github.com/hexley21/handy/pkg/logger"
	"github.com/jackc/pgx/v5/pgxpool"
)

type server struct {
	cfg           *config.Config
	logger        logger.Logger
	dbPool        *pgxpool.Pool
	http          *http.Server
	router        chi.Router
	snowflakeNode *snowflake.Node
	hasher        util.Hasher
	authService   service.AuthService
	userService   service.UserService
}

func NewServer(
	cfg *config.Config,
	logger logger.Logger,
	dbPool *pgxpool.Pool,
	snowflakeNode *snowflake.Node,
	hasher util.Hasher,
) *server {
	userRepository := repository.NewUserRepository(dbPool, snowflakeNode)

	emailService := service.NewEmailService(logger)

	authService := service.NewAuthService(
		userRepository,
		emailService,
		dbPool,
		hasher,
	)

	mux := chi.NewRouter()

	http := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.HTTP.Port),
		Handler:      mux,
		IdleTimeout:  cfg.Server.HTTP.IdleTimeout,
		ReadTimeout:  cfg.Server.HTTP.ReadTimeout,
		WriteTimeout: cfg.Server.HTTP.WriteTimeout,
	}

	return &server{
		cfg,
		logger,
		dbPool,
		http,
		mux,
		snowflakeNode,
		hasher,
		authService,
		service.NewUserService(userRepository),
	}
}

func (s *server) Run() error {
	v1_http.NewRouter(s.logger, s.authService, s.userService).MapV1Routes(s.router)

	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello World"))
	})

	s.logger.Info("Server is starting...")
	return s.http.ListenAndServe()
}

func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := s.http.Shutdown(ctx); err != nil {
		return err
	}

	if err := postgres.Close(s.dbPool); err != nil {
		return err
	}

	return nil
}
