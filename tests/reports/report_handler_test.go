package tests

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend-residuos-app/internal/handlers"
	"backend-residuos-app/internal/models"
	"backend-residuos-app/pkg/middleware"
	"backend-residuos-app/tests/mocks"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

// ReportHandlerTestSuite define la suite de tests
type ReportHandlerTestSuite struct {
	suite.Suite
	router            *gin.Engine
	reportService     *mocks.MockReportService
	cedulaService     *mocks.MockCedulaValidationService
	photoService      *mocks.MockPhotoUploadService
	validationService *mocks.MockValidationService
	handler           *handlers.ReportHandler
}

// SetupTest configura cada test
func (suite *ReportHandlerTestSuite) SetupTest() {
	gin.SetMode(gin.TestMode)

	// Crear mocks
	suite.reportService = new(mocks.MockReportService)
	suite.cedulaService = new(mocks.MockCedulaValidationService)
	suite.photoService = new(mocks.MockPhotoUploadService)
	suite.validationService = new(mocks.MockValidationService)

	// Crear handler
	suite.handler = handlers.NewReportHandler(
		suite.reportService,
		suite.cedulaService,
		suite.photoService,
		suite.validationService,
	)

	// Configurar router
	suite.router = gin.New()
	suite.router.Use(func(c *gin.Context) {
		// Mock del middleware de autenticación
		mockUser := &models.User{
			ID:       1,
			RoleID:   1, // Ciudadano
			Email:    stringPtr("test@example.com"),
			FullName: stringPtr("Test User"),
			Cedula:   stringPtr("1713175071"),
		}
		c.Set(middleware.UserContextKey, mockUser)
		c.Next()
	})

	suite.router.POST("/reports/create", suite.handler.CreateReport)
	suite.router.POST("/reports/validate-cedula", suite.handler.ValidateCedula)
	suite.router.GET("/reports/my-reports", suite.handler.GetReportsByUser)
	suite.router.GET("/reports/:id", suite.handler.GetReportByID)
}

// TestCreateReport_Success prueba la creación exitosa de un reporte
func (suite *ReportHandlerTestSuite) TestCreateReport_Success() {
	// Preparar datos de entrada
	request := models.CreateReportRequest{
		Title:       "Basura en la calle",
		Description: "Hay mucha basura acumulada en la esquina",
		Category:    "limpieza_publica",
		Location: models.Location{
			Latitude:  -0.1807,
			Longitude: -78.4678,
		},
		Address:        stringPtr("Av. Amazonas y Naciones Unidas"),
		Priority:       "media",
		CedulaValidate: true,
	}

	// Configurar mocks
	suite.validationService.On("ValidateStruct", mock.AnythingOfType("*models.CreateReportRequest")).
		Return([]models.ValidationError{})

	suite.cedulaService.On("ValidateCedula", "1713175071").
		Return(&models.CedulaValidationResponse{
			IsValid: true,
			Cedula:  "1713175071",
			Message: "Cédula válida",
			Source:  "algorithm",
		}, nil)

	expectedReport := &models.Report{
		ID:          1,
		UserID:      1,
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		Location:    request.Location,
		Address:     request.Address,
		Status:      "pending",
		Priority:    request.Priority,
		SyncStatus:  "synced",
	}

	suite.reportService.On("CreateReport", mock.AnythingOfType("*models.Report")).
		Return(expectedReport, nil)

	// Preparar request HTTP
	jsonBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/reports/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusCreated, w.Code)

	var response models.CreateReportResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedReport.ID, response.ID)
	assert.Equal(suite.T(), expectedReport.Title, response.Title)
	assert.Equal(suite.T(), "Reporte creado exitosamente", response.Message)

	// Verificar que se llamaron los mocks
	suite.validationService.AssertExpectations(suite.T())
	suite.cedulaService.AssertExpectations(suite.T())
	suite.reportService.AssertExpectations(suite.T())
}

// TestCreateReport_ValidationError prueba errores de validación
func (suite *ReportHandlerTestSuite) TestCreateReport_ValidationError() {
	request := models.CreateReportRequest{
		Title:       "Ab",   // Muy corto
		Description: "Desc", // Muy corto
		Category:    "invalid_category",
		Location:    models.Location{Latitude: 200, Longitude: 200}, // Inválido
	}

	// Configurar mock para devolver errores de validación
	validationErrors := []models.ValidationError{
		{Field: "title", Message: "El título debe tener al menos 5 caracteres"},
		{Field: "description", Message: "La descripción debe tener al menos 10 caracteres"},
		{Field: "category", Message: "Categoría inválida"},
	}

	suite.validationService.On("ValidateStruct", mock.AnythingOfType("*models.CreateReportRequest")).
		Return(validationErrors)

	// Preparar request HTTP
	jsonBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/reports/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Errores de validación", response["error"])
	assert.Equal(suite.T(), "VALIDATION_ERROR", response["code"])
	assert.NotNil(suite.T(), response["validation_errors"])

	suite.validationService.AssertExpectations(suite.T())
}

