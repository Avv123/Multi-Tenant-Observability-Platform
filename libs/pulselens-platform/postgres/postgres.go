package postgres

import (
	"github.com/omniful/pulselens-platform/netutil"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(netutil.NormalizeDSNHost(dsn)), &gorm.Config{})
}
