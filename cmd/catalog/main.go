package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/internal/common/rest"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	e := echo.New()

	e.Logger = logger.NewZapLogger(cfg.Logging, cfg.Server.IsProd)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello World")
	})

	e.HTTPErrorHandler = func(err error, c echo.Context) {
		if apiErr, ok := err.(rest.ErrorResponse); ok {
			c.JSON(apiErr.Status, apiErr)
			c.Logger().Error(err)
			return
		}
		c.Logger().Error(err)
		c.JSON(http.StatusInternalServerError, "Internal server error")
	}

	e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", cfg.Server.HttpPort)))
}
