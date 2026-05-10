package backpressure

import (
	"context"
	"fmt"
	"sort"

	"github.com/go-redis/redis/v8"
)

type SnapshotRow struct {
	Queue      string `json:"queue"`
	Pending    int64  `json:"pending"`
	Threshold  int64  `json:"threshold"`
	Overloaded bool   `json:"overloaded"`
}

type Controller struct {
	client *redis.Client
	prefix string
}

func New(client *redis.Client, prefix string) *Controller {
	return &Controller{client: client, prefix: prefix}
}

func (c *Controller) Reserve(ctx context.Context, queue string, limit int64, amount int64) (bool, int64, error) {
	if c == nil || c.client == nil {
		return true, limit, nil
	}
	if amount <= 0 {
		return true, limit, nil
	}

	key := c.key(queue)
	current, err := c.client.IncrBy(ctx, key, amount).Result()
	if err != nil {
		return false, 0, err
	}
	if limit > 0 && current > limit {
		if _, releaseErr := c.client.DecrBy(ctx, key, amount).Result(); releaseErr != nil {
			return false, 0, releaseErr
		}
		return false, 0, nil
	}
	return true, maxInt64(limit-current, 0), nil
}

func (c *Controller) Release(ctx context.Context, queue string, amount int64) error {
	if c == nil || c.client == nil || amount <= 0 {
		return nil
	}

	key := c.key(queue)
	next, err := c.client.DecrBy(ctx, key, amount).Result()
	if err != nil {
		if err == redis.Nil {
			return nil
		}
		return err
	}
	if next >= 0 {
		return nil
	}
	return c.client.Set(ctx, key, 0, 0).Err()
}

func (c *Controller) Pending(ctx context.Context, queue string) (int64, error) {
	if c == nil || c.client == nil {
		return 0, nil
	}

	value, err := c.client.Get(ctx, c.key(queue)).Int64()
	if err == redis.Nil {
		return 0, nil
	}
	return value, err
}

func (c *Controller) Snapshot(ctx context.Context, thresholds map[string]int64) ([]SnapshotRow, error) {
	rows := make([]SnapshotRow, 0, len(thresholds))
	for queue, threshold := range thresholds {
		pending, err := c.Pending(ctx, queue)
		if err != nil {
			return nil, err
		}
		rows = append(rows, SnapshotRow{
			Queue:      queue,
			Pending:    pending,
			Threshold:  threshold,
			Overloaded: threshold > 0 && pending >= threshold,
		})
	}
	sort.Slice(rows, func(left, right int) bool {
		return rows[left].Queue < rows[right].Queue
	})
	return rows, nil
}

func (c *Controller) key(queue string) string {
	return fmt.Sprintf("%s:%s", c.prefix, queue)
}

func maxInt64(left int64, right int64) int64 {
	if left > right {
		return left
	}
	return right
}
