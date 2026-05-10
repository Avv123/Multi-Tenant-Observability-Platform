package services

import (
	"context"

	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	platformbackpressure "github.com/omniful/pulselens-platform/backpressure"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-processing-service/pkg/cache"
)

func (s *Service) releasePending(ctx context.Context, eventType pulsetelemetry.EventType) {
	controller := platformbackpressure.New(cache.Get(), "queue:pending")
	_ = controller.Release(ctx, retryTargetTopic(eventType), 1)
}

func queueThresholds() map[string]int64 {
	return map[string]int64{
		config.GetString("kafka.topics.logs"):    config.GetInt64("backpressure.logsMaxPending"),
		config.GetString("kafka.topics.metrics"): config.GetInt64("backpressure.metricsMaxPending"),
		config.GetString("kafka.topics.traces"):  config.GetInt64("backpressure.tracesMaxPending"),
		config.GetString("kafka.topics.custom"):  config.GetInt64("backpressure.customMaxPending"),
	}
}
