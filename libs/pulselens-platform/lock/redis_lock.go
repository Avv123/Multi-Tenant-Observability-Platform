package lock

import (
	"context"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type RedisLock struct {
	client *goredis.Client
}

const releaseIfOwnerScript = `
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`

func NewRedisLock(client *goredis.Client) *RedisLock {
	return &RedisLock{client: client}
}

func (r *RedisLock) Acquire(ctx context.Context, key string, value string, ttl time.Duration) (bool, error) {
	return r.client.SetNX(ctx, key, value, ttl).Result()
}

func (r *RedisLock) Release(ctx context.Context, key string) error {
	return r.client.Del(ctx, key).Err()
}

func (r *RedisLock) ReleaseOwner(ctx context.Context, key string, value string) (bool, error) {
	result, err := r.client.Eval(ctx, releaseIfOwnerScript, []string{key}, value).Int64()
	if err != nil {
		return false, err
	}
	return result == 1, nil
}
