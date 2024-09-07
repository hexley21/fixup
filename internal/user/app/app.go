package app

import (
	"context"
	"fmt"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	v1_http "github.com/hexley21/fixup/internal/user/delivery/http/v1"
	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/hexley21/fixup/internal/user/service"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/encryption"
	"github.com/hexley21/fixup/pkg/hasher"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/hexley21/fixup/pkg/rest"
)

type services struct {
	authService service.AuthService
	userService service.UserService
}

type server struct {
	cfg           *config.Config
	dbPool        *pgxpool.Pool
	echo          *echo.Echo
	snowflakeNode *snowflake.Node
	hasher        hasher.Hasher
	encryptor     encryption.Encryptor
	services      services
}

func NewServer(
	cfg *config.Config,
	logger echo.Logger,
	validator echo.Validator,
	dbPool *pgxpool.Pool,
	s3Bucket s3.Bucket,
	cdnFileInvalidator cdn.FileInvalidator,
	snowflakeNode *snowflake.Node,
	hasher hasher.Hasher,
	encryptor encryption.Encryptor,
	mailer mailer.Mailer,
	emailAddress string,
) *server {
	cloudFrontURLSigner := cdn.NewCloudFrontURLSigner(cfg.AWS.CDN)

	userRepository := repository.NewUserRepository(dbPool, snowflakeNode)
	providerRepository := repository.NewProviderRepository(dbPool)

	authService := service.NewAuthService(
		userRepository,
		providerRepository,
		dbPool,
		hasher,
		encryptor,
		mailer,
		emailAddress,
		cloudFrontURLSigner,
	)

	userService := service.NewUserService(
		userRepository,
		s3Bucket,
		cdnFileInvalidator,
		cloudFrontURLSigner,
	)

	e := echo.New()

	e.Logger = logger
	e.HTTPErrorHandler = httpErrorHandler
	e.Validator = validator

	return &server{
		cfg,
		dbPool,
		e,
		snowflakeNode,
		hasher,
		encryptor,
		services{
			authService: authService,
			userService: userService,
		},
	}
}

func (s *server) Run() error {
	s.echo.Use(middleware.Logger())
	s.echo.Use(middleware.Recover())
	s.echo.Use(middleware.CORS())

	v1_http.NewRouter(
		s.cfg.JWT,
		s.services.authService,
		s.services.userService,
	).MapV1Routes(s.echo)

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
	c.Logger().Error(err)
	if apiErr, ok := err.(rest.ErrorResponse); ok {
		c.JSON(apiErr.Status, apiErr)
		return
	}
	c.JSON(http.StatusInternalServerError, rest.NewInternalServerError(err))
}
