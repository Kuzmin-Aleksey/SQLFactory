package redis

import (
	"SQLFactory/pkg/failure"
	"context"
	"errors"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

type Adapter struct {
	client *redis.Client
}

func NewAdapter(client *redis.Client) *Adapter {
	return &Adapter{client}
}

func (c *Adapter) Set(ctx context.Context, key string, v int, ttl time.Duration) error {
	if err := c.client.Set(ctx, key, v, ttl).Err(); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (c *Adapter) Get(ctx context.Context, key string) (int, error) {
	res := c.client.Get(ctx, key)
	if err := res.Err(); err != nil {
		if errors.Is(err, redis.Nil) {
			return 0, nil
		}
		return 0, failure.NewInternalError(err)
	}

	v, _ := strconv.Atoi(res.Val())

	return v, nil
}

func (c *Adapter) Del(ctx context.Context, key string) error {
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return failure.NewInternalError(err)
	}
	return nil
}

func (c *Adapter) TTL(ctx context.Context, key string) time.Duration {
	return c.client.TTL(ctx, key).Val()
}
