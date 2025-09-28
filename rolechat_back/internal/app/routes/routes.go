package routes

import (
	"rolechat_back/internal/app/middleware"
	"rolechat_back/internal/handler"
	"rolechat_back/internal/repository"
	"rolechat_back/internal/service"
	"rolechat_back/pkg/ai"
	"rolechat_back/pkg/config"
	"rolechat_back/pkg/logger"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {

	repo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(repo, cfg)
	userHandler := handler.NewUserHandler(userSvc)

	chatRepo := repository.NewChatRepository(db)
	chatSvc := service.NewChatService(chatRepo)
	chatHandler := handler.NewChatHandler(chatSvc)
	zhipuKey := cfg.APIKey.ZhipuAI
	var aiHandler *handler.AIHandler
	if zhipuKey != "" {
		zc := ai.NewZhipuClient(zhipuKey)
		aiSvc := service.NewAIService(chatRepo, zc)
		aiHandler = handler.NewAIHandler(chatSvc, aiSvc)
	}

	api := r.Group("/api")
	{
		api.GET("/health", userHandler.Health)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.Refresh)
		}

		secure := api.Group("")
		secure.Use(middleware.AuthMiddleware(cfg))
		secure.GET("/me", userHandler.GetProfile)
		secure.POST("/chat/message", chatHandler.SendMessage)
		secure.GET("/chat/topics", chatHandler.ListTopics)
		secure.GET("/chat/topics/limit", chatHandler.ListTopicsWithLimit)
		secure.GET("/chat/topics/:id/messages", chatHandler.ListMessages)
		if aiHandler != nil {
			secure.POST("/chat/role-reply", aiHandler.RoleReply)
			secure.POST("/chat/role-reply/stream", aiHandler.StreamRoleReply)
		} else {
			logger.Warn("ZHIPU AI key not configured; /api/chat/role-reply and /api/chat/role-reply/stream not registered")
		}
	}
}
