package services

import (
	"context"
	"fmt"
	"log"
	"strings"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/repositories"
	"backend-residuos-app/pkg/firebase"

	"firebase.google.com/go/v4/auth"
)

// SyncService gestiona la sincronización entre Firebase y PostgreSQL
type SyncService struct {
	firebaseClient *firebase.Firebase
	userRepo       *repositories.UserRepository
	reportRepo     *repositories.ReportRepository
	auditRepo      *repositories.AuditRepository
}

// NewSyncService crea una nueva instancia del servicio de sincronización
func NewSyncService(
	firebaseClient *firebase.Firebase,
	userRepo *repositories.UserRepository,
	reportRepo *repositories.ReportRepository,
	auditRepo *repositories.AuditRepository,
) *SyncService {
	return &SyncService{
		firebaseClient: firebaseClient,
		userRepo:       userRepo,
		reportRepo:     reportRepo,
		auditRepo:      auditRepo,
	}
}

// UserSyncRequest estructura para sincronizar usuario
type UserSyncRequest struct {
	FirebaseUID string  `json:"firebase_uid" binding:"required"`
	Email       *string `json:"email,omitempty"`
	DisplayName *string `json:"display_name,omitempty"`
	Cedula      *string `json:"cedula,omitempty"`
	Phone       *string `json:"phone,omitempty"`
	RoleName    string  `json:"role_name,omitempty"` // Default: "ciudadano"
}

// SyncUserFromFirebase sincroniza un usuario desde Firebase Authentication a PostgreSQL
func (ss *SyncService) SyncUserFromFirebase(ctx context.Context, req UserSyncRequest) (*models.User, error) {
	// Verificar que el usuario existe en Firebase
	firebaseUser, err := ss.firebaseClient.GetUser(ctx, req.FirebaseUID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user from Firebase: %v", err)
	}

	// Usar datos de Firebase si no se proporcionan en la request
	if req.Email == nil && firebaseUser.Email != "" {
		req.Email = &firebaseUser.Email
	}
	if req.DisplayName == nil && firebaseUser.DisplayName != "" {
		req.DisplayName = &firebaseUser.DisplayName
	}

	// Verificar si el usuario ya existe en PostgreSQL
	existingUser, err := ss.userRepo.GetUserByFirebaseUID(ctx, req.FirebaseUID)
	if err != nil && !strings.Contains(err.Error(), "user not found") {
		return nil, fmt.Errorf("failed to check existing user: %v", err)
	}

	var user *models.User

	if existingUser == nil {
		// Usuario no existe, crearlo
		user, err = ss.createNewUser(ctx, req, firebaseUser)
		if err != nil {
			return nil, fmt.Errorf("failed to create new user: %v", err)
		}

		// Log de auditoría
		ss.logAuditEvent(ctx, "user_created", "user", user.ID, user.ID,
			map[string]interface{}{
				"firebase_uid": user.FirebaseUID,
				"email":        user.Email,
				"role":         user.Role.Name,
			})

		log.Printf("New user created and synced: %s (ID: %d)", req.FirebaseUID, user.ID)
	} else {
		// Usuario existe, actualizarlo si es necesario
		user, err = ss.updateExistingUser(ctx, existingUser, req)
		if err != nil {
			return nil, fmt.Errorf("failed to update existing user: %v", err)
		}

		log.Printf("User synced: %s (ID: %d)", req.FirebaseUID, user.ID)
	}

	return user, nil
}

// createNewUser crea un nuevo usuario en PostgreSQL
func (ss *SyncService) createNewUser(ctx context.Context, req UserSyncRequest, firebaseUser *auth.UserRecord) (*models.User, error) {
	// Determinar el rol
	roleName := req.RoleName
	if roleName == "" {
		roleName = models.RoleCiudadano // Rol por defecto
	}

	role, err := ss.userRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %v", err)
	}

	// Crear usuario
	user := &models.User{
		FirebaseUID: req.FirebaseUID,
		Email:       req.Email,
		RoleID:      role.ID,
		Cedula:      req.Cedula,
		Phone:       req.Phone,
		FullName:    req.DisplayName,
		IsActive:    true,
		Role:        role,
	}

	err = ss.userRepo.CreateUser(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user in database: %v", err)
	}

	return user, nil
}

