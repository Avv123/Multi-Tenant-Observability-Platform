package quota

import (
	"context"
	"fmt"
	"time"

	goredis "github.com/go-redis/redis/v8"
)

type DailyCounter struct {
	client *goredis.Client
	prefix string
}

func NewDailyCounter(client *goredis.Client, prefix string) *DailyCounter {
	return &DailyCounter{
		client: client,
		prefix: prefix,
	}
}

func (q *DailyCounter) Reserve(ctx context.Context, key string, limit int64, amount int64) (bool, int64, error) {
	if amount <= 0 {
		return true, limit, nil
	}

	dayKey := fmt.Sprintf("%s:%s:%s", q.prefix, time.Now().UTC().Format("2006-01-02"), key)
	total, err := q.client.IncrBy(ctx, dayKey, amount).Result()
	if err != nil {
		return false, 0, err
	}

	if total == amount {
		untilTomorrow := time.Until(time.Now().UTC().Truncate(24 * time.Hour).Add(24 * time.Hour))
		if expireErr := q.client.Expire(ctx, dayKey, untilTomorrow).Err(); expireErr != nil {
			return false, 0, expireErr
		}
	}

	remaining := limit - total
	return total <= limit, remaining, nil
}
