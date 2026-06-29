package appinit

import (
	"context"
	"fmt"
	"time"

	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-platform/logging"
	platformreadiness "github.com/Avv123/pulselens-platform/readiness"
)

func Readiness(ctx context.Context) []platformreadiness.DependencyStatus {
	return platformreadiness.Run(ctx, 5*time.Second,
		platformreadiness.Check{
			Name: "tenant-service",
			Kind: "http",
			Fn: func(ctx context.Context) error {
				return platformreadiness.CheckHTTP(ctx, config.GetString("tenantService.baseUrl")+"/ready", nil)
			},
		},
		platformreadiness.Check{
			Name: "kafka",
			Kind: "queue",
			Fn: func(ctx context.Context) error {
				return platformreadiness.CheckKafka(ctx, config.GetStringSlice("kafka.brokers"))
			},
		},
		platformreadiness.Check{
			Name: "redis",
			Kind: "cache",
			Fn: func(ctx context.Context) error {
				return platformreadiness.CheckRedis(ctx, firstRedisHost(), config.GetInt("redis.db"))
			},
		},
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
		return fmt.Errorf("ingest-service preflight failed")
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
	if !config.HasEnvOverride("tenantService.internalToken") && config.GetString("tenantService.internalToken") == "pulselens-internal-token" {
		logging.Warnf("using default tenant-service internal token; set %s for safer local operation", config.EnvKey("tenantService.internalToken"))
	}
}
