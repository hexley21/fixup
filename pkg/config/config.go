package config

import (
	"math"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		IsProd   bool
		HTTP     HTTP
		Postgres Postgres
		Redis    Redis
		JWT      JWT
		Argon2   Argon2
		Logging  Logging
	}

	HTTP struct {
		Port int `yaml:"port"`
		IdleTimeout time.Duration `yaml:"idle_timeout"`
		ReadTimeout time.Duration `yaml:"read_timeout"`
		WriteTimeout time.Duration `yaml:"write_timeout"`
	}

	Postgres struct {
		Port              int    `yaml:"port"`
		Host              string `yaml:"host"`
		DBName            string `yaml:"db_name"`
		User              string
		Password          string
		SslMode           string
		MaxConns          int32 `yaml:"max-connections"`
		MinConns          int32 `yaml:"min-connections"`
		HealthCheckPeriod time.Duration `yaml:"healthcheck-period"`
		MaxConnLifetime   time.Duration `yaml:"max-conn-lifetime"`
		MaxConnIdleTime   time.Duration `yaml:"max-conn-idle-time"`
	}

	Redis struct {
		Port        int    `yaml:"port"`
		Host        string `yaml:"host"`
		DBName      int    `yaml:"db_name"`
		User        string
		Password    string
		SslMode     bool
		DialTimeout time.Duration
	}

	JWT struct {
		AccessSecret  string
		AccessTTL     time.Duration `yaml:"access_ttl"`
		RefreshSecret string
		RefreshTTL    time.Duration `yaml:"refresh_ttl"`
	}

	Argon2 struct {
		SaltLen    uint32 `yaml:"salt_len"`
		KeyLen     uint32 `yaml:"key_len"`
		Time       uint32 `yaml:"time"`
		Memory     uint32 `yaml:"memory"`
		Threads    uint8  `yaml:"threads"`
		Breakpoint int
	}

	Logging struct {
		LogLevel      string `yaml:"level"`
		CallerEnabled bool   `yaml:"caller_enabled"`
	}
)

func LoadConfig(configDir string) (*Config, error) {
	var cfg Config

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}

	return parseConfig(configDir, &cfg)
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
	if err := godotenv.Load(); err != nil {
		return err
	}

	cfg.JWT.AccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	cfg.JWT.RefreshSecret = os.Getenv("JWT_REFRESH_SECRET")

	cfg.Postgres.User = os.Getenv("POSTGRES_USER")
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Postgres.SslMode = os.Getenv("POSTGRES_SSL_MODE")

	cfg.Redis.User = os.Getenv("REDIS_USER")
	cfg.Redis.Password = os.Getenv("REDIS_PASSWORD")
	cfg.Redis.SslMode = os.Getenv("REDIS_SSL_MODE") == "true"

	cfg.IsProd = os.Getenv("IS_PROD") == "true"

	return nil
}
