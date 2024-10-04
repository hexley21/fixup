package config

import (
	"context"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		Server       Server
		HTTP         HTTP
		Postgres     Postgres
		Redis        Redis
		AWS          AWS
		JWT          JWT
		Argon2       Argon2
		AesEncryptor AesEncryptor
		Mailer       Mailer
		Logging      Logging
	}

	Server struct {
		ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
		InstanceId      int64         `yaml:"instance_id"`
		Email           string
		IsProd          bool
	}

	HTTP struct {
		Port         int           `yaml:"port"`
		CorsOrigins  string        `yaml:"cors_origins"`
		IdleTimeout  time.Duration `yaml:"idle_timeout"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
	}

	Postgres struct {
		Port              int    `yaml:"port"`
		Host              string `yaml:"host"`
		DBName            string `yaml:"db_name"`
		User              string
		Password          string
		SslMode           string
		MaxConns          int32         `yaml:"max-connections"`
		MinConns          int32         `yaml:"min-connections"`
		HealthCheckPeriod time.Duration `yaml:"healthcheck-period"`
		MaxConnLifetime   time.Duration `yaml:"max-conn-lifetime"`
		MaxConnIdleTime   time.Duration `yaml:"max-conn-idle-time"`
	}

	Redis struct {
		Password     string
		Addresses    string        `yaml:"addresses"`
		MinIdleConn  int           `yaml:"min_idle_conn"`
		PoolSize     int           `yaml:"pool_size"`
		ReadTimeout  time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
		PoolTimeout  time.Duration `yaml:"pool_timeout"`
	}

	AWS struct {
		AWSCfg AWSCfg
		S3     S3
		CDN    CDN
	}

	AWSCfg struct {
		Region          string `yaml:"region"`
		AccessKeyID     string
		SecretAccessKey string
	}

	S3 struct {
		Bucket         string `yaml:"bucket"`
		RandomNameSize int    `yaml:"random_name_size"`
	}

	CDN struct {
		UrlFmt         string        `yaml:"url_fmt"`
		Expiry         time.Duration `yaml:"expiry"`
		DistributionId string
		PrivateKey     *rsa.PrivateKey
		KeyPairId      string
	}

	JWT struct {
		AccessSecret       string
		AccessTTL          time.Duration `yaml:"access_ttl"`
		RefreshSecret      string
		RefreshTTL         time.Duration `yaml:"refresh_ttl"`
		VerificationSecret string
		VerificationTTL    time.Duration `yaml:"verification_ttl"`
	}

	Mailer struct {
		Host     string
		Port     int
		User     string
		Password string
	}

	Argon2 struct {
		SaltLen    uint32 `yaml:"salt_len"`
		KeyLen     uint32 `yaml:"key_len"`
		Time       uint32 `yaml:"time"`
		Memory     uint32 `yaml:"memory"`
		Threads    uint8  `yaml:"threads"`
		Breakpoint int
	}

	AesEncryptor struct {
		Key string
	}

	Logging struct {
		LogLevel      string `yaml:"level"`
		CallerEnabled bool   `yaml:"caller_enabled"`
		LogFile       string `yaml:"log_file"`
	}
)

func LoadConfig(configDir string) (*Config, error) {
	var cfg Config

	if err := parseKeys(&cfg); err != nil {
		return nil, err
	}

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}

	return parseConfig(configDir, &cfg)
}

func (cfg AWSCfg) LoadDefaultConfig(ctx context.Context) (aws.Config, error) {
	return config.LoadDefaultConfig(
		ctx,
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				cfg.AccessKeyID,
				cfg.SecretAccessKey,
				"",
			)),
		config.WithRegion(cfg.Region),
	)
}

func parseConfig(configDir string, cfg *Config) (*Config, error) {
	yamlFile, err := os.ReadFile(configDir)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(yamlFile, &cfg); err != nil {
		return nil, err
	}

	cfg.Argon2.Breakpoint = int(math.Ceil(79.0 * 4.0 / 3.0))

	return cfg, nil
}

func parseEnv(cfg *Config) error {
	err := godotenv.Load()
	if err != nil {
		return err
	}

	cfg.JWT.AccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	cfg.JWT.RefreshSecret = os.Getenv("JWT_REFRESH_SECRET")
	cfg.JWT.VerificationSecret = os.Getenv("JWT_VERIFICATION_SECRET")

	cfg.Postgres.User = os.Getenv("POSTGRES_USER")
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Postgres.SslMode = os.Getenv("POSTGRES_SSL_MODE")

	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")

	cfg.AWS.AWSCfg.AccessKeyID = os.Getenv("AWS_AC_ID")
	cfg.AWS.AWSCfg.SecretAccessKey = os.Getenv("AWS_SECRET_AC")

	cfg.AWS.CDN.KeyPairId = os.Getenv("CDN_KP_ID")
	cfg.AWS.CDN.DistributionId = os.Getenv("CDN_DISTRIBUTION_ID")

	cfg.Server.IsProd = os.Getenv("IS_PROD") == "true"

	cfg.Mailer.Host = os.Getenv("SMTP_HOST")
	cfg.Mailer.Port, err = strconv.Atoi(os.Getenv("SMTP_PORT"))
	if err != nil {
		return err
	}

	cfg.Mailer.User = os.Getenv("SMTP_USER")
	cfg.Server.Email = cfg.Mailer.User
	cfg.Mailer.Password = os.Getenv("SMTP_PASSWORD")

	cfg.AesEncryptor.Key = os.Getenv("DATA_ENCRYPTION_KEY")

	return nil
}

func parseKeys(cfg *Config) error {
	pkFile, err := os.ReadFile("./keys/cdn/private_key.pem")
	if err != nil {
		return err
	}

	block, _ := pem.Decode(pkFile)
	if block == nil {
		return errors.New("failed to decode PEM block")
	}

	if block.Type != "PRIVATE KEY" {
		return fmt.Errorf("unsupported block type: %s", block.Type)
	}

	privateKey, err := x509.ParsePKCS8PrivateKey(block.Bytes)
	if err != nil {
		return fmt.Errorf("error parsing PKCS#8 private key: %w", err)
	}

	rsaKey, ok := privateKey.(*rsa.PrivateKey)
	if !ok {
		return errors.New("private key is not an RSA key")
	}

	cfg.AWS.CDN.PrivateKey = rsaKey

	return nil
}
