package redis

import (
	"context"
	"strings"

	"github.com/hexley21/fixup/pkg/config"
	"github.com/redis/go-redis/v9"
)

func NewClient(config *config.Redis) (*redis.ClusterClient, error) {
	client := redis.NewClusterClient(&redis.ClusterOptions{
		Addrs:          strings.Split(config.Addresses, ","),
		Password:       config.Password,
		ReadOnly:       true,
		RouteByLatency: true,
		MinIdleConns:   config.MinIdleConn,
		PoolSize:       config.PoolSize,
		ReadTimeout:    config.ReadTimeout,
		WriteTimeout:   config.WriteTimeout,
		PoolTimeout:    config.PoolTimeout,
	})
	ctx := context.Background()
	_, err := client.Ping(ctx).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	}
	return client, nil
}