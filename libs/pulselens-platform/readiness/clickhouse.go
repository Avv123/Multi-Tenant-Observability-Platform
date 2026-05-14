package readiness

import (
	"context"

	platformclickhouse "github.com/omniful/pulselens-platform/clickhouse"
)

func CheckClickHouse(ctx context.Context, enabled bool, baseURL, database, username, password string) error {
	client := platformclickhouse.New(enabled, baseURL, database, username, password)
	return client.Ping(ctx)
}
