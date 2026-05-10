package cache

import (
	"sync"

	goredis "github.com/go-redis/redis/v8"
)

var (
	client *goredis.Client
	once   sync.Once
)

func Set(redisClient *goredis.Client) {
	once.Do(func() {
		client = redisClient
	})
}

func Get() *goredis.Client {
	return client
}
