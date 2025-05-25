package cache

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/hoangdv99/morgana/internal/configs"
	"github.com/hoangdv99/morgana/internal/utils"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Client interface {
	Set(ctx context.Context, key string, data any, ttl time.Duration) error
	Get(ctx context.Context, key string) (any, error)
	AddToSet(ctx context.Context, key string, data ...any) error
	IsDataInSet(ctx context.Context, key string, data any) (bool, error)
}

type client struct {
	redisClient *redis.Client
	logger      *zap.Logger
}

func NewClient(
	cacheConfig configs.Cache,
	logger *zap.Logger,
) Client {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cacheConfig.Address,
		Username: cacheConfig.Username,
		Password: cacheConfig.Password,
	})

	return &client{
		redisClient: redisClient,
		logger:      logger,
	}
}

func (c client) AddToSet(ctx context.Context, key string, data ...any) error {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key)).
		With(zap.Any("data", data))

	if err := c.redisClient.SAdd(ctx, key, data...).Err(); err != nil {
		logger.With(zap.Error(err)).Error("failed to set data into set inside cache")
		return err
	}

	return nil
}

func (c client) Get(ctx context.Context, key string) (any, error) {
	logger := utils.LoggerWithContext(ctx, c.logger).With(zap.String("key", key))
	data, err := c.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, ErrCacheMiss
		}
		logger.With(zap.Error(err)).Error("failed to get data from cache")
		return nil, status.Errorf(codes.Internal, "failed to get data from cache: %+v", err)
	}

	return data, nil
}

func (c client) IsDataInSet(ctx context.Context, key string, data any) (bool, error) {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key)).
		With(zap.Any("data", data))

	result, err := c.redisClient.SIsMember(ctx, key, data).Result()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to check if data is member of set inside cache")
		return false, err
	}

	return result, nil
}

func (c client) Set(ctx context.Context, key string, data any, ttl time.Duration) error {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("key", key)).
		With(zap.Any("data", data)).
		With(zap.Duration("ttl", ttl))

	err := c.redisClient.Set(ctx, key, data, ttl).Err()
	if err != nil {
		logger.With(zap.Error(err)).Error("failed to set data in cache")
		return status.Errorf(codes.Internal, "failed to set data in cache: %v", err)
	}

	return nil
}
