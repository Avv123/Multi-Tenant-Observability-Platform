package services

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/IBM/sarama"
	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	"github.com/omniful/pulselens-platform/cacheversion"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/idgen"
	platformkafka "github.com/omniful/pulselens-platform/kafka"
	"github.com/omniful/pulselens-platform/lock"
	"github.com/omniful/pulselens-platform/logging"
	"github.com/omniful/pulselens-processing-service/internal/telemetry/models"
	"github.com/omniful/pulselens-processing-service/internal/telemetry/repositories"
	"github.com/omniful/pulselens-processing-service/pkg/archive"
	"github.com/omniful/pulselens-processing-service/pkg/cache"
	"github.com/omniful/pulselens-processing-service/pkg/postgres"
	"github.com/omniful/pulselens-processing-service/pkg/producer"
)

type Service struct {
	repository *repositories.Repository
}

func New() *Service {
	return &Service{repository: repositories.NewRepository(postgres.Get())}
}

func RunWorkers(ctx context.Context) error {
	service := New()
	topics := []string{
		config.GetString("kafka.topics.logs"),
		config.GetString("kafka.topics.metrics"),
		config.GetString("kafka.topics.traces"),
		config.GetString("kafka.topics.custom"),
		config.GetString("kafka.topics.retry"),
		config.GetString("kafka.topics.retryScheduled"),
	}
	errorChannel := make(chan error, 2)

	go func() {
		errorChannel <- platformkafka.ConsumeGroup(ctx, config.GetStringSlice("kafka.brokers"), config.GetString("kafka.groupId"), topics, service.HandleMessage)
	}()
	go func() {
		errorChannel <- service.DispatchRetryEvents(ctx)
	}()
	go func() {
		errorChannel <- service.RunCleanup(ctx)
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errorChannel:
		return err
	}
}

func (s *Service) HandleMessage(ctx context.Context, message *sarama.ConsumerMessage) error {
	var envelope pulsetelemetry.Envelope
	if err := json.Unmarshal(message.Value, &envelope); err != nil {
		return err
	}

	if message.Topic == config.GetString("kafka.topics.retryScheduled") {
		return s.ScheduleRetryEvent(ctx, envelope)
	}

	dedupeState, err := cache.Get().Get(ctx, dedupeKey(envelope.TenantID, envelope.EventID)).Result()
	if err == nil && dedupeState == "done" {
		return nil
	}

	if err = s.persistEnvelope(ctx, envelope); err != nil {
		if isDuplicateError(err) {
			_ = cache.Get().Set(ctx, dedupeKey(envelope.TenantID, envelope.EventID), "done", 24*time.Hour).Err()
			s.releasePending(ctx, envelope.EventType)
			return nil
		}
		return s.retryOrDLQ(ctx, envelope, err)
	}
	_ = cache.Get().Set(ctx, dedupeKey(envelope.TenantID, envelope.EventID), "done", 24*time.Hour).Err()
	s.releasePending(ctx, envelope.EventType)
	_ = s.repository.IncrementUsage(ctx, envelope.TenantID, envelope.ServiceID, string(envelope.EventType))
	// B2: archive write is decoupled from the Kafka consumer goroutine.
	// A slow or unavailable MinIO will no longer stall processing throughput.
	go func(env pulsetelemetry.Envelope) {
		archCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if archErr := s.archiveEnvelope(archCtx, env); archErr != nil {
			logging.Errorf("archive write failed tenant=%s event=%s err=%v", env.TenantID, env.EventID, archErr)
		}
	}(envelope)
	bumpTelemetryScopes(ctx, envelope)
	return nil
}

