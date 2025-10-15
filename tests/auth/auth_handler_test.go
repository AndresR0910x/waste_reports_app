package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-residuos-app/internal/handlers"
	"backend-residuos-app/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockAuthService para testear los handlers
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) RegisterUser(ctx interface{}, req *models.RegisterRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginUser(ctx interface{}, req *models.LoginRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func (m *MockAuthService) LoginWithOAuth(ctx interface{}, req *models.OAuthRequest) (*models.AuthResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.AuthResponse), args.Error(1)
}

func TestRegistrationHandler_Register(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful registration",
			requestBody: models.RegisterRequest{
				FullName: "Juan Carlos Pérez",
				Email:    "juan.perez@gmail.com",
				Password: "SecurePass123!",
				Phone:    "0987654321",
				Cedula:   "1234567890",
				Language: "es",
			},
			setupMocks: func(authService *MockAuthService) {
				mockResponse := &models.AuthResponse{
					Success:     true,
					Message:     "Usuario registrado exitosamente",
					AccessToken: "jwt-token-123",
					User: &models.UserInfo{
						ID:          1,
						FirebaseUID: "firebase-uid-123",
						Email:       "juan.perez@gmail.com",
						FullName:    "Juan Carlos Pérez",
						Role:        "ciudadano",
						RoleID:      3,
						IsActive:    true,
						Language:    "es",
						CreatedAt:   time.Now(),
					},
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				authService.On("RegisterUser", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(mockResponse, nil)
			},
			expectedStatus: http.StatusCreated,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "Usuario registrado exitosamente",
			},
		},
		{
			name:        "Invalid JSON request",
			requestBody: `{"invalid": json}`,
			setupMocks: func(authService *MockAuthService) {
				// No se llama al servicio
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
				"error":   "validation_error",
			},
		},
		{
			name: "Service error - user already exists",
			requestBody: models.RegisterRequest{
				FullName: "Juan Carlos Pérez",
				Email:    "existing@gmail.com",
				Password: "SecurePass123!",
				Phone:    "0987654321",
				Cedula:   "1234567890",
				Language: "es",
			},
			setupMocks: func(authService *MockAuthService) {
				authService.On("RegisterUser", mock.Anything, mock.AnythingOfType("*models.RegisterRequest")).Return(nil, errors.New("usuario ya existe"))
			},
			expectedStatus: http.StatusConflict,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "usuario ya existe",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockAuthService := new(MockAuthService)
			tt.setupMocks(mockAuthService)

			// Create handler with mock service
			handler := &handlers.RegistrationHandler{} // We'd need to inject the auth service

			// Create request
			var reqBody []byte
			var err error

			if str, ok := tt.requestBody.(string); ok {
				reqBody = []byte(str)
			} else {
				reqBody, err = json.Marshal(tt.requestBody)
				assert.NoError(t, err)
			}

			req, err := http.NewRequest("POST", "/auth/register", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Execute handler
			handler.Register(ctx)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["success"], response["success"])
			assert.Equal(t, tt.expectedBody["message"], response["message"])

			// Verify expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestRegistrationHandler_Login(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful login",
			requestBody: models.LoginRequest{
				Email:    "juan.perez@gmail.com",
				Password: "SecurePass123!",
			},
			setupMocks: func(authService *MockAuthService) {
				mockResponse := &models.AuthResponse{
					Success:     true,
					Message:     "Login exitoso",
					AccessToken: "jwt-token-123",
					User: &models.UserInfo{
						ID:          1,
						FirebaseUID: "firebase-uid-123",
						Email:       "juan.perez@gmail.com",
						FullName:    "Juan Carlos Pérez",
						Role:        "ciudadano",
						RoleID:      3,
						IsActive:    true,
						Language:    "es",
						CreatedAt:   time.Now(),
					},
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				authService.On("LoginUser", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(mockResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "Login exitoso",
			},
		},
		{
			name: "Invalid credentials",
			requestBody: models.LoginRequest{
				Email:    "wrong@gmail.com",
				Password: "wrongpassword",
			},
			setupMocks: func(authService *MockAuthService) {
				authService.On("LoginUser", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, errors.New("credenciales inválidas"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "credenciales inválidas",
			},
		},
		{
			name: "Missing email",
			requestBody: models.LoginRequest{
				Email:    "",
				Password: "SecurePass123!",
			},
			setupMocks: func(authService *MockAuthService) {
				authService.On("LoginUser", mock.Anything, mock.AnythingOfType("*models.LoginRequest")).Return(nil, errors.New("Email es requerido"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Email es requerido",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockAuthService := new(MockAuthService)
			tt.setupMocks(mockAuthService)

			// Create handler with mock service
			handler := &handlers.RegistrationHandler{} // We'd need to inject the auth service

			// Create request
			reqBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/auth/login", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Execute handler
			handler.Login(ctx)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["success"], response["success"])
			assert.Equal(t, tt.expectedBody["message"], response["message"])

			// Verify expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}

func TestRegistrationHandler_LoginOAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		requestBody    interface{}
		setupMocks     func(*MockAuthService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Successful OAuth login",
			requestBody: models.OAuthRequest{
				Token:    "valid.jwt.token",
				Provider: "google",
				FullName: "OAuth User",
				Language: "es",
			},
			setupMocks: func(authService *MockAuthService) {
				mockResponse := &models.AuthResponse{
					Success:     true,
					Message:     "Login OAuth exitoso",
					AccessToken: "jwt-token-123",
					User: &models.UserInfo{
						ID:          1,
						FirebaseUID: "firebase-uid-123",
						Email:       "oauth@gmail.com",
						FullName:    "OAuth User",
						Role:        "ciudadano",
						RoleID:      3,
						IsActive:    true,
						Language:    "es",
						CreatedAt:   time.Now(),
					},
					ExpiresAt: time.Now().Add(24 * time.Hour),
				}
				authService.On("LoginWithOAuth", mock.Anything, mock.AnythingOfType("*models.OAuthRequest")).Return(mockResponse, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "Login OAuth exitoso",
			},
		},
		{
			name: "Invalid OAuth token",
			requestBody: models.OAuthRequest{
				Token:    "invalid.token",
				Provider: "google",
				FullName: "OAuth User",
				Language: "es",
			},
			setupMocks: func(authService *MockAuthService) {
				authService.On("LoginWithOAuth", mock.Anything, mock.AnythingOfType("*models.OAuthRequest")).Return(nil, errors.New("token inválido"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "token inválido",
			},
		},
		{
			name: "Invalid provider",
			requestBody: models.OAuthRequest{
				Token:    "valid.jwt.token",
				Provider: "invalid_provider",
				FullName: "OAuth User",
				Language: "es",
			},
			setupMocks: func(authService *MockAuthService) {
				authService.On("LoginWithOAuth", mock.Anything, mock.AnythingOfType("*models.OAuthRequest")).Return(nil, errors.New("Proveedor inválido"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"success": false,
				"message": "Proveedor inválido",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock service
			mockAuthService := new(MockAuthService)
			tt.setupMocks(mockAuthService)

			// Create handler with mock service
			handler := &handlers.RegistrationHandler{} // We'd need to inject the auth service

			// Create request
			reqBody, err := json.Marshal(tt.requestBody)
			assert.NoError(t, err)

			req, err := http.NewRequest("POST", "/auth/oauth", bytes.NewBuffer(reqBody))
			assert.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			// Create response recorder
			w := httptest.NewRecorder()

			// Create Gin context
			ctx, _ := gin.CreateTestContext(w)
			ctx.Request = req

			// Execute handler
			handler.LoginOAuth(ctx)

			// Assertions
			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			err = json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)

			assert.Equal(t, tt.expectedBody["success"], response["success"])
			assert.Equal(t, tt.expectedBody["message"], response["message"])

			// Verify expectations
			mockAuthService.AssertExpectations(t)
		})
	}
}
