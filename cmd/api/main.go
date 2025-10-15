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

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
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
// @description     API para la gestión de reportes de residuos con integración Firebase y PostgreSQL funciono!
// @termsOfService  http://swagger.io/terms/

// @contact.name   API Support
// @contact.url    http://www.example.com/support
// @contact.email  support@example.com

// @license.name  Apache 2.0
// @license.url   http://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8081
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
	authService := services.NewAuthService(firebaseClient, userRepo)
	notificationService := services.NewNotificationService(firebaseClient)
	syncService := services.NewSyncService(firebaseClient, userRepo, reportRepo, auditRepo)

	// Nuevos servicios para reportes
	validationService := services.NewValidationService()
	cedulaService := services.NewCedulaValidationService(cfg.CedulaAPIURL)

	// Configurar Firebase Storage
	photoService, err := services.NewPhotoUploadService(cfg.FirebaseStorageBucket, cfg.FirebaseCredentialsPath)
	if err != nil {
		log.Printf("Warning: Failed to initialize photo upload service: %v", err)
		// Continuar sin el servicio de fotos si no está disponible
	}

	reportService := services.NewReportService(reportRepo, userRepo)

	// Inicializar handlers
	authHandler := handlers.NewAuthHandler(authService)
	registrationHandler := handlers.NewRegistrationHandler(authService)
	notificationHandler := handlers.NewNotificationHandler(notificationService)
	reportHandler := handlers.NewReportHandler(reportService, cedulaService, photoService, validationService)

	if cfg.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// Configuración CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",
		"http://localhost:3001",
		"http://localhost:8080",
		"http://127.0.0.1:3000",
		"http://127.0.0.1:8080",
		"https://localhost:3000",
		"https://127.0.0.1:3000",
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Length",
		"Content-Type",
		"Authorization",
		"Accept",
		"Accept-Encoding",
		"Accept-Language",
		"X-CSRF-Token",
		"X-Requested-With",
		"Cache-Control",
	}
	config.ExposeHeaders = []string{"Content-Length"}
	config.AllowCredentials = true
	config.MaxAge = 12 * 3600 // 12 hours

	router.Use(cors.New(config))

	// Documentación Swagger
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	public := router.Group("/api/v1")
	{
		public.GET("/health", healthCheck)

		// Endpoint de verificación de Firebase
		public.GET("/auth/firebase-status", func(c *gin.Context) {
			c.JSON(200, gin.H{
				"firebase_project_id": cfg.FirebaseProjectID,
				"firebase_configured": firebaseClient != nil,
				"message":             "Firebase connection status",
			})
		})

		// Nuevos endpoints de autenticación
		public.POST("/auth/register", registrationHandler.Register)
		public.POST("/auth/login", registrationHandler.Login)
		public.POST("/auth/oauth", registrationHandler.LoginOAuth)

		// Endpoints existentes (mantener compatibilidad)
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

		// Endpoints públicos para reportes
		reports := public.Group("/reports")
		{
			reports.POST("/validate-cedula", reportHandler.ValidateCedula)
		}
	}

	protected := router.Group("/api/v1")
	protected.Use(middleware.FirebaseAuth(firebaseClient))
	{
		auth := protected.Group("/auth")
		{
			// Nuevos endpoints protegidos
			auth.POST("/verify", registrationHandler.VerifyToken)

			// Endpoints existentes
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

		// Endpoints protegidos para reportes
		reports := protected.Group("/reports")
		{
			reports.POST("/create", reportHandler.CreateReport)
			reports.GET("/my-reports", reportHandler.GetReportsByUser)
			reports.GET("/:id", reportHandler.GetReportByID)
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
