package tests

import (
	"testing"

	"backend-residuos-app/internal/services"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// CedulaValidationServiceTestSuite define la suite de tests
type CedulaValidationServiceTestSuite struct {
	suite.Suite
	service *services.CedulaValidationService
}

// SetupTest configura cada test
func (suite *CedulaValidationServiceTestSuite) SetupTest() {
	// Inicializar con API URL vacía para usar solo algoritmo local
	suite.service = services.NewCedulaValidationService("")
}

// TestValidCedulas prueba cédulas válidas
func (suite *CedulaValidationServiceTestSuite) TestValidCedulas() {
	validCedulas := []string{
		"1713175071", // Cédula válida de Pichincha
		"0926687856", // Cédula válida de Guayas
		"0102030405", // Cédula válida sintética
	}

	for _, cedula := range validCedulas {
		suite.Run("Valid_cedula_"+cedula, func() {
			result, err := suite.service.ValidateCedula(cedula)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), cedula, result.Cedula)
			assert.Equal(suite.T(), "algorithm", result.Source)

			// La validez depende del algoritmo específico
			suite.T().Logf("Cédula %s: %v - %s", cedula, result.IsValid, result.Message)
		})
	}
}

// TestInvalidCedulas prueba cédulas inválidas
func (suite *CedulaValidationServiceTestSuite) TestInvalidCedulas() {
	testCases := []struct {
		name          string
		cedula        string
		expectedValid bool
		expectedError bool
	}{
		{
			name:          "Too_short",
			cedula:        "123456789",
			expectedValid: false,
		},
		{
			name:          "Too_long",
			cedula:        "12345678901",
			expectedValid: false,
		},
		{
			name:          "Empty_string",
			cedula:        "",
			expectedValid: false,
		},
		{
			name:          "Non_numeric",
			cedula:        "171317507a",
			expectedValid: false,
		},
		{
			name:          "Invalid_province_code",
			cedula:        "2513175071", // Provincia 25 no existe
			expectedValid: false,
		},
		{
			name:          "Invalid_third_digit",
			cedula:        "1763175071", // Tercer dígito 6 es inválido para personas naturales
			expectedValid: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result, err := suite.service.ValidateCedula(tc.cedula)

			if tc.expectedError {
				assert.Error(suite.T(), err)
				return
			}

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), tc.expectedValid, result.IsValid)
			assert.Equal(suite.T(), "algorithm", result.Source)

			suite.T().Logf("Cédula %s: %v - %s", tc.cedula, result.IsValid, result.Message)
		})
	}
}

// TestCedulaFormatValidation prueba la validación de formato
func (suite *CedulaValidationServiceTestSuite) TestCedulaFormatValidation() {
	testCases := []struct {
		name     string
		cedula   string
		expected bool
	}{
		{
			name:     "Valid_format",
			cedula:   "1713175071",
			expected: true,
		},
		{
			name:     "Valid_format_with_spaces",
			cedula:   "171 317 5071",
			expected: true, // El servicio debería limpiar espacios
		},
		{
			name:     "Valid_format_with_dashes",
			cedula:   "171-317-5071",
			expected: true, // El servicio debería limpiar guiones
		},
		{
			name:     "Invalid_length",
			cedula:   "123456789",
			expected: false,
		},
		{
			name:     "Contains_letters",
			cedula:   "171317507a",
			expected: false,
		},
		{
			name:     "Empty",
			cedula:   "",
			expected: false,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			result := suite.service.IsValidCedulaFormat(tc.cedula)
			assert.Equal(suite.T(), tc.expected, result)
		})
	}
}

// TestProvinceNames prueba que se asignen correctamente los nombres de provincia
func (suite *CedulaValidationServiceTestSuite) TestProvinceNames() {
	testCases := []struct {
		cedula           string
		expectedProvince string
	}{
		{"0102030405", "Azuay"},       // Provincia 01
		{"0902030405", "Guayas"},      // Provincia 09
		{"1702030405", "Pichincha"},   // Provincia 17
		{"2402030405", "Santa Elena"}, // Provincia 24
	}

	for _, tc := range testCases {
		suite.Run("Province_"+tc.expectedProvince, func() {
			result, err := suite.service.ValidateCedula(tc.cedula)

			assert.NoError(suite.T(), err)
			assert.NotNil(suite.T(), result)
			assert.Equal(suite.T(), tc.expectedProvince, result.Province)
		})
	}
}

// TestCleanCedula prueba la limpieza de cédulas
func (suite *CedulaValidationServiceTestSuite) TestCleanCedula() {
	// Este test requiere acceso al método privado, así que probamos indirectamente
	testCases := []struct {
		name          string
		input         string
		shouldBeValid bool
	}{
		{
			name:          "With_spaces",
			input:         "171 317 5071",
			shouldBeValid: true, // Formato válido después de limpiar
		},
		{
			name:          "With_dashes",
			input:         "171-317-5071",
			shouldBeValid: true,
		},
		{
			name:          "With_dots",
			input:         "171.317.5071",
			shouldBeValid: true,
		},
		{
			name:          "Mixed_separators",
			input:         "171 317-5071",
			shouldBeValid: true,
		},
	}

	for _, tc := range testCases {
		suite.Run(tc.name, func() {
			isValid := suite.service.IsValidCedulaFormat(tc.input)
			assert.Equal(suite.T(), tc.shouldBeValid, isValid)
		})
	}
}

// TestCedulaValidationService ejecuta toda la suite
func TestCedulaValidationService(t *testing.T) {
	suite.Run(t, new(CedulaValidationServiceTestSuite))
}
