package runtime

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/omniful/pulselens-platform/idgen"
)

type Heartbeat struct {
	ServiceName string            `json:"service_name"`
	InstanceID  string            `json:"instance_id"`
	Mode        string            `json:"mode"`
	Port        string            `json:"port"`
	PID         int               `json:"pid"`
	StartedAt   time.Time         `json:"started_at"`
	LastSeenAt  time.Time         `json:"last_seen_at"`
	TTLSeconds  int               `json:"ttl_seconds"`
	Status      string            `json:"status"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type HeartbeatOptions struct {
	ServiceName string
	Mode        string
	Port        string
	Metadata    map[string]string
	Interval    time.Duration
	TTL         time.Duration
}

func Start(ctx context.Context, client *redis.Client, options HeartbeatOptions) {
	if client == nil || options.ServiceName == "" {
		return
	}

	interval := options.Interval
	if interval <= 0 {
		interval = 5 * time.Second
	}
	ttl := options.TTL
	if ttl <= interval {
		ttl = interval * 3
	}

	instanceID := idgen.New("instance")
	startedAt := time.Now().UTC()
	emit := func() {
		row := Heartbeat{
			ServiceName: options.ServiceName,
			InstanceID:  instanceID,
			Mode:        options.Mode,
			Port:        options.Port,
			PID:         os.Getpid(),
			StartedAt:   startedAt,
			LastSeenAt:  time.Now().UTC(),
			TTLSeconds:  int(ttl / time.Second),
			Metadata:    options.Metadata,
		}
		payload, err := json.Marshal(row)
		if err != nil {
			return
		}
		_ = client.Set(ctx, heartbeatKey(instanceID), payload, ttl).Err()
	}

	emit()
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				emit()
			}
		}
	}()
}

func List(ctx context.Context, client *redis.Client) ([]Heartbeat, error) {
	if client == nil {
		return nil, nil
	}

	keys := make([]string, 0)
	var cursor uint64
	for {
		batch, nextCursor, err := client.Scan(ctx, cursor, heartbeatKey("*"), 50).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, batch...)
		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	if len(keys) == 0 {
		return []Heartbeat{}, nil
	}

	rows := make([]Heartbeat, 0, len(keys))
	now := time.Now().UTC()
	for _, key := range keys {
		payload, err := client.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var row Heartbeat
		if err = json.Unmarshal([]byte(payload), &row); err != nil {
			continue
		}
		row.Status = statusFor(row.LastSeenAt, row.TTLSeconds, now)
		rows = append(rows, row)
	}

	sort.Slice(rows, func(left, right int) bool {
		if rows[left].ServiceName == rows[right].ServiceName {
			return rows[left].InstanceID < rows[right].InstanceID
		}
		return rows[left].ServiceName < rows[right].ServiceName
	})
	return rows, nil
}

func heartbeatKey(instanceID string) string {
	return fmt.Sprintf("runtime:heartbeat:%s", instanceID)
}

func statusFor(lastSeenAt time.Time, ttlSeconds int, now time.Time) string {
	if ttlSeconds <= 0 {
		ttlSeconds = 15
	}
	age := now.Sub(lastSeenAt)
	ttl := time.Duration(ttlSeconds) * time.Second
	switch {
	case age > ttl:
		return "down"
	case age > ttl/2:
		return "degraded"
	default:
		return "healthy"
	}
}
