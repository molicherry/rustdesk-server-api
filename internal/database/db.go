package database

import (
	"fmt"

	"github.com/rustdesk/rustdesk-api-server/config"
	"github.com/rustdesk/rustdesk-api-server/internal/model"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init initializes the database connection and runs auto-migration.
func Init(cfg config.DatabaseConfig) error {
	if cfg.Type != "sqlite" {
		return fmt.Errorf("unsupported database type: %s (only sqlite is supported)", cfg.Type)
	}

	var err error
	DB, err = gorm.Open(sqlite.Open(cfg.Path), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return fmt.Errorf("failed to connect to database: %w", err)
	}

	return nil
}

// Migrate runs AutoMigrate for all models.
func Migrate() error {
	return DB.AutoMigrate(
		&model.User{},
		&model.UserToken{},
		&model.Peer{},
		&model.AddressBook{},
		&model.Tag{},
		&model.DeviceGroup{},
		&model.AuditConn{},
		&model.AuditFile{},
		&model.LoginLog{},
		&model.TfaSetupToken{},
		&model.EmailVerification{},
		&model.PasswordReset{},
		&model.Organization{},
		&model.UserOrganization{},
	)
}

// GetDB returns the global database instance.
func GetDB() *gorm.DB {
	return DB
}
