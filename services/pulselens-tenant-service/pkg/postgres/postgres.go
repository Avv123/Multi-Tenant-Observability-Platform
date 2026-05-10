package postgres

import (
	"sync"

	"gorm.io/gorm"
)

var (
	db   *gorm.DB
	once sync.Once
)

func Get() *gorm.DB {
	return db
}

func Set(database *gorm.DB) {
	once.Do(func() {
		db = database
	})
}
