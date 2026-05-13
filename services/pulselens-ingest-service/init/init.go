package appinit

import (
	"context"

	"github.com/omniful/pulselens-ingest-service/pkg/cache"
	"github.com/omniful/pulselens-ingest-service/pkg/producer"
	"github.com/omniful/pulselens-ingest-service/pkg/tenantclient"
	"github.com/omniful/pulselens-platform/config"
	platformkafka "github.com/omniful/pulselens-platform/kafka"
	"github.com/omniful/pulselens-platform/logging"
	platformredis "github.com/omniful/pulselens-platform/redis"
)

func Initialize(_ context.Context) {
	logging.Initialize()

	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6381"}
	}
	cache.Set(platformredis.New(hosts[0], config.GetInt("redis.db")))

	kafkaProducer, err := platformkafka.NewProducer(config.GetStringSlice("kafka.brokers"))
	if err != nil {
		logging.Fatalf("failed to initialize kafka producer: %v", err)
	}
	producer.Set(kafkaProducer)

	tenantclient.Set(tenantclient.New(
		config.GetString("tenantService.baseUrl"),
		config.GetString("tenantService.internalToken"),
	))
}
