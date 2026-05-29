package appinit

import (
	"context"

	"github.com/omniful/pulselens-platform/config"
	"github.com/omniful/pulselens-platform/logging"
	platformpostgres "github.com/omniful/pulselens-platform/postgres"
	platformredis "github.com/omniful/pulselens-platform/redis"
	tenantmodels "github.com/omniful/pulselens-tenant-service/internal/tenants/models"
	"github.com/omniful/pulselens-tenant-service/pkg/postgres"
	tenantredis "github.com/omniful/pulselens-tenant-service/pkg/redis"
	"golang.org/x/crypto/bcrypt"
)

func Initialize(ctx context.Context) {
	initializeLog()
	initializeRedis()
	initializePostgres(ctx)
	initializeSchema()
	seedDemoData(ctx)
}


func seedDemoData(ctx context.Context) {
	db := postgres.Get()

	// 1. Create Demo Tenant
	demoSlug := "demo"
	var existingTenant tenantmodels.Tenant
	if err := db.Where("slug = ?", demoSlug).First(&existingTenant).Error; err != nil {
		logging.Infof("Seeding demo tenant...")
		demoTenant := tenantmodels.Tenant{
			ID:            "tenant_demo",
			Name:          "Demo Workspace",
			Slug:          demoSlug,
			Plan:          tenantmodels.PlanEnterprise,
			IngestQuota:   1000000,
			RetentionDays: 90,
		}
		if err := db.Create(&demoTenant).Error; err != nil {
			logging.Errorf("Failed to seed demo tenant: %v", err)
			return
		}
		existingTenant = demoTenant
	}

	// 2. Create Demo User
	demoEmail := "demo@pulselens.io"
	var existingUser tenantmodels.User
	if err := db.Where("email = ?", demoEmail).First(&existingUser).Error; err != nil {
		logging.Infof("Seeding demo user...")
		hash, _ := bcrypt.GenerateFromPassword([]byte("pulselens123"), bcrypt.DefaultCost)
		demoUser := tenantmodels.User{
			ID:           "user_demo",
			TenantID:     existingTenant.ID,
			Name:         "Demo User",
			Email:        demoEmail,
			PasswordHash: string(hash),
			Role:         "tenant_admin",
		}
		db.Create(&demoUser)
	}

	// 2b. Create Platform Admin User
	adminEmail := "admin@pulselens.io"
	var existingAdmin tenantmodels.User
	if err := db.Where("email = ?", adminEmail).First(&existingAdmin).Error; err != nil {
		logging.Infof("Seeding platform admin...")
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin"), bcrypt.DefaultCost)
		platformAdmin := tenantmodels.User{
			ID:           "user_admin",
			TenantID:     existingTenant.ID,
			Name:         "Platform Admin",
			Email:        adminEmail,
			PasswordHash: string(hash),
			Role:         "super_admin",
		}
		db.Create(&platformAdmin)
	}

	// 3. Create Demo Service
	var existingSvc tenantmodels.Service
	if err := db.Where("tenant_id = ? AND name = ?", existingTenant.ID, "web-app").First(&existingSvc).Error; err != nil {
		demoSvc := tenantmodels.Service{
			ID:          "svc_demo",
			TenantID:    existingTenant.ID,
			Name:        "web-app",
			Environment: "production",
			Tags:        "[]",
		}
		db.Create(&demoSvc)
	}
}

func initializeLog() {
	logging.Initialize()
}

func initializeRedis() {
	hosts := config.GetStringSlice("redis.hosts")
	if len(hosts) == 0 {
		hosts = []string{"localhost:6381"}
	}
	client := platformredis.New(hosts[0], config.GetInt("redis.db"))
	tenantredis.Set(client)
}

func initializePostgres(ctx context.Context) {
	db, err := platformpostgres.Open(config.GetString("postgres.dsn"))
	if err != nil {
		logging.Fatalf("unable to initialise postgres: %v", err)
	}
	postgres.Set(db)

	sqlDB, err := db.DB()
	if err != nil {
		logging.Fatalf("unable to get postgres sql db: %v", err)
	}

	err = sqlDB.PingContext(ctx)
	if err != nil {
		logging.Fatalf("unable to ping postgres: %v", err)
	}
}

func initializeSchema() {
	db := postgres.Get()
	err := db.AutoMigrate(
		&tenantmodels.Tenant{},
		&tenantmodels.Service{},
		&tenantmodels.APIKey{},
		&tenantmodels.User{},
		&tenantmodels.AuditLog{},
	)
	if err != nil {
		logging.Fatalf("unable to automigrate tenant schema: %v", err)
	}
}
