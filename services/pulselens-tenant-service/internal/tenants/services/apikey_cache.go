package services

// B1: Redis read-through cache for ResolvedAPIKey.
// Every ingest request previously did 3 sequential DB round-trips (GetAPIKeyByHash,
// GetTenant, GetService). Now the first resolution populates a Redis entry keyed by
// the SHA-256 hash of the raw API key. Subsequent calls within the TTL return the
// cached value without touching the database.
//
// Cache is invalidated explicitly in RevokeAPIKey and RotateAPIKey so hot keys are
// never served stale after a security operation.

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	pulsetenant "github.com/Avv123/pulselens-common/tenant"
	goredis "github.com/go-redis/redis/v8"
	tenantredis "github.com/Avv123/pulselens-tenant-service/pkg/redis"
)

const apiKeyCacheTTL = 60 * time.Second

func apiKeyCacheKey(keyHash string) string {
	return fmt.Sprintf("apicache:%s", keyHash)
}

func getResolvedFromCache(ctx context.Context, keyHash string) (pulsetenant.ResolvedAPIKey, bool) {
	client := tenantredis.Get()
	if client == nil {
		return pulsetenant.ResolvedAPIKey{}, false
	}
	raw, err := client.Get(ctx, apiKeyCacheKey(keyHash)).Result()
	if err != nil {
		return pulsetenant.ResolvedAPIKey{}, false
	}
	var resolved pulsetenant.ResolvedAPIKey
	if json.Unmarshal([]byte(raw), &resolved) != nil {
		return pulsetenant.ResolvedAPIKey{}, false
	}
	return resolved, true
}

func setResolvedInCache(ctx context.Context, keyHash string, resolved pulsetenant.ResolvedAPIKey) {
	client := tenantredis.Get()
	if client == nil {
		return
	}
	payload, err := json.Marshal(resolved)
	if err != nil {
		return
	}
	_ = client.Set(ctx, apiKeyCacheKey(keyHash), string(payload), apiKeyCacheTTL).Err()
}

func invalidateResolvedCache(ctx context.Context, keyHash string) {
	client := tenantredis.Get()
	if client == nil {
		return
	}
	_ = client.Del(ctx, apiKeyCacheKey(keyHash)).Err()
}

// touchAPIKeyAsync fires a background goroutine so last_used_at updates never
// block the hot ingest path (B13).
func touchAPIKeyAsync(client *goredis.Client, repo interface {
	TouchAPIKey(context.Context, string) error
}, keyID string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()
		_ = repo.TouchAPIKey(ctx, keyID)
	}()
}
