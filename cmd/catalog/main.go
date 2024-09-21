package main

import (
	"log"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/cmd/util/shutdown"
	"github.com/hexley21/fixup/internal/catalog/app"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/validator"
)

// @title Catalog Microservice
// @version 1.0.0-alpha0
// @description Handles catalog operations
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:8080
// @BasePath /v1
// @schemes http
//
// @securityDefinitions.apikey access_token
// @in header
// @name Authorization
// @securityDefinitions.apikey refresh_token
// @in header
// @name Authorization
func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	zapLogger := logger.NewZapLogger(cfg.Logging, cfg.Server.IsProd)
	playgroundValidator := validator.NewValidator()

	pgPool, err := postgres.NewPool(&cfg.Postgres)
	if err != nil {
		zapLogger.Fatal(err)
	}

	snowflakeNode, err := snowflake.NewNode(cfg.Server.InstanceId)
	if err != nil {
		zapLogger.Fatal(err)
	}

	server := app.NewServer(
		cfg,
		zapLogger,
		playgroundValidator,
		pgPool,
		snowflakeNode,
		cfg.Mailer.User,
	)

	shutdownError := make(chan error)
	go shutdown.NotifyShutdown(server, zapLogger, shutdownError)

	if err := <-shutdownError; err != nil {
		zapLogger.Error(err)
	}

	zapLogger.Info("server stopped")
}
