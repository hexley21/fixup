package repository

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type VerificationRepository interface {
    IsTokenUsed(ctx context.Context, token string) (bool, error)
    SetTokenUsed(ctx context.Context, token string, ttl time.Duration) error
}

type verificationRepositoryImpl struct {
	redis redis.UniversalClient
}

func NewVerificationRepository(redis redis.UniversalClient) *verificationRepositoryImpl {
	return &verificationRepositoryImpl{
		redis: redis,
	}
}

func (r *verificationRepositoryImpl) IsTokenUsed(ctx context.Context, token string) (bool, error) {
    exists, err := r.redis.Exists(ctx, token).Result()
    if err != nil {
        return false, err
    }

    return exists == 1, nil
}

func (r *verificationRepositoryImpl) SetTokenUsed(ctx context.Context, token string, ttl time.Duration) error {
    success, err := r.redis.SetNX(ctx, token, "", ttl).Result()
    if err != nil {
        return err
    }

    if !success {
        return redis.TxFailedErr
    }

    return nil
}
