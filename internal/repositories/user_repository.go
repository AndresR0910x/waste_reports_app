package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/pkg/database"
)

type UserRepository struct {
	db *database.DB
}

func NewUserRepository(db *database.DB) *UserRepository {
	return &UserRepository{db: db}
}

// CreateUser crea un nuevo usuario sincronizado con Firebase
func (ur *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	query := `
		INSERT INTO users (firebase_uid, email, role_id, cedula, phone, full_name, device_tokens)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at, updated_at, anonymous_id`

	err := ur.db.QueryRowContext(ctx, query,
		user.FirebaseUID, user.Email, user.RoleID, user.Cedula,
		user.Phone, user.FullName, user.DeviceTokens).Scan(
		&user.ID, &user.CreatedAt, &user.UpdatedAt, &user.AnonymousID)

	if err != nil {
		return fmt.Errorf("failed to create user: %v", err)
	}

	return nil
}

// GetUserByFirebaseUID obtiene un usuario por su UID de Firebase
func (ur *UserRepository) GetUserByFirebaseUID(ctx context.Context, firebaseUID string) (*models.User, error) {
	query := `
		SELECT u.id, u.firebase_uid, u.email, u.role_id, u.cedula, u.phone, 
			   u.full_name, u.is_active, u.last_login, u.device_tokens, 
			   u.created_at, u.updated_at, u.anonymous_id,
			   r.name as role_name, r.description as role_description, r.permissions as role_permissions
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.firebase_uid = $1 AND u.is_active = true`

	user := &models.User{Role: &models.Role{}}

	err := ur.db.QueryRowContext(ctx, query, firebaseUID).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.RoleID, &user.Cedula,
		&user.Phone, &user.FullName, &user.IsActive, &user.LastLogin, &user.DeviceTokens,
		&user.CreatedAt, &user.UpdatedAt, &user.AnonymousID,
		&user.Role.Name, &user.Role.Description, &user.Role.Permissions)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	user.Role.ID = user.RoleID
	return user, nil
}

// GetUserByCedula obtiene un usuario por su cédula
func (ur *UserRepository) GetUserByCedula(ctx context.Context, cedula string) (*models.User, error) {
	query := `
		SELECT u.id, u.firebase_uid, u.email, u.role_id, u.cedula, u.phone, 
			   u.full_name, u.is_active, u.last_login, u.device_tokens, 
			   u.created_at, u.updated_at, u.anonymous_id,
			   r.name as role_name, r.description as role_description, r.permissions as role_permissions
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.cedula = $1 AND u.is_active = true`

	user := &models.User{Role: &models.Role{}}

	err := ur.db.QueryRowContext(ctx, query, cedula).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.RoleID, &user.Cedula,
		&user.Phone, &user.FullName, &user.IsActive, &user.LastLogin, &user.DeviceTokens,
		&user.CreatedAt, &user.UpdatedAt, &user.AnonymousID,
		&user.Role.Name, &user.Role.Description, &user.Role.Permissions)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	user.Role.ID = user.RoleID
	return user, nil
}

// GetUserByID obtiene un usuario por su ID
func (ur *UserRepository) GetUserByID(ctx context.Context, userID int) (*models.User, error) {
	query := `
		SELECT u.id, u.firebase_uid, u.email, u.role_id, u.cedula, u.phone, 
			   u.full_name, u.is_active, u.last_login, u.device_tokens, 
			   u.created_at, u.updated_at, u.anonymous_id,
			   r.name as role_name, r.description as role_description, r.permissions as role_permissions
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.id = $1 AND u.is_active = true`

	user := &models.User{Role: &models.Role{}}
	err := ur.db.QueryRowContext(ctx, query, userID).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.RoleID, &user.Cedula,
		&user.Phone, &user.FullName, &user.IsActive, &user.LastLogin, &user.DeviceTokens,
		&user.CreatedAt, &user.UpdatedAt, &user.AnonymousID,
		&user.Role.Name, &user.Role.Description, &user.Role.Permissions,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user: %v", err)
	}

	user.Role.ID = user.RoleID
	return user, nil
}

