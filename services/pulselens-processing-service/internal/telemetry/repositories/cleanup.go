package repositories

import (
	"context"
	"time"

	"github.com/Avv123/pulselens-processing-service/internal/telemetry/models"
)

func (r *Repository) CreateCleanupRun(ctx context.Context, row *models.CleanupRun) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) DeleteOldLogs(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and occurred_at < ?", tenantID, cutoff).Delete(&models.LogEvent{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldMetrics(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and occurred_at < ?", tenantID, cutoff).Delete(&models.MetricPoint{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldTraces(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and occurred_at < ?", tenantID, cutoff).Delete(&models.TraceSpan{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldCustomEvents(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and occurred_at < ?", tenantID, cutoff).Delete(&models.CustomEvent{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldDLQ(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and received_at < ?", tenantID, cutoff).Delete(&models.DeadLetterEvent{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldRetries(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and created_at < ?", tenantID, cutoff).Delete(&models.RetryEvent{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldTelemetryRollups(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and bucket_start < ?", tenantID, cutoff).Delete(&models.TelemetryRollupMinute{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldMetricRollups(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and bucket_start < ?", tenantID, cutoff).Delete(&models.MetricRollupMinute{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldServiceHealthRollups(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and bucket_start < ?", tenantID, cutoff).Delete(&models.ServiceHealthRollupMinute{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldLogSeverityRollups(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and bucket_start < ?", tenantID, cutoff).Delete(&models.LogSeverityRollupMinute{})
	return result.RowsAffected, result.Error
}

func (r *Repository) DeleteOldTraceLatencyRollups(ctx context.Context, tenantID string, cutoff time.Time) (int64, error) {
	result := r.db.WithContext(ctx).Where("tenant_id = ? and bucket_start < ?", tenantID, cutoff).Delete(&models.TraceLatencyRollupMinute{})
	return result.RowsAffected, result.Error
}

func (r *Repository) ListExpiredArchiveRecords(ctx context.Context, tenantID string, cutoff time.Time, limit int) ([]models.ArchiveRecord, error) {
	rows := make([]models.ArchiveRecord, 0)
	err := r.db.WithContext(ctx).
		Where("tenant_id = ? and created_at < ?", tenantID, cutoff).
		Order("created_at asc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) DeleteArchiveRecords(ctx context.Context, ids []string) (int64, error) {
	if len(ids) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).Where("id in ?", ids).Delete(&models.ArchiveRecord{})
	return result.RowsAffected, result.Error
}

func (r *Repository) CountArchiveRecordsForObject(ctx context.Context, archiveBucket string, archiveKey string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&models.ArchiveRecord{}).Where("archive_bucket = ? and archive_key = ?", archiveBucket, archiveKey).Count(&count).Error
	return count, err
}
