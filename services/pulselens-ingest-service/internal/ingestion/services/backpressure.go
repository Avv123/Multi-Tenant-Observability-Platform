package services

import (
	"context"

	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	"github.com/omniful/pulselens-ingest-service/pkg/cache"
	platformbackpressure "github.com/omniful/pulselens-platform/backpressure"
	"github.com/omniful/pulselens-platform/config"
)

func (s *Service) reserveQueues(ctx context.Context, events []pulsetelemetry.ClientEvent) (map[string]int64, error) {
	controller := platformbackpressure.New(cache.Get(), "queue:pending")
	counts := batchTopicCounts(events)
	reserved := make(map[string]int64)

	for topic, amount := range counts {
		allowed, _, err := controller.Reserve(ctx, topic, queueThreshold(topic), amount)
		if err != nil {
			s.releaseQueues(ctx, reserved)
			return nil, err
		}
		if !allowed {
			s.releaseQueues(ctx, reserved)
			return nil, errPipelineOverloaded()
		}
		reserved[topic] = amount
	}

	return reserved, nil
}

func (s *Service) releaseQueues(ctx context.Context, reserved map[string]int64) {
	controller := platformbackpressure.New(cache.Get(), "queue:pending")
	for topic, amount := range reserved {
		_ = controller.Release(ctx, topic, amount)
	}
}

func batchTopicCounts(events []pulsetelemetry.ClientEvent) map[string]int64 {
	counts := make(map[string]int64)
	for _, event := range events {
		counts[topicFor(event.EventType)]++
	}
	return counts
}

func queueThreshold(topic string) int64 {
	switch topic {
	case config.GetString("kafka.topics.metrics"):
		return config.GetInt64("backpressure.metricsMaxPending")
	case config.GetString("kafka.topics.traces"):
		return config.GetInt64("backpressure.tracesMaxPending")
	case config.GetString("kafka.topics.custom"):
		return config.GetInt64("backpressure.customMaxPending")
	default:
		return config.GetInt64("backpressure.logsMaxPending")
	}
}
