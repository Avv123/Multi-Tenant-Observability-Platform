package redis

import (
	"sync"

	goredis "github.com/go-redis/redis/v8"
)

var (
	client *goredis.Client
	once   sync.Once
)

func Set(c *goredis.Client) {
	once.Do(func() {
		client = c
	})
}

func Get() *goredis.Client {
	return client
}
