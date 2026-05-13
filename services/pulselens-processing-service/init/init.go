package appinit

import (
	"context"

	platformclickhouse "github.com/omniful/pulselens-platform/clickhouse"
	"github.com/omniful/pulselens-platform/config"
	platformkafka "github.com/omniful/pulselens-platform/kafka"
	"github.com/omniful/pulselens-platform/logging"
	platformobjectstore "github.com/omniful/pulselens-platform/objectstore"
	platformpostgres "github.com/omniful/pulselens-platform/postgres"
	platformredis "github.com/omniful/pulselens-platform/redis"
	telemetrymodels "github.com/omniful/pulselens-processing-service/internal/telemetry/models"
	"github.com/omniful/pulselens-processing-service/pkg/archive"
	"github.com/omniful/pulselens-processing-service/pkg/cache"
	serviceclickhouse "github.com/omniful/pulselens-processing-service/pkg/clickhouse"
	"github.com/omniful/pulselens-processing-service/pkg/postgres"
	"github.com/omniful/pulselens-processing-service/pkg/producer"
)

func Initialize(ctx context.Context) {
	logging.Initialize()

	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6381"}
	}
	cache.Set(platformredis.New(hosts[0], config.GetInt("redis.db")))

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

	kafkaProducer, err := platformkafka.NewProducer(config.GetStringSlice("kafka.brokers"))
	if err != nil {
		logging.Fatalf("failed to initialize kafka producer: %v", err)
	}
	producer.Set(kafkaProducer)
	initializeObjectStore(ctx)
	initializeClickHouse(ctx)

	if err = database.AutoMigrate(
		&telemetrymodels.LogEvent{},
		&telemetrymodels.MetricPoint{},
		&telemetrymodels.TraceSpan{},
		&telemetrymodels.CustomEvent{},
		&telemetrymodels.DeadLetterEvent{},
		&telemetrymodels.UsageCounter{},
		&telemetrymodels.RetryEvent{},
		&telemetrymodels.ArchiveRecord{},
		&telemetrymodels.CleanupRun{},
		&telemetrymodels.TelemetryRollupMinute{},
		&telemetrymodels.MetricRollupMinute{},
		&telemetrymodels.ServiceHealthRollupMinute{},
		&telemetrymodels.LogSeverityRollupMinute{},
		&telemetrymodels.TraceLatencyRollupMinute{},
	); err != nil {
		logging.Fatalf("failed to automigrate telemetry schema: %v", err)
	}
}

func initializeClickHouse(ctx context.Context) {
	client := platformclickhouse.New(
		config.GetBool("clickhouse.enabled"),
		config.GetString("clickhouse.baseUrl"),
		config.GetString("clickhouse.database"),
		config.GetString("clickhouse.username"),
		config.GetString("clickhouse.password"),
	)
	serviceclickhouse.Set(client)
	if client == nil || !client.Enabled() {
		return
	}
	if err := client.Ping(ctx); err != nil {
		logging.Fatalf("failed to ping clickhouse: %v", err)
	}
	if err := serviceclickhouse.EnsureTelemetrySchema(ctx, client); err != nil {
		logging.Fatalf("failed to ensure clickhouse schema: %v", err)
	}
}

func initializeObjectStore(ctx context.Context) {
	client, err := platformobjectstore.New(
		config.GetBool("archive.enabled"),
		config.GetString("archive.endpoint"),
		config.GetString("archive.region"),
		config.GetString("archive.accessKey"),
		config.GetString("archive.secretKey"),
		config.GetString("archive.bucket"),
		config.GetString("archive.prefix"),
		config.GetBool("archive.forcePathStyle"),
	)
	if err != nil {
		logging.Fatalf("failed to initialize archive object store: %v", err)
	}
	if err = client.EnsureBucket(ctx); err != nil {
		logging.Fatalf("failed to ensure archive bucket: %v", err)
	}
	archive.Set(client)
}
