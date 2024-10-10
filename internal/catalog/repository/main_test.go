package repository_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	infra "github.com/hexley21/fixup/pkg/infra/postgres"
	pg_tt "github.com/hexley21/fixup/pkg/infra/postgres/testcontainer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	connURL = ""
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	migrationPath := flag.String("mp", "", "Migration Path")
	flag.Parse()
	if *migrationPath == "" {
		log.Print("Continueing without database migration")
		os.Exit(0)
	}

	image, config := pg_tt.GetConfig()
	container, err := postgres.Run(ctx, image, config...)
	if err != nil {
		log.Fatalln("failed to load container:", err)
	}

	connURL, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalln("failed to get database connection string:", err)
	}

	migrate, err := infra.Migrate(connURL, "file://"+*migrationPath)
	if err != nil {
		log.Fatalln("failed to migrate db: ", err)
	}

	res := m.Run()

	migrate.Drop()

	os.Exit(res)
}

func cleanupPostgres(ctx context.Context, dbPool *pgxpool.Pool) {
	_, err := dbPool.Exec(ctx, "TRUNCATE TABLE category_types CASCADE")
	dbPool.Close()
	if err != nil {
		log.Fatalln("failed to cleanup database:", err)
	}
}

func getPgPool(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		log.Fatalln("failed to get database pool:", err)
	}

	return pool
}
