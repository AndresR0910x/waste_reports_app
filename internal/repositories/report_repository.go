package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/pkg/database"
)

type ReportRepository struct {
	db *database.DB
}

func NewReportRepository(db *database.DB) *ReportRepository {
	return &ReportRepository{db: db}
}

// CreateReport crea un nuevo reporte
func (rr *ReportRepository) CreateReport(ctx context.Context, report *models.Report) error {
	query := `
		INSERT INTO reports (user_id, title, description, category, location, address, 
							status, priority, photo_url, evidence_photos)
		VALUES ($1, $2, $3, $4, ST_GeogFromText($5), $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	locationWKT := fmt.Sprintf("POINT(%f %f)", report.Location.Longitude, report.Location.Latitude)

	err := rr.db.QueryRowContext(ctx, query,
		report.UserID, report.Title, report.Description, report.Category,
		locationWKT, report.Address, report.Status, report.Priority,
		report.PhotoURL, report.EvidencePhotos).Scan(
		&report.ID, &report.CreatedAt, &report.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create report: %v", err)
	}

	return nil
}

// GetReportByID obtiene un reporte por ID
func (rr *ReportRepository) GetReportByID(ctx context.Context, id int) (*models.Report, error) {
	query := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.evidence_photos, r.assigned_operator_id, 
			   r.estimated_resolution, r.actual_resolution, r.citizen_satisfaction_rating,
			   r.citizen_feedback, r.sync_status, r.created_at, r.updated_at,
			   u.firebase_uid, u.email, u.full_name, u.cedula,
			   op.firebase_uid as operator_firebase_uid, op.email as operator_email, 
			   op.full_name as operator_full_name
		FROM reports r
		JOIN users u ON r.user_id = u.id
		LEFT JOIN users op ON r.assigned_operator_id = op.id
		WHERE r.id = $1`

	report := &models.Report{
		User:             &models.User{},
		AssignedOperator: &models.User{},
	}

	var locationText string
	var operatorFirebaseUID, operatorEmail, operatorFullName sql.NullString

	err := rr.db.QueryRowContext(ctx, query, id).Scan(
		&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
		&locationText, &report.Address, &report.Status, &report.Priority,
		&report.PhotoURL, &report.EvidencePhotos, &report.AssignedOperatorID,
		&report.EstimatedResolution, &report.ActualResolution, &report.CitizenSatisfactionRating,
		&report.CitizenFeedback, &report.SyncStatus, &report.CreatedAt, &report.UpdatedAt,
		&report.User.FirebaseUID, &report.User.Email, &report.User.FullName, &report.User.Cedula,
		&operatorFirebaseUID, &operatorEmail, &operatorFullName)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("report not found")
		}
		return nil, fmt.Errorf("failed to get report: %v", err)
	}

	// Parse location
	if err := report.Location.Scan(locationText); err != nil {
		return nil, fmt.Errorf("failed to parse location: %v", err)
	}

	// Set assigned operator if exists
	if operatorFirebaseUID.Valid {
		report.AssignedOperator.FirebaseUID = operatorFirebaseUID.String
		report.AssignedOperator.Email = &operatorEmail.String
		report.AssignedOperator.FullName = &operatorFullName.String
	} else {
		report.AssignedOperator = nil
	}

	return report, nil
}

