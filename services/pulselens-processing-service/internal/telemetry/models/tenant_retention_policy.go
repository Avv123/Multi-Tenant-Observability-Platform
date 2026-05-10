package models

type TenantRetentionPolicy struct {
	ID            string `gorm:"column:id" json:"id"`
	RetentionDays int    `gorm:"column:retention_days" json:"retention_days"`
}
