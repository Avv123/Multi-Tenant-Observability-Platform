package postgres

import "gorm.io/gorm"

var client *gorm.DB

func Set(db *gorm.DB) {
	client = db
}

func Get() *gorm.DB {
	return client
}
