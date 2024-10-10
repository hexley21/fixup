package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/cmd/util/shutdown"
	"github.com/hexley21/fixup/internal/catalog/server"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/logger/zap_logger"
	"github.com/hexley21/fixup/pkg/validator/playground_validator"
)

// @title Catalog Microservice
// @version 1.0.0-alpha0
// @description Handles catalog operations
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:80
// @BasePath /v1
// @schemes http
//
// @securityDefinitions.apikey access_token
// @in header
// @name Authorization
func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	zapLogger := zap_logger.New(cfg.Logging, cfg.Server.IsProd)
	playgroundValidator := playground_validator.New()

	pgPool, err := postgres.NewPool(&cfg.Postgres)
	if err != nil {
		zapLogger.Fatal(err)
	}

	snowflakeNode, err := snowflake.NewNode(cfg.Server.InstanceId)
	if err != nil {
		zapLogger.Fatal(err)
	}

	catalogServer := server.NewServer(
		cfg,
		pgPool,
		zapLogger,
		snowflakeNode,
		playgroundValidator,
	)

	shutdownChan := make(chan struct{})
	go shutdown.NotifyShutdown(catalogServer, zapLogger, shutdownChan)

	log.Print("Catalog service started...")
	if !errors.Is(catalogServer.Run(), http.ErrServerClosed) {
		zapLogger.Fatal(err)
	}

	zapLogger.Info("Catalog service stopped...")
}
