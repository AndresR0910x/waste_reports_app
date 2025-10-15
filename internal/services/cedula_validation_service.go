package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"backend-residuos-app/internal/models"
)

// CedulaValidationService maneja la validación de cédulas ecuatorianas
type CedulaValidationService struct {
	externalAPIURL string
	httpClient     *http.Client
}

// NewCedulaValidationService crea una nueva instancia del servicio
func NewCedulaValidationService(externalAPIURL string) *CedulaValidationService {
	return &CedulaValidationService{
		externalAPIURL: externalAPIURL,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// ExternalAPIResponse representa la respuesta de la API externa
type ExternalAPIResponse struct {
	Success bool `json:"success"`
	Valid   bool `json:"valid"`
	Data    struct {
		Cedula   string `json:"cedula"`
		FullName string `json:"full_name"`
		Province string `json:"province"`
	} `json:"data"`
	Message string `json:"message"`
}

// ValidateCedula valida una cédula ecuatoriana usando API externa con fallback al algoritmo local
func (s *CedulaValidationService) ValidateCedula(cedula string) (*models.CedulaValidationResponse, error) {
	// Limpiar y validar formato básico
	cleanCedula := s.cleanCedula(cedula)
	if len(cleanCedula) != 10 {
		return &models.CedulaValidationResponse{
			IsValid: false,
			Cedula:  cedula,
			Message: "La cédula debe tener exactamente 10 dígitos",
			Source:  "validation",
		}, nil
	}

	// Intentar validación con API externa primero
	if s.externalAPIURL != "" {
		if response, err := s.validateWithExternalAPI(cleanCedula); err == nil {
			return response, nil
		}
		// Si falla la API externa, continuar con algoritmo local
	}

	// Fallback: usar algoritmo local de validación
	return s.validateWithLocalAlgorithm(cleanCedula), nil
}

// validateWithExternalAPI intenta validar usando una API externa
func (s *CedulaValidationService) validateWithExternalAPI(cedula string) (*models.CedulaValidationResponse, error) {
	// Preparar request
	requestBody := map[string]string{
		"cedula": cedula,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("error marshaling request: %w", err)
	}

	// Realizar petición HTTP
	req, err := http.NewRequest("POST", s.externalAPIURL, bytes.NewBuffer(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "CleanCity-Backend/1.0")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status code: %d", resp.StatusCode)
	}

	// Parsear respuesta
	var apiResponse ExternalAPIResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}

	// Construir respuesta
	response := &models.CedulaValidationResponse{
		IsValid: apiResponse.Success && apiResponse.Valid,
		Cedula:  cedula,
		Message: apiResponse.Message,
		Source:  "api",
	}

	if apiResponse.Valid {
		response.FullName = apiResponse.Data.FullName
		response.Province = apiResponse.Data.Province
	}

	return response, nil
}

// validateWithLocalAlgorithm implementa el algoritmo de validación de cédula ecuatoriana
func (s *CedulaValidationService) validateWithLocalAlgorithm(cedula string) *models.CedulaValidationResponse {
	// Validar que todos sean dígitos
	for _, char := range cedula {
		if char < '0' || char > '9' {
			return &models.CedulaValidationResponse{
				IsValid: false,
				Cedula:  cedula,
				Message: "La cédula debe contener solo números",
				Source:  "algorithm",
			}
		}
	}

	// Validar provincia (los primeros 2 dígitos)
	province := cedula[:2]
	provinceNum, _ := strconv.Atoi(province)
	if provinceNum < 1 || provinceNum > 24 {
		return &models.CedulaValidationResponse{
			IsValid: false,
			Cedula:  cedula,
			Message: "Código de provincia inválido",
			Source:  "algorithm",
		}
	}

	// Validar tercer dígito (debe ser menor a 6 para personas naturales)
	thirdDigit, _ := strconv.Atoi(string(cedula[2]))
	if thirdDigit >= 6 {
		return &models.CedulaValidationResponse{
			IsValid: false,
			Cedula:  cedula,
			Message: "Tercer dígito inválido para persona natural",
			Source:  "algorithm",
		}
	}

	// Algoritmo de verificación módulo 10
	isValid := s.validateWithModule10Algorithm(cedula)

	response := &models.CedulaValidationResponse{
		IsValid:  isValid,
		Cedula:   cedula,
		Source:   "algorithm",
		Province: s.getProvinceName(provinceNum),
	}

	if isValid {
		response.Message = "Cédula válida según algoritmo ecuatoriano"
	} else {
		response.Message = "Cédula inválida según algoritmo de verificación"
	}

	return response
}

// validateWithModule10Algorithm implementa el algoritmo módulo 10 para cédulas ecuatorianas
func (s *CedulaValidationService) validateWithModule10Algorithm(cedula string) bool {
	// Los primeros 9 dígitos para el cálculo
	digits := make([]int, 9)
	for i := 0; i < 9; i++ {
		digits[i], _ = strconv.Atoi(string(cedula[i]))
	}

	// Dígito verificador (último dígito)
	checkDigit, _ := strconv.Atoi(string(cedula[9]))

	// Aplicar algoritmo módulo 10
	sum := 0
	for i := 0; i < 9; i++ {
		value := digits[i]

		// Multiplicar por 2 si la posición es par (0, 2, 4, 6, 8)
		if i%2 == 0 {
			value *= 2
			// Si el resultado es mayor a 9, restar 9
			if value > 9 {
				value -= 9
			}
		}

		sum += value
	}

	// Calcular el dígito verificador esperado
	remainder := sum % 10
	expectedCheckDigit := 0
	if remainder != 0 {
		expectedCheckDigit = 10 - remainder
	}

	return checkDigit == expectedCheckDigit
}

// cleanCedula limpia la cédula removiendo espacios y caracteres especiales
func (s *CedulaValidationService) cleanCedula(cedula string) string {
	// Remover espacios, guiones y otros caracteres no numéricos
	cleaned := strings.ReplaceAll(cedula, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	cleaned = strings.ReplaceAll(cleaned, ".", "")

	return cleaned
}

// getProvinceName devuelve el nombre de la provincia basado en el código
func (s *CedulaValidationService) getProvinceName(code int) string {
	provinces := map[int]string{
		1:  "Azuay",
		2:  "Bolívar",
		3:  "Cañar",
		4:  "Carchi",
		5:  "Cotopaxi",
		6:  "Chimborazo",
		7:  "El Oro",
		8:  "Esmeraldas",
		9:  "Guayas",
		10: "Imbabura",
		11: "Loja",
		12: "Los Ríos",
		13: "Manabí",
		14: "Morona Santiago",
		15: "Napo",
		16: "Pastaza",
		17: "Pichincha",
		18: "Tungurahua",
		19: "Zamora Chinchipe",
		20: "Galápagos",
		21: "Sucumbíos",
		22: "Orellana",
		23: "Santo Domingo de los Tsáchilas",
		24: "Santa Elena",
	}

	if name, exists := provinces[code]; exists {
		return name
	}
	return "Desconocida"
}

// IsValidCedulaFormat verifica si una cédula tiene el formato correcto
func (s *CedulaValidationService) IsValidCedulaFormat(cedula string) bool {
	cleanCedula := s.cleanCedula(cedula)
	if len(cleanCedula) != 10 {
		return false
	}

	// Verificar que todos los caracteres sean dígitos
	for _, char := range cleanCedula {
		if char < '0' || char > '9' {
			return false
		}
	}

	return true
}
