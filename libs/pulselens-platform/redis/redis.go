package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/omniful/pulselens-platform/netutil"
)

func New(addr string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: netutil.NormalizeHostPort(addr),
		DB:   db,
	})
}
