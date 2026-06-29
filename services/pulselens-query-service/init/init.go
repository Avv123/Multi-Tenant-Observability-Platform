package appinit

import (
	"context"

	platformclickhouse "github.com/Avv123/pulselens-platform/clickhouse"
	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-platform/logging"
	platformpostgres "github.com/Avv123/pulselens-platform/postgres"
	platformredis "github.com/Avv123/pulselens-platform/redis"
	observabilitymodels "github.com/Avv123/pulselens-query-service/internal/observability/models"
	queryclickhouse "github.com/Avv123/pulselens-query-service/pkg/clickhouse"
	"github.com/Avv123/pulselens-query-service/pkg/postgres"
	queryredis "github.com/Avv123/pulselens-query-service/pkg/redis"
)

func Initialize(ctx context.Context) {
	logging.Initialize()
	initializeRedis()
	initializeClickHouse(ctx)

	database, err := platformpostgres.Open(config.GetString("postgres.dsn"))
	if err != nil {
		logging.Fatalf("failed to connect postgres: %v", err)
	}
	postgres.Set(database)

	sqlDB, err := database.DB()
	if err != nil {
		logging.Fatalf("failed to access sql db: %v", err)
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		logging.Fatalf("failed to ping postgres: %v", err)
	}

	if err = database.AutoMigrate(&observabilitymodels.SavedQuery{}, &observabilitymodels.Dashboard{}, &observabilitymodels.CleanupRun{}); err != nil {
		logging.Fatalf("failed to automigrate query schema: %v", err)
	}
}

func initializeRedis() {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6381"}
	}
	queryredis.Set(platformredis.New(hosts[0], config.GetInt("redis.db")))
}

func initializeClickHouse(ctx context.Context) {
	client := platformclickhouse.New(
		config.GetBool("clickhouse.enabled"),
		config.GetString("clickhouse.baseUrl"),
		config.GetString("clickhouse.database"),
		config.GetString("clickhouse.username"),
		config.GetString("clickhouse.password"),
	)
	queryclickhouse.Set(client)
	if client == nil || !client.Enabled() {
		return
	}
	if err := client.Ping(ctx); err != nil {
		logging.Fatalf("failed to ping clickhouse: %v", err)
	}
}
