package main

import (
	"billing-note/internal/handlers"
	"billing-note/internal/middleware"
	"billing-note/internal/repository"
	"billing-note/internal/services"
	"billing-note/pkg/config"
	"billing-note/pkg/database"
	"billing-note/pkg/logger"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	// Initialize logger
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}
	jsonLogs := os.Getenv("LOG_FORMAT") == "json"
	logger.Init(logLevel, jsonLogs)

	logger.Info("Starting Billing Note Server...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.WithError(err).Fatal("Failed to load configuration")
	}
	logger.Info("Configuration loaded successfully")

	// Connect to database
	if err := database.Connect(&cfg.Database); err != nil {
		logger.WithError(err).Fatal("Failed to connect to database")
	}
	defer database.Close()
	logger.Info("Database connected successfully")

	// Initialize repositories
	logger.Debug("Initializing repositories...")
	userRepo := repository.NewUserRepository(database.GetDB())
	categoryRepo := repository.NewCategoryRepository(database.GetDB())
	transactionRepo := repository.NewTransactionRepository(database.GetDB())
	logger.Debug("Repositories initialized")

	// Initialize services
	logger.Debug("Initializing services...")
	authService := services.NewAuthService(userRepo, cfg.JWT.Secret, cfg.JWT.Expiry)
	transactionService := services.NewTransactionService(transactionRepo)

	// Initialize PDF password service
	pdfPasswordService, err := services.NewPDFPasswordService(database.GetDB(), cfg.Encryption.Key)
	if err != nil {
		logger.WithError(err).Fatal("Failed to initialize PDF password service")
	}

	// Initialize upload service
	uploadService := services.NewUploadService(database.GetDB(), pdfPasswordService, cfg.Upload.Dir)

	// Initialize category keyword service
	catKeywordRepo := repository.NewCategoryKeywordRepository(database.GetDB())
	catKeywordService := services.NewCategoryKeywordService(catKeywordRepo, categoryRepo, database.GetDB())
	uploadService.SetCategoryKeywordService(catKeywordService)
	logger.Debug("Services initialized")

	// Initialize handlers
	logger.Debug("Initializing handlers...")
	authHandler := handlers.NewAuthHandler(authService)
	transactionHandler := handlers.NewTransactionHandler(transactionService)
	categoryHandler := handlers.NewCategoryHandler(categoryRepo)
	pdfPasswordHandler := handlers.NewPDFPasswordHandler(pdfPasswordService)
	uploadHandler := handlers.NewUploadHandler(uploadService)
	catKeywordHandler := handlers.NewCategoryKeywordHandler(catKeywordService)
	logger.Debug("Handlers initialized")

	// Setup Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New() // Use gin.New() instead of gin.Default() to have full control over middleware

	// Middleware - Add logging middleware first
	r.Use(gin.Recovery()) // Panic recovery
	r.Use(middleware.LoggingMiddleware()) // Custom logging
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

		// Category Keyword Rules
		api.GET("/category-keywords", catKeywordHandler.List)
		api.POST("/category-keywords", catKeywordHandler.Add)
		api.PUT("/category-keywords/batch", catKeywordHandler.BatchSet)
		api.DELETE("/category-keywords/:id", catKeywordHandler.Delete)
		api.POST("/category-keywords/init-defaults", catKeywordHandler.InitDefaults)
		api.POST("/category-keywords/reclassify", catKeywordHandler.Reclassify)
	}

	// Start server
	addr := fmt.Sprintf(":%s", cfg.Server.Port)
	logger.WithFields(logger.Fields{
		"port": cfg.Server.Port,
		"mode": cfg.Server.Mode,
	}).Info("Server starting")

	if err := r.Run(addr); err != nil {
		logger.WithError(err).Fatal("Failed to start server")
	}
}