// UpdateUser actualiza un usuario
func (ur *UserRepository) UpdateUser(ctx context.Context, user *models.User) error {
	query := `
		UPDATE users 
		SET email = $2, role_id = $3, cedula = $4, phone = $5, full_name = $6, 
			device_tokens = $7, is_active = $8, updated_at = CURRENT_TIMESTAMP
		WHERE firebase_uid = $1`

	result, err := ur.db.ExecContext(ctx, query,
		user.FirebaseUID, user.Email, user.RoleID, user.Cedula,
		user.Phone, user.FullName, user.DeviceTokens, user.IsActive)

	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

// UpdateLastLogin actualiza el último login del usuario
func (ur *UserRepository) UpdateLastLogin(ctx context.Context, firebaseUID string) error {
	query := `UPDATE users SET last_login = CURRENT_TIMESTAMP WHERE firebase_uid = $1`

	_, err := ur.db.ExecContext(ctx, query, firebaseUID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %v", err)
	}

	return nil
}

// GetUsersByRole obtiene usuarios por rol
func (ur *UserRepository) GetUsersByRole(ctx context.Context, roleName string) ([]*models.User, error) {
	query := `
		SELECT u.id, u.firebase_uid, u.email, u.role_id, u.cedula, u.phone, 
			   u.full_name, u.is_active, u.last_login, u.device_tokens, 
			   u.created_at, u.updated_at, u.anonymous_id,
			   r.name as role_name, r.description as role_description, r.permissions as role_permissions
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE r.name = $1 AND u.is_active = true
		ORDER BY u.created_at DESC`

	rows, err := ur.db.QueryContext(ctx, query, roleName)
	if err != nil {
		return nil, fmt.Errorf("failed to get users by role: %v", err)
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		user := &models.User{Role: &models.Role{}}

		err := rows.Scan(
			&user.ID, &user.FirebaseUID, &user.Email, &user.RoleID, &user.Cedula,
			&user.Phone, &user.FullName, &user.IsActive, &user.LastLogin, &user.DeviceTokens,
			&user.CreatedAt, &user.UpdatedAt, &user.AnonymousID,
			&user.Role.Name, &user.Role.Description, &user.Role.Permissions)

		if err != nil {
			return nil, fmt.Errorf("failed to scan user: %v", err)
		}

		user.Role.ID = user.RoleID
		users = append(users, user)
	}

	return users, nil
}

// AddDeviceToken agrega un token de dispositivo al usuario
func (ur *UserRepository) AddDeviceToken(ctx context.Context, firebaseUID, token string) error {
	// Primero verificar si el token ya existe
	query := `
		UPDATE users 
		SET device_tokens = array_append(
			array_remove(device_tokens, $2), $2
		)
		WHERE firebase_uid = $1`

	_, err := ur.db.ExecContext(ctx, query, firebaseUID, token)
	if err != nil {
		return fmt.Errorf("failed to add device token: %v", err)
	}

	return nil
}

// RemoveDeviceToken remueve un token de dispositivo del usuario
func (ur *UserRepository) RemoveDeviceToken(ctx context.Context, firebaseUID, token string) error {
	query := `
		UPDATE users 
		SET device_tokens = array_remove(device_tokens, $2)
		WHERE firebase_uid = $1`

	_, err := ur.db.ExecContext(ctx, query, firebaseUID, token)
	if err != nil {
		return fmt.Errorf("failed to remove device token: %v", err)
	}

	return nil
}

// GetAllRoles obtiene todos los roles disponibles
func (ur *UserRepository) GetAllRoles(ctx context.Context) ([]*models.Role, error) {
	query := `SELECT id, name, description, permissions, created_at, updated_at FROM roles ORDER BY name`

	rows, err := ur.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get roles: %v", err)
	}
	defer rows.Close()

	var roles []*models.Role
	for rows.Next() {
		role := &models.Role{}

		err := rows.Scan(&role.ID, &role.Name, &role.Description,
			&role.Permissions, &role.CreatedAt, &role.UpdatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan role: %v", err)
		}

		roles = append(roles, role)
	}

	return roles, nil
}

// GetRoleByName obtiene un rol por nombre
func (ur *UserRepository) GetRoleByName(ctx context.Context, name string) (*models.Role, error) {
	query := `SELECT id, name, description, permissions, created_at, updated_at FROM roles WHERE name = $1`

	role := &models.Role{}
	err := ur.db.QueryRowContext(ctx, query, name).Scan(
		&role.ID, &role.Name, &role.Description,
		&role.Permissions, &role.CreatedAt, &role.UpdatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role: %v", err)
	}

	return role, nil
}

// SyncUserWithFirebase sincroniza los datos del usuario con Firebase
func (ur *UserRepository) SyncUserWithFirebase(ctx context.Context, firebaseUID string, email *string, displayName *string) error {
	// Buscar si el usuario ya existe
	existingUser, err := ur.GetUserByFirebaseUID(ctx, firebaseUID)
	if err != nil && !strings.Contains(err.Error(), "user not found") {
		return fmt.Errorf("failed to check existing user: %v", err)
	}

	if existingUser == nil {
		// Usuario no existe, crear uno nuevo con rol de ciudadano por defecto
		role, err := ur.GetRoleByName(ctx, models.RoleCiudadano)
		if err != nil {
			return fmt.Errorf("failed to get default role: %v", err)
		}

		newUser := &models.User{
			FirebaseUID: firebaseUID,
			Email:       email,
			FullName:    displayName,
			RoleID:      role.ID,
			IsActive:    true,
		}

		return ur.CreateUser(ctx, newUser)
	} else {
		// Usuario existe, actualizar información si es necesario
		if (email != nil && existingUser.Email != email) ||
			(displayName != nil && existingUser.FullName != displayName) {
			existingUser.Email = email
			existingUser.FullName = displayName
			return ur.UpdateUser(ctx, existingUser)
		}
	}

	return nil
}

// GetUserByEmail obtiene un usuario por su email
func (ur *UserRepository) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `
		SELECT u.id, u.firebase_uid, u.email, u.role_id, u.cedula, u.phone, 
			   u.full_name, u.is_active, u.last_login, u.device_tokens, 
			   u.created_at, u.updated_at, u.anonymous_id
		FROM users u
		WHERE u.email = $1`

	user := &models.User{}
	err := ur.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.FirebaseUID, &user.Email, &user.RoleID, &user.Cedula,
		&user.Phone, &user.FullName, &user.IsActive, &user.LastLogin,
		&user.DeviceTokens, &user.CreatedAt, &user.UpdatedAt, &user.AnonymousID,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by email: %v", err)
	}

	return user, nil
}

// GetRoleByID obtiene un rol por su ID
func (ur *UserRepository) GetRoleByID(ctx context.Context, roleID int) (*models.Role, error) {
	query := `
		SELECT id, name, description, permissions, created_at, updated_at
		FROM roles 
		WHERE id = $1`

	role := &models.Role{}
	err := ur.db.QueryRowContext(ctx, query, roleID).Scan(
		&role.ID, &role.Name, &role.Description, &role.Permissions,
		&role.CreatedAt, &role.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role by ID: %v", err)
	}

	return role, nil
}
