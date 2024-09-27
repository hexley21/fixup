package repository_test

import (
	"context"
	"testing"
	"time"

	"github.com/hexley21/fixup/internal/user/repository"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

const (
	token    = "pGWYACFvicVbLH3A0VRjZrCFu2hcGF9d"
	tokenTTL = time.Hour
)

func TestIsTokenUsed_Used(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisClient := getRedisClient(t)
	defer setupRedisCleanup(t, redisClient, redisContainer)

	_, err := redisClient.Set(ctx, token, "", tokenTTL).Result()
	if err != nil {
		t.Fatalf("Failed to set token: %v", err)
	}

	repo := repository.NewVerificationRepository(redisClient)

	isUsed, err := repo.IsTokenUsed(ctx, token)
	assert.NoError(t, err)
	assert.True(t, isUsed)
}

func TestIsTokenUsed_NotUsed(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisClient := getRedisClient(t)
	defer setupRedisCleanup(t, redisClient, redisContainer)

	repo := repository.NewVerificationRepository(redisClient)

	isUsed, err := repo.IsTokenUsed(ctx, token)
	assert.NoError(t, err)
	assert.False(t, isUsed)
}

func TestSetTokenUsed_Success(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisClient := getRedisClient(t)
	defer setupRedisCleanup(t, redisClient, redisContainer)

	repo := repository.NewVerificationRepository(redisClient)

	err := repo.SetTokenUsed(ctx, token, tokenTTL)
	if assert.NoError(t, err) {
		val, err := redisClient.Get(ctx, token).Result()
		assert.NoError(t, err)
		assert.Equal(t, "", val)
	}
}

func TestSetTokenUsed_AlreadySet(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisClient := getRedisClient(t)
	defer setupRedisCleanup(t, redisClient, redisContainer)

	repo := repository.NewVerificationRepository(redisClient)

	_, err := redisClient.Set(ctx, token, "", tokenTTL).Result()
	if assert.NoError(t, err) {
		err = repo.SetTokenUsed(ctx, token, tokenTTL)
		assert.ErrorIs(t, redis.TxFailedErr, err)
	}
}

func TestSetTokenUsed_Expired(t *testing.T) {
	ctx := context.Background()
	redisContainer, redisClient := getRedisClient(t)
	defer setupRedisCleanup(t, redisClient, redisContainer)

	repo := repository.NewVerificationRepository(redisClient)

	_, err := redisClient.Set(ctx, token, "", time.Second).Result()
	time.Sleep(time.Second * 2)
	if assert.NoError(t, err) {
		err := repo.SetTokenUsed(ctx, token, tokenTTL)
		assert.NoError(t, err)
	}
}
