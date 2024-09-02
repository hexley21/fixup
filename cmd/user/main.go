package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/handy/cmd/util/shutdown"
	"github.com/hexley21/handy/internal/user/app"
	"github.com/hexley21/handy/pkg/config"
	"github.com/hexley21/handy/pkg/encryption/aes"
	"github.com/hexley21/handy/pkg/hasher/argon2"
	"github.com/hexley21/handy/pkg/infra/postgres"
	"github.com/hexley21/handy/pkg/logger"
	"github.com/hexley21/handy/pkg/mailer/gomail"
)

func main() {
	cfg, err := config.LoadConfig("./config/config.yml")
	if err != nil {
		log.Fatalf("could not load config: %v\n", err)
	}

	zapLogger := logger.NewZapLogger(cfg.Logging, cfg.Server.IsProd)

	pgPool, err := postgres.InitPool(&cfg.Postgres)
	if err != nil {
		zapLogger.Fatal(err)
	}

	snowflakeNode, err := snowflake.NewNode(cfg.Server.InstanceId)
	if err != nil {
		zapLogger.Fatal(err)
	}

	goMailer := gomail.NewGoMailer(&cfg.Mailer)
	argon2Hasher := argon2.NewHasher(cfg.Argon2)
	aesEncryption := aes.NewAesEncryptor(cfg.AesEncryptor.Key)
	
	server := app.NewServer(cfg, zapLogger, pgPool, snowflakeNode, argon2Hasher, aesEncryption, goMailer, cfg.Mailer.User)

	shutdownError := make(chan error)
	go shutdown.NotifyShutdown(server, zapLogger, shutdownError)

	err = server.Run()
	if !errors.Is(err, http.ErrServerClosed) {
		zapLogger.Error(err)
	}

	if err := <-shutdownError; err != nil {
		zapLogger.Error(err)
	}

	zapLogger.Info("server stopped")
}
