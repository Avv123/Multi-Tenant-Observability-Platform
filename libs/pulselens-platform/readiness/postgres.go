package readiness

import (
	"context"

	platformpostgres "github.com/Avv123/pulselens-platform/postgres"
)

func CheckPostgres(ctx context.Context, dsn string) error {
	db, err := platformpostgres.Open(dsn)
	if err != nil {
		return err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer func() { _ = sqlDB.Close() }()
	return sqlDB.PingContext(ctx)
}