func (s *Service) persistEnvelope(ctx context.Context, envelope pulsetelemetry.Envelope) error {
	payloadBytes, err := json.Marshal(envelope.Payload)
	if err != nil {
		return err
	}

	var primaryErr error
	traceDurationMS := durationFromPayload(envelope.Payload)
	switch envelope.EventType {
	case pulsetelemetry.EventTypeMetric:
		primaryErr = s.repository.CreateMetric(ctx, &models.MetricPoint{
			EventID:     envelope.EventID,
			TenantID:    envelope.TenantID,
			ServiceID:   envelope.ServiceID,
			ServiceName: envelope.ServiceName,
			Environment: envelope.Environment,
			ShardID:     envelope.ShardID,
			MetricName:  stringValue(envelope.Payload["metric_name"]),
			Value:       floatValue(envelope.Payload["value"]),
			Payload:     string(payloadBytes),
			OccurredAt:  envelope.OccurredAt,
			ReceivedAt:  envelope.ReceivedAt,
		})
	case pulsetelemetry.EventTypeTrace:
		primaryErr = s.repository.CreateTrace(ctx, &models.TraceSpan{
			EventID:      envelope.EventID,
			TenantID:     envelope.TenantID,
			ServiceID:    envelope.ServiceID,
			ServiceName:  envelope.ServiceName,
			Environment:  envelope.Environment,
			ShardID:      envelope.ShardID,
			TraceID:      envelope.TraceID,
			SpanID:       stringValue(envelope.Payload["span_id"]),
			ParentSpanID: stringValue(envelope.Payload["parent_span_id"]),
			Operation:    stringValue(envelope.Payload["operation"]),
			Status:       stringValue(envelope.Payload["status"]),
			DurationMS:   traceDurationMS,
			Payload:      string(payloadBytes),
			OccurredAt:   envelope.OccurredAt,
			ReceivedAt:   envelope.ReceivedAt,
		})
	case pulsetelemetry.EventTypeCustom:
		primaryErr = s.repository.CreateCustom(ctx, &models.CustomEvent{
			EventID:     envelope.EventID,
			TenantID:    envelope.TenantID,
			ServiceID:   envelope.ServiceID,
			ServiceName: envelope.ServiceName,
			Environment: envelope.Environment,
			ShardID:     envelope.ShardID,
			Payload:     string(payloadBytes),
			OccurredAt:  envelope.OccurredAt,
			ReceivedAt:  envelope.ReceivedAt,
		})
	default:
		primaryErr = s.repository.CreateLog(ctx, &models.LogEvent{
			EventID:     envelope.EventID,
			TenantID:    envelope.TenantID,
			ServiceID:   envelope.ServiceID,
			ServiceName: envelope.ServiceName,
			Environment: envelope.Environment,
			ShardID:     envelope.ShardID,
			TraceID:     envelope.TraceID,
			Severity:    envelope.Severity,
			Message:     stringValue(envelope.Payload["message"]),
			Payload:     string(payloadBytes),
			OccurredAt:  envelope.OccurredAt,
			ReceivedAt:  envelope.ReceivedAt,
		})
	}

	if primaryErr != nil && !isDuplicateError(primaryErr) {
		return primaryErr
	}
	if isDuplicateError(primaryErr) {
		return primaryErr
	}
	if err = s.persistToClickHouse(ctx, envelope, payloadBytes); err != nil {
		return err
	}
	if err = s.repository.IncrementRollups(ctx, envelope); err != nil {
		return err
	}
	return nil
}

func (s *Service) retryOrDLQ(ctx context.Context, envelope pulsetelemetry.Envelope, processErr error) error {
	maxRetry := config.GetInt("retry.maxCount")
	if envelope.RetryCount < maxRetry {
		envelope.RetryCount++
		envelope.NextAttemptAt = time.Now().UTC().Add(nextRetryDelay(envelope.RetryCount))
		payload, err := json.Marshal(envelope)
		if err != nil {
			return err
		}
		return producer.Get().Publish(ctx, config.GetString("kafka.topics.retryScheduled"), envelope.TenantID+":"+envelope.ServiceID, payload)
	}

	payload, _ := json.Marshal(envelope)
	if err := s.repository.CreateDLQ(ctx, &models.DeadLetterEvent{
		EventID:    envelope.EventID,
		TenantID:   envelope.TenantID,
		ServiceID:  envelope.ServiceID,
		EventType:  string(envelope.EventType),
		Reason:     processErr.Error(),
		Payload:    string(payload),
		RetryCount: envelope.RetryCount,
		ReceivedAt: time.Now().UTC(),
	}); err != nil {
		return err
	}
	s.releasePending(ctx, envelope.EventType)
	return nil
}

