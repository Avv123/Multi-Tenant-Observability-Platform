package appinit

import (
	"context"

	"github.com/omniful/pulselens-archive-service/internal/replay/models"
	serviceobjectstore "github.com/omniful/pulselens-archive-service/pkg/objectstore"
	"github.com/omniful/pulselens-archive-service/pkg/postgres"
	"github.com/omniful/pulselens-archive-service/pkg/producer"
	serviceredis "github.com/omniful/pulselens-archive-service/pkg/redis"
	"github.com/omniful/pulselens-platform/config"
	platformkafka "github.com/omniful/pulselens-platform/kafka"
	"github.com/omniful/pulselens-platform/logging"
	platformobjectstore "github.com/omniful/pulselens-platform/objectstore"
	platformpostgres "github.com/omniful/pulselens-platform/postgres"
	platformredis "github.com/omniful/pulselens-platform/redis"
)

func Initialize(ctx context.Context) {
	logging.Initialize()
	initializeRedis()
	initializePostgres(ctx)
	initializeProducer()
	initializeObjectStore(ctx)
	initializeSchema()
}

func initializeRedis() {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6379"}
	}
	serviceredis.Set(platformredis.New(hosts[0], config.GetInt("redis.db")))
}

func initializePostgres(ctx context.Context) {
	db, err := platformpostgres.Open(config.GetString("postgres.dsn"))
	if err != nil {
		logging.Fatalf("unable to initialize postgres: %v", err)
	}
	postgres.Set(db)
	sqlDB, err := db.DB()
	if err != nil {
		logging.Fatalf("unable to get sql db: %v", err)
	}
	if err = sqlDB.PingContext(ctx); err != nil {
		logging.Fatalf("unable to ping postgres: %v", err)
	}
}

func initializeProducer() {
	prod, err := platformkafka.NewProducer(config.GetStringSlice("kafka.brokers"))
	if err != nil {
		logging.Fatalf("unable to initialize kafka producer: %v", err)
	}
	producer.Set(prod)
}

func initializeSchema() {
	if err := postgres.Get().AutoMigrate(&models.ReplayJob{}); err != nil {
		logging.Fatalf("unable to automigrate replay schema: %v", err)
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
		logging.Fatalf("unable to initialize archive object store: %v", err)
	}
	if err = client.EnsureBucket(ctx); err != nil {
		logging.Fatalf("unable to ensure archive bucket: %v", err)
	}
	serviceobjectstore.Set(client)
}
