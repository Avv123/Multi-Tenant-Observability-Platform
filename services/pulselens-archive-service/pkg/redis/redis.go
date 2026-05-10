package redis

import goredis "github.com/go-redis/redis/v8"

var client *goredis.Client

func Set(db *goredis.Client) {
	client = db
}

func Get() *goredis.Client {
	return client
}
