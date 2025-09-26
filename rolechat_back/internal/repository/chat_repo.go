package repository

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"rolechat_back/internal/models"

	"gorm.io/gorm"
)

type ChatRepository interface {
	CreateTopic(userID uint, title string) (*models.Topic, error)
	GetTopicByID(id uint) (*models.Topic, error)
	ListTopicsByUser(userID uint, limit, offset int) ([]models.Topic, error)
	CreateMessage(topicID uint, role, content string) (*models.Message, error)
	ListMessagesByTopic(topicID uint, limit, offset int) ([]models.Message, error)
}

type chatRepository struct {
	db *gorm.DB
}

func NewChatRepository(db *gorm.DB) ChatRepository {
	return &chatRepository{db: db}
}

func (r *chatRepository) CreateTopic(userID uint, title string) (*models.Topic, error) {
	if title == "" {
		title = "Untitled"
	}
	title = strings.ReplaceAll(strings.TrimSpace(title), "\n", " ")
	if utf8.RuneCountInString(title) > 80 {
		runes := []rune(title)
		title = string(runes[:80])
	}
	t := &models.Topic{UserID: userID, Title: title}
	if err := r.db.WithContext(context.Background()).Create(t).Error; err != nil {
		return nil, err
	}
	return t, nil
}

func (r *chatRepository) GetTopicByID(id uint) (*models.Topic, error) {
	var t models.Topic
	if err := r.db.WithContext(context.Background()).First(&t, id).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *chatRepository) ListTopicsByUser(userID uint, limit, offset int) ([]models.Topic, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	var topics []models.Topic
	err := r.db.WithContext(context.Background()).Where("user_id = ?", userID).
		Order("updated_at DESC").
		Limit(limit).Offset(offset).Find(&topics).Error
	return topics, err
}

func (r *chatRepository) CreateMessage(topicID uint, role, content string) (*models.Message, error) {
	if content == "" {
		return nil, errors.New("content required")
	}
	m := &models.Message{TopicID: topicID, Role: role, Content: content}
	if err := r.db.WithContext(context.Background()).Create(m).Error; err != nil {
		return nil, err
	}
	_ = r.db.WithContext(context.Background()).Model(&models.Topic{}).Where("id = ?", topicID).Update("updated_at", gorm.Expr("NOW()"))
	return m, nil
}

func (r *chatRepository) ListMessagesByTopic(topicID uint, limit, offset int) ([]models.Message, error) {
	if limit <= 0 || limit > 200 {
		limit = 100
	}
	var msgs []models.Message
	err := r.db.WithContext(context.Background()).Where("topic_id = ?", topicID).
		Order("id ASC").Limit(limit).Offset(offset).Find(&msgs).Error
	return msgs, err
}