func (s *Service) ScheduleRetryEvent(ctx context.Context, envelope pulsetelemetry.Envelope) error {
	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	id := fmt.Sprintf("%s:%d", envelope.EventID, envelope.RetryCount)
	return s.repository.CreateRetryEvent(ctx, &models.RetryEvent{
		ID:            id,
		EventID:       envelope.EventID,
		TenantID:      envelope.TenantID,
		ServiceID:     envelope.ServiceID,
		EventType:     string(envelope.EventType),
		Payload:       string(payload),
		RetryCount:    envelope.RetryCount,
		Status:        "pending",
		NextAttemptAt: envelope.NextAttemptAt,
	})
}

// B3: DispatchRetryEvents now holds a Redis distributed lock so that if two
// processing-service instances run simultaneously (e.g. during a rolling restart)
// they do not double-dispatch the same retry events.
func (s *Service) DispatchRetryEvents(ctx context.Context) error {
	redisLock := lock.NewRedisLock(cache.Get())
	lockKey := "lock:retry-dispatcher"
	lockOwner := idgen.New("dispatcher")

	interval := time.Duration(config.GetInt("replay.pollSeconds")) * time.Second
	if interval <= 0 {
		interval = 3 * time.Second
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		acquired, lockErr := redisLock.Acquire(ctx, lockKey, lockOwner, interval*3)
		if lockErr == nil && acquired {
			if dispatchErr := s.dispatchDueRetryEvents(ctx); dispatchErr != nil {
				logging.Errorf("retry dispatch error: %v", dispatchErr)
			}
			_, _ = redisLock.ReleaseOwner(ctx, lockKey, lockOwner)
		}

		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
		}
	}
}

func (s *Service) dispatchDueRetryEvents(ctx context.Context) error {
	rows, err := s.repository.ListDueRetryEvents(ctx, 100)
	if err != nil {
		return err
	}

	for _, row := range rows {
		var envelope pulsetelemetry.Envelope
		if unmarshalErr := json.Unmarshal([]byte(row.Payload), &envelope); unmarshalErr != nil {
			_ = s.repository.MarkRetryEventStatus(ctx, row.ID, "failed", unmarshalErr.Error(), nil)
			continue
		}

		payload, marshalErr := json.Marshal(envelope)
		if marshalErr != nil {
			_ = s.repository.MarkRetryEventStatus(ctx, row.ID, "failed", marshalErr.Error(), nil)
			continue
		}

		if publishErr := producer.Get().Publish(ctx, retryTargetTopic(envelope.EventType), envelope.TenantID+":"+envelope.ServiceID, payload); publishErr != nil {
			_ = s.repository.MarkRetryEventStatus(ctx, row.ID, "failed", publishErr.Error(), nil)
			continue
		}

		dispatchedAt := time.Now().UTC()
		_ = s.repository.MarkRetryEventStatus(ctx, row.ID, "dispatched", "", &dispatchedAt)
	}
	return nil
}

