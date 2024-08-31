package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	v1_http "github.com/hexley21/handy/internal/user/delivery/http/v1"
	"github.com/hexley21/handy/internal/user/delivery/http/v1/jwt"
	"github.com/hexley21/handy/internal/user/repository"
	"github.com/hexley21/handy/internal/user/service"
	"github.com/hexley21/handy/internal/user/util"
	"github.com/hexley21/handy/pkg/config"
	"github.com/hexley21/handy/pkg/infra/postgres"
	"github.com/hexley21/handy/pkg/mailer"
	"github.com/hexley21/handy/pkg/rest"
)

type server struct {
	cfg           *config.Config
	dbPool        *pgxpool.Pool
	echo          *echo.Echo
	snowflakeNode *snowflake.Node
	hasher        util.Hasher
	authJwt       jwt.AuthJwt
	authService   service.AuthService
	userService   service.UserService
}

func NewServer(
	cfg *config.Config,
	logger echo.Logger,
	dbPool *pgxpool.Pool,
	snowflakeNode *snowflake.Node,
	hasher util.Hasher,
	mailer mailer.Mailer,
	emailAddress string,
) *server {
	userRepository := repository.NewUserRepository(dbPool, snowflakeNode)

	authService := service.NewAuthService(
		userRepository,
		dbPool,
		hasher,
		mailer,
		emailAddress,
	)

	e := echo.New()

	e.Logger = logger
	e.HTTPErrorHandler = httpErrorHandler

	return &server{
		cfg,
		dbPool,
		e,
		snowflakeNode,
		hasher,
		jwt.NewAuthJwtImpl(&cfg.JWT),
		authService,
		service.NewUserService(userRepository),
	}
}

func (s *server) Run() error {
	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.CORS())

	v1_http.NewRouter(s.authService, s.userService, s.authJwt).MapV1Routes(s.echo)

	s.echo.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})

	return s.echo.Start(fmt.Sprintf(":%d", s.cfg.Server.HttpPort))
}

func (s *server) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), s.cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := s.echo.Shutdown(ctx); err != nil {
		return err
	}

	if err := postgres.Close(s.dbPool); err != nil {
		return err
	}

	return nil
}

func httpErrorHandler(err error, c echo.Context) {
	if apiErr, ok := err.(rest.ErrorResponse); ok {
		c.JSON(apiErr.Status, apiErr)
		c.Logger().Error(err)
		return
	}
	// c.Logger().Error(err)
	c.JSON(http.StatusInternalServerError, rest.InternalServerError)
}
