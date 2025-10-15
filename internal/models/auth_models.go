package models

import (
	"time"
)

// AuthRequest representa una solicitud de autenticación
type AuthRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
}

// RegisterRequest representa una solicitud de registro
type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required,min=6" example:"password123"`
	FullName string `json:"full_name" binding:"required" example:"Juan Pérez"`
	Phone    string `json:"phone,omitempty" example:"+51987654321"`
	Cedula   string `json:"cedula,omitempty" example:"1234567891"`
	Language string `json:"language,omitempty" example:"es" default:"es"`
}

// LoginRequest representa una solicitud de login
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email" example:"user@example.com"`
	Password string `json:"password" binding:"required" example:"password123"`
}

// OAuthRequest representa una solicitud OAuth
type OAuthRequest struct {
	Token    string `json:"token" binding:"required" example:"google_oauth_token"`
	Provider string `json:"provider" binding:"required" example:"google"`
	FullName string `json:"full_name,omitempty" example:"Juan Pérez"`
	Language string `json:"language,omitempty" example:"es" default:"es"`
}

// AuthResponse representa una respuesta de autenticación exitosa
type AuthResponse struct {
	Success      bool      `json:"success" example:"true"`
	Message      string    `json:"message" example:"Authentication successful"`
	AccessToken  string    `json:"access_token" example:"eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string    `json:"refresh_token,omitempty" example:"refresh_token_here"`
	User         *UserInfo `json:"user"`
	ExpiresAt    time.Time `json:"expires_at" example:"2024-12-31T23:59:59Z"`
}

// UserInfo representa información básica del usuario
type UserInfo struct {
	ID          int       `json:"id" example:"1"`
	FirebaseUID string    `json:"firebase_uid" example:"firebase_uid_123"`
	Email       string    `json:"email" example:"user@example.com"`
	FullName    string    `json:"full_name" example:"Juan Pérez"`
	Role        string    `json:"role" example:"ciudadano"`
	RoleID      int       `json:"role_id" example:"3"`
	IsActive    bool      `json:"is_active" example:"true"`
	Language    string    `json:"language" example:"es"`
	CreatedAt   time.Time `json:"created_at" example:"2024-01-01T00:00:00Z"`
}

// ErrorResponse representa una respuesta de error
type ErrorResponse struct {
	Success bool        `json:"success" example:"false"`
	Error   string      `json:"error" example:"validation_error"`
	Message string      `json:"message" example:"Invalid email format"`
	Details interface{} `json:"details,omitempty"`
}

// ValidationError representa errores de validación específicos
type ValidationError struct {
	Field   string `json:"field" example:"email"`
	Tag     string `json:"tag" example:"required"`
	Message string `json:"message" example:"Email is required"`
}
