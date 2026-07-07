// Herlin AI Assistant - Backend Service
// Copyright 2026 Herlin AI. All rights reserved.

package database

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func Initialize(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName, cfg.SSLMode,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Auto migrate models
	if err := db.AutoMigrate(
		&models.User{},
		&models.Conversation{},
		&models.Message{},
		&models.Memory{},
		&models.Document{},
		&models.Agent{},
		&models.Notification{},
	); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	DB = db
	log.Println("Database connected and migrated successfully")
	return db, nil
}

func GetDB() *gorm.DB {
	return DB
}

func DBMiddleware(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	}
}
