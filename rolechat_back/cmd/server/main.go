package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"rolechat_back/internal/app/routes"
	"rolechat_back/internal/repository"
	"rolechat_back/pkg/config"
	"rolechat_back/pkg/logger"
)

func main() {

	cfg, err := config.LoadConfig("./configs/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	if err := logger.InitLogger(cfg.Log.Level); err != nil {
		log.Fatalf("Failed to initialize logger: %v", err)
	}
	logger.Info("Logger initialized successfully")

	db, err := repository.InitDB(&cfg.Database)
	if err != nil {
		logger.Fatal("Failed to connect to database", "error", err)
	}
	logger.Info("Database connected successfully")

	r := gin.Default()

	config_cors := cors.DefaultConfig()

	config_cors.AllowOrigins = []string{"*"}
	config_cors.AllowMethods = []string{"GET", "POST", "DELETE", "PUT", "OPTIONS"}
	config_cors.AllowHeaders = []string{"Origin", "Content-Type", "Authorization"}
	r.Use(cors.New(config_cors))

	routes.SetupRoutes(r, db, cfg)
	logger.Info("Routes registered successfully")

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	go func() {
		logger.Info("Starting server", "address", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", "error", err)
		}
	}()

	quit := make(chan os.Signal, 1)

	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", "error", err)
	}

	logger.Info("Server exited properly")
}
