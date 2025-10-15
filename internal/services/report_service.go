package services

import (
	"context"
	"fmt"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/internal/repositories"
)

// ReportService maneja la lógica de negocio para reportes
type ReportService struct {
	reportRepo *repositories.ReportRepository
	userRepo   *repositories.UserRepository
}

// NewReportService crea una nueva instancia del servicio
func NewReportService(reportRepo *repositories.ReportRepository, userRepo *repositories.UserRepository) *ReportService {
	return &ReportService{
		reportRepo: reportRepo,
		userRepo:   userRepo,
	}
}

// CreateReport crea un nuevo reporte
func (s *ReportService) CreateReport(report *models.Report) (*models.Report, error) {
	ctx := context.Background()

	// Validar que el usuario existe
	user, err := s.userRepo.GetUserByID(ctx, report.UserID)
	if err != nil {
		return nil, fmt.Errorf("usuario no encontrado: %w", err)
	}

	// Validar categoría
	if !s.isValidCategory(report.Category) {
		return nil, fmt.Errorf("categoría inválida: %s", report.Category)
	}

	// Validar prioridad
	if !s.isValidPriority(report.Priority) {
		return nil, fmt.Errorf("prioridad inválida: %s", report.Priority)
	}

	// Validar estado
	if !s.isValidStatus(report.Status) {
		return nil, fmt.Errorf("estado inválido: %s", report.Status)
	}

	// Validar ubicación
	if err := s.validateLocation(report.Location); err != nil {
		return nil, fmt.Errorf("ubicación inválida: %w", err)
	}

	// Crear el reporte
	if err := s.reportRepo.CreateReport(ctx, report); err != nil {
		return nil, fmt.Errorf("error creando reporte: %w", err)
	}

	// Asignar información del usuario
	report.User = user

	return report, nil
}

// GetReportByID obtiene un reporte por su ID
func (s *ReportService) GetReportByID(id int) (*models.Report, error) {
	ctx := context.Background()
	return s.reportRepo.GetReportByID(ctx, id)
}

// GetReportsByUserID obtiene reportes de un usuario específico
func (s *ReportService) GetReportsByUserID(userID int, limit, offset int) ([]*models.Report, error) {
	ctx := context.Background()

	// Validar límites
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	return s.reportRepo.GetReportsByUserID(ctx, userID, limit, offset)
}

// GetReportsCountByUserID obtiene el total de reportes de un usuario
func (s *ReportService) GetReportsCountByUserID(userID int) (int, error) {
	ctx := context.Background()
	return s.reportRepo.GetReportsCountByUserID(ctx, userID)
}

// GetReportsByLocation obtiene reportes cerca de una ubicación
func (s *ReportService) GetReportsByLocation(lat, lng float64, radiusKm float64, limit int) ([]*models.Report, error) {
	ctx := context.Background()

	// Validar parámetros
	if lat < -90 || lat > 90 {
		return nil, fmt.Errorf("latitud inválida: %f", lat)
	}
	if lng < -180 || lng > 180 {
		return nil, fmt.Errorf("longitud inválida: %f", lng)
	}
	if radiusKm <= 0 || radiusKm > 50 { // Máximo 50km de radio
		radiusKm = 5 // Default 5km
	}
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	return s.reportRepo.GetReportsByLocation(ctx, lat, lng, radiusKm, limit)
}

// UpdateReportStatus actualiza el estado de un reporte
func (s *ReportService) UpdateReportStatus(reportID int, newStatus string, operatorID *int) error {
	ctx := context.Background()

	// Validar estado
	if !s.isValidStatus(newStatus) {
		return fmt.Errorf("estado inválido: %s", newStatus)
	}

	// Obtener reporte actual
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("reporte no encontrado: %w", err)
	}

	// Actualizar estado
	report.Status = newStatus
	if operatorID != nil {
		report.AssignedOperatorID = operatorID
	}

	return s.reportRepo.UpdateReport(ctx, report)
}

// AssignOperator asigna un operador a un reporte
func (s *ReportService) AssignOperator(reportID int, operatorID int) error {
	ctx := context.Background()

	// Verificar que el operador existe y tiene el rol correcto
	operator, err := s.userRepo.GetUserByID(ctx, operatorID)
	if err != nil {
		return fmt.Errorf("operador no encontrado: %w", err)
	}

	// Verificar rol de operador (asumiendo que RoleID 2 es operador)
	if operator.RoleID != 2 {
		return fmt.Errorf("el usuario no tiene rol de operador")
	}

	// Obtener reporte
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("reporte no encontrado: %w", err)
	}

	// Asignar operador
	report.AssignedOperatorID = &operatorID
	if report.Status == "pending" {
		report.Status = "in_progress"
	}

	return s.reportRepo.UpdateReport(ctx, report)
}

