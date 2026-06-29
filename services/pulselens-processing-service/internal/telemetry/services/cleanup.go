package services

import (
	"context"
	"time"

	"github.com/Avv123/pulselens-platform/cacheversion"
	"github.com/Avv123/pulselens-platform/config"
	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-platform/logging"
	"github.com/Avv123/pulselens-processing-service/internal/telemetry/models"
	"github.com/Avv123/pulselens-processing-service/pkg/archive"
	"github.com/Avv123/pulselens-processing-service/pkg/cache"
)

func (s *Service) RunCleanup(ctx context.Context) error {
	interval := time.Duration(config.GetInt("retention.cleanupIntervalSeconds")) * time.Second
	if interval <= 0 {
		interval = time.Minute
	}

	if err := s.cleanupExpiredData(ctx); err != nil {
		logging.Errorf("cleanup run failed: %v", err)
	}

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil
		case <-ticker.C:
			if err := s.cleanupExpiredData(ctx); err != nil {
				logging.Errorf("cleanup run failed: %v", err)
			}
		}
	}
}

func (s *Service) cleanupExpiredData(ctx context.Context) error {
	startedAt := time.Now().UTC()
	row := models.CleanupRun{
		ID:        idgen.New("cleanup"),
		Status:    "completed",
		StartedAt: startedAt,
	}

	defaultTelemetryDays := config.GetInt("retention.telemetryDays")
	if defaultTelemetryDays <= 0 {
		defaultTelemetryDays = 7
	}
	defaultArchiveDays := config.GetInt("retention.archiveDays")
	if defaultArchiveDays <= 0 {
		defaultArchiveDays = 14
	}

	var telemetryDeleted int64
	var archiveDeleted int64
	var fileDeleteErrors int64

	policies, err := s.repository.ListTenantRetentionPolicies(ctx)
	if err != nil {
		row.Status = "failed"
		row.ErrorMessage = err.Error()
		row.CompletedAt = time.Now().UTC()
		_ = s.repository.CreateCleanupRun(ctx, &row)
		return err
	}

	deleteOps := []func(context.Context, string, time.Time) (int64, error){
		s.repository.DeleteOldLogs,
		s.repository.DeleteOldMetrics,
		s.repository.DeleteOldTraces,
		s.repository.DeleteOldCustomEvents,
		s.repository.DeleteOldDLQ,
		s.repository.DeleteOldRetries,
		s.repository.DeleteOldTelemetryRollups,
		s.repository.DeleteOldMetricRollups,
		s.repository.DeleteOldServiceHealthRollups,
		s.repository.DeleteOldLogSeverityRollups,
		s.repository.DeleteOldTraceLatencyRollups,
	}

	for _, policy := range policies {
		telemetryDays := policy.RetentionDays
		if telemetryDays <= 0 {
			telemetryDays = defaultTelemetryDays
		}
		archiveDays := maxInt(policy.RetentionDays, defaultArchiveDays)
		telemetryCutoff := time.Now().UTC().AddDate(0, 0, -telemetryDays)
		archiveCutoff := time.Now().UTC().AddDate(0, 0, -archiveDays)

		for _, deleteOp := range deleteOps {
			deleted, deleteErr := deleteOp(ctx, policy.ID, telemetryCutoff)
			if deleteErr != nil {
				row.Status = "failed"
				row.ErrorMessage = deleteErr.Error()
				row.CompletedAt = time.Now().UTC()
				row.TelemetryDeleted = telemetryDeleted
				row.ArchiveDeleted = archiveDeleted
				row.FileDeleteErrors = fileDeleteErrors
				_ = s.repository.CreateCleanupRun(ctx, &row)
				return deleteErr
			}
			telemetryDeleted += deleted
		}

		if clickhouseErr := s.cleanupClickHouseTenant(ctx, policy.ID, telemetryCutoff); clickhouseErr != nil {
			row.Status = "failed"
			row.ErrorMessage = clickhouseErr.Error()
			row.CompletedAt = time.Now().UTC()
			row.TelemetryDeleted = telemetryDeleted
			row.ArchiveDeleted = archiveDeleted
			row.FileDeleteErrors = fileDeleteErrors
			_ = s.repository.CreateCleanupRun(ctx, &row)
			return clickhouseErr
		}

		archiveRows, archiveErr := s.repository.ListExpiredArchiveRecords(ctx, policy.ID, archiveCutoff, 500)
		if archiveErr != nil {
			row.Status = "failed"
			row.ErrorMessage = archiveErr.Error()
			row.CompletedAt = time.Now().UTC()
			row.TelemetryDeleted = telemetryDeleted
			row.ArchiveDeleted = archiveDeleted
			row.FileDeleteErrors = fileDeleteErrors
			_ = s.repository.CreateCleanupRun(ctx, &row)
			return archiveErr
		}

		archiveIDs := make([]string, 0, len(archiveRows))
		type archiveObject struct {
			bucket string
			key    string
		}
		archiveObjects := make(map[archiveObject]struct{})
		for _, archiveRow := range archiveRows {
			archiveIDs = append(archiveIDs, archiveRow.ID)
			archiveObjects[archiveObject{bucket: archiveRow.ArchiveBucket, key: archiveRow.ArchiveKey}] = struct{}{}
		}

		deletedArchives, deleteErr := s.repository.DeleteArchiveRecords(ctx, archiveIDs)
		if deleteErr != nil {
			row.Status = "failed"
			row.ErrorMessage = deleteErr.Error()
			row.CompletedAt = time.Now().UTC()
			row.TelemetryDeleted = telemetryDeleted
			row.ArchiveDeleted = archiveDeleted
			row.FileDeleteErrors = fileDeleteErrors
			_ = s.repository.CreateCleanupRun(ctx, &row)
			return deleteErr
		}
		archiveDeleted += deletedArchives

		for object := range archiveObjects {
			remaining, countErr := s.repository.CountArchiveRecordsForObject(ctx, object.bucket, object.key)
			if countErr != nil {
				fileDeleteErrors++
				continue
			}
			if remaining == 0 {
				if removeErr := archive.Get().Delete(ctx, object.bucket, object.key); removeErr != nil {
					fileDeleteErrors++
				}
			}
		}

		cacheversion.BumpMany(ctx, cache.Get(), policy.ID,
			cacheversion.ScopeTelemetryOverview,
			cacheversion.ScopeLogs,
			cacheversion.ScopeMetrics,
			cacheversion.ScopeTraces,
			cacheversion.ScopeServiceHealth,
			cacheversion.ScopeLogAnalytics,
			cacheversion.ScopeMetricAnalytics,
			cacheversion.ScopeTraceAnalytics,
		)
	}

	row.TelemetryDeleted = telemetryDeleted
	row.ArchiveDeleted = archiveDeleted
	row.FileDeleteErrors = fileDeleteErrors
	row.CompletedAt = time.Now().UTC()
	return s.repository.CreateCleanupRun(ctx, &row)
}

func maxInt(left int, right int) int {
	if left > right {
		return left
	}
	return right
}
