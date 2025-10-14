package main

import (
	"fmt"
	"log"

	_ "backend-residuos-app/docs"
	"backend-residuos-app/internal/handlers"
	"backend-residuos-app/internal/repositories"
	"backend-residuos-app/internal/services"
	"backend-residuos-app/pkg/config"
	"backend-residuos-app/pkg/database"
	"backend-residuos-app/pkg/firebase"
	"backend-residuos-app/pkg/middleware"
	"backend-residuos-app/pkg/migration"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"
	swaggerFiles "github.com/swaggo/files"
)

// healthCheck godoc
// @Summary      Health check endpoint
// @Description  Checks if the API is running properly
// @Tags         health
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}
// @Router       /health [get]
func healthCheck(c *gin.Context) {
	c.JSON(200, gin.H{
		"status":  "healthy",
		"service": "backend-residuos-app",
	})
}

// @title           Backend Residuos API
// @version         1.0
// @description     API para la gestión de reportes de residuos con integración Firebase y PostgreSQL
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080
// @BasePath  /api/v1

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

func main() {
	cfg := config.LoadConfig()

	firebaseClient, err := firebase.InitializeFirebase(
		cfg.FirebaseCredentialsPath,
		cfg.FirebaseProjectID,
	)
	if err != nil {
		log.Fatalf("Failed to initialize Firebase: %v", err)
	}
	log.Println("Firebase initialized successfully")

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()
	log.Println("Database connected successfully")

	// Ejecutar migraciones automáticamente
	migrationManager := migration.NewMigrationManager(db, cfg)
	if err := migrationManager.RunMigrations(); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}
	log.Println("Database migrations completed")

	// Inicializar repositorios
	userRepo := repositories.NewUserRepository(db)
	reportRepo := repositories.NewReportRepository(db)
	auditRepo := repositories.NewAuditRepository(db)

	// Inicializar servicios
	authService := services.NewAuthService(firebaseClient)
	notificationService := services.NewNotificationService(firebaseClient)
	syncService := services.NewSyncService(firebaseClient, userRepo, reportRepo, auditRepo)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Documentación Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := router.Group("/api/v1")
	{
		public.GET("/health", healthCheck)
		public.POST("/auth/users", authHandler.CreateUser)
		public.GET("/auth/users/email/:email", authHandler.GetUserByEmail)

		// Ruta para sincronizar usuario con Firebase
		public.POST("/sync/user", func(c *gin.Context) {
			var req services.UserSyncRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(400, gin.H{"error": "Invalid request", "message": err.Error()})
				return
			}

			user, err := syncService.SyncUserFromFirebase(c.Request.Context(), req)
			if err != nil {
				c.JSON(500, gin.H{"error": "Sync failed", "message": err.Error()})
				return
			}

			c.JSON(200, gin.H{"success": true, "data": user, "message": "User synced successfully"})
		})
	}

	protected := router.Group("/api/v1")
	protected.Use(middleware.FirebaseAuth(firebaseClient))
	{
		auth := protected.Group("/auth")
		{
			auth.GET("/profile", authHandler.GetProfile)
			auth.PUT("/profile", authHandler.UpdateProfile)
			auth.DELETE("/account", authHandler.DeleteAccount)
		}

		notifications := protected.Group("/notifications")
		{
			notifications.POST("/send", notificationHandler.SendNotification)
			notifications.POST("/multicast", notificationHandler.SendMulticast)
			notifications.POST("/topic", notificationHandler.SendToTopic)
			notifications.POST("/subscribe", notificationHandler.SubscribeToTopic)
			notifications.POST("/unsubscribe", notificationHandler.UnsubscribeFromTopic)
			notifications.POST("/waste-reminder", notificationHandler.SendWasteCollectionReminder)
		}
	}

	log.Printf("Server starting on port %s", cfg.ServerPort)
	log.Printf("Environment: %s", cfg.Environment)
	log.Printf("Database: %s:%s/%s", cfg.DBHost, cfg.DBPort, cfg.DBName)
	log.Printf("Firebase Project ID: %s", cfg.FirebaseProjectID)

	if err := router.Run(fmt.Sprintf(":%s", cfg.ServerPort)); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
