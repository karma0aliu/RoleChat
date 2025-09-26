package repository

import (
	"fmt"
	"net/url"
	"time"

	"rolechat_back/internal/models"
	"rolechat_back/pkg/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(dbCfg *config.DatabaseConfig) (*gorm.DB, error) {
	password := url.QueryEscape(dbCfg.Password)
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s TimeZone=UTC", dbCfg.Host, dbCfg.Port, dbCfg.User, password, dbCfg.DBName, dbCfg.SSLMode)
	fmt.Printf("[DB] Connecting with host=%s port=%d user=%s db=%s sslmode=%s\n", dbCfg.Host, dbCfg.Port, dbCfg.User, dbCfg.DBName, dbCfg.SSLMode)
	var db *gorm.DB
	var err error
	const maxRetries = 6
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err == nil {
			sqlDB, pingErr := db.DB()
			if pingErr == nil {
				if pErr := sqlDB.Ping(); pErr == nil {
					break
				} else {
					err = pErr
				}
			} else {
				err = pingErr
			}
		}
		time.Sleep(time.Duration(1+i) * time.Second)
	}
	if err != nil {
		return nil, fmt.Errorf("connect db failed after retries: %w", err)
	}
	if err := db.AutoMigrate(&models.User{}, &models.Topic{}, &models.Message{}, &models.RefreshToken{}, &models.RolePersona{}); err != nil {
		return nil, err
	}
	return db, nil
}
