package repositories

import (
	"context"
	"time"

	"github.com/Avv123/pulselens-platform/idgen"
	"github.com/Avv123/pulselens-processing-service/internal/telemetry/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateLog(ctx context.Context, row *models.LogEvent) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) CreateMetric(ctx context.Context, row *models.MetricPoint) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) CreateTrace(ctx context.Context, row *models.TraceSpan) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) CreateCustom(ctx context.Context, row *models.CustomEvent) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) CreateDLQ(ctx context.Context, row *models.DeadLetterEvent) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) CreateRetryEvent(ctx context.Context, row *models.RetryEvent) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(row).Error
}

func (r *Repository) ListDueRetryEvents(ctx context.Context, limit int) ([]models.RetryEvent, error) {
	rows := make([]models.RetryEvent, 0)
	err := r.db.WithContext(ctx).
		Where("status = ? and next_attempt_at <= ?", "pending", time.Now().UTC()).
		Order("next_attempt_at asc").
		Limit(limit).
		Find(&rows).Error
	return rows, err
}

func (r *Repository) MarkRetryEventStatus(ctx context.Context, id string, status string, lastError string, dispatchedAt *time.Time) error {
	updates := map[string]interface{}{
		"status":     status,
		"last_error": lastError,
		"updated_at": time.Now().UTC(),
	}
	if dispatchedAt != nil {
		updates["dispatched_at"] = *dispatchedAt
	}
	return r.db.WithContext(ctx).Model(&models.RetryEvent{}).Where("id = ?", id).Updates(updates).Error
}

func (r *Repository) CreateArchiveRecord(ctx context.Context, row *models.ArchiveRecord) error {
	return r.db.WithContext(ctx).Create(row).Error
}

func (r *Repository) IncrementUsage(ctx context.Context, tenantID, serviceID, signalType string) error {
	return r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns: []clause.Column{
			{Name: "tenant_id"},
			{Name: "service_id"},
			{Name: "signal_type"},
			{Name: "usage_date"},
		},
		DoUpdates: clause.Assignments(map[string]interface{}{
			"event_count": gorm.Expr("usage_counters.event_count + 1"),
			"updated_at":  time.Now().UTC(),
		}),
	}).Create(&models.UsageCounter{
		ID:         idgen.New("usage"),
		TenantID:   tenantID,
		ServiceID:  serviceID,
		SignalType: signalType,
		UsageDate:  time.Now().UTC().Format("2006-01-02"),
		EventCount: 1,
		UpdatedAt:  time.Now().UTC(),
	}).Error
}