// TestCreateReport_InvalidCedula prueba cédula inválida
func (suite *ReportHandlerTestSuite) TestCreateReport_InvalidCedula() {
	request := models.CreateReportRequest{
		Title:       "Basura en la calle",
		Description: "Hay mucha basura acumulada en la esquina",
		Category:    "limpieza_publica",
		Location: models.Location{
			Latitude:  -0.1807,
			Longitude: -78.4678,
		},
		CedulaValidate: true,
	}

	// Configurar mocks
	suite.validationService.On("ValidateStruct", mock.AnythingOfType("*models.CreateReportRequest")).
		Return([]models.ValidationError{})

	suite.cedulaService.On("ValidateCedula", "1713175071").
		Return(&models.CedulaValidationResponse{
			IsValid: false,
			Cedula:  "1713175071",
			Message: "Cédula inválida según algoritmo de verificación",
			Source:  "algorithm",
		}, nil)

	// Preparar request HTTP
	jsonBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/reports/create", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusForbidden, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Cédula inválida", response["error"])
	assert.Equal(suite.T(), "INVALID_CEDULA", response["code"])

	suite.validationService.AssertExpectations(suite.T())
	suite.cedulaService.AssertExpectations(suite.T())
}

// TestValidateCedula_Success prueba la validación exitosa de cédula
func (suite *ReportHandlerTestSuite) TestValidateCedula_Success() {
	request := models.CedulaValidationRequest{
		Cedula: "1713175071",
	}

	expectedResponse := &models.CedulaValidationResponse{
		IsValid:  true,
		Cedula:   "1713175071",
		FullName: "Juan Pérez",
		Province: "Pichincha",
		Message:  "Cédula válida",
		Source:   "algorithm",
	}

	// Configurar mocks
	suite.cedulaService.On("IsValidCedulaFormat", "1713175071").Return(true)
	suite.cedulaService.On("ValidateCedula", "1713175071").Return(expectedResponse, nil)

	// Preparar request HTTP
	jsonBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/reports/validate-cedula", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.CedulaValidationResponse
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedResponse.IsValid, response.IsValid)
	assert.Equal(suite.T(), expectedResponse.Cedula, response.Cedula)
	assert.Equal(suite.T(), expectedResponse.Message, response.Message)

	suite.cedulaService.AssertExpectations(suite.T())
}

// TestValidateCedula_InvalidFormat prueba formato inválido de cédula
func (suite *ReportHandlerTestSuite) TestValidateCedula_InvalidFormat() {
	request := models.CedulaValidationRequest{
		Cedula: "123456789", // Muy corto
	}

	// Configurar mock
	suite.cedulaService.On("IsValidCedulaFormat", "123456789").Return(false)

	// Preparar request HTTP
	jsonBody, _ := json.Marshal(request)
	req := httptest.NewRequest("POST", "/reports/validate-cedula", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Formato de cédula inválido", response["error"])
	assert.Equal(suite.T(), "INVALID_FORMAT", response["code"])

	suite.cedulaService.AssertExpectations(suite.T())
}

// TestGetReportsByUser_Success prueba obtener reportes del usuario
func (suite *ReportHandlerTestSuite) TestGetReportsByUser_Success() {
	expectedReports := []*models.Report{
		{
			ID:          1,
			UserID:      1,
			Title:       "Reporte 1",
			Description: "Descripción 1",
			Status:      "pending",
		},
		{
			ID:          2,
			UserID:      1,
			Title:       "Reporte 2",
			Description: "Descripción 2",
			Status:      "resolved",
		},
	}

	// Configurar mocks
	suite.reportService.On("GetReportsByUserID", 1, 10, 0).Return(expectedReports, nil)
	suite.reportService.On("GetReportsCountByUserID", 1).Return(2, nil)

	// Preparar request HTTP
	req := httptest.NewRequest("GET", "/reports/my-reports", nil)
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response["reports"])
	assert.NotNil(suite.T(), response["pagination"])

	suite.reportService.AssertExpectations(suite.T())
}

// TestGetReportByID_Success prueba obtener un reporte por ID
func (suite *ReportHandlerTestSuite) TestGetReportByID_Success() {
	expectedReport := &models.Report{
		ID:          1,
		UserID:      1,
		Title:       "Reporte de prueba",
		Description: "Descripción detallada",
		Status:      "pending",
	}

	// Configurar mock
	suite.reportService.On("GetReportByID", 1).Return(expectedReport, nil)

	// Preparar request HTTP
	req := httptest.NewRequest("GET", "/reports/1", nil)
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response models.Report
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), expectedReport.ID, response.ID)
	assert.Equal(suite.T(), expectedReport.Title, response.Title)

	suite.reportService.AssertExpectations(suite.T())
}

// TestGetReportByID_NotFound prueba reporte no encontrado
func (suite *ReportHandlerTestSuite) TestGetReportByID_NotFound() {
	// Configurar mock
	suite.reportService.On("GetReportByID", 999).Return(nil, errors.New("report not found"))

	// Preparar request HTTP
	req := httptest.NewRequest("GET", "/reports/999", nil)
	w := httptest.NewRecorder()

	// Ejecutar request
	suite.router.ServeHTTP(w, req)

	// Verificar respuesta
	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), "Reporte no encontrado", response["error"])
	assert.Equal(suite.T(), "NOT_FOUND", response["code"])

	suite.reportService.AssertExpectations(suite.T())
}

// TestReportHandler ejecuta toda la suite
func TestReportHandler(t *testing.T) {
	suite.Run(t, new(ReportHandlerTestSuite))
}

// Función auxiliar
func stringPtr(s string) *string {
	return &s
}
