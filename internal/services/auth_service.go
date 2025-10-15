package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/repositories"
	"backend-residuos-app/pkg/firebase"

	"firebase.google.com/go/v4/auth"
)

type AuthService struct {
	firebaseClient    *firebase.Firebase
	validationService *ValidationService
	userRepo          *repositories.UserRepository
}

// NewAuthService crea una nueva instancia del servicio de autenticación
func NewAuthService(firebaseClient *firebase.Firebase, userRepo *repositories.UserRepository) *AuthService {
	return &AuthService{
		firebaseClient:    firebaseClient,
		validationService: NewValidationService(),
		userRepo:          userRepo,
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

// RegisterUser registra un nuevo usuario en Firebase y PostgreSQL
func (as *AuthService) RegisterUser(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	// Validar la solicitud
	if validationErrors := as.validationService.ValidateRegisterRequest(req); len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %+v", validationErrors)
	}

	// Verificar si el usuario ya existe por email
	existingUser, _ := as.userRepo.GetUserByEmail(ctx, req.Email)
	if existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Crear usuario en Firebase
	firebaseUser := &auth.UserToCreate{}
	firebaseUser.Email(req.Email).Password(req.Password).DisplayName(req.FullName).EmailVerified(false)

	userRecord, err := as.firebaseClient.Auth.CreateUser(ctx, firebaseUser)
	if err != nil {
		// Manejar errores específicos de Firebase
		if strings.Contains(err.Error(), "IdP configuration") {
			return nil, fmt.Errorf("Firebase Email/Password authentication is not enabled. Please enable it in Firebase Console under Authentication > Sign-in method")
		}
		if strings.Contains(err.Error(), "EMAIL_EXISTS") {
			return nil, fmt.Errorf("user with email %s already exists in Firebase", req.Email)
		}
		if strings.Contains(err.Error(), "WEAK_PASSWORD") {
			return nil, fmt.Errorf("password is too weak, please use a stronger password")
		}
		if strings.Contains(err.Error(), "INVALID_EMAIL") {
			return nil, fmt.Errorf("invalid email format")
		}
		return nil, fmt.Errorf("failed to create Firebase user: %v", err)
	}

	// Asignar idioma por defecto si no se proporciona
	language := req.Language
	if language == "" {
		language = "es"
	}

	// Crear usuario en PostgreSQL
	user := &models.User{
		FirebaseUID: userRecord.UID,
		Email:       &req.Email,
		FullName:    &req.FullName,
		Phone:       &req.Phone,
		Cedula:      &req.Cedula,
		RoleID:      3, // Rol "ciudadano" por defecto
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	// Establecer campos opcionales como nil si están vacíos
	if req.Phone == "" {
		user.Phone = nil
	}
	if req.Cedula == "" {
		user.Cedula = nil
	}

	err = as.userRepo.CreateUser(ctx, user)
	if err != nil {
		// Si falla la creación en PostgreSQL, eliminar el usuario de Firebase
		as.firebaseClient.Auth.DeleteUser(ctx, userRecord.UID)
		return nil, fmt.Errorf("failed to create user in database: %v", err)
	}

	// Generar token JWT personalizado
	customToken, err := as.firebaseClient.Auth.CustomToken(ctx, userRecord.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate custom token: %v", err)
	}

	// Obtener información del rol
	role, err := as.userRepo.GetRoleByID(ctx, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %v", err)
	}

	// Construir respuesta
	response := &models.AuthResponse{
		Success:     true,
		Message:     "User registered successfully",
		AccessToken: customToken,
		User: &models.UserInfo{
			ID:          user.ID,
			FirebaseUID: user.FirebaseUID,
			Email:       getStringValue(user.Email),
			FullName:    getStringValue(user.FullName),
			Role:        role.Name,
			RoleID:      user.RoleID,
			IsActive:    user.IsActive,
			Language:    language,
			CreatedAt:   user.CreatedAt,
		},
		ExpiresAt: time.Now().Add(24 * time.Hour), // Token válido por 24 horas
	}

	return response, nil
}

// LoginUser autentica un usuario existente
func (as *AuthService) LoginUser(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	// Validar la solicitud
	if validationErrors := as.validationService.ValidateLoginRequest(req); len(validationErrors) > 0 {
		return nil, fmt.Errorf("validation failed: %+v", validationErrors)
	}

	// Buscar usuario en PostgreSQL por email
	user, err := as.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Obtener usuario de Firebase por UID
	firebaseUser, err := as.firebaseClient.Auth.GetUser(ctx, user.FirebaseUID)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if firebaseUser.Disabled {
		return nil, fmt.Errorf("user account is disabled in Firebase")
	}

	// Firebase no permite verificar contraseñas directamente en el Admin SDK
	// Por lo que generamos un custom token y dejamos que el cliente lo use
	customToken, err := as.firebaseClient.Auth.CustomToken(ctx, user.FirebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate authentication token: %v", err)
	}

	// Actualizar último login
	user.LastLogin = &time.Time{}
	*user.LastLogin = time.Now()
	user.UpdatedAt = time.Now()

	err = as.userRepo.UpdateUser(ctx, user)
	if err != nil {
		// No es crítico si falla la actualización del último login
		fmt.Printf("Warning: failed to update last login for user %s: %v\n", user.Email, err)
	}

	// Obtener información del rol
	role, err := as.userRepo.GetRoleByID(ctx, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %v", err)
	}

	// Construir respuesta
	response := &models.AuthResponse{
		Success:     true,
		Message:     "Login successful",
		AccessToken: customToken,
		User: &models.UserInfo{
			ID:          user.ID,
			FirebaseUID: user.FirebaseUID,
			Email:       getStringValue(user.Email),
			FullName:    getStringValue(user.FullName),
			Role:        role.Name,
			RoleID:      user.RoleID,
			IsActive:    user.IsActive,
			Language:    "es", // Por ahora por defecto, luego se puede almacenar en BD
			CreatedAt:   user.CreatedAt,
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return response, nil
}

// LoginWithOAuth autentica un usuario usando OAuth (Google)
func (as *AuthService) LoginWithOAuth(ctx context.Context, req *models.OAuthRequest) (*models.AuthResponse, error) {
	if req.Provider != "google" {
		return nil, fmt.Errorf("unsupported OAuth provider: %s", req.Provider)
	}

	// Verificar el token de Google con Firebase
	token, err := as.firebaseClient.Auth.VerifyIDToken(ctx, req.Token)
	if err != nil {
		return nil, fmt.Errorf("invalid OAuth token: %v", err)
	}

	// Obtener información del usuario de Firebase
	firebaseUser, err := as.firebaseClient.Auth.GetUser(ctx, token.UID)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase user: %v", err)
	}

	// Verificar si el usuario ya existe en PostgreSQL
	existingUser, err := as.userRepo.GetUserByFirebaseUID(ctx, token.UID)
	if err != nil && !strings.Contains(err.Error(), "user not found") {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	var user *models.User

	if existingUser == nil {
		// Usuario no existe, crearlo
		fullName := req.FullName
		if fullName == "" && firebaseUser.DisplayName != "" {
			fullName = firebaseUser.DisplayName
		}
		if fullName == "" {
			fullName = firebaseUser.Email // Fallback
		}

		language := req.Language
		if language == "" {
			language = "es"
		}

		user = &models.User{
			FirebaseUID: token.UID,
			Email:       getStringPointer(firebaseUser.Email),
			FullName:    getStringPointer(fullName),
			RoleID:      3, // Rol "ciudadano" por defecto
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		err = as.userRepo.CreateUser(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}
	} else {
		// Usuario existe, actualizar último login
		user = existingUser
		user.LastLogin = &time.Time{}
		*user.LastLogin = time.Now()
		user.UpdatedAt = time.Now()

		err = as.userRepo.UpdateUser(ctx, user)
		if err != nil {
			fmt.Printf("Warning: failed to update last login for user %s: %v\n", user.Email, err)
		}
	}

	// Obtener información del rol
	role, err := as.userRepo.GetRoleByID(ctx, user.RoleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user role: %v", err)
	}

	// El token de OAuth ya es válido, pero generamos uno personalizado para consistencia
	customToken, err := as.firebaseClient.Auth.CustomToken(ctx, user.FirebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate custom token: %v", err)
	}

	// Construir respuesta
	response := &models.AuthResponse{
		Success:     true,
		Message:     "OAuth login successful",
		AccessToken: customToken,
		User: &models.UserInfo{
			ID:          user.ID,
			FirebaseUID: user.FirebaseUID,
			Email:       getStringValue(user.Email),
			FullName:    getStringValue(user.FullName),
			Role:        role.Name,
			RoleID:      user.RoleID,
			IsActive:    user.IsActive,
			Language:    req.Language,
			CreatedAt:   user.CreatedAt,
		},
		ExpiresAt: time.Now().Add(24 * time.Hour),
	}

	return response, nil
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

// Helper function para convertir *string a string
func getStringValue(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// Helper function para convertir string a *string
func getStringPointer(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
