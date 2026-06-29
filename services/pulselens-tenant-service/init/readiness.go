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
			Name: "postgres",
			Kind: "database",
			Fn: func(ctx context.Context) error {
				return platformreadiness.CheckPostgres(ctx, config.GetString("postgres.dsn"))
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
		return fmt.Errorf("tenant-service preflight failed")
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
	if !config.HasEnvOverride("jwt.secret") && config.GetString("jwt.secret") == "pulselens-local-secret" {
		logging.Warnf("using default local jwt secret; set %s for safer local operation", config.EnvKey("jwt.secret"))
	}
	if !config.HasEnvOverride("internal.token") && config.GetString("internal.token") == "pulselens-internal-token" {
		logging.Warnf("using default internal token; set %s for safer local operation", config.EnvKey("internal.token"))
	}
}
