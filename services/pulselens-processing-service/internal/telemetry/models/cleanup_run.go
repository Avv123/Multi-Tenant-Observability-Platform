package models

import "time"

type CleanupRun struct {
	ID               string    `gorm:"primaryKey;type:text" json:"id"`
	Status           string    `gorm:"index;not null" json:"status"`
	TelemetryDeleted int64     `json:"telemetry_deleted"`
	ArchiveDeleted   int64     `json:"archive_deleted"`
	FileDeleteErrors int64     `json:"file_delete_errors"`
	ErrorMessage     string    `gorm:"type:text" json:"error_message"`
	StartedAt        time.Time `json:"started_at"`
	CompletedAt      time.Time `json:"completed_at"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}