// GetReportsByUserID obtiene reportes de un usuario específico
func (rr *ReportRepository) GetReportsByUserID(ctx context.Context, userID int, limit, offset int) ([]*models.Report, error) {
	query := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.created_at, r.updated_at
		FROM reports r
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := rr.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by user: %w", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report := &models.Report{}
		var locationText string

		err := rows.Scan(
			&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
			&locationText, &report.Address, &report.Status, &report.Priority,
			&report.PhotoURL, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}

		// Parse location
		if err := report.Location.Scan(locationText); err != nil {
			return nil, fmt.Errorf("failed to parse location: %w", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetReportsCountByUserID obtiene el total de reportes de un usuario
func (rr *ReportRepository) GetReportsCountByUserID(ctx context.Context, userID int) (int, error) {
	query := `SELECT COUNT(*) FROM reports WHERE user_id = $1`

	var count int
	err := rr.db.QueryRowContext(ctx, query, userID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get reports count: %w", err)
	}

	return count, nil
}

// GetReportsByLocation obtiene reportes cerca de una ubicación
func (rr *ReportRepository) GetReportsByLocation(ctx context.Context, lat, lng float64, radiusKm float64, limit int) ([]*models.Report, error) {
	query := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.created_at, r.updated_at,
			   ST_Distance(r.location::geography, ST_Point($2, $1)::geography) as distance
		FROM reports r
		WHERE ST_DWithin(r.location::geography, ST_Point($2, $1)::geography, $3 * 1000)
		ORDER BY distance ASC
		LIMIT $4`

	rows, err := rr.db.QueryContext(ctx, query, lat, lng, radiusKm, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by location: %w", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report := &models.Report{}
		var locationText string
		var distance float64

		err := rows.Scan(
			&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
			&locationText, &report.Address, &report.Status, &report.Priority,
			&report.PhotoURL, &report.CreatedAt, &report.UpdatedAt, &distance,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}

		// Parse location
		if err := report.Location.Scan(locationText); err != nil {
			return nil, fmt.Errorf("failed to parse location: %w", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// UpdateReport actualiza un reporte
func (rr *ReportRepository) UpdateReport(ctx context.Context, report *models.Report) error {
	query := `
		UPDATE reports 
		SET title = $2, description = $3, category = $4, location = ST_GeogFromText($5),
			address = $6, status = $7, priority = $8, photo_url = $9,
			assigned_operator_id = $10, estimated_resolution = $11, actual_resolution = $12,
			citizen_satisfaction_rating = $13, citizen_feedback = $14, updated_at = NOW()
		WHERE id = $1`

	locationWKT := fmt.Sprintf("POINT(%f %f)", report.Location.Longitude, report.Location.Latitude)

	_, err := rr.db.ExecContext(ctx, query,
		report.ID, report.Title, report.Description, report.Category, locationWKT,
		report.Address, report.Status, report.Priority, report.PhotoURL,
		report.AssignedOperatorID, report.EstimatedResolution, report.ActualResolution,
		report.CitizenSatisfactionRating, report.CitizenFeedback)

	if err != nil {
		return fmt.Errorf("failed to update report: %w", err)
	}

	return nil
}

// DeleteReport elimina un reporte
func (rr *ReportRepository) DeleteReport(ctx context.Context, id int) error {
	query := `DELETE FROM reports WHERE id = $1`

	result, err := rr.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete report: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("report not found")
	}

	return nil
}

// GetReportsWithFilters obtiene reportes con filtros
func (rr *ReportRepository) GetReportsWithFilters(ctx context.Context, userID *int, status *string, category *string, limit, offset int) ([]*models.Report, error) {
	query := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.created_at, r.updated_at
		FROM reports r
		WHERE 1=1`

	args := []interface{}{}
	argIndex := 1

	if userID != nil {
		query += fmt.Sprintf(" AND r.user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	if status != nil {
		query += fmt.Sprintf(" AND r.status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	if category != nil {
		query += fmt.Sprintf(" AND r.category = $%d", argIndex)
		args = append(args, *category)
		argIndex++
	}

	query += " ORDER BY r.created_at DESC"
	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := rr.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports with filters: %w", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report := &models.Report{}
		var locationText string

		err := rows.Scan(
			&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
			&locationText, &report.Address, &report.Status, &report.Priority,
			&report.PhotoURL, &report.CreatedAt, &report.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %w", err)
		}

		// Parse location
		if err := report.Location.Scan(locationText); err != nil {
			return nil, fmt.Errorf("failed to parse location: %w", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetReportsCountWithFilters obtiene el total de reportes con filtros
func (rr *ReportRepository) GetReportsCountWithFilters(ctx context.Context, userID *int, status *string, category *string) (int, error) {
	query := `SELECT COUNT(*) FROM reports WHERE 1=1`
	args := []interface{}{}
	argIndex := 1

	if userID != nil {
		query += fmt.Sprintf(" AND user_id = $%d", argIndex)
		args = append(args, *userID)
		argIndex++
	}

	if status != nil {
		query += fmt.Sprintf(" AND status = $%d", argIndex)
		args = append(args, *status)
		argIndex++
	}

	if category != nil {
		query += fmt.Sprintf(" AND category = $%d", argIndex)
		args = append(args, *category)
		argIndex++
	}

	var count int
	err := rr.db.QueryRowContext(ctx, query, args...).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get reports count: %w", err)
	}

	return count, nil
}

// GetReportsByUser obtiene reportes de un usuario
func (rr *ReportRepository) GetReportsByUser(ctx context.Context, userID int, limit, offset int) ([]*models.Report, error) {
	query := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.evidence_photos, r.assigned_operator_id, 
			   r.estimated_resolution, r.actual_resolution, r.citizen_satisfaction_rating,
			   r.citizen_feedback, r.sync_status, r.created_at, r.updated_at
		FROM reports r
		WHERE r.user_id = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := rr.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by user: %v", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report := &models.Report{}
		var locationText string

		err := rows.Scan(
			&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
			&locationText, &report.Address, &report.Status, &report.Priority,
			&report.PhotoURL, &report.EvidencePhotos, &report.AssignedOperatorID,
			&report.EstimatedResolution, &report.ActualResolution, &report.CitizenSatisfactionRating,
			&report.CitizenFeedback, &report.SyncStatus, &report.CreatedAt, &report.UpdatedAt)

		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %v", err)
		}

		// Parse location
		if err := report.Location.Scan(locationText); err != nil {
			return nil, fmt.Errorf("failed to parse location: %v", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// GetReportsByStatus obtiene reportes por estado
func (rr *ReportRepository) GetReportsByStatus(ctx context.Context, status string, limit, offset int) ([]*models.Report, error) {
	query := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.evidence_photos, r.assigned_operator_id, 
			   r.estimated_resolution, r.actual_resolution, r.citizen_satisfaction_rating,
			   r.citizen_feedback, r.sync_status, r.created_at, r.updated_at,
			   u.firebase_uid, u.email, u.full_name, u.cedula
		FROM reports r
		JOIN users u ON r.user_id = u.id
		WHERE r.status = $1
		ORDER BY r.created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := rr.db.QueryContext(ctx, query, status, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by status: %v", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report := &models.Report{User: &models.User{}}
		var locationText string

		err := rows.Scan(
			&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
			&locationText, &report.Address, &report.Status, &report.Priority,
			&report.PhotoURL, &report.EvidencePhotos, &report.AssignedOperatorID,
			&report.EstimatedResolution, &report.ActualResolution, &report.CitizenSatisfactionRating,
			&report.CitizenFeedback, &report.SyncStatus, &report.CreatedAt, &report.UpdatedAt,
			&report.User.FirebaseUID, &report.User.Email, &report.User.FullName, &report.User.Cedula)

		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %v", err)
		}

		// Parse location
		if err := report.Location.Scan(locationText); err != nil {
			return nil, fmt.Errorf("failed to parse location: %v", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// UpdateReportStatus actualiza el estado de un reporte
func (rr *ReportRepository) UpdateReportStatus(ctx context.Context, reportID int, newStatus string, operatorID *int) error {
	query := `
		UPDATE reports 
		SET status = $2, assigned_operator_id = $3, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1`

	result, err := rr.db.ExecContext(ctx, query, reportID, newStatus, operatorID)
	if err != nil {
		return fmt.Errorf("failed to update report status: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("report not found")
	}

	return nil
}

// GetReportsByFilters obtiene reportes con filtros múltiples
func (rr *ReportRepository) GetReportsByFilters(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*models.Report, error) {
	baseQuery := `
		SELECT r.id, r.user_id, r.title, r.description, r.category, 
			   ST_AsText(r.location) as location_text, r.address, r.status, r.priority,
			   r.photo_url, r.evidence_photos, r.assigned_operator_id, 
			   r.estimated_resolution, r.actual_resolution, r.citizen_satisfaction_rating,
			   r.citizen_feedback, r.sync_status, r.created_at, r.updated_at,
			   u.firebase_uid, u.email, u.full_name, u.cedula
		FROM reports r
		JOIN users u ON r.user_id = u.id`

	var conditions []string
	var args []interface{}
	argIndex := 1

	// Construir condiciones dinámicamente
	if status, ok := filters["status"].(string); ok && status != "" {
		conditions = append(conditions, fmt.Sprintf("r.status = $%d", argIndex))
		args = append(args, status)
		argIndex++
	}

	if category, ok := filters["category"].(string); ok && category != "" {
		conditions = append(conditions, fmt.Sprintf("r.category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if priority, ok := filters["priority"].(string); ok && priority != "" {
		conditions = append(conditions, fmt.Sprintf("r.priority = $%d", argIndex))
		args = append(args, priority)
		argIndex++
	}

	if operatorID, ok := filters["assigned_operator_id"].(int); ok && operatorID > 0 {
		conditions = append(conditions, fmt.Sprintf("r.assigned_operator_id = $%d", argIndex))
		args = append(args, operatorID)
		argIndex++
	}

	// Agregar condiciones WHERE si existen
	if len(conditions) > 0 {
		baseQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Agregar ORDER BY y LIMIT/OFFSET
	baseQuery += fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	rows, err := rr.db.QueryContext(ctx, baseQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get reports by filters: %v", err)
	}
	defer rows.Close()

	var reports []*models.Report
	for rows.Next() {
		report := &models.Report{User: &models.User{}}
		var locationText string

		err := rows.Scan(
			&report.ID, &report.UserID, &report.Title, &report.Description, &report.Category,
			&locationText, &report.Address, &report.Status, &report.Priority,
			&report.PhotoURL, &report.EvidencePhotos, &report.AssignedOperatorID,
			&report.EstimatedResolution, &report.ActualResolution, &report.CitizenSatisfactionRating,
			&report.CitizenFeedback, &report.SyncStatus, &report.CreatedAt, &report.UpdatedAt,
			&report.User.FirebaseUID, &report.User.Email, &report.User.FullName, &report.User.Cedula)

		if err != nil {
			return nil, fmt.Errorf("failed to scan report: %v", err)
		}

		// Parse location
		if err := report.Location.Scan(locationText); err != nil {
			return nil, fmt.Errorf("failed to parse location: %v", err)
		}

		reports = append(reports, report)
	}

	return reports, nil
}

// AddReportUpdate agrega una actualización al reporte
func (rr *ReportRepository) AddReportUpdate(ctx context.Context, update *models.ReportUpdate) error {
	query := `
		INSERT INTO report_updates (report_id, operator_id, previous_status, new_status, 
								   update_type, comments, evidence_url, evidence_photos, 
								   internal_notes, estimated_completion)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at`

	err := rr.db.QueryRowContext(ctx, query,
		update.ReportID, update.OperatorID, update.PreviousStatus, update.NewStatus,
		update.UpdateType, update.Comments, update.EvidenceURL, update.EvidencePhotos,
		update.InternalNotes, update.EstimatedCompletion).Scan(
		&update.ID, &update.CreatedAt)

	if err != nil {
		return fmt.Errorf("failed to add report update: %v", err)
	}

	return nil
}

// GetReportUpdates obtiene las actualizaciones de un reporte
func (rr *ReportRepository) GetReportUpdates(ctx context.Context, reportID int) ([]*models.ReportUpdate, error) {
	query := `
		SELECT ru.id, ru.report_id, ru.operator_id, ru.previous_status, ru.new_status,
			   ru.update_type, ru.comments, ru.evidence_url, ru.evidence_photos,
			   ru.internal_notes, ru.estimated_completion, ru.created_at,
			   u.firebase_uid, u.email, u.full_name
		FROM report_updates ru
		JOIN users u ON ru.operator_id = u.id
		WHERE ru.report_id = $1
		ORDER BY ru.created_at ASC`

	rows, err := rr.db.QueryContext(ctx, query, reportID)
	if err != nil {
		return nil, fmt.Errorf("failed to get report updates: %v", err)
	}
	defer rows.Close()

	var updates []*models.ReportUpdate
	for rows.Next() {
		update := &models.ReportUpdate{Operator: &models.User{}}

		err := rows.Scan(
			&update.ID, &update.ReportID, &update.OperatorID, &update.PreviousStatus, &update.NewStatus,
			&update.UpdateType, &update.Comments, &update.EvidenceURL, &update.EvidencePhotos,
			&update.InternalNotes, &update.EstimatedCompletion, &update.CreatedAt,
			&update.Operator.FirebaseUID, &update.Operator.Email, &update.Operator.FullName)

		if err != nil {
			return nil, fmt.Errorf("failed to scan report update: %v", err)
		}

		updates = append(updates, update)
	}

	return updates, nil
}
