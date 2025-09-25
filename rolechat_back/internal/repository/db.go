package repository

import (
	"fmt"
	"net/url"

	"rolechat_back/internal/models"
	"rolechat_back/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// InitDB initializes the PostgreSQL database connection and runs migrations.
func InitDB(dbCfg *config.DatabaseConfig) (*gorm.DB, error) {
	password := url.QueryEscape(dbCfg.Password)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC", dbCfg.Host, dbCfg.Port, dbCfg.User, password, dbCfg.DBName, dbCfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		return nil, err
	}
	if err := db.AutoMigrate(&models.User{}, &models.Topic{}, &models.Message{}, &models.RefreshToken{}); err != nil {
		return nil, err
	}
	return db, nil
}
