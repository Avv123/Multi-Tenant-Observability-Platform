package appinit

import (
	"context"

	"github.com/Avv123/pulselens-ingest-service/pkg/cache"
	"github.com/Avv123/pulselens-ingest-service/pkg/producer"
	"github.com/Avv123/pulselens-ingest-service/pkg/tenantclient"
	"github.com/Avv123/pulselens-platform/config"
	platformkafka "github.com/Avv123/pulselens-platform/kafka"
	"github.com/Avv123/pulselens-platform/logging"
	platformredis "github.com/Avv123/pulselens-platform/redis"
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
