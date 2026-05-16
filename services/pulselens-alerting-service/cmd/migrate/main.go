package main

import (
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	alertModels "github.com/omniful/pulselens-alerting-service/internal/alerts/models"
)

func main() {
	dsn := "host=pulselens-postgres user=omniful password=omniful dbname=pulselens port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("failed to connect database: %v", err)
	}

	log.Println("Migrating Alert models...")
	if err := db.AutoMigrate(
		&alertModels.AlertPolicy{},
		&alertModels.AlertRule{},
		&alertModels.NotificationChannel{},
		&alertModels.NotificationDelivery{},
		&alertModels.Incident{},
		&alertModels.IncidentEvent{},
		&alertModels.IncidentComment{},
	); err != nil {
		log.Fatalf("alert migration failed: %v", err)
	}

	log.Println("Migrations completed successfully.")
}
