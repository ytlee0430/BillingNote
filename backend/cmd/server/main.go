package main

import (
	"billing-note/internal/handlers"
	"billing-note/internal/middleware"
	"billing-note/internal/repository"
	"billing-note/internal/services"
	"billing-note/pkg/config"
	"billing-note/pkg/database"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Connect to database
	if err := database.Connect(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Initialize repositories
	userRepo := repository.NewUserRepository(database.GetDB())
	categoryRepo := repository.NewCategoryRepository(database.GetDB())
	transactionRepo := repository.NewTransactionRepository(database.GetDB())

	// Initialize services
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiry)
	transactionService := services.NewTransactionService(transactionRepo)

	// Initialize PDF password service
	pdfPasswordService, err := services.NewPDFPasswordService(database.GetDB(), cfg.Encryption.Key)
	if err != nil {
		log.Fatalf("Failed to initialize PDF password service: %v", err)
	}

	// Initialize upload service
	uploadService := services.NewUploadService(database.GetDB(), pdfPasswordService, cfg.Upload.Dir)

	// Initialize handlers
	authHandler := handlers.NewAuthHandler(authService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	pdfPasswordHandler := handlers.NewPDFPasswordHandler(pdfPasswordService)
	uploadHandler := handlers.NewUploadHandler(uploadService)

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.Default()

	// Middleware
	r.Use(middleware.CORSMiddleware(cfg.Server.AllowOrigins))

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Auth routes (public)
	auth := r.Group("/api/auth")
	{
		auth.POST("/register", authHandler.Register)
		auth.POST("/login", authHandler.Login)
	}

	// Protected routes
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware(cfg.JWT.Secret))
	{
		// Auth
		api.GET("/auth/me", authHandler.Me)

		// Categories
		api.GET("/categories", categoryHandler.GetAll)
		api.GET("/categories/type/:type", categoryHandler.GetByType)

		// Transactions
		api.POST("/transactions", transactionHandler.Create)
		api.GET("/transactions", transactionHandler.List)
		api.GET("/transactions/:id", transactionHandler.Get)
		api.PUT("/transactions/:id", transactionHandler.Update)
		api.DELETE("/transactions/:id", transactionHandler.Delete)

		// Stats
		api.GET("/stats/monthly", transactionHandler.GetMonthlyStats)
		api.GET("/stats/category", transactionHandler.GetCategoryStats)

		// PDF Upload
		api.POST("/upload/pdf", uploadHandler.UploadAndParse)
		api.POST("/transactions/import", uploadHandler.Import)

		// PDF Password Settings
		api.GET("/settings/pdf-passwords", pdfPasswordHandler.List)
		api.POST("/settings/pdf-passwords", pdfPasswordHandler.Set)
		api.PUT("/settings/pdf-passwords", pdfPasswordHandler.SetMultiple)
		api.DELETE("/settings/pdf-passwords/:priority", pdfPasswordHandler.Delete)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	log.Printf("Server starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
