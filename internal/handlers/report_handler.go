package handlers

import (
	"net/http"
	"strconv"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/services"
	"backend-residuos-app/pkg/middleware"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ReportHandler maneja las operaciones relacionadas con reportes
type ReportHandler struct {
	reportService     *services.ReportService
	cedulaService     *services.CedulaValidationService
	photoService      *services.PhotoUploadService
	validationService *services.ValidationService
	validator         *validator.Validate
}

// NewReportHandler crea una nueva instancia del handler
func NewReportHandler(
	reportService *services.ReportService,
	cedulaService *services.CedulaValidationService,
	photoService *services.PhotoUploadService,
	validationService *services.ValidationService,
) *ReportHandler {
	return &ReportHandler{
		reportService:     reportService,
		cedulaService:     cedulaService,
		photoService:      photoService,
		validationService: validationService,
		validator:         validator.New(),
	}
}

// CreateReport godoc
// @Summary Crear un nuevo reporte
// @Description Crea un nuevo reporte de problema urbano con validación de cédula opcional para usuarios ciudadanos
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param report body models.CreateReportRequest true "Datos del reporte"
// @Success 201 {object} models.CreateReportResponse "Reporte creado exitosamente"
// @Failure 400 {object} map[string]interface{} "Error de validación"
// @Failure 401 {object} map[string]interface{} "No autorizado"
// @Failure 403 {object} map[string]interface{} "Validación de cédula requerida"
// @Failure 500 {object} map[string]interface{} "Error interno del servidor"
// @Router /reports/create [post]
func (h *ReportHandler) CreateReport(c *gin.Context) {
	// Obtener información del usuario autenticado
	userInfo, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "Usuario no autenticado",
			"code":    "UNAUTHORIZED",
			"details": "No se pudo obtener información del usuario",
		})
		return
	}

	user, ok := userInfo.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error interno",
			"code":    "INTERNAL_ERROR",
			"details": "Error al procesar información del usuario",
		})
		return
	}

	// Parsear el request
	var request models.CreateReportRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Validar datos básicos
	if validationErrors := h.validationService.ValidateStruct(&request); len(validationErrors) > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":             "Errores de validación",
			"code":              "VALIDATION_ERROR",
			"validation_errors": validationErrors,
		})
		return
	}

	// Validar cédula si es requerido para usuarios con rol "User" (ciudadanos)
	if user.RoleID == 1 && request.CedulaValidate { // Asumiendo que RoleID 1 es "User/Ciudadano"
		if user.Cedula == nil || *user.Cedula == "" {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Validación de cédula requerida",
				"code":    "CEDULA_REQUIRED",
				"details": "Los ciudadanos deben tener una cédula validada para crear reportes",
			})
			return
		}

		// Validar que la cédula del usuario sea válida
		cedulaValidation, err := h.cedulaService.ValidateCedula(*user.Cedula)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error validando cédula",
				"code":    "CEDULA_VALIDATION_ERROR",
				"details": err.Error(),
			})
			return
		}

		if !cedulaValidation.IsValid {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "Cédula inválida",
				"code":    "INVALID_CEDULA",
				"details": cedulaValidation.Message,
			})
			return
		}
	}

	// Procesar foto si está presente
	var photoURL *string
	if request.PhotoBase64 != nil && *request.PhotoBase64 != "" {
		// Validar tamaño de la imagen (5MB máximo)
		if len(*request.PhotoBase64) > 5*1024*1024 { // Aproximado para base64
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "Imagen demasiado grande",
				"code":    "IMAGE_TOO_LARGE",
				"details": "La imagen no debe superar los 5MB",
			})
			return
		}

		// Subir foto a Firebase Storage
		uploadedURL, err := h.photoService.UploadPhotoFromBase64(
			*request.PhotoBase64,
			"report_photo.jpg",
			user.ID,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Error subiendo imagen",
				"code":    "PHOTO_UPLOAD_ERROR",
				"details": err.Error(),
			})
			return
		}
		photoURL = &uploadedURL
	}

	// Crear el modelo de reporte
	report := &models.Report{
		UserID:      user.ID,
		Title:       request.Title,
		Description: request.Description,
		Category:    request.Category,
		Location:    request.Location,
		Address:     request.Address,
		Status:      "pending", // Estado inicial
		Priority:    request.Priority,
		PhotoURL:    photoURL,
		SyncStatus:  "synced",
	}

	// Si no se especifica prioridad, usar "media" como default
	if report.Priority == "" {
		report.Priority = "media"
	}

	// Crear el reporte en la base de datos
	createdReport, err := h.reportService.CreateReport(report)
	if err != nil {
		// Si hubo error y se subió una foto, intentar eliminarla
		if photoURL != nil {
			_ = h.photoService.DeletePhoto(*photoURL)
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error creando reporte",
			"code":    "REPORT_CREATION_ERROR",
			"details": err.Error(),
		})
		return
	}

	// Preparar respuesta
	response := models.CreateReportResponse{
		ID:          createdReport.ID,
		Title:       createdReport.Title,
		Description: createdReport.Description,
		Category:    createdReport.Category,
		Location:    createdReport.Location,
		Address:     createdReport.Address,
		Status:      createdReport.Status,
		Priority:    createdReport.Priority,
		PhotoURL:    createdReport.PhotoURL,
		CreatedAt:   createdReport.CreatedAt,
		Message:     "Reporte creado exitosamente",
	}

	c.JSON(http.StatusCreated, response)
}

