package appinit

import (
	"context"

	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/logging"
	platformpostgres "github.com/omniful/pulselens-platform/postgres"
	platformredis "github.com/omniful/pulselens-platform/redis"
	tenantmodels "github.com/omniful/pulselens-tenant-service/internal/tenants/models"
	"github.com/omniful/pulselens-tenant-service/pkg/postgres"
	tenantredis "github.com/omniful/pulselens-tenant-service/pkg/redis"
)

func Initialize(ctx context.Context) {
	initializeLog()
	initializeRedis()
	initializePostgres(ctx)
	initializeSchema()
}

func initializeLog() {
	logging.Initialize()
}

func initializeRedis() {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6381"}
	}
	client := platformredis.New(hosts[0], config.GetInt("redis.db"))
	tenantredis.Set(client)
}

func initializePostgres(ctx context.Context) {
	db, err := platformpostgres.Open(config.GetString("postgres.dsn"))
	if err != nil {
		logging.Fatalf("unable to initialise postgres: %v", err)
	}
	postgres.Set(db)

	sqlDB, err := db.DB()
	if err != nil {
		logging.Fatalf("unable to get postgres sql db: %v", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		logging.Fatalf("unable to ping postgres: %v", err)
	}
}

func initializeSchema() {
	db := postgres.Get()
	err := db.AutoMigrate(
		&tenantmodels.Tenant{},
		&tenantmodels.Service{},
		&tenantmodels.APIKey{},
		&tenantmodels.User{},
		&tenantmodels.AuditLog{},
	)
	if err != nil {
		logging.Fatalf("unable to automigrate tenant schema: %v", err)
	}
}
