package redis

import "github.com/go-redis/redis/v8"

func New(addr string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: addr,
		DB:   db,
	})
}
