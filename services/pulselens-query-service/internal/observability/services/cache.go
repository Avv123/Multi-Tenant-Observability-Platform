package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/omniful/pulselens-platform/cacheversion"
	"github.com/omniful/pulselens-platform/config"
	queryredis "github.com/omniful/pulselens-query-service/pkg/redis"
)

func cacheTTL() time.Duration {
	seconds := config.GetInt("cache.ttlSeconds")
	if seconds <= 0 {
		seconds = 5
	}
	return time.Duration(seconds) * time.Second
}

func cacheKey(ctx context.Context, prefix string, tenantID string, scope string, payload any) string {
	body, _ := json.Marshal(payload)
	sum := sha1.Sum(body)
	version := cacheversion.Current(ctx, queryredis.Get(), tenantID, scope)
	return fmt.Sprintf("querycache:%s:%s:%s", prefix, version, hex.EncodeToString(sum[:]))
}

func cachedValue[T any](ctx context.Context, key string, loader func() T) T {
	client := queryredis.Get()
	if client == nil {
		return loader()
	}

	if payload, err := client.Get(ctx, key).Bytes(); err == nil {
		var row T
		if json.Unmarshal(payload, &row) == nil {
			return row
		}
	}

	row := loader()
	if payload, err := json.Marshal(row); err == nil {
		_ = client.Set(ctx, key, payload, cacheTTL()).Err()
	}
	return row
}
