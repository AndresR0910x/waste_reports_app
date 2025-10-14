package handlers

import (
	"net/http"

	"backend-residuos-app/internal/services"
	"backend-residuos-app/pkg/middleware"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	authService *services.AuthService
}

// NewAuthHandler crea una nueva instancia del handler de autenticación
func NewAuthHandler(authService *services.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// GetProfile godoc
// @Summary      Get user profile
// @Description  Get the profile of the authenticated user
// @Tags         auth
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  map[string]interface{}
// @Failure      401  {object}  map[string]interface{}
// @Failure      404  {object}  map[string]interface{}
// @Failure      500  {object}  map[string]interface{}
// @Router       /auth/profile [get]
func (ah *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "No user ID found in context",
		})
		return
	}

	profile, err := ah.authService.GetUserProfile(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}

// CreateUser crea un nuevo usuario (solo para administradores)
func (ah *AuthHandler) CreateUser(c *gin.Context) {
	var req services.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	profile, err := ah.authService.CreateUser(c.Request.Context(), req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to create user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    profile,
		"message": "User created successfully",
	})
}

// UpdateProfile actualiza el perfil del usuario autenticado
func (ah *AuthHandler) UpdateProfile(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "No user ID found in context",
		})
		return
	}

	var req services.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Invalid request",
			"message": err.Error(),
		})
		return
	}

	profile, err := ah.authService.UpdateUser(c.Request.Context(), userID, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to update user",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
		"message": "Profile updated successfully",
	})
}

// DeleteAccount elimina la cuenta del usuario autenticado
func (ah *AuthHandler) DeleteAccount(c *gin.Context) {
	userID, exists := middleware.GetUserID(c)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "User not authenticated",
			"message": "No user ID found in context",
		})
		return
	}

	err := ah.authService.DeleteUser(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Failed to delete account",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Account deleted successfully",
	})
}

// GetUserByEmail obtiene un usuario por email (solo para administradores)
func (ah *AuthHandler) GetUserByEmail(c *gin.Context) {
	email := c.Param("email")
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Email parameter required",
			"message": "Email parameter is missing",
		})
		return
	}

	profile, err := ah.authService.GetUserByEmail(c.Request.Context(), email)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "User not found",
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    profile,
	})
}
