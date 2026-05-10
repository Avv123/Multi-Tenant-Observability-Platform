package services

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	pulsetenant "github.com/omniful/pulselens-common/tenant"
	"github.com/omniful/pulselens-ingest-service/internal/ingestion/responses"
	"github.com/omniful/pulselens-ingest-service/internal/ingestion/validators"
	"github.com/omniful/pulselens-ingest-service/pkg/cache"
	pulselens_error "github.com/omniful/pulselens-ingest-service/pkg/error"
	"github.com/omniful/pulselens-ingest-service/pkg/producer"
	"github.com/omniful/pulselens-ingest-service/pkg/tenantclient"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/errs"
	"github.com/omniful/pulselens-platform/idgen"
	"github.com/omniful/pulselens-platform/quota"
	"github.com/omniful/pulselens-platform/ratelimit"
	"github.com/omniful/pulselens-platform/sharding"
)

type Service struct {
	rateLimiter *ratelimit.FixedWindow
	quota       *quota.DailyCounter
}

var pipelineOverloaded = errors.New("pipeline overloaded")

func New() *Service {
	window := time.Duration(config.GetInt("rateLimit.windowSeconds")) * time.Second
	if window <= 0 {
		window = time.Minute
	}

	defaultLimit := config.GetInt64("rateLimit.defaultPerWindow")
	if defaultLimit <= 0 {
		defaultLimit = 500
	}

	return &Service{
		rateLimiter: ratelimit.NewFixedWindow(cache.Get(), "ingest", defaultLimit, window),
		quota:       quota.NewDailyCounter(cache.Get(), "quota"),
	}
}

func (s *Service) Ingest(ctx context.Context, apiKey string, request *pulsetelemetry.BatchIngestRequest) (responses.IngestResponse, errs.CustomError) {
	if strings.TrimSpace(apiKey) == "" {
		return responses.IngestResponse{}, errs.New(pulselens_error.Unauthorized, "missing api key")
	}
	if customError := validators.ValidateBatch(request); customError.Exists() {
		return responses.IngestResponse{}, customError
	}

	resolved, err := tenantclient.Get().ResolveAPIKey(ctx, apiKey)
	if err != nil {
		return responses.IngestResponse{}, errs.New(pulselens_error.Unauthorized, err.Error())
	}
	if !resolved.Active {
		return responses.IngestResponse{}, errs.New(pulselens_error.Forbidden, "inactive api key")
	}
	if !hasScope(resolved.Scopes, pulsetenant.ScopeIngest) {
		return responses.IngestResponse{}, errs.New(pulselens_error.Forbidden, "api key cannot ingest")
	}

	batchSize := int64(len(request.Events))
	allowed, remainingRate, err := s.rateLimiter.AllowN(ctx, resolved.TenantID+":"+resolved.ServiceID, batchSize)
	if err != nil {
		return responses.IngestResponse{}, errs.New(pulselens_error.InternalServer, err.Error())
	}
	if !allowed {
		return responses.IngestResponse{}, errs.New(pulselens_error.TooMany, "rate limit exceeded")
	}

	quotaLimit := resolved.IngestQuota
	if quotaLimit <= 0 {
		quotaLimit = 100000
	}
	allowed, remainingQuota, err := s.quota.Reserve(ctx, resolved.TenantID, quotaLimit, batchSize)
	if err != nil {
		return responses.IngestResponse{}, errs.New(pulselens_error.InternalServer, err.Error())
	}
	if !allowed {
		return responses.IngestResponse{}, errs.New(pulselens_error.TooMany, "daily ingest quota exceeded")
	}

	accepted := 0
	reservedQueues, reserveErr := s.reserveQueues(ctx, request.Events)
	if reserveErr != nil {
		if errors.Is(reserveErr, pipelineOverloaded) {
			return responses.IngestResponse{}, errs.New(pulselens_error.TooMany, "ingest pipeline is under backpressure")
		}
		return responses.IngestResponse{}, errs.New(pulselens_error.InternalServer, reserveErr.Error())
	}
	publishedByTopic := map[string]int64{}
	for _, event := range request.Events {
		envelope := buildEnvelope(resolved, event)
		payload, marshalErr := json.Marshal(envelope)
		if marshalErr != nil {
			releaseUnpublished(ctx, s, reservedQueues, publishedByTopic)
			return responses.IngestResponse{}, errs.New(pulselens_error.BadRequest, marshalErr.Error())
		}
		topic := topicFor(event.EventType)
		if err = producer.Get().Publish(ctx, topic, resolved.TenantID+":"+resolved.ServiceID, payload); err != nil {
			releaseUnpublished(ctx, s, reservedQueues, publishedByTopic)
			return responses.IngestResponse{}, errs.New(pulselens_error.InternalServer, err.Error())
		}
		publishedByTopic[topic]++
		accepted++
	}

	return responses.IngestResponse{
		Requested:           len(request.Events),
		Accepted:            accepted,
		RemainingRate:       remainingRate,
		RemainingDailyQuota: remainingQuota,
	}, errs.CustomError{}
}

func releaseUnpublished(ctx context.Context, service *Service, reserved map[string]int64, published map[string]int64) {
	toRelease := make(map[string]int64)
	for topic, amount := range reserved {
		if amount > published[topic] {
			toRelease[topic] = amount - published[topic]
		}
	}
	service.releaseQueues(ctx, toRelease)
}

func errPipelineOverloaded() error {
	return pipelineOverloaded
}

func hasScope(scopes []pulsetenant.APIKeyScope, expected pulsetenant.APIKeyScope) bool {
	for _, scope := range scopes {
		if scope == expected {
			return true
		}
	}
	return false
}

func buildEnvelope(resolved pulsetenant.ResolvedAPIKey, event pulsetelemetry.ClientEvent) pulsetelemetry.Envelope {
	eventID := event.EventID
	if eventID == "" {
		eventID = idgen.New("evt")
	}

	occurredAt := event.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now().UTC()
	}

	schemaVersion := event.SchemaVersion
	if schemaVersion == "" {
		schemaVersion = "v1"
	}

	return pulsetelemetry.Envelope{
		EventID:       eventID,
		TenantID:      resolved.TenantID,
		TenantName:    resolved.TenantName,
		ServiceID:     resolved.ServiceID,
		ServiceName:   resolved.ServiceName,
		Environment:   resolved.Environment,
		ShardID:       sharding.BucketForKey(resolved.TenantID, config.GetInt("partitioning.shardBuckets")),
		EventType:     event.EventType,
		SchemaVersion: schemaVersion,
		OccurredAt:    occurredAt,
		ReceivedAt:    time.Now().UTC(),
		TraceID:       event.TraceID,
		Severity:      event.Severity,
		Payload:       event.Payload,
	}
}

func topicFor(eventType pulsetelemetry.EventType) string {
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
