package repo

import (
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

func Open(dsn string) (*gorm.DB, error) {
	config := &gorm.Config{
		NamingStrategy: &schema.NamingStrategy{
			TablePrefix: "9_", // add prefix for all tables
		},
	}

	return gorm.Open(postgres.Open(dsn), config)
}
