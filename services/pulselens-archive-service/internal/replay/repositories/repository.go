package repositories

import (
	"context"
	"encoding/json"
	"time"

	"github.com/omniful/pulselens-archive-service/internal/replay/models"
	pulsetelemetry "github.com/omniful/pulselens-common/telemetry"
	"gorm.io/gorm"
)

type ArchiveRecord struct {
	ID            string    `gorm:"column:id"`
	EventID       string    `gorm:"column:event_id"`
	TenantID      string    `gorm:"column:tenant_id"`
	ServiceID     string    `gorm:"column:service_id"`
	EventType     string    `gorm:"column:event_type"`
	Payload       string    `gorm:"column:payload"`
	ArchiveBucket string    `gorm:"column:archive_bucket"`
	ArchiveKey    string    `gorm:"column:archive_key"`
	ArchivePath   string    `gorm:"column:archive_path"`
	OccurredAt    time.Time `gorm:"column:occurred_at"`
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateReplayJob(ctx context.Context, row *models.ReplayJob) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) ListReplayJobs(ctx context.Context, tenantID string) ([]models.ReplayJob, error) {
	rows := make([]models.ReplayJob, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListPendingReplayJobs(ctx context.Context, limit int) ([]models.ReplayJob, error) {
	rows := make([]models.ReplayJob, 0)
	err := r.db.WithContext(ctx).Where("status = ?", "pending").Order("created_at asc").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) UpdateReplayJob(ctx context.Context, row *models.ReplayJob) error {
	return r.db.WithContext(ctx).Save(row).Error
}

func (r *Repository) CountReplayJobs(ctx context.Context, tenantID string) int64 {
	var count int64
	r.db.WithContext(ctx).Model(&models.ReplayJob{}).Where("tenant_id = ?", tenantID).Count(&count)
	return count
}

func (r *Repository) CountArchivedEvents(ctx context.Context, tenantID string) int64 {
	var count int64
	r.db.WithContext(ctx).Table("archive_records").Where("tenant_id = ?", tenantID).Count(&count)
	return count
}

func (r *Repository) CountArchiveObjects(ctx context.Context, tenantID string) int64 {
	type result struct {
		Count int64 `gorm:"column:count"`
	}
	var row result
	r.db.WithContext(ctx).
		Table("archive_records").
		Select("count(distinct archive_key) as count").
		Where("tenant_id = ?", tenantID).
		Scan(&row)
	return row.Count
}

func (r *Repository) ListArchivedEvents(ctx context.Context, tenantID string, serviceID string, eventType string, startTime time.Time, endTime time.Time, limit int) ([]ArchiveRecord, error) {
	rows := make([]ArchiveRecord, 0)
	query := r.db.WithContext(ctx).
		Table("archive_records").
		Where("tenant_id = ? and occurred_at >= ? and occurred_at <= ?", tenantID, startTime, endTime)
	if serviceID != "" {
		query = query.Where("service_id = ?", serviceID)
	}
	if eventType != "" {
		query = query.Where("event_type = ?", eventType)
	}
	err := query.Order("occurred_at asc").Limit(limit).Find(&rows).Error
	return rows, err
}

func (r *Repository) ParseEnvelope(record ArchiveRecord) (pulsetelemetry.Envelope, error) {
	var envelope pulsetelemetry.Envelope
	err := json.Unmarshal([]byte(record.Payload), &envelope)
	return envelope, err
}