// updateExistingUser actualiza un usuario existente
func (ss *SyncService) updateExistingUser(ctx context.Context, existingUser *models.User, req UserSyncRequest) (*models.User, error) {
	hasChanges := false
	oldValues := make(map[string]interface{})
	newValues := make(map[string]interface{})

	// Verificar cambios en email
	if req.Email != nil && (existingUser.Email == nil || *existingUser.Email != *req.Email) {
		oldValues["email"] = existingUser.Email
		newValues["email"] = req.Email
		existingUser.Email = req.Email
		hasChanges = true
	}

	// Verificar cambios en nombre
	if req.DisplayName != nil && (existingUser.FullName == nil || *existingUser.FullName != *req.DisplayName) {
		oldValues["full_name"] = existingUser.FullName
		newValues["full_name"] = req.DisplayName
		existingUser.FullName = req.DisplayName
		hasChanges = true
	}

	// Verificar cambios en cédula
	if req.Cedula != nil && (existingUser.Cedula == nil || *existingUser.Cedula != *req.Cedula) {
		oldValues["cedula"] = existingUser.Cedula
		newValues["cedula"] = req.Cedula
		existingUser.Cedula = req.Cedula
		hasChanges = true
	}

	// Verificar cambios en teléfono
	if req.Phone != nil && (existingUser.Phone == nil || *existingUser.Phone != *req.Phone) {
		oldValues["phone"] = existingUser.Phone
		newValues["phone"] = req.Phone
		existingUser.Phone = req.Phone
		hasChanges = true
	}

	// Actualizar rol si se especifica y es diferente
	if req.RoleName != "" && existingUser.Role.Name != req.RoleName {
		role, err := ss.userRepo.GetRoleByName(ctx, req.RoleName)
		if err != nil {
			return nil, fmt.Errorf("failed to get new role: %v", err)
		}

		oldValues["role"] = existingUser.Role.Name
		newValues["role"] = role.Name
		existingUser.RoleID = role.ID
		existingUser.Role = role
		hasChanges = true
	}

	if hasChanges {
		err := ss.userRepo.UpdateUser(ctx, existingUser)
		if err != nil {
			return nil, fmt.Errorf("failed to update user: %v", err)
		}

		// Log de auditoría
		ss.logAuditEvent(ctx, "user_updated", "user", existingUser.ID, existingUser.ID,
			map[string]interface{}{
				"old_values": oldValues,
				"new_values": newValues,
			})
	}

	return existingUser, nil
}

// ValidateUserRole valida que un usuario tenga permisos para realizar una acción
func (ss *SyncService) ValidateUserRole(ctx context.Context, firebaseUID string, requiredPermissions []string) (*models.User, error) {
	user, err := ss.userRepo.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	if !user.IsActive {
		return nil, fmt.Errorf("user account is disabled")
	}

	// Verificar permisos
	if user.Role.Name == models.RoleAdministrador {
		// Los administradores tienen todos los permisos
		return user, nil
	}

	// Verificar permisos específicos
	for _, required := range requiredPermissions {
		if !contains(user.Role.Permissions, required) {
			return nil, fmt.Errorf("user does not have required permission: %s", required)
		}
	}

	return user, nil
}

// SyncUserDeviceToken sincroniza un token de dispositivo FCM
func (ss *SyncService) SyncUserDeviceToken(ctx context.Context, firebaseUID, token string, deviceInfo map[string]interface{}) error {
	// Agregar token al usuario en PostgreSQL
	err := ss.userRepo.AddDeviceToken(ctx, firebaseUID, token)
	if err != nil {
		return fmt.Errorf("failed to add device token: %v", err)
	}

	// Log de auditoría
	user, _ := ss.userRepo.GetUserByFirebaseUID(ctx, firebaseUID)
	if user != nil {
		ss.logAuditEvent(ctx, "device_token_added", "user", user.ID, user.ID,
			map[string]interface{}{
				"token":       token,
				"device_info": deviceInfo,
			})
	}

	return nil
}

// RemoveUserDeviceToken remueve un token de dispositivo
func (ss *SyncService) RemoveUserDeviceToken(ctx context.Context, firebaseUID, token string) error {
	err := ss.userRepo.RemoveDeviceToken(ctx, firebaseUID, token)
	if err != nil {
		return fmt.Errorf("failed to remove device token: %v", err)
	}

	// Log de auditoría
	user, _ := ss.userRepo.GetUserByFirebaseUID(ctx, firebaseUID)
	if user != nil {
		ss.logAuditEvent(ctx, "device_token_removed", "user", user.ID, user.ID,
			map[string]interface{}{
				"token": token,
			})
	}

	return nil
}

// GetUserPermissions obtiene los permisos de un usuario
func (ss *SyncService) GetUserPermissions(ctx context.Context, firebaseUID string) ([]string, error) {
	user, err := ss.userRepo.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %v", err)
	}

	if user.Role.Name == models.RoleAdministrador {
		return []string{"all"}, nil
	}

	return user.Role.Permissions, nil
}

// UpdateUserLastLogin actualiza el último login del usuario
func (ss *SyncService) UpdateUserLastLogin(ctx context.Context, firebaseUID string, sessionInfo map[string]interface{}) error {
	err := ss.userRepo.UpdateLastLogin(ctx, firebaseUID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %v", err)
	}

	// Log de auditoría
	user, _ := ss.userRepo.GetUserByFirebaseUID(ctx, firebaseUID)
	if user != nil {
		ss.logAuditEvent(ctx, "user_login", "user", user.ID, user.ID, sessionInfo)
	}

	return nil
}

// logAuditEvent registra un evento de auditoría
func (ss *SyncService) logAuditEvent(ctx context.Context, eventType, entityType string, entityID, userID int, details map[string]interface{}) {
	if ss.auditRepo == nil {
		return
	}

	auditLog := &models.AuditLog{
		EventType:     eventType,
		EntityType:    &entityType,
		EntityID:      &entityID,
		UserID:        &userID,
		ActionDetails: details,
	}

	// Este método no debería fallar la operación principal
	if err := ss.auditRepo.CreateAuditLog(ctx, auditLog); err != nil {
		log.Printf("Failed to create audit log: %v", err)
	}
}

// Función helper para verificar si un slice contiene un string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
