package appinit

import (
	"context"

	alertmodels "github.com/Avv123/pulselens-alerting-service/internal/alerts/models"
	serviceclickhouse "github.com/Avv123/pulselens-alerting-service/pkg/clickhouse"
	servicepostgres "github.com/Avv123/pulselens-alerting-service/pkg/postgres"
	serviceredis "github.com/Avv123/pulselens-alerting-service/pkg/redis"
	platformclickhouse "github.com/Avv123/pulselens-platform/clickhouse"
	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-platform/logging"
	platformpostgres "github.com/Avv123/pulselens-platform/postgres"
	platformredis "github.com/Avv123/pulselens-platform/redis"
)

func Initialize(ctx context.Context) {
	logging.Initialize()
	initializeRedis()
	initializeClickHouse(ctx)
	initializePostgres(ctx)
	initializeSchema()
}

func initializeRedis() {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6381"}
	}
	serviceredis.Set(platformredis.New(hosts[0], config.GetInt("redis.db")))
}

func initializePostgres(ctx context.Context) {
	db, err := platformpostgres.Open(config.GetString("postgres.dsn"))
	if err != nil {
		logging.Fatalf("unable to initialise postgres: %v", err)
	}
	servicepostgres.Set(db)

	sqlDB, err := db.DB()
	if err != nil {
		logging.Fatalf("unable to get sql db: %v", err)
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		logging.Fatalf("unable to ping postgres: %v", err)
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
		logging.Fatalf("unable to ping clickhouse: %v", err)
	}
}

func initializeSchema() {
	if err := servicepostgres.Get().AutoMigrate(
		&alertmodels.AlertPolicy{},
		&alertmodels.AlertRule{},
		&alertmodels.Incident{},
		&alertmodels.IncidentEvent{},
		&alertmodels.NotificationChannel{},
		&alertmodels.NotificationDelivery{},
		&alertmodels.IncidentComment{},
	); err != nil {
		logging.Fatalf("unable to automigrate alert schema: %v", err)
	}
}