// ValidateCedula godoc
// @Summary Validar cédula ecuatoriana
// @Description Valida una cédula ecuatoriana usando API externa con fallback a algoritmo local
// @Tags reports
// @Accept json
// @Produce json
// @Param cedula body models.CedulaValidationRequest true "Número de cédula"
// @Success 200 {object} models.CedulaValidationResponse "Resultado de validación"
// @Failure 400 {object} map[string]interface{} "Error de validación"
// @Failure 500 {object} map[string]interface{} "Error interno del servidor"
// @Router /reports/validate-cedula [post]
func (h *ReportHandler) ValidateCedula(c *gin.Context) {
	var request models.CedulaValidationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Datos inválidos",
			"code":    "INVALID_REQUEST",
			"details": err.Error(),
		})
		return
	}

	// Validar formato básico
	if !h.cedulaService.IsValidCedulaFormat(request.Cedula) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "Formato de cédula inválido",
			"code":    "INVALID_FORMAT",
			"details": "La cédula debe tener exactamente 10 dígitos",
		})
		return
	}

	// Validar cédula
	result, err := h.cedulaService.ValidateCedula(request.Cedula)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error validando cédula",
			"code":    "VALIDATION_ERROR",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetReportsByUser godoc
// @Summary Obtener reportes del usuario
// @Description Obtiene todos los reportes creados por el usuario autenticado
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param page query int false "Número de página" default(1)
// @Param limit query int false "Cantidad por página" default(10)
// @Success 200 {object} map[string]interface{} "Lista de reportes"
// @Failure 401 {object} map[string]interface{} "No autorizado"
// @Failure 500 {object} map[string]interface{} "Error interno del servidor"
// @Router /reports/my-reports [get]
func (h *ReportHandler) GetReportsByUser(c *gin.Context) {
	// Obtener información del usuario
	userInfo, exists := c.Get(middleware.UserContextKey)
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Usuario no autenticado",
			"code":  "UNAUTHORIZED",
		})
		return
	}

	user, ok := userInfo.(*models.User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Error interno",
			"code":  "INTERNAL_ERROR",
		})
		return
	}

	// Parsear parámetros de paginación
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 10
	}

	offset := (page - 1) * limit

	// Obtener reportes del usuario
	reports, err := h.reportService.GetReportsByUserID(user.ID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error obteniendo reportes",
			"code":    "FETCH_ERROR",
			"details": err.Error(),
		})
		return
	}

	// Obtener total de reportes para paginación
	total, err := h.reportService.GetReportsCountByUserID(user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error obteniendo total de reportes",
			"code":    "COUNT_ERROR",
			"details": err.Error(),
		})
		return
	}

	totalPages := (total + limit - 1) / limit

	c.JSON(http.StatusOK, gin.H{
		"reports": reports,
		"pagination": gin.H{
			"current_page": page,
			"per_page":     limit,
			"total":        total,
			"total_pages":  totalPages,
			"has_next":     page < totalPages,
			"has_prev":     page > 1,
		},
	})
}

// GetReportByID godoc
// @Summary Obtener reporte por ID
// @Description Obtiene un reporte específico por su ID
// @Tags reports
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID del reporte"
// @Success 200 {object} models.Report "Reporte encontrado"
// @Failure 400 {object} map[string]interface{} "ID inválido"
// @Failure 404 {object} map[string]interface{} "Reporte no encontrado"
// @Failure 500 {object} map[string]interface{} "Error interno del servidor"
// @Router /reports/{id} [get]
func (h *ReportHandler) GetReportByID(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "ID inválido",
			"code":    "INVALID_ID",
			"details": "El ID debe ser un número entero",
		})
		return
	}

	report, err := h.reportService.GetReportByID(id)
	if err != nil {
		if err.Error() == "report not found" {
			c.JSON(http.StatusNotFound, gin.H{
				"error": "Reporte no encontrado",
				"code":  "NOT_FOUND",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Error obteniendo reporte",
			"code":    "FETCH_ERROR",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, report)
}