func (s *Service) archiveEnvelope(ctx context.Context, envelope pulsetelemetry.Envelope) error {
	location, err := archive.Get().Archive(ctx, envelope)
	if err != nil || location.Key == "" {
		return err
	}
	payload, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return s.repository.CreateArchiveRecord(ctx, &models.ArchiveRecord{
		ID:            idgen.New("archive"),
		EventID:       envelope.EventID,
		TenantID:      envelope.TenantID,
		ServiceID:     envelope.ServiceID,
		EventType:     string(envelope.EventType),
		Payload:       string(payload),
		ArchiveBucket: location.Bucket,
		ArchiveKey:    location.Key,
		ArchivePath:   location.URI,
		OccurredAt:    envelope.OccurredAt,
	})
}

func retryTargetTopic(eventType pulsetelemetry.EventType) string {
	switch eventType {
	case pulsetelemetry.EventTypeMetric:
		return config.GetString("kafka.topics.metrics")
	case pulsetelemetry.EventTypeTrace:
		return config.GetString("kafka.topics.traces")
	case pulsetelemetry.EventTypeCustom:
		return config.GetString("kafka.topics.custom")
	default:
		return config.GetString("kafka.topics.logs")
	}
}

// nextRetryDelay computes exponential backoff with ±25% jitter.
// B8: without jitter a mass failure event schedules thousands of retries at the
// exact same timestamp, creating a thundering herd on recovery. Jitter spreads
// them across a window so the system recovers gracefully.
func nextRetryDelay(retryCount int) time.Duration {
	base := time.Duration(config.GetInt("retry.baseDelaySeconds")) * time.Second
	if base <= 0 {
		base = 15 * time.Second
	}
	delay := base * time.Duration(1<<(retryCount-1))
	if delay > 10*time.Minute {
		delay = 10 * time.Minute
	}
	// Add ±25% jitter: randomise within [0.75*delay, 1.25*delay].
	jitter := time.Duration(rand.Int63n(int64(delay / 2)))
	if rand.Intn(2) == 0 {
		delay -= jitter / 2
	} else {
		delay += jitter / 2
	}
	return delay
}

func dedupeKey(tenantID, eventID string) string {
	return fmt.Sprintf("dedupe:%s:%s", tenantID, eventID)
}

func isDuplicateError(err error) bool {
	if err == nil {
		return false
	}
	message := strings.ToLower(err.Error())
	return strings.Contains(message, "duplicate key") || strings.Contains(message, "unique constraint")
}

func stringValue(value interface{}) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func floatValue(value interface{}) float64 {
	switch typed := value.(type) {
	case float64:
		return typed
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	default:
		return 0
	}
}

func durationFromPayload(payload map[string]interface{}) int64 {
	startTime, startOK := timeValue(payload["start_time"])
	endTime, endOK := timeValue(payload["end_time"])
	if startOK && endOK && !endTime.Before(startTime) {
		return endTime.Sub(startTime).Milliseconds()
	}
	return 0
}

func timeValue(value interface{}) (time.Time, bool) {
	switch typed := value.(type) {
	case time.Time:
		return typed.UTC(), true
	case string:
		parsed, err := time.Parse(time.RFC3339, typed)
		if err == nil {
			return parsed.UTC(), true
		}
	}
	return time.Time{}, false
}

func bumpTelemetryScopes(ctx context.Context, envelope pulsetelemetry.Envelope) {
	baseScopes := []string{cacheversion.ScopeTelemetryOverview, cacheversion.ScopeServiceHealth}
	switch envelope.EventType {
	case pulsetelemetry.EventTypeMetric:
		cacheversion.BumpMany(ctx, cache.Get(), envelope.TenantID, append(baseScopes, cacheversion.ScopeMetrics, cacheversion.ScopeMetricAnalytics)...)
	case pulsetelemetry.EventTypeTrace:
		cacheversion.BumpMany(ctx, cache.Get(), envelope.TenantID, append(baseScopes, cacheversion.ScopeTraces, cacheversion.ScopeTraceAnalytics)...)
	default:
		cacheversion.BumpMany(ctx, cache.Get(), envelope.TenantID, append(baseScopes, cacheversion.ScopeLogs, cacheversion.ScopeLogAnalytics)...)
	}
}
