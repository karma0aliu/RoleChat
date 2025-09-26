package handler

import (
	"net/http"
	"strconv"

	"rolechat_back/internal/service"

	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	Chat service.ChatService
}

func NewChatHandler(s service.ChatService) *ChatHandler {
	return &ChatHandler{Chat: s}
}

type sendMessageRequest struct {
	TopicID uint   `json:"topic_id"`
	Role    string `json:"role"`
	Content string `json:"content" binding:"required"`
}

func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req sendMessageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)
	topic, msg, newTopic, err := h.Chat.AddMessage(userID, req.TopicID, req.Role, req.Content)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"new_topic": newTopic,
		"topic":     gin.H{"id": topic.ID, "title": topic.Title},
		"message":   gin.H{"id": msg.ID, "role": msg.Role, "content": msg.Content},
	})
}

func (h *ChatHandler) ListTopics(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)
	topics, err := h.Chat.ListUserTopics(userID, 100, 0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res := make([]gin.H, 0, len(topics))
	for _, t := range topics {
		res = append(res, gin.H{"id": t.ID, "title": t.Title, "updated_at": t.UpdatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"topics": res})
}

func (h *ChatHandler) ListMessages(c *gin.Context) {
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)
	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid id"})
		return
	}
	msgs, err := h.Chat.ListTopicMessages(userID, uint(id64), 200, 0)
	if err != nil {
		status := http.StatusInternalServerError
		if err.Error() == "forbidden" {
			status = http.StatusForbidden
		}
		c.JSON(status, gin.H{"error": err.Error()})
		return
	}
	res := make([]gin.H, 0, len(msgs))
	for _, m := range msgs {
		res = append(res, gin.H{"id": m.ID, "role": m.Role, "content": m.Content, "created_at": m.CreatedAt})
	}
	c.JSON(http.StatusOK, gin.H{"messages": res})
}
