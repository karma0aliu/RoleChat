package service

import (
	"errors"
	"rolechat_back/internal/models"
	"rolechat_back/internal/repository"
)

type ChatService interface {
	AddMessage(userID uint, topicID uint, role, content string) (topic *models.Topic, msg *models.Message, newTopic bool, err error)
	ListUserTopics(userID uint, limit, offset int) ([]models.Topic, error)
	ListTopicMessages(userID uint, topicID uint, limit, offset int) ([]models.Message, error)
}

type chatService struct {
	chatRepo repository.ChatRepository
}

func NewChatService(r repository.ChatRepository) ChatService {
	return &chatService{chatRepo: r}
}

func (s *chatService) AddMessage(userID uint, topicID uint, role, content string) (*models.Topic, *models.Message, bool, error) {
	if role == "" {
		role = "user"
	}
	var (
		t        *models.Topic
		m        *models.Message
		err      error
		newTopic bool
	)
	if topicID == 0 {
		t, err = s.chatRepo.CreateTopic(userID, content)
		if err != nil {
			return nil, nil, false, err
		}
		newTopic = true
	} else {
		t, err = s.chatRepo.GetTopicByID(topicID)
		if err != nil {
			return nil, nil, false, err
		}
		if t.UserID != userID {
			return nil, nil, false, errors.New("forbidden")
		}
	}
	m, err = s.chatRepo.CreateMessage(t.ID, role, content)
	if err != nil {
		return nil, nil, newTopic, err
	}
	return t, m, newTopic, nil
}

func (s *chatService) ListUserTopics(userID uint, limit, offset int) ([]models.Topic, error) {
	return s.chatRepo.ListTopicsByUser(userID, limit, offset)
}

func (s *chatService) ListTopicMessages(userID uint, topicID uint, limit, offset int) ([]models.Message, error) {
	t, err := s.chatRepo.GetTopicByID(topicID)
	if err != nil {
		return nil, err
	}
	if t.UserID != userID {
		return nil, errors.New("forbidden")
	}
	return s.chatRepo.ListMessagesByTopic(topicID, limit, offset)
}
