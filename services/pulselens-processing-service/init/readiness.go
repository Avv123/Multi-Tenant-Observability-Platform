package appinit

import (
	"context"
	"fmt"
	"time"

	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/logging"
	platformreadiness "github.com/omniful/pulselens-platform/readiness"
)

func Readiness(ctx context.Context) []platformreadiness.DependencyStatus {
	return platformreadiness.Run(ctx, 5*time.Second,
		platformreadiness.Check{Name: "kafka", Kind: "queue", Fn: func(ctx context.Context) error {
			return platformreadiness.CheckKafka(ctx, config.GetStringSlice("kafka.brokers"))
		}},
		platformreadiness.Check{Name: "postgres", Kind: "database", Fn: func(ctx context.Context) error {
			return platformreadiness.CheckPostgres(ctx, config.GetString("postgres.dsn"))
		}},
		platformreadiness.Check{Name: "clickhouse", Kind: "analytics", Fn: func(ctx context.Context) error {
			return platformreadiness.CheckClickHouse(ctx, config.GetBool("clickhouse.enabled"), config.GetString("clickhouse.baseUrl"), config.GetString("clickhouse.database"), config.GetString("clickhouse.username"), config.GetString("clickhouse.password"))
		}},
		platformreadiness.Check{Name: "redis", Kind: "cache", Fn: func(ctx context.Context) error {
			return platformreadiness.CheckRedis(ctx, firstRedisHost(), config.GetInt("redis.db"))
		}},
		platformreadiness.Check{Name: "minio", Kind: "objectstore", Fn: func(ctx context.Context) error {
			return platformreadiness.CheckObjectStore(ctx, config.GetBool("archive.enabled"), config.GetString("archive.endpoint"), config.GetString("archive.region"), config.GetString("archive.accessKey"), config.GetString("archive.secretKey"), config.GetString("archive.bucket"), config.GetString("archive.prefix"), config.GetBool("archive.forcePathStyle"))
		}},
	)
}

func Preflight(ctx context.Context) error {
	warnInsecureDefaults()
	rows := Readiness(ctx)
	for _, row := range rows {
		if row.Status == "healthy" {
			logging.Infof("preflight dependency healthy service=%s dependency=%s kind=%s", config.GetString("service.name"), row.Name, row.Kind)
			continue
		}
		logging.Errorf("preflight dependency down service=%s dependency=%s kind=%s error=%s", config.GetString("service.name"), row.Name, row.Kind, row.Error)
	}
	if !platformreadiness.AllHealthy(rows) {
		return fmt.Errorf("processing-service preflight failed")
	}
	return nil
}

func firstRedisHost() string {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		return "localhost:6381"
	}
	return hosts[0]
}

func warnInsecureDefaults() {
	if !config.HasEnvOverride("archive.accessKey") && config.GetString("archive.accessKey") == "test" {
		logging.Warnf("using default archive access key; set %s for safer local operation", config.EnvKey("archive.accessKey"))
	}
	if !config.HasEnvOverride("archive.secretKey") && config.GetString("archive.secretKey") == "test" {
		logging.Warnf("using default archive secret key; set %s for safer local operation", config.EnvKey("archive.secretKey"))
	}
}
