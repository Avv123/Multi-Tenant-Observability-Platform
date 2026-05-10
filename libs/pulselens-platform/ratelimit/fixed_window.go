package ratelimit

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type FixedWindow struct {
	client *goredis.Client
	window time.Duration
	limit  int64
	prefix string
}

func NewFixedWindow(client *goredis.Client, prefix string, limit int64, window time.Duration) *FixedWindow {
	return &FixedWindow{
		client: client,
		window: window,
		limit:  limit,
		prefix: prefix,
	}
}

func (r *FixedWindow) Allow(ctx context.Context, key string) (bool, int64, error) {
	return r.AllowN(ctx, key, 1)
}

func (r *FixedWindow) AllowN(ctx context.Context, key string, amount int64) (bool, int64, error) {
	if amount <= 0 {
		return true, r.limit, nil
	}

	redisKey := fmt.Sprintf("%s:%d:%s", r.prefix, time.Now().Unix()/int64(r.window.Seconds()), key)
	count, err := r.client.IncrBy(ctx, redisKey, amount).Result()
	if err != nil {
		return false, 0, err
	}

	if count == amount {
		if err = r.client.Expire(ctx, redisKey, r.window).Err(); err != nil {
			return false, 0, err
		}
	}

	remaining := r.limit - count
	return count <= r.limit, remaining, nil
}
