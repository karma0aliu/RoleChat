package service

import (
	"fmt"
	"rolechat_back/internal/models"
	"rolechat_back/internal/repository"
	"rolechat_back/pkg/ai"
)

type AIService interface {
	GenerateRoleReply(userID uint, topicID uint, persona *models.RolePersona, userMessage string) (replyText string, audioBase64 string, err error)
	StreamRoleReply(userID uint, topicID uint, persona *models.RolePersona, userMessage string) (<-chan string, <-chan error, error)
}

type aiService struct {
	chatRepo repository.ChatRepository
	client   *ai.ZhipuClient
}

func NewAIService(chatRepo repository.ChatRepository, client *ai.ZhipuClient) AIService {
	return &aiService{chatRepo: chatRepo, client: client}
}

func (s *aiService) GenerateRoleReply(userID, topicID uint, persona *models.RolePersona, userMessage string) (string, string, error) {
	msgs, err := s.chatRepo.ListMessagesByTopic(topicID, 30, 0)
	if err != nil {
		return "", "", err
	}
	chatMsgs := make([]ai.ChatMessage, 0, len(msgs)+3)
	if persona != nil && persona.SystemPrompt != "" {
		chatMsgs = append(chatMsgs, ai.ChatMessage{Role: "system", Content: persona.SystemPrompt})
	}
	for _, m := range msgs {
		role := m.Role
		if role == "" {
			role = "user"
		}
		chatMsgs = append(chatMsgs, ai.ChatMessage{Role: role, Content: m.Content})
	}
	chatMsgs = append(chatMsgs, ai.ChatMessage{Role: "user", Content: userMessage})
	if persona != nil && persona.Voice != "" {
		text, audio, err := s.client.ChatVoice(chatMsgs)
		if err == nil && text != "" {
			return text, audio, nil
		}
	}
	text, err := s.client.Chat(chatMsgs)
	if err != nil {
		return "", "", fmt.Errorf("chat generation failed: %w", err)
	}
	voice := "default"
	if persona != nil && persona.Voice != "" {
		voice = persona.Voice
	}
	audio, err := s.client.TextToSpeech(text, voice)
	if err != nil {
		return text, "", fmt.Errorf("tts failed: %w", err)
	}
	return text, audio, nil
}

func (s *aiService) StreamRoleReply(userID uint, topicID uint, persona *models.RolePersona, userMessage string) (<-chan string, <-chan error, error) {
	msgs, err := s.chatRepo.ListMessagesByTopic(topicID, 30, 0)
	if err != nil {
		return nil, nil, err
	}
	chatMsgs := make([]ai.ChatMessage, 0, len(msgs)+3)
	if persona != nil && persona.SystemPrompt != "" {
		chatMsgs = append(chatMsgs, ai.ChatMessage{Role: "system", Content: persona.SystemPrompt})
	}
	for _, m := range msgs {
		role := m.Role
		if role == "" {
			role = "user"
		}
		chatMsgs = append(chatMsgs, ai.ChatMessage{Role: role, Content: m.Content})
	}
	chatMsgs = append(chatMsgs, ai.ChatMessage{Role: "user", Content: userMessage})
	tokenCh, errCh := s.client.ChatStream(chatMsgs)
	return tokenCh, errCh, nil
}
