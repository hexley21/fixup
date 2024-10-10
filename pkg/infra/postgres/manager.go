package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/hexley21/fixup/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var errNoConnection = errors.New("no connection")

func NewPool(cfg *config.Postgres) (*pgxpool.Pool, error) {
	dataSourceName := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.DBName,
		cfg.Password,
	)

	poolCfg, err := pgxpool.ParseConfig(dataSourceName)
	if err != nil {
		return nil, err
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	pgPool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		return nil, err
	}

	if err := pgPool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pgPool, nil
}

func Migrate(url string, migrationPath string) (*migrate.Migrate, error) {
	m, err := migrate.New(migrationPath, url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect migrator: %w", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return nil, fmt.Errorf("failed to migrate up: %w", err)
	}

	return m, nil
}

func Close(pgPool *pgxpool.Pool) error {
	if pgPool == nil {
		return errNoConnection
	}

	pgPool.Close()
	log.Println("database was closed")

	return nil
}
