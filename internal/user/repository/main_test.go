package repository_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/bwmarrin/snowflake"
	infra "github.com/hexley21/fixup/pkg/infra/postgres"
	"github.com/hexley21/fixup/pkg/infra/postgres/testcontainer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

var (
	connURL = ""
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	image, config := testcontainer.GetConfig()
	container, err := postgres.Run(ctx, image, config...)
	if err != nil {
		log.Fatalln("failed to load container:", err)
	}

	connURL, err = container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		log.Fatalln("failed to get database connection string:", err)
	}

	migrationPath := flag.String("mp", "", "Migration Path")
	flag.Parse()
	if *migrationPath == "" {
		log.Fatalln("migration path flag should not be empty")
	}

	migrate, err := infra.Migrate(connURL, "file://"+*migrationPath)
	if err != nil {
		log.Fatalln("failed to migrate db: ", err)
	}

	res := m.Run()

	migrate.Drop()

	os.Exit(res)
}

func setupDatabaseCleanup(t *testing.T, ctx context.Context, dbPool *pgxpool.Pool) {
	t.Cleanup(func() {
		_, err := dbPool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
		dbPool.Close()
		if err != nil {
			log.Fatalln("failed to cleanup database:", err)
		}
	})
}

func getDbPool(ctx context.Context) *pgxpool.Pool {
	pool, err := pgxpool.New(ctx, connURL)
	if err != nil {
		log.Fatalln("failed to get database pool:", err)
	}

	return pool
}

func getSnowflakeNode() *snowflake.Node {
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalln("failed to get snowflake node:", err)
	}

	return node
}
