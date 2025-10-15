package handlers

import (
	"net/http"
	"strings"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/services"

	"github.com/gin-gonic/gin"
)

type RegistrationHandler struct {
	authService *services.AuthService
}

// NewRegistrationHandler crea una nueva instancia del handler de autenticación
func NewRegistrationHandler(authService *services.AuthService) *RegistrationHandler {
	return &RegistrationHandler{
		authService: authService,
	}
}

// Register godoc
// @Summary      Register a new user
// @Description  Register a new user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.RegisterRequest  true  "Registration request"
// @Success      201      {object}  models.AuthResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      409      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/register [post]
func (ah *RegistrationHandler) Register(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "validation_error",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Registrar usuario
	response, err := ah.authService.RegisterUser(c.Request.Context(), &req)
	if err != nil {
		errorMsg := err.Error()

		// Errores de validación
		if strings.Contains(errorMsg, "validation failed") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "validation_error",
				Message: errorMsg,
			})
			return
		}

		// Usuario ya existe
		if strings.Contains(errorMsg, "already exists") {
			c.JSON(http.StatusConflict, models.ErrorResponse{
				Success: false,
				Error:   "user_exists",
				Message: errorMsg,
			})
			return
		}

		// Errores de configuración de Firebase
		if strings.Contains(errorMsg, "Firebase Email/Password authentication is not enabled") {
			c.JSON(http.StatusServiceUnavailable, models.ErrorResponse{
				Success: false,
				Error:   "firebase_config_error",
				Message: "Authentication service is not properly configured",
				Details: errorMsg,
			})
			return
		}

		// Errores de contraseña débil
		if strings.Contains(errorMsg, "password is too weak") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "weak_password",
				Message: errorMsg,
			})
			return
		}

		// Errores de email inválido
		if strings.Contains(errorMsg, "invalid email") {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "invalid_email",
				Message: errorMsg,
			})
			return
		}

		// Error interno del servidor
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "registration_failed",
			Message: "Failed to register user",
			Details: errorMsg,
		})
		return
	}

	c.JSON(http.StatusCreated, response)
}

// Login godoc
// @Summary      Login user
// @Description  Authenticate user with email and password
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.LoginRequest  true  "Login request"
// @Success      200      {object}  models.AuthResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/login [post]
func (ah *RegistrationHandler) Login(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "validation_error",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Autenticar usuario
	response, err := ah.authService.LoginUser(c.Request.Context(), &req)
	if err != nil {
		// Verificar si es un error de validación
		if len(err.Error()) > 0 && err.Error()[:10] == "validation" {
			c.JSON(http.StatusBadRequest, models.ErrorResponse{
				Success: false,
				Error:   "validation_error",
				Message: err.Error(),
			})
			return
		}

		// Verificar si son credenciales inválidas
		if err.Error() == "invalid credentials" || err.Error() == "user account is disabled" ||
			err.Error() == "user account is disabled in Firebase" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   "invalid_credentials",
				Message: err.Error(),
			})
			return
		}

		// Error interno del servidor
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "login_failed",
			Message: "Failed to authenticate user",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// LoginOAuth godoc
// @Summary      OAuth login
// @Description  Authenticate user with OAuth provider (Google)
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        request  body      models.OAuthRequest  true  "OAuth login request"
// @Success      200      {object}  models.AuthResponse
// @Failure      400      {object}  models.ErrorResponse
// @Failure      401      {object}  models.ErrorResponse
// @Failure      500      {object}  models.ErrorResponse
// @Router       /auth/oauth [post]
func (ah *RegistrationHandler) LoginOAuth(c *gin.Context) {
	var req models.OAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Success: false,
			Error:   "validation_error",
			Message: "Invalid request format",
			Details: err.Error(),
		})
		return
	}

	// Autenticar con OAuth
	response, err := ah.authService.LoginWithOAuth(c.Request.Context(), &req)
	if err != nil {
		// Verificar si es un token inválido
		if len(err.Error()) >= 7 && (err.Error()[:7] == "invalid" || err.Error()[:11] == "unsupported") {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Success: false,
				Error:   "invalid_token",
				Message: err.Error(),
			})
			return
		}

		// Error interno del servidor
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Success: false,
			Error:   "oauth_failed",
			Message: "Failed to authenticate with OAuth",
			Details: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, response)
}

// VerifyToken godoc
// @Summary      Verify JWT token
// @Description  Verify if the provided JWT token is valid
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  models.ErrorResponse
// @Failure      500  {object}  models.ErrorResponse
// @Router       /auth/verify [post]
func (ah *RegistrationHandler) VerifyToken(c *gin.Context) {
	// El middleware de Firebase ya verificó el token
	// Si llegamos aquí, el token es válido

	// Obtener información del token del contexto
	firebaseUID, exists := c.Get("firebase_uid")
	if !exists {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Success: false,
			Error:   "no_token_info",
			Message: "No token information found",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Token is valid",
		"data": gin.H{
			"firebase_uid": firebaseUID,
			"verified":     true,
		},
	})
}
