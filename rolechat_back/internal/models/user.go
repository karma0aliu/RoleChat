package models

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID           uint   `gorm:"primaryKey"`
	Email        string `gorm:"uniqueIndex;not null"`
	PasswordHash string `gorm:"not null"`
	Role         string `gorm:"not null;default:user"`
	Nickname     string `gorm:"not null;"`
	Status       string `gorm:"not null;"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
	DeletedAt    gorm.DeletedAt `gorm:"index"`
}

func (User) TableName() string {
	return "users"
}

type RefreshToken struct {
	ID         uint      `gorm:"primaryKey"`
	UserID     uint      `gorm:"index;not null"`
	TokenID    string    `gorm:"uniqueIndex;size:64;not null"` // jti
	TokenHash  string    `gorm:"not null"`
	ExpiresAt  time.Time `gorm:"index"`
	Revoked    bool      `gorm:"default:false"`
	ReplacedBy string    `gorm:"size:64"`
	CreatedAt  time.Time
}
