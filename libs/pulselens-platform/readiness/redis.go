package readiness

import (
	"context"

	platformredis "github.com/omniful/pulselens-platform/redis"
)

func CheckRedis(ctx context.Context, addr string, db int) error {
	client := platformredis.New(addr, db)
	defer func() { _ = client.Close() }()
	return client.Ping(ctx).Err()
}
