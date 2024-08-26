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

func InitPool(cfg *config.Postgres) *pgxpool.Pool {
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
		log.Fatalf("could not parse config: %v\n", err)
	}

	poolCfg.MaxConns = cfg.MaxConns
	poolCfg.MinConns = cfg.MinConns
	poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	poolCfg.HealthCheckPeriod = cfg.HealthCheckPeriod

	connPool, err := pgxpool.NewWithConfig(context.Background(), poolCfg)
	if err != nil {
		log.Fatalf("unable to create connection pool: %v\n", err)
	}

	if err := connPool.Ping(context.Background()); err != nil {
		log.Fatalf("database is offline: %v\n", err)
	}

	return connPool
}

func Close(connPool *pgxpool.Pool) error {
	if connPool == nil {
		return ErrNoConnection
	}

	connPool.Close()
	log.Println("database was closed")

	return nil
}
