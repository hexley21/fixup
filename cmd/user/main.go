package main

import (
	"log"

	"github.com/hexley21/handy/pkg/config"
	"github.com/hexley21/handy/pkg/logger"
	"github.com/hexley21/handy/pkg/logger/zap"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	var zapLogger logger.Logger = zap.InitLogger(cfg.Logging, cfg.IsProd)

	zapLogger.Debug("Hello World")
}
