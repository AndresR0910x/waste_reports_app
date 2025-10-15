package auth

import (
	"strings"
	"testing"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/services"

	"github.com/stretchr/testify/assert"
)

func TestValidationService_ValidateRegisterRequest(t *testing.T) {
	validator := services.NewValidationService()

	tests := []struct {
		name        string
		request     *models.RegisterRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid registration request",
			request: &models.RegisterRequest{
				FullName: "Juan Carlos Pérez",
				Email:    "juan.perez@gmail.com",
				Password: "SecurePass123!",
				Phone:    "+51987654321",
				Cedula:   "1234567890",
				Language: "es",
			},
			expectError: false,
		},
		{
			name: "Empty full name",
			request: &models.RegisterRequest{
				FullName: "",
				Email:    "juan.perez@gmail.com",
				Password: "SecurePass123!",
				Phone:    "+51987654321",
				Cedula:   "1234567890",
				Language: "es",
			},
			expectError: true,
			errorMsg:    "Full name is required",
		},
		{
			name: "Invalid email format",
			request: &models.RegisterRequest{
				FullName: "Juan Carlos Pérez",
				Email:    "invalid-email",
				Password: "SecurePass123!",
				Phone:    "+51987654321",
				Cedula:   "1234567890",
				Language: "es",
			},
			expectError: true,
			errorMsg:    "Invalid email format",
		},
		{
			name: "Weak password",
			request: &models.RegisterRequest{
				FullName: "Juan Carlos Pérez",
				Email:    "juan.perez@gmail.com",
				Password: "123",
				Phone:    "+51987654321",
				Cedula:   "1234567890",
				Language: "es",
			},
			expectError: true,
			errorMsg:    "Password must be at least",
		},
		{
			name: "Invalid phone format",
			request: &models.RegisterRequest{
				FullName: "Juan Carlos Pérez",
				Email:    "juan.perez@gmail.com",
				Password: "SecurePass123!",
				Phone:    "12345",
				Cedula:   "1234567890",
				Language: "es",
			},
			expectError: true,
			errorMsg:    "Invalid phone format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateRegisterRequest(tt.request)

			if tt.expectError {
				assert.NotEmpty(t, errors, "Se esperaba al menos un error de validación")
				found := false
				for _, err := range errors {
					if strings.Contains(strings.ToLower(err.Message), strings.ToLower(tt.errorMsg)) {
						found = true
						break
					}
				}
				assert.True(t, found, "No se encontró el mensaje de error esperado: %s", tt.errorMsg)
			} else {
				assert.Empty(t, errors, "No se esperaban errores de validación")
			}
		})
	}
}

func TestValidationService_ValidateLoginRequest(t *testing.T) {
	validator := services.NewValidationService()

	tests := []struct {
		name        string
		request     *models.LoginRequest
		expectError bool
		errorMsg    string
	}{
		{
			name: "Valid login request",
			request: &models.LoginRequest{
				Email:    "juan.perez@gmail.com",
				Password: "SecurePass123!",
			},
			expectError: false,
		},
		{
			name: "Empty email",
			request: &models.LoginRequest{
				Email:    "",
				Password: "SecurePass123!",
			},
			expectError: true,
			errorMsg:    "Email is required",
		},
		{
			name: "Invalid email format",
			request: &models.LoginRequest{
				Email:    "invalid-email",
				Password: "SecurePass123!",
			},
			expectError: true,
			errorMsg:    "Invalid email format",
		},
		{
			name: "Empty password",
			request: &models.LoginRequest{
				Email:    "juan.perez@gmail.com",
				Password: "",
			},
			expectError: true,
			errorMsg:    "Password is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errors := validator.ValidateLoginRequest(tt.request)

			if tt.expectError {
				assert.NotEmpty(t, errors, "Se esperaba al menos un error de validación")
				found := false
				for _, err := range errors {
					if strings.Contains(strings.ToLower(err.Message), strings.ToLower(tt.errorMsg)) {
						found = true
						break
					}
				}
				assert.True(t, found, "No se encontró el mensaje de error esperado: %s", tt.errorMsg)
			} else {
				assert.Empty(t, errors, "No se esperaban errores de validación")
			}
		})
	}
}

// Nota: ValidateOAuthRequest no existe en el ValidationService actual
// Los tests de OAuth se manejarían directamente en el AuthService
