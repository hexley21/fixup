package repository_test

import (
	"context"
	"flag"
	"log"
	"os"
	"testing"

	"github.com/bwmarrin/snowflake"
	infra "github.com/hexley21/fixup/pkg/infra/postgres"
	pg_tt "github.com/hexley21/fixup/pkg/infra/postgres/testcontainer"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	redis_tt "github.com/testcontainers/testcontainers-go/modules/redis"
)

var (
	connURL = ""
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	migrationPath := flag.String("mp", "", "Migration Path")
	flag.Parse()
	if *migrationPath == "" {
		log.Print("Continuing without database migration")
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
	_, err := dbPool.Exec(ctx, "TRUNCATE TABLE users CASCADE")
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

func getRedisClient(t *testing.T) (*redis_tt.RedisContainer, *redis.Client) {
	ctx := context.Background()

	redisContainer, err := redis_tt.Run(ctx,
		"docker.io/redis:7",
		redis_tt.WithSnapshotting(10, 1),
		redis_tt.WithLogLevel(redis_tt.LogLevelVerbose),
	)

	if err != nil {
		t.Fatalf("Failed to start container: %v", err)
	}

	endpoint, err := redisContainer.Endpoint(ctx, "")
	if err != nil {
		t.Fatalf("Failed to get container endpoint: %v", err)
	}

	return redisContainer, redis.NewClient(&redis.Options{Addr: endpoint})
}

func setupRedisCleanup(t *testing.T, client *redis.Client, container *redis_tt.RedisContainer) {
	if err := client.Close(); err != nil {
		t.Fatalf("Failed to close client: %v", err)
	}

	if err := container.Terminate(context.Background()); err != nil {
		t.Fatalf("Failed to terminate container: %v", err)
	}
}

func getSnowflakeNode() *snowflake.Node {
	node, err := snowflake.NewNode(1)
	if err != nil {
		log.Fatalln("failed to get snowflake node:", err)
	}

	return node
}
