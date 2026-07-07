package main

import (
	"log"
	"os"

	"github.com/herlin-ai/herlin-assistant/config"
	"github.com/herlin-ai/herlin-assistant/database"
	"github.com/herlin-ai/herlin-assistant/internal/auth"
	"github.com/herlin-ai/herlin-assistant/internal/chat"
	"github.com/herlin-ai/herlin-assistant/internal/users"
	"github.com/herlin-ai/herlin-assistant/middleware"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.Initialize(cfg.Database)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Set Gin mode
	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create router
	router := gin.Default()

	// Initialize handlers
	userHandler := users.NewHandler(db, cfg)
	authHandler := auth.NewHandler(db, cfg)
	chatHandler := chat.NewHandler(db, cfg)

	// Middleware
	router.Use(database.DBMiddleware(db))
	router.Use(middleware.CORSMiddleware())

	// Public routes
	public := router.Group("/api/v1/public")
	{
		public.POST("/register", userHandler.Register)
		public.POST("/login", authHandler.Login)
		public.POST("/refresh", authHandler.RefreshToken)
	}

	// Protected routes
	protected := router.Group("/api/v1")
	protected.Use(auth.AuthMiddleware(cfg.JWT.Secret))
	{
		// User routes
		protected.GET("/users/me", userHandler.GetCurrentUser)
		protected.PUT("/users/me", userHandler.UpdateUser)
		
		// Chat routes
		protected.POST("/chat", chatHandler.SendMessage)
		protected.GET("/conversations", chatHandler.GetConversations)
		protected.POST("/conversations", chatHandler.CreateConversation)
		protected.GET("/conversations/:id", chatHandler.GetConversation)
		protected.PUT("/conversations/:id", chatHandler.UpdateConversation)
		protected.DELETE("/conversations/:id", chatHandler.DeleteConversation)
		protected.GET("/conversations/:id/messages", chatHandler.GetMessages)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = cfg.Server.Port
	}
	
	log.Printf("Starting Herlin AI Assistant server on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
