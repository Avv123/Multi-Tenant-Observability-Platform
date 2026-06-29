package redis

import (
	"github.com/go-redis/redis/v8"
	"github.com/Avv123/pulselens-platform/netutil"
)

func New(addr string, db int) *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: netutil.NormalizeHostPort(addr),
		DB:   db,
	})
}
