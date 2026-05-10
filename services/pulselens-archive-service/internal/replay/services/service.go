package services

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/omniful/pulselens-archive-service/internal/replay/models"
	"github.com/omniful/pulselens-archive-service/internal/replay/repositories"
	replayrequests "github.com/omniful/pulselens-archive-service/internal/replay/requests"
	replayresponses "github.com/omniful/pulselens-archive-service/internal/replay/responses"
	serviceobjectstore "github.com/omniful/pulselens-archive-service/pkg/objectstore"
	"github.com/omniful/pulselens-archive-service/pkg/postgres"
	"github.com/omniful/pulselens-archive-service/pkg/producer"
	commonauth "github.com/omniful/pulselens-common/auth"
	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/errs"
	"github.com/omniful/pulselens-platform/idgen"
)

type Service struct {
	repository *repositories.Repository
}

func New() *Service {
	return &Service{repository: repositories.NewRepository(postgres.Get())}
}

func (s *Service) CreateReplayJob(ctx context.Context, claims *commonauth.Claims, request *replayrequests.CreateReplayJobRequest) (models.ReplayJob, errs.CustomError) {
	startTime, err := time.Parse(time.RFC3339, request.StartTime)
	if err != nil {
		return models.ReplayJob{}, errs.New("BAD_REQUEST", "start_time must be RFC3339")
	}
	endTime, err := time.Parse(time.RFC3339, request.EndTime)
	if err != nil {
		return models.ReplayJob{}, errs.New("BAD_REQUEST", "end_time must be RFC3339")
	}
	if endTime.Before(startTime) {
		return models.ReplayJob{}, errs.New("BAD_REQUEST", "end_time must be after start_time")
	}

	row := models.ReplayJob{
		ID:          idgen.New("replay"),
		TenantID:    claims.TenantID,
		ServiceID:   request.ServiceID,
		EventType:   strings.ToLower(request.EventType),
		StartTime:   startTime.UTC(),
		EndTime:     endTime.UTC(),
		Status:      "pending",
		RequestedBy: claims.UserID,
	}
	if err = s.repository.CreateReplayJob(ctx, &row); err != nil {
		return models.ReplayJob{}, errs.New("INTERNAL_SERVER_ERROR", err.Error())
	}
	return row, errs.CustomError{}
}

func (s *Service) ListReplayJobs(ctx context.Context, claims *commonauth.Claims) ([]models.ReplayJob, errs.CustomError) {
	rows, err := s.repository.ListReplayJobs(ctx, claims.TenantID)
	if err != nil {
		return nil, errs.New("INTERNAL_SERVER_ERROR", err.Error())
	}
	return rows, errs.CustomError{}
}

func (s *Service) Stats(ctx context.Context, claims *commonauth.Claims) replayresponses.ArchiveStatsResponse {
	return replayresponses.ArchiveStatsResponse{
		ReplayJobCount:     s.repository.CountReplayJobs(ctx, claims.TenantID),
		ArchivedEvents:     s.repository.CountArchivedEvents(ctx, claims.TenantID),
		ArchiveObjectCount: s.repository.CountArchiveObjects(ctx, claims.TenantID),
	}
}

func (s *Service) RunReplayJobs(ctx context.Context) error {
	jobs, err := s.repository.ListPendingReplayJobs(ctx, 10)
	if err != nil {
		return err
	}
	for _, job := range jobs {
		if runErr := s.runReplayJob(ctx, &job); runErr != nil {
			now := time.Now().UTC()
			job.Status = "failed"
			job.CompletedAt = &now
			job.ErrorMessage = runErr.Error()
			_ = s.repository.UpdateReplayJob(ctx, &job)
		}
	}
	return nil
}

func (s *Service) runReplayJob(ctx context.Context, job *models.ReplayJob) error {
	now := time.Now().UTC()
	job.Status = "in_progress"
	job.StartedAt = &now
	if err := s.repository.UpdateReplayJob(ctx, job); err != nil {
		return err
	}

	rows, err := s.repository.ListArchivedEvents(ctx, job.TenantID, job.ServiceID, job.EventType, job.StartTime, job.EndTime, 2000)
	if err != nil {
		return err
	}

	var replayCount int64
	for _, row := range rows {
		envelope, err := s.parseArchivedEnvelope(ctx, row)
		if err != nil {
			return err
		}
		originalEventID := envelope.EventID
		envelope.EventID = idgen.New("replayevt")
		envelope.RetryCount = 0
		envelope.NextAttemptAt = time.Time{}
		if envelope.Payload == nil {
			envelope.Payload = map[string]interface{}{}
		}
		envelope.Payload["replayed_from_event_id"] = originalEventID
		payload, marshalErr := json.Marshal(envelope)
		if marshalErr != nil {
			return marshalErr
		}
		if publishErr := producer.Get().Publish(ctx, replayTopic(envelope.EventType), envelope.TenantID+":"+envelope.ServiceID, payload); publishErr != nil {
			return publishErr
		}
		replayCount++
	}

	completedAt := time.Now().UTC()
	job.Status = "completed"
	job.CompletedAt = &completedAt
	job.ReplayCount = replayCount
	job.ErrorMessage = ""
	return s.repository.UpdateReplayJob(ctx, job)
}

func (s *Service) parseArchivedEnvelope(ctx context.Context, row repositories.ArchiveRecord) (pulsetelemetry.Envelope, error) {
	store := serviceobjectstore.Get()
	if store != nil && store.Enabled() && strings.TrimSpace(row.ArchiveKey) != "" {
		payload, err := store.GetObject(ctx, row.ArchiveBucket, row.ArchiveKey)
		if err != nil {
			return pulsetelemetry.Envelope{}, err
		}
		var envelope pulsetelemetry.Envelope
		if unmarshalErr := json.Unmarshal(payload, &envelope); unmarshalErr != nil {
			return pulsetelemetry.Envelope{}, unmarshalErr
		}
		return envelope, nil
	}
	return s.repository.ParseEnvelope(row)
}

func replayTopic(eventType pulsetelemetry.EventType) string {
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
