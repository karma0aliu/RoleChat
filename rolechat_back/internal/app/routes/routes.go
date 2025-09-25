package routes

import (
	"rolechat_back/internal/app/middleware"
	"rolechat_back/internal/handler"
	"rolechat_back/internal/repository"
	"rolechat_back/internal/service"
	"rolechat_back/pkg/config"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func SetupRoutes(r *gin.Engine, db *gorm.DB, cfg *config.Config) {

	repo := repository.NewUserRepository(db)
	userSvc := service.NewUserService(repo, cfg)
	userHandler := handler.NewUserHandler(userSvc)

	api := r.Group("/api")
	{
		api.GET("/health", userHandler.Health)
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
			auth.POST("/refresh", userHandler.Refresh)
		}
		// example authenticated group
		secure := api.Group("")
		secure.Use(middleware.AuthMiddleware(cfg))
		secure.GET("/me", userHandler.GetProfile)
	}
}
