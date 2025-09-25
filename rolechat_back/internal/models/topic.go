package models

import (
	"time"

	"gorm.io/gorm"
)

type Topic struct {
	ID        uint      `gorm:"primaryKey"`
	UserID    uint      `gorm:"not null;index"`
	Title     string    `gorm:"type:varchar(255)"`
	Messages  []Message `gorm:"foreignKey:TopicID"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (Topic) TableName() string {
	return "topics"
}
