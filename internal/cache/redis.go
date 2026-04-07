package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Del(ctx context.Context, key string) error
	Close() error
}

type cacheImpl struct {
	client *redis.Client
}

func NewCache(ctx context.Context, addr string) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &cacheImpl{client: client}, nil
}

func (c *cacheImpl) Get(ctx context.Context, key string) (string, error) {
	return c.client.Get(ctx, key).Result()
}

func (c *cacheImpl) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	return c.client.Set(ctx, key, value, ttl).Err()
}

func (c *cacheImpl) Del(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}

func (c *cacheImpl) Close() error {
	return c.client.Close()
}
