package postgres

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/hexley21/handy/pkg/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNoConnection = errors.New("no connection")

func InitPool(cfg *config.Postgres) (*pgxpool.Pool, error) {
	var dataSourceName string
	dataSourceName = fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s",
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

func Close(pgPool *pgxpool.Pool) error {
	if pgPool == nil {
		return ErrNoConnection
	}

	pgPool.Close()
	log.Println("database was closed")

	return nil
}
