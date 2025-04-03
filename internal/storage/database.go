package storage

import (
	"fmt"
	"heart-rate-server/internal/config"
	"heart-rate-server/internal/models"
	"os"
	"path/filepath"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dbDir := filepath.Dir(cfg.DBDSN)
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %v", err)
	}

	db, err := gorm.Open(sqlite.Open(cfg.DBDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect database: %v", err)
	}

	if err := db.AutoMigrate(&models.User{}); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %v", err)
	}

	return db, nil
}
