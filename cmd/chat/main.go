package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/logger/zap_logger"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	zapLogger := zap_logger.New(cfg.Logging, cfg.Server.IsProd)
	
	mux := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.HTTP.Port),
		Handler:      chi.NewMux(),
		IdleTimeout:  cfg.HTTP.IdleTimeout,
		ReadTimeout:  cfg.HTTP.ReadTimeout,
		WriteTimeout: cfg.HTTP.WriteTimeout,
	}

	zapLogger.Debug("Server Started")
	if !errors.Is(mux.ListenAndServe(), http.ErrServerClosed) {
		zapLogger.Fatal(err)
	}

	zapLogger.Info("Server Stopped")
	// playgroundValidator := playground_validator.New()

	// pgPool, err := postgres.NewPool(&cfg.Postgres)
	// if err != nil {
	// 	zapLogger.Fatal(err)
	// }

	// snowflakeNode, err := snowflake.NewNode(cfg.Server.InstanceId)
	// if err != nil {
	// 	zapLogger.Fatal(err)
	// }

	
}
