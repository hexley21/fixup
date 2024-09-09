package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/cmd/util/shutdown"
	"github.com/hexley21/fixup/internal/user/app"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/encryption/aes"
	"github.com/hexley21/fixup/pkg/hasher/argon2"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/hexley21/fixup/pkg/logger"
	"github.com/hexley21/fixup/pkg/mailer/gomail"
	"github.com/hexley21/fixup/pkg/validator"
)

// @title User Microservice
// @version 1.0.0-alpha0
// @description Handles user and authentication operations
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

	awsS3Bucket, err := s3.NewClient(cfg.AWS.AWSCfg, cfg.AWS.S3)
	if err != nil {
		zapLogger.Fatal(err)
	}

	awsCloudFrontCdn, err := cdn.NewClient(cfg.AWS.AWSCfg, cfg.AWS.CDN)
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

	server := app.NewServer(
		cfg,
		zapLogger,
		playgroundValidator,
		pgPool,
		awsS3Bucket,
		awsCloudFrontCdn,
		snowflakeNode,
		argon2Hasher,
		aesEncryption,
		goMailer,
		cfg.Mailer.User,
	)

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
