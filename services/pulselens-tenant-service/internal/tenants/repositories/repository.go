package repositories

import (
	"context"

	"github.com/omniful/pulselens-tenant-service/internal/tenants/models"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateTenant(ctx context.Context, tenant *models.Tenant, user *models.User, auditLog *models.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(tenant).Error; err != nil {
			return err
		}
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if auditLog != nil {
			if err := tx.Create(auditLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListTenants(ctx context.Context) ([]models.Tenant, error) {
	rows := make([]models.Tenant, 0)
	err := r.db.WithContext(ctx).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetTenant(ctx context.Context, tenantID string) (models.Tenant, error) {
	var row models.Tenant
	err := r.db.WithContext(ctx).Where("id = ?", tenantID).First(&row).Error
	return row, err
}

func (r *Repository) CreateService(ctx context.Context, service *models.Service, auditLog *models.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(service).Error; err != nil {
			return err
		}
		if auditLog != nil {
			if err := tx.Create(auditLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListServices(ctx context.Context, tenantID string) ([]models.Service, error) {
	rows := make([]models.Service, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetService(ctx context.Context, serviceID string) (models.Service, error) {
	var row models.Service
	err := r.db.WithContext(ctx).Where("id = ?", serviceID).First(&row).Error
	return row, err
}

func (r *Repository) CreateAPIKey(ctx context.Context, key *models.APIKey, auditLog *models.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(key).Error; err != nil {
			return err
		}
		if auditLog != nil {
			if err := tx.Create(auditLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListAPIKeys(ctx context.Context, tenantID string) ([]models.APIKey, error) {
	rows := make([]models.APIKey, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) GetAPIKeyByHash(ctx context.Context, hash string) (models.APIKey, error) {
	var row models.APIKey
	err := r.db.WithContext(ctx).Where("key_hash = ?", hash).First(&row).Error
	return row, err
}

func (r *Repository) TouchAPIKey(ctx context.Context, keyID string) error {
	return r.db.WithContext(ctx).Model(&models.APIKey{}).Where("id = ?", keyID).Update("last_used_at", gorm.Expr("now()")).Error
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (models.User, error) {
	var row models.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&row).Error
	return row, err
}

func (r *Repository) GetUserByID(ctx context.Context, userID string) (models.User, error) {
	var row models.User
	err := r.db.WithContext(ctx).Where("id = ?", userID).First(&row).Error
	return row, err
}

func (r *Repository) CreateUser(ctx context.Context, user *models.User, auditLog *models.AuditLog) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(user).Error; err != nil {
			return err
		}
		if auditLog != nil {
			if err := tx.Create(auditLog).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *Repository) ListUsers(ctx context.Context, tenantID string) ([]models.User, error) {
	rows := make([]models.User, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}

func (r *Repository) ListAuditLogs(ctx context.Context, tenantID string) ([]models.AuditLog, error) {
	rows := make([]models.AuditLog, 0)
	err := r.db.WithContext(ctx).Where("tenant_id = ?", tenantID).Order("created_at desc").Find(&rows).Error
	return rows, err
}
