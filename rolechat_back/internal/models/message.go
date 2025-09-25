package models

import (
	"time"

	"gorm.io/gorm"
)

type Message struct {
	ID        uint   `gorm:"primaryKey"`
	TopicID   uint   `gorm:"not null;index"`
	Role      string `gorm:"type:varchar(20);not null"`
	Content   string `gorm:"type:text;not null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Message) TableName() string {
	return "messages"
}
