package config

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/joho/godotenv"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		HTTP     HTTP
		Postgres Postgres
		Auth     Auth
	}

	HTTP struct {
		Port    int      `yaml:"port"`
		Origins []string `yaml:"origins"`
	}

	Postgres struct {
		Port     int    `yaml:"port"`
		Host     string `yaml:"host"`
		Db       string
		User     string
		Password string
	}

	Auth struct {
		JWT    JWT
		Hasher Argon2
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
)

func Init(configDir string) (*Config, error) {
	var cfg Config

	if err := parseEnv(&cfg); err != nil {
		return nil, err
	}

	return parseConfig(configDir, &cfg)
}

func parseConfig(configDir string, cfg *Config) (*Config, error) {
	yamlFile, err := os.Open(configDir)
	if err != nil {
		return nil, err
	}

	defer yamlFile.Close()

	decorder := yaml.NewDecoder(yamlFile)

	if err = decorder.Decode(&cfg); err != nil {
		return nil, err
	}

	cfg.Auth.Hasher.Breakpoint = int(math.Ceil(79.0 * 4.0 / 3.0))

	return cfg, nil
}

func parseEnv(cfg *Config) error {
	if err := godotenv.Load(); err != nil {
		return err
	}

	cfg.Auth.JWT.AccessSecret = os.Getenv("JWT_ACCESS_SECRET")
	cfg.Auth.JWT.RefreshSecret = os.Getenv("JWT_REFRESH_SECRET")

	cfg.Postgres.User = os.Getenv("POSTGRES_USER")
	cfg.Postgres.Password = os.Getenv("POSTGRES_PASSWORD")
	cfg.Postgres.Db = os.Getenv("POSTGRES_DB")

	return nil
}

func GetDbSource(cfg *Config) string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Postgres.User,
		cfg.Postgres.Password,
		cfg.Postgres.Host,
		cfg.Postgres.Port,
		cfg.Postgres.Db,
	)
}
