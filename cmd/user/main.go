package main

import (
	"errors"
	"log"
	"net/http"

	"github.com/bwmarrin/snowflake"
	"github.com/hexley21/fixup/cmd/util/shutdown"
	"github.com/hexley21/fixup/internal/user/server"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/hexley21/fixup/pkg/encryption/aes"
	"github.com/hexley21/fixup/pkg/hasher/argon2"
	"github.com/hexley21/fixup/pkg/infra/cdn"
	"github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/redis"
	"github.com/hexley21/fixup/pkg/infra/s3"
	"github.com/hexley21/fixup/pkg/logger/zap_logger"
	"github.com/hexley21/fixup/pkg/mailer"
	"github.com/hexley21/fixup/pkg/mailer/gomail"
	"github.com/hexley21/fixup/pkg/validator/playground_validator"
)

// @title User Microservice
// @version 1.0.0-alpha0
// @description Handles user and authentication operations
// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html
// @host localhost:80
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
		log.Fatalf("Could not load config: %v\n", err)
	}

	zapLogger := zap_logger.New(cfg.Logging, cfg.Server.IsProd)
	playgroundValidator := playground_validator.New()

	pgPool, err := postgres.NewPool(&cfg.Postgres)
	if err != nil {
		zapLogger.Fatal(err)
	}

	redisCluster, err := redis.NewClient(&cfg.Redis)
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

	var goMailer mailer.Mailer
	if cfg.Server.IsProd {
		goMailer = gomail.New(&cfg.Mailer)
	} else {
		goMailer = gomail.NewDev(&cfg.Mailer)
	}
	argon2Hasher := argon2.NewHasher(cfg.Argon2)
	aesEncryption := aes.NewAesEncryptor(cfg.AesEncryptor.Key)

	userServer := server.NewServer(
		cfg,
		pgPool,
		redisCluster,
		zapLogger,
		snowflakeNode,
		playgroundValidator,
		awsS3Bucket,
		awsCloudFrontCdn,
		argon2Hasher,
		aesEncryption,
		goMailer,
	)

	shutdownChan := make(chan struct{})
	go shutdown.NotifyShutdown(userServer, zapLogger, shutdownChan)

	log.Print("User service started...")
	if !errors.Is(userServer.Run(), http.ErrServerClosed) {
		zapLogger.Fatal(err)
	}

	zapLogger.Info("User service stopped...")
}
