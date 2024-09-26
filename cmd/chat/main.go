package main

import (
	"log"

	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/logger/zap_logger"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	zapLogger := zap_logger.New(cfg.Logging, cfg.Server.IsProd)
	zapLogger.Debug("")
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
