package services

import (
	"context"
	"fmt"
	"time"

	"github.com/omniful/pulselens-platform/config"
	platformkafka "github.com/omniful/pulselens-platform/kafka"
	platformobjectstore "github.com/omniful/pulselens-platform/objectstore"
	platformredis "github.com/omniful/pulselens-platform/redis"
	platformruntime "github.com/omniful/pulselens-platform/runtime"
	observabilitymodels "github.com/omniful/pulselens-query-service/internal/observability/models"
	observabilityresponses "github.com/omniful/pulselens-query-service/internal/observability/responses"
	queryclickhouse "github.com/omniful/pulselens-query-service/pkg/clickhouse"
	"github.com/omniful/pulselens-query-service/pkg/postgres"
	queryredis "github.com/omniful/pulselens-query-service/pkg/redis"
)

func (s *Service) PlatformRuntime(ctx context.Context) ([]platformruntime.Heartbeat, error) {
	return platformruntime.List(ctx, queryredis.Get())
}

func (s *Service) CleanupRuns(ctx context.Context, limit int) ([]observabilitymodels.CleanupRun, error) {
	return s.repository.ListCleanupRuns(ctx, limit)
}

func (s *Service) DependencyHealth(ctx context.Context) []observabilityresponses.DependencyHealthRow {
	now := time.Now().UTC().Format(time.RFC3339)
	rows := make([]observabilityresponses.DependencyHealthRow, 0, 5)

	rows = append(rows, dependencyRow("redis", "cache", now, func() error {
		return freshRedisPing(ctx)
	}))
	rows = append(rows, dependencyRow("postgres", "database", now, func() error {
		sqlDB, err := postgres.Get().DB()
		if err != nil {
			return err
		}
		return sqlDB.PingContext(ctx)
	}))
	rows = append(rows, dependencyRow("clickhouse", "analytics", now, func() error {
		return queryclickhouse.Get().Ping(ctx)
	}))
	rows = append(rows, dependencyRow("kafka", "queue", now, func() error {
		return platformkafka.Ping(ctx, config.GetStringSlice("kafka.brokers"))
	}))
	rows = append(rows, dependencyRow("minio", "objectstore", now, func() error {
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
			return err
		}
		return client.Ping(ctx)
	}))
	return rows
}

func (s *Service) KafkaLag(ctx context.Context) ([]observabilityresponses.KafkaLagRow, error) {
	groupID := config.GetString("kafka.groups.processing")
	rows, err := platformkafka.ConsumerGroupLag(config.GetStringSlice("kafka.brokers"), groupID, []string{
		config.GetString("kafka.topics.logs"),
		config.GetString("kafka.topics.metrics"),
		config.GetString("kafka.topics.traces"),
		config.GetString("kafka.topics.custom"),
		config.GetString("kafka.topics.retry"),
		config.GetString("kafka.topics.retryScheduled"),
	})
	if err != nil {
		return nil, err
	}

	result := make([]observabilityresponses.KafkaLagRow, 0, len(rows))
	for _, row := range rows {
		result = append(result, observabilityresponses.KafkaLagRow(row))
	}
	return result, nil
}



func freshRedisPing(ctx context.Context) error {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		return fmt.Errorf("redis hosts not configured")
	}
	client := platformredis.New(hosts[0], config.GetInt("redis.db"))
	defer client.Close()
	return client.Ping(ctx).Err()
}

func dependencyRow(name string, depType string, checkedAt string, check func() error) observabilityresponses.DependencyHealthRow {
	row := observabilityresponses.DependencyHealthRow{
		Name:      name,
		Type:      depType,
		Status:    "healthy",
		Message:   "ok",
		CheckedAt: checkedAt,
	}
	if err := check(); err != nil {
		row.Status = "down"
		row.Message = err.Error()
	}
	return row
}
