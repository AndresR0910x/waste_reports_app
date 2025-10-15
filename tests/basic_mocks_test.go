package tests

import (
	"testing"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/tests/mocks"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// BasicTestSuite define la suite de tests básicos
type BasicTestSuite struct {
	suite.Suite
}

// TestCedulaValidationMock prueba que los mocks de validación de cédula funcionen
func (suite *BasicTestSuite) TestCedulaValidationMock() {
	// Crear mock
	mockService := new(mocks.MockCedulaValidationService)

	// Configurar expectativa
	expectedResponse := &models.CedulaValidationResponse{
		IsValid:  true,
		Cedula:   "1713175071",
		FullName: "Juan Pérez",
		Province: "Pichincha",
		Message:  "Cédula válida",
		Source:   "algorithm",
	}

	mockService.On("ValidateCedula", "1713175071").Return(expectedResponse, nil)

	// Ejecutar función
	result, err := mockService.ValidateCedula("1713175071")

	// Verificar resultado
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), result)
	assert.Equal(suite.T(), expectedResponse.IsValid, result.IsValid)
	assert.Equal(suite.T(), expectedResponse.Cedula, result.Cedula)
	assert.Equal(suite.T(), expectedResponse.FullName, result.FullName)

	// Verificar que se llamó el mock
	mockService.AssertExpectations(suite.T())
}

// TestValidationServiceMock prueba que el mock de validación funcione
func (suite *BasicTestSuite) TestValidationServiceMock() {
	// Crear mock
	mockService := new(mocks.MockValidationService)

	// Configurar expectativa
	validationErrors := []models.ValidationError{
		{Field: "title", Message: "El título es requerido"},
	}

	mockService.On("ValidateStruct", &models.CreateReportRequest{}).Return(validationErrors)

	// Ejecutar función
	result := mockService.ValidateStruct(&models.CreateReportRequest{})

	// Verificar resultado
	assert.NotNil(suite.T(), result)
	assert.Len(suite.T(), result, 1)
	assert.Equal(suite.T(), "title", result[0].Field)
	assert.Equal(suite.T(), "El título es requerido", result[0].Message)

	// Verificar que se llamó el mock
	mockService.AssertExpectations(suite.T())
}

// TestBasicMocks ejecuta la suite de tests básicos
func TestBasicMocks(t *testing.T) {
	suite.Run(t, new(BasicTestSuite))
}
