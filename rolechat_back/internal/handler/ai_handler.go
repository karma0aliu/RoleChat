package handler

import (
	"fmt"
	"net/http"
	"rolechat_back/internal/models"
	"rolechat_back/internal/service"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gin-gonic/gin"
)

type AIHandler struct {
	ChatSvc service.ChatService
	AISvc   service.AIService
}

type RoleReplyRequest struct {
	TopicID     uint   `json:"topic_id"`
	PersonaName string `json:"persona_name"`
	RoleID      string `json:"role_id"` // 前端发送的角色ID
	Content     string `json:"content" binding:"required"`
}

var demoPersonas = map[string]*models.RolePersona{
	"导师":       {Name: "导师", SystemPrompt: "你是一个耐心的中文导师, 给出循序渐进的讲解, 语言温和。", Voice: "mentor"},
	"搞笑":       {Name: "搞笑", SystemPrompt: "你是一名幽默搞笑的朋友, 回答要轻松, 可以加表情, 但保持有用信息。", Voice: "fun"},
	"严肃":       {Name: "严肃", SystemPrompt: "你是一位严谨专业的顾问, 用正式语气回答, 避免多余的感叹。", Voice: "serious"},
	"hermione": {Name: "赫敏", SystemPrompt: "你就是赫敏·格兰杰本人。请始终用第一人称'我'来回答，绝不使用第三人称。我是格兰芬多学院的学生，哈利和罗恩是我最好的朋友。我聪明博学，热爱学习，来自麻瓜家庭。", Voice: "hermione"},
	"harry":    {Name: "哈利", SystemPrompt: "你就是哈利·波特本人。请始终用第一人称'我'来回答，绝不使用第三人称。我是哈利·波特，被称为'大难不死的男孩'，格兰芬多学院学生。赫敏和罗恩是我最好的朋友。我在德思礼家长大，经历过很多冒险。", Voice: "harry"},
	"ron":      {Name: "罗恩", SystemPrompt: "你就是罗恩·韦斯莱本人。请始终用第一人称'我'来回答，绝不使用第三人称。我是罗恩·韦斯莱，来自韦斯莱家族，格兰芬多学院学生。哈利和赫敏是我最好的朋友。我有很多兄弟姐妹，擅长巫师棋。", Voice: "ron"},
}

func NewAIHandler(chatSvc service.ChatService, aiSvc service.AIService) *AIHandler {
	return &AIHandler{ChatSvc: chatSvc, AISvc: aiSvc}
}

func (h *AIHandler) RoleReply(c *gin.Context) {
	var req RoleReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	// 优先使用role_id，如果没有则使用persona_name
	roleKey := req.RoleID
	if roleKey == "" {
		roleKey = req.PersonaName
	}

	persona := demoPersonas[roleKey]
	if persona == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role/persona: " + roleKey})
		return
	}

	_, userMsg, _, err := h.ChatSvc.AddMessage(userID, req.TopicID, "user", req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	replyText, audio64, err := h.AISvc.GenerateRoleReply(userID, req.TopicID, persona, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, assistantMsg, _, err := h.ChatSvc.AddMessage(userID, req.TopicID, "assistant", replyText)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"user_message":      gin.H{"id": userMsg.ID, "content": userMsg.Content},
		"assistant_message": gin.H{"id": assistantMsg.ID, "content": assistantMsg.Content},
		"audio_base64":      audio64,
		"persona":           req.PersonaName,
	})
}

func (h *AIHandler) StreamRoleReply(c *gin.Context) {
	var req RoleReplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userIDVal, _ := c.Get("userID")
	userID := userIDVal.(uint)

	// 优先使用role_id，如果没有则使用persona_name
	roleKey := req.RoleID
	if roleKey == "" {
		roleKey = req.PersonaName
	}

	persona := demoPersonas[roleKey]
	if persona == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid role/persona: " + roleKey})
		return
	}

	topic, _, _, err := h.ChatSvc.AddMessage(userID, req.TopicID, "user", req.Content)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	effectiveTopicID := topic.ID
	tokenCh, errCh, err := h.AISvc.StreamRoleReply(userID, effectiveTopicID, persona, req.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Writer.Header().Set("Content-Type", "text/event-stream; charset=utf-8")
	c.Writer.Header().Set("Cache-Control", "no-cache, no-transform")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Flush()

	var builder strings.Builder
	lastFlushed := 0
	flush := func(force bool) {
		current := builder.String()
		if len(current) == lastFlushed {
			return
		}
		segment := current[lastFlushed:]
		if !force {
			rc := utf8.RuneCountInString(segment)
			if rc < 12 {
				r, _ := utf8.DecodeLastRuneInString(segment)
				if !strings.ContainsRune("。！？!?，,；;\n", r) {
					return
				}
			}
		}
		esc := strings.ReplaceAll(segment, "\n", "\\n")
		fmt.Fprintf(c.Writer, "data: %s\n\n", esc)
		c.Writer.Flush()
		lastFlushed = len(current)
	}

	heartbeat := time.NewTicker(10 * time.Second)
	defer heartbeat.Stop()

	for {
		select {
		case tk, ok := <-tokenCh:
			if !ok {
				full := builder.String()
				if full != "" {
					_, _, _, _ = h.ChatSvc.AddMessage(userID, effectiveTopicID, "assistant", full)
				}
				flush(true)
				fmt.Fprintf(c.Writer, "data: %s\n\n", "[DONE]")
				c.Writer.Flush()
				return
			}
			builder.WriteString(tk)
			flush(false)
		case e, ok := <-errCh:
			if ok && e != nil {
				fmt.Fprintf(c.Writer, "event: error\n")
				fmt.Fprintf(c.Writer, "data: %s\n\n", strings.ReplaceAll(e.Error(), "\n", " "))
				c.Writer.Flush()
			}
			return
		case <-heartbeat.C:
			fmt.Fprintf(c.Writer, ": ping\n\n")
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			partial := builder.String()
			if partial != "" && lastFlushed < len(partial) {
				_, _, _, _ = h.ChatSvc.AddMessage(userID, effectiveTopicID, "assistant", partial)
			}
			return
		}
	}
}