// GetReportsWithFilters obtiene reportes con filtros
func (s *ReportService) GetReportsWithFilters(userID *int, status *string, category *string, limit, offset int) ([]*models.Report, int, error) {
	ctx := context.Background()

	// Validar límites
	if limit <= 0 || limit > 100 {
		limit = 10
	}
	if offset < 0 {
		offset = 0
	}

	// Validar filtros
	if status != nil && !s.isValidStatus(*status) {
		return nil, 0, fmt.Errorf("estado inválido: %s", *status)
	}
	if category != nil && !s.isValidCategory(*category) {
		return nil, 0, fmt.Errorf("categoría inválida: %s", *category)
	}

	// Obtener reportes
	reports, err := s.reportRepo.GetReportsWithFilters(ctx, userID, status, category, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("error obteniendo reportes: %w", err)
	}

	// Obtener total
	total, err := s.reportRepo.GetReportsCountWithFilters(ctx, userID, status, category)
	if err != nil {
		return nil, 0, fmt.Errorf("error obteniendo total: %w", err)
	}

	return reports, total, nil
}

// DeleteReport elimina un reporte (solo para el creador o administradores)
func (s *ReportService) DeleteReport(reportID int, userID int, userRole int) error {
	ctx := context.Background()

	// Obtener reporte
	report, err := s.reportRepo.GetReportByID(ctx, reportID)
	if err != nil {
		return fmt.Errorf("reporte no encontrado: %w", err)
	}

	// Verificar permisos: solo el creador o administradores pueden eliminar
	if report.UserID != userID && userRole != 3 { // Asumiendo que RoleID 3 es admin
		return fmt.Errorf("no tienes permisos para eliminar este reporte")
	}

	// No permitir eliminar reportes en progreso o resueltos
	if report.Status == "in_progress" || report.Status == "resolved" {
		return fmt.Errorf("no se puede eliminar un reporte en progreso o resuelto")
	}

	return s.reportRepo.DeleteReport(ctx, reportID)
}

// Métodos de validación

// isValidCategory valida si una categoría es válida
func (s *ReportService) isValidCategory(category string) bool {
	validCategories := []string{
		"residuos_organicos",
		"residuos_reciclables",
		"residuos_peligrosos",
		"limpieza_publica",
		"mantenimiento_urbano",
		"otros",
	}

	for _, valid := range validCategories {
		if category == valid {
			return true
		}
	}
	return false
}

// isValidPriority valida si una prioridad es válida
func (s *ReportService) isValidPriority(priority string) bool {
	validPriorities := []string{"baja", "media", "alta", "urgente"}

	for _, valid := range validPriorities {
		if priority == valid {
			return true
		}
	}
	return false
}

// isValidStatus valida si un estado es válido
func (s *ReportService) isValidStatus(status string) bool {
	validStatuses := []string{"pending", "in_progress", "resolved", "rejected"}

	for _, valid := range validStatuses {
		if status == valid {
			return true
		}
	}
	return false
}

// validateLocation valida una ubicación
func (s *ReportService) validateLocation(location models.Location) error {
	if location.Latitude < -90 || location.Latitude > 90 {
		return fmt.Errorf("latitud inválida: %f", location.Latitude)
	}
	if location.Longitude < -180 || location.Longitude > 180 {
		return fmt.Errorf("longitud inválida: %f", location.Longitude)
	}
	return nil
}

// GetReportStatistics obtiene estadísticas de reportes
func (s *ReportService) GetReportStatistics(userID *int) (map[string]interface{}, error) {
	ctx := context.Background()

	stats := make(map[string]interface{})

	// Total de reportes
	total, err := s.reportRepo.GetReportsCountWithFilters(ctx, userID, nil, nil)
	if err != nil {
		return nil, fmt.Errorf("error obteniendo total: %w", err)
	}
	stats["total"] = total

	// Reportes por estado
	statusStats := make(map[string]int)
	statuses := []string{"pending", "in_progress", "resolved", "rejected"}
	for _, status := range statuses {
		count, err := s.reportRepo.GetReportsCountWithFilters(ctx, userID, &status, nil)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo estadísticas de estado: %w", err)
		}
		statusStats[status] = count
	}
	stats["by_status"] = statusStats

	// Reportes por categoría
	categoryStats := make(map[string]int)
	categories := []string{
		"residuos_organicos", "residuos_reciclables", "residuos_peligrosos",
		"limpieza_publica", "mantenimiento_urbano", "otros",
	}
	for _, category := range categories {
		count, err := s.reportRepo.GetReportsCountWithFilters(ctx, userID, nil, &category)
		if err != nil {
			return nil, fmt.Errorf("error obteniendo estadísticas de categoría: %w", err)
		}
		categoryStats[category] = count
	}
	stats["by_category"] = categoryStats

	return stats, nil
}
