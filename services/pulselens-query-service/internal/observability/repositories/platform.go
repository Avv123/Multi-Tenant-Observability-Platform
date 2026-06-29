package repositories

import (
	"context"

	observabilitymodels "github.com/Avv123/pulselens-query-service/internal/observability/models"
)

func (r *Repository) ListCleanupRuns(ctx context.Context, limit int) ([]observabilitymodels.CleanupRun, error) {
	rows := make([]observabilitymodels.CleanupRun, 0)
	err := r.db.WithContext(ctx).Order("started_at desc").Limit(limit).Find(&rows).Error
	return rows, err
}
