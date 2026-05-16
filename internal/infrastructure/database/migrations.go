package database

import (
	"log"
	"time"

	"github.com/go-gormigrate/gormigrate/v2"
	"gorm.io/gorm"
)

// Migrate runs the versioned database migrations.
func Migrate(db *gorm.DB) error {
	m := gormigrate.New(db, gormigrate.DefaultOptions, []*gormigrate.Migration{
		{
			ID: "202605161100_create_notifications_table",
			Migrate: func(tx *gorm.DB) error {
				// We define the model struct here just for the migration to ensure it exactly matches the state at this point in time.
				type Notification struct {
					ID        string    `gorm:"primaryKey;type:varchar(36)"`
					BatchID   *string   `gorm:"index;type:varchar(36)"`
					Recipient string    `gorm:"not null;index"`
					Channel   string    `gorm:"not null"`
					Content   string    `gorm:"type:text;not null"`
					Priority  string    `gorm:"not null"`
					Status    string    `gorm:"not null;index"`
					CreatedAt time.Time `gorm:"not null"`
					UpdatedAt time.Time `gorm:"not null"`
				}
				return tx.AutoMigrate(&Notification{})
			},
			Rollback: func(tx *gorm.DB) error {
				return tx.Migrator().DropTable("notifications")
			},
		},
	})

	if err := m.Migrate(); err != nil {
		log.Printf("Migration failed: %v", err)
		return err
	}
	log.Println("Migrations ran successfully")
	return nil
}
