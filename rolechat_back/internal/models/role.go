package models

import "time"

type RolePersona struct {
	ID           uint   `gorm:"primaryKey"`
	Name         string `gorm:"type:varchar(50);uniqueIndex"`
	SystemPrompt string `gorm:"type:text"`
	Voice        string `gorm:"type:varchar(50)"`
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (RolePersona) TableName() string { return "role_personas" }
