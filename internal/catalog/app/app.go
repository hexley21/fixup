package app

import (
	"context"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"

	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/infra/postgres"
)

type server struct {
	cfg           *config.Config
	dbPool        *pgxpool.Pool
	echo          *echo.Echo
	snowflakeNode *snowflake.Node
}

func NewServer(
	cfg *config.Config,
	logger echo.Logger,
	validator echo.Validator,
	dbPool *pgxpool.Pool,
	snowflakeNode *snowflake.Node,
	emailAddress string,
) *server {
	e := echo.New()

	e.Logger = logger
	e.HTTPErrorHandler = httpErrorHandler
	e.Validator = validator

	return &server{
		cfg,
		dbPool,
		e,
		snowflakeNode,
	}
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
	if apiErr, ok := err.(*rest.ErrorResponse); ok {
		c.JSON(apiErr.Status, apiErr)
		return
	}
	c.JSON(http.StatusInternalServerError, rest.NewInternalServerError(err))
}
