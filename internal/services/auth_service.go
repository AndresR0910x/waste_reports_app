package services

import (
	"context"
	"fmt"
	"time"

	"backend-residuos-app/pkg/firebase"

	"firebase.google.com/go/v4/auth"
)

type AuthService struct {
	firebaseClient *firebase.Firebase
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(firebaseClient *firebase.Firebase) *AuthService {
	return &AuthService{
		firebaseClient: firebaseClient,
	}
}

// UserProfile estructura para el perfil de usuario
type UserProfile struct {
	UID           string `json:"uid"`
	Email         string `json:"email"`
	DisplayName   string `json:"display_name,omitempty"`
	PhotoURL      string `json:"photo_url,omitempty"`
	EmailVerified bool   `json:"email_verified"`
	Disabled      bool   `json:"disabled"`
	CreatedAt     string `json:"created_at"`
}

// CreateUserRequest estructura para crear un usuario
type CreateUserRequest struct {
	Email       string `json:"email" binding:"required,email"`
	Password    string `json:"password" binding:"required,min=6"`
	DisplayName string `json:"display_name,omitempty"`
	PhotoURL    string `json:"photo_url,omitempty"`
}

// UpdateUserRequest estructura para actualizar un usuario
type UpdateUserRequest struct {
	Email       string `json:"email,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	PhotoURL    string `json:"photo_url,omitempty"`
	Disabled    *bool  `json:"disabled,omitempty"`
}

// VerifyToken verifica un token de ID de Firebase
func (as *AuthService) VerifyToken(ctx context.Context, idToken string) (*auth.Token, error) {
	token, err := as.firebaseClient.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, fmt.Errorf("invalid token: %v", err)
	}
	return token, nil
}

// GetUserProfile obtiene el perfil de un usuario por UID
func (as *AuthService) GetUserProfile(ctx context.Context, uid string) (*UserProfile, error) {
	user, err := as.firebaseClient.GetUser(ctx, uid)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	profile := &UserProfile{
		UID:           user.UID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		PhotoURL:      user.PhotoURL,
		EmailVerified: user.EmailVerified,
		Disabled:      user.Disabled,
		CreatedAt:     time.Unix(user.UserMetadata.CreationTimestamp/1000, 0).Format("2006-01-02T15:04:05Z"),
	}

	return profile, nil
}

// CreateUser crea un nuevo usuario
func (as *AuthService) CreateUser(ctx context.Context, req CreateUserRequest) (*UserProfile, error) {
	params := (&auth.UserToCreate{}).
		Email(req.Email).
		Password(req.Password).
		DisplayName(req.DisplayName).
		PhotoURL(req.PhotoURL).
		EmailVerified(false).
		Disabled(false)

	user, err := as.firebaseClient.CreateUser(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %v", err)
	}

	profile := &UserProfile{
		UID:           user.UID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		PhotoURL:      user.PhotoURL,
		EmailVerified: user.EmailVerified,
		Disabled:      user.Disabled,
		CreatedAt:     time.Unix(user.UserMetadata.CreationTimestamp/1000, 0).Format("2006-01-02T15:04:05Z"),
	}

	return profile, nil
}

// UpdateUser actualiza un usuario existente
func (as *AuthService) UpdateUser(ctx context.Context, uid string, req UpdateUserRequest) (*UserProfile, error) {
	params := &auth.UserToUpdate{}

	if req.Email != "" {
		params = params.Email(req.Email)
	}
	if req.DisplayName != "" {
		params = params.DisplayName(req.DisplayName)
	}
	if req.PhotoURL != "" {
		params = params.PhotoURL(req.PhotoURL)
	}
	if req.Disabled != nil {
		params = params.Disabled(*req.Disabled)
	}

	user, err := as.firebaseClient.UpdateUser(ctx, uid, params)
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %v", err)
	}

	profile := &UserProfile{
		UID:           user.UID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		PhotoURL:      user.PhotoURL,
		EmailVerified: user.EmailVerified,
		Disabled:      user.Disabled,
		CreatedAt:     time.Unix(user.UserMetadata.CreationTimestamp/1000, 0).Format("2006-01-02T15:04:05Z"),
	}

	return profile, nil
}

// DeleteUser elimina un usuario
func (as *AuthService) DeleteUser(ctx context.Context, uid string) error {
	err := as.firebaseClient.DeleteUser(ctx, uid)
	if err != nil {
		return fmt.Errorf("failed to delete user: %v", err)
	}
	return nil
}

// GetUserByEmail obtiene un usuario por email
func (as *AuthService) GetUserByEmail(ctx context.Context, email string) (*UserProfile, error) {
	user, err := as.firebaseClient.Auth.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	profile := &UserProfile{
		UID:           user.UID,
		Email:         user.Email,
		DisplayName:   user.DisplayName,
		PhotoURL:      user.PhotoURL,
		EmailVerified: user.EmailVerified,
		Disabled:      user.Disabled,
		CreatedAt:     time.Unix(user.UserMetadata.CreationTimestamp/1000, 0).Format("2006-01-02T15:04:05Z"),
	}

	return profile, nil
}

// SetCustomClaims establece claims personalizados para un usuario
func (as *AuthService) SetCustomClaims(ctx context.Context, uid string, claims map[string]interface{}) error {
	err := as.firebaseClient.Auth.SetCustomUserClaims(ctx, uid, claims)
	if err != nil {
		return fmt.Errorf("failed to set custom claims: %v", err)
	}
	return nil
}

// GenerateEmailVerificationLink genera un enlace de verificación de email
func (as *AuthService) GenerateEmailVerificationLink(ctx context.Context, email string) (string, error) {
	link, err := as.firebaseClient.Auth.EmailVerificationLink(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to generate email verification link: %v", err)
	}
	return link, nil
}

// GeneratePasswordResetLink genera un enlace de restablecimiento de contraseña
func (as *AuthService) GeneratePasswordResetLink(ctx context.Context, email string) (string, error) {
	link, err := as.firebaseClient.Auth.PasswordResetLink(ctx, email)
	if err != nil {
		return "", fmt.Errorf("failed to generate password reset link: %v", err)
	}
	return link, nil
}
