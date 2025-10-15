package services

import (
	"regexp"
	"strings"
	"unicode"

	"backend-residuos-app/internal/models"
)

// ValidationService maneja las validaciones de entrada
type ValidationService struct{}

// NewValidationService crea una nueva instancia del servicio de validación
func NewValidationService() *ValidationService {
	return &ValidationService{}
}

// ValidateRegisterRequest valida una solicitud de registro
func (v *ValidationService) ValidateRegisterRequest(req *models.RegisterRequest) []models.ValidationError {
	var errors []models.ValidationError

	// Validar email
	if err := v.validateEmail(req.Email); err != nil {
		errors = append(errors, *err)
	}

	// Validar contraseña
	if err := v.validatePassword(req.Password); err != nil {
		errors = append(errors, *err)
	}

	// Validar nombre completo
	if err := v.validateFullName(req.FullName); err != nil {
		errors = append(errors, *err)
	}

	// Validar teléfono (opcional pero si se proporciona debe ser válido)
	if req.Phone != "" {
		if err := v.validatePhone(req.Phone); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validar cédula (opcional pero si se proporciona debe ser válida)
	if req.Cedula != "" {
		if err := v.validateCedula(req.Cedula); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validar idioma
	if err := v.validateLanguage(req.Language); err != nil {
		errors = append(errors, *err)
	}

	return errors
}

// ValidateLoginRequest valida una solicitud de login
func (v *ValidationService) ValidateLoginRequest(req *models.LoginRequest) []models.ValidationError {
	var errors []models.ValidationError

	// Validar email
	if err := v.validateEmail(req.Email); err != nil {
		errors = append(errors, *err)
	}

	// Validar que la contraseña no esté vacía
	if strings.TrimSpace(req.Password) == "" {
		errors = append(errors, models.ValidationError{
			Field:   "password",
			Tag:     "required",
			Message: "Password is required",
		})
	}

	return errors
}

// validateEmail valida el formato del email
func (v *ValidationService) validateEmail(email string) *models.ValidationError {
	if strings.TrimSpace(email) == "" {
		return &models.ValidationError{
			Field:   "email",
			Tag:     "required",
			Message: "Email is required",
		}
	}

	// Regex más estricto para email
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return &models.ValidationError{
			Field:   "email",
			Tag:     "email",
			Message: "Invalid email format",
		}
	}

	return nil
}

// validatePassword valida la fortaleza de la contraseña
func (v *ValidationService) validatePassword(password string) *models.ValidationError {
	if len(password) < 6 {
		return &models.ValidationError{
			Field:   "password",
			Tag:     "min",
			Message: "Password must be at least 6 characters long",
		}
	}

	if len(password) > 128 {
		return &models.ValidationError{
			Field:   "password",
			Tag:     "max",
			Message: "Password must be no more than 128 characters long",
		}
	}

	// Verificar que contenga al menos una letra y un número
	hasLetter := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsLetter(char):
			hasLetter = true
		case unicode.IsNumber(char):
			hasNumber = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasLetter {
		return &models.ValidationError{
			Field:   "password",
			Tag:     "format",
			Message: "Password must contain at least one letter",
		}
	}

	if !hasNumber {
		return &models.ValidationError{
			Field:   "password",
			Tag:     "format",
			Message: "Password must contain at least one number",
		}
	}

	// Opcional: requerir caracteres especiales para mayor seguridad
	if !hasSpecial {
		return &models.ValidationError{
			Field:   "password",
			Tag:     "format",
			Message: "Password should contain at least one special character for better security",
		}
	}

	return nil
}

// validateFullName valida el nombre completo
func (v *ValidationService) validateFullName(fullName string) *models.ValidationError {
	trimmed := strings.TrimSpace(fullName)
	if trimmed == "" {
		return &models.ValidationError{
			Field:   "full_name",
			Tag:     "required",
			Message: "Full name is required",
		}
	}

	if len(trimmed) < 2 {
		return &models.ValidationError{
			Field:   "full_name",
			Tag:     "min",
			Message: "Full name must be at least 2 characters long",
		}
	}

	if len(trimmed) > 255 {
		return &models.ValidationError{
			Field:   "full_name",
			Tag:     "max",
			Message: "Full name must be no more than 255 characters long",
		}
	}

	// Verificar que solo contenga letras, espacios y algunos caracteres especiales
	nameRegex := regexp.MustCompile(`^[a-zA-ZáéíóúÁÉÍÓÚñÑüÜ\s'-]+$`)
	if !nameRegex.MatchString(trimmed) {
		return &models.ValidationError{
			Field:   "full_name",
			Tag:     "format",
			Message: "Full name contains invalid characters",
		}
	}

	return nil
}

// validatePhone valida el formato del teléfono
func (v *ValidationService) validatePhone(phone string) *models.ValidationError {
	trimmed := strings.TrimSpace(phone)
	if trimmed == "" {
		return nil // El teléfono es opcional
	}

	// Regex para teléfonos peruanos: +51XXXXXXXXX o 9XXXXXXXX
	phoneRegex := regexp.MustCompile(`^(\+51|51)?[9][0-9]{8}$`)
	if !phoneRegex.MatchString(strings.ReplaceAll(trimmed, " ", "")) {
		return &models.ValidationError{
			Field:   "phone",
			Tag:     "format",
			Message: "Invalid phone format. Use format: +51987654321 or 987654321",
		}
	}

	return nil
}

// validateCedula valida el formato de la cédula
func (v *ValidationService) validateCedula(cedula string) *models.ValidationError {
	trimmed := strings.TrimSpace(cedula)
	if trimmed == "" {
		return nil // La cédula es opcional
	}

	// Regex para DNI peruano: 8 dígitos
	cedulaRegex := regexp.MustCompile(`^[0-9]{10}$`)
	if !cedulaRegex.MatchString(trimmed) {
		return &models.ValidationError{
			Field:   "cedula",
			Tag:     "format",
			Message: "Invalid cedula format. Must be 10 digits",
		}
	}

	return nil
}

// validateLanguage valida el código de idioma
func (v *ValidationService) validateLanguage(language string) *models.ValidationError {
	if language == "" {
		return nil // Se asignará "es" por defecto
	}

	supportedLanguages := []string{"es", "qu", "en"}
	for _, lang := range supportedLanguages {
		if language == lang {
			return nil
		}
	}

	return &models.ValidationError{
		Field:   "language",
		Tag:     "enum",
		Message: "Language must be one of: es, qu, en",
	}
}

// ValidateStruct valida un struct usando las validaciones definidas
func (v *ValidationService) ValidateStruct(s interface{}) []models.ValidationError {
	var errors []models.ValidationError

	// Validar según el tipo de struct
	switch req := s.(type) {
	case *models.CreateReportRequest:
		errors = v.validateCreateReportRequest(req)
	default:
		// Para otros tipos, retornar slice vacío por ahora
		return errors
	}

	return errors
}

// validateCreateReportRequest valida una solicitud de creación de reporte
func (v *ValidationService) validateCreateReportRequest(req *models.CreateReportRequest) []models.ValidationError {
	var errors []models.ValidationError

	// Validar título
	if req.Title == "" {
		errors = append(errors, models.ValidationError{
			Field:   "title",
			Tag:     "required",
			Message: "El título es requerido",
		})
	} else if len(req.Title) < 5 {
		errors = append(errors, models.ValidationError{
			Field:   "title",
			Tag:     "min",
			Message: "El título debe tener al menos 5 caracteres",
		})
	} else if len(req.Title) > 100 {
		errors = append(errors, models.ValidationError{
			Field:   "title",
			Tag:     "max",
			Message: "El título no puede tener más de 100 caracteres",
		})
	}

	// Validar descripción
	if req.Description == "" {
		errors = append(errors, models.ValidationError{
			Field:   "description",
			Tag:     "required",
			Message: "La descripción es requerida",
		})
	} else if len(req.Description) < 10 {
		errors = append(errors, models.ValidationError{
			Field:   "description",
			Tag:     "min",
			Message: "La descripción debe tener al menos 10 caracteres",
		})
	} else if len(req.Description) > 1000 {
		errors = append(errors, models.ValidationError{
			Field:   "description",
			Tag:     "max",
			Message: "La descripción no puede tener más de 1000 caracteres",
		})
	}

	// Validar categoría
	validCategories := []string{
		"limpieza_publica", "agua_potable", "alcantarillado", "alumbrado_publico",
		"vias_transporte", "seguridad_ciudadana", "ruido_ambiental",
		"residuos_peligrosos", "areas_verdes", "otros",
	}

	if req.Category == "" {
		errors = append(errors, models.ValidationError{
			Field:   "category",
			Tag:     "required",
			Message: "La categoría es requerida",
		})
	} else {
		validCategory := false
		for _, cat := range validCategories {
			if req.Category == cat {
				validCategory = true
				break
			}
		}
		if !validCategory {
			errors = append(errors, models.ValidationError{
				Field:   "category",
				Tag:     "oneof",
				Message: "Categoría inválida",
			})
		}
	}

	// Validar prioridad
	validPriorities := []string{"baja", "media", "alta", "urgente"}
	if req.Priority == "" {
		errors = append(errors, models.ValidationError{
			Field:   "priority",
			Tag:     "required",
			Message: "La prioridad es requerida",
		})
	} else {
		validPriority := false
		for _, priority := range validPriorities {
			if req.Priority == priority {
				validPriority = true
				break
			}
		}
		if !validPriority {
			errors = append(errors, models.ValidationError{
				Field:   "priority",
				Tag:     "oneof",
				Message: "Prioridad inválida",
			})
		}
	}

	// Validar ubicación (latitude debe estar entre -90 y 90, longitude entre -180 y 180)
	if req.Location.Latitude < -90 || req.Location.Latitude > 90 {
		errors = append(errors, models.ValidationError{
			Field:   "location.latitude",
			Tag:     "range",
			Message: "La latitud debe estar entre -90 y 90",
		})
	}

	if req.Location.Longitude < -180 || req.Location.Longitude > 180 {
		errors = append(errors, models.ValidationError{
			Field:   "location.longitude",
			Tag:     "range",
			Message: "La longitud debe estar entre -180 y 180",
		})
	}

	return errors
}
