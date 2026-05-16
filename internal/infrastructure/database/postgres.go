package database

import (
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewPostgresDB establishes a connection to the PostgreSQL database using GORM.
func NewPostgresDB(dsn string) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	maxRetries := 5

	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
			Logger: logger.Default.LogMode(logger.Info),
		})
		if err == nil {
			return db, nil
		}
		log.Printf("Failed to connect to database (attempt %d/%d), retrying in 5 seconds...", i+1, maxRetries)
		time.Sleep(5 * time.Second)
	}

	return nil, err
}
