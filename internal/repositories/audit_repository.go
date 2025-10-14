package repositories

import (
	"context"
	"fmt"
	"net"

	"backend-residuos-app/internal/models"
	"backend-residuos-app/pkg/database"
)

type AuditRepository struct {
	db *database.DB
}

func NewAuditRepository(db *database.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// CreateAuditLog crea una entrada de auditoría
func (ar *AuditRepository) CreateAuditLog(ctx context.Context, auditLog *models.AuditLog) error {
	query := `
		INSERT INTO audit_logs (event_type, entity_type, entity_id, user_id, firebase_uid,
							   anonymous_id, ip_address, user_agent, action_details, 
							   old_values, new_values, session_id, device_info, location_info)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, event_time`

	err := ar.db.QueryRowContext(ctx, query,
		auditLog.EventType, auditLog.EntityType, auditLog.EntityID, auditLog.UserID,
		auditLog.FirebaseUID, auditLog.AnonymousID, auditLog.IPAddress, auditLog.UserAgent,
		auditLog.ActionDetails, auditLog.OldValues, auditLog.NewValues, auditLog.SessionID,
		auditLog.DeviceInfo, auditLog.LocationInfo).Scan(
		&auditLog.ID, &auditLog.EventTime)

	if err != nil {
		return fmt.Errorf("failed to create audit log: %v", err)
	}

	return nil
}

// GetAuditLogsByUser obtiene logs de auditoría por usuario
func (ar *AuditRepository) GetAuditLogsByUser(ctx context.Context, userID int, limit, offset int) ([]*models.AuditLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, firebase_uid,
			   anonymous_id, ip_address, user_agent, action_details, old_values,
			   new_values, event_time, session_id, device_info, location_info
		FROM audit_logs
		WHERE user_id = $1
		ORDER BY event_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := ar.db.QueryContext(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by user: %v", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}

		err := rows.Scan(
			&log.ID, &log.EventType, &log.EntityType, &log.EntityID, &log.UserID,
			&log.FirebaseUID, &log.AnonymousID, &log.IPAddress, &log.UserAgent,
			&log.ActionDetails, &log.OldValues, &log.NewValues, &log.EventTime,
			&log.SessionID, &log.DeviceInfo, &log.LocationInfo)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %v", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditLogsByEntity obtiene logs de auditoría por entidad
func (ar *AuditRepository) GetAuditLogsByEntity(ctx context.Context, entityType string, entityID int, limit, offset int) ([]*models.AuditLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, firebase_uid,
			   anonymous_id, ip_address, user_agent, action_details, old_values,
			   new_values, event_time, session_id, device_info, location_info
		FROM audit_logs
		WHERE entity_type = $1 AND entity_id = $2
		ORDER BY event_time DESC
		LIMIT $3 OFFSET $4`

	rows, err := ar.db.QueryContext(ctx, query, entityType, entityID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by entity: %v", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}

		err := rows.Scan(
			&log.ID, &log.EventType, &log.EntityType, &log.EntityID, &log.UserID,
			&log.FirebaseUID, &log.AnonymousID, &log.IPAddress, &log.UserAgent,
			&log.ActionDetails, &log.OldValues, &log.NewValues, &log.EventTime,
			&log.SessionID, &log.DeviceInfo, &log.LocationInfo)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %v", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditLogsByEventType obtiene logs por tipo de evento
func (ar *AuditRepository) GetAuditLogsByEventType(ctx context.Context, eventType string, limit, offset int) ([]*models.AuditLog, error) {
	query := `
		SELECT id, event_type, entity_type, entity_id, user_id, firebase_uid,
			   anonymous_id, ip_address, user_agent, action_details, old_values,
			   new_values, event_time, session_id, device_info, location_info
		FROM audit_logs
		WHERE event_type = $1
		ORDER BY event_time DESC
		LIMIT $2 OFFSET $3`

	rows, err := ar.db.QueryContext(ctx, query, eventType, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get audit logs by event type: %v", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}

		err := rows.Scan(
			&log.ID, &log.EventType, &log.EntityType, &log.EntityID, &log.UserID,
			&log.FirebaseUID, &log.AnonymousID, &log.IPAddress, &log.UserAgent,
			&log.ActionDetails, &log.OldValues, &log.NewValues, &log.EventTime,
			&log.SessionID, &log.DeviceInfo, &log.LocationInfo)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %v", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// CreateAuditLogFromContext crea un log de auditoría extrayendo información del contexto HTTP
func (ar *AuditRepository) CreateAuditLogFromContext(ctx context.Context, eventType, entityType string, entityID, userID *int, firebaseUID *string, actionDetails map[string]interface{}) error {
	auditLog := &models.AuditLog{
		EventType:     eventType,
		EntityType:    &entityType,
		EntityID:      entityID,
		UserID:        userID,
		FirebaseUID:   firebaseUID,
		ActionDetails: actionDetails,
	}

	// Extraer información del contexto si está disponible
	if ipStr := ctx.Value("ip_address"); ipStr != nil {
		if ip := net.ParseIP(ipStr.(string)); ip != nil {
			ipString := ip.String()
			auditLog.IPAddress = &ipString
		}
	}

	if userAgent := ctx.Value("user_agent"); userAgent != nil {
		userAgentStr := userAgent.(string)
		auditLog.UserAgent = &userAgentStr
	}

	if sessionID := ctx.Value("session_id"); sessionID != nil {
		sessionIDStr := sessionID.(string)
		auditLog.SessionID = &sessionIDStr
	}

	if deviceInfo := ctx.Value("device_info"); deviceInfo != nil {
		if deviceMap, ok := deviceInfo.(map[string]interface{}); ok {
			auditLog.DeviceInfo = deviceMap
		}
	}

	if locationInfo := ctx.Value("location_info"); locationInfo != nil {
		if locationMap, ok := locationInfo.(map[string]interface{}); ok {
			auditLog.LocationInfo = locationMap
		}
	}

	// Generar ID anónimo si no se especifica
	if auditLog.AnonymousID == "" {
		if anonymousID := ctx.Value("anonymous_id"); anonymousID != nil {
			auditLog.AnonymousID = anonymousID.(string)
		} else {
			// Generar uno nuevo (esto debería manejarse mejor en producción)
			auditLog.AnonymousID = "unknown"
		}
	}

	return ar.CreateAuditLog(ctx, auditLog)
}

// GetSecurityEvents obtiene eventos de seguridad importantes
func (ar *AuditRepository) GetSecurityEvents(ctx context.Context, limit, offset int) ([]*models.AuditLog, error) {
	securityEvents := []string{
		"user_login", "user_logout", "failed_login", "password_reset",
		"account_locked", "permission_denied", "suspicious_activity",
		"data_export", "user_role_changed", "admin_action",
	}

	// Construir query con IN clause
	placeholders := ""
	args := make([]interface{}, len(securityEvents))
	for i, event := range securityEvents {
		if i > 0 {
			placeholders += ","
		}
		placeholders += fmt.Sprintf("$%d", i+1)
		args[i] = event
	}

	query := fmt.Sprintf(`
		SELECT id, event_type, entity_type, entity_id, user_id, firebase_uid,
			   anonymous_id, ip_address, user_agent, action_details, old_values,
			   new_values, event_time, session_id, device_info, location_info
		FROM audit_logs
		WHERE event_type IN (%s)
		ORDER BY event_time DESC
		LIMIT $%d OFFSET $%d`, placeholders, len(securityEvents)+1, len(securityEvents)+2)

	args = append(args, limit, offset)

	rows, err := ar.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get security events: %v", err)
	}
	defer rows.Close()

	var logs []*models.AuditLog
	for rows.Next() {
		log := &models.AuditLog{}

		err := rows.Scan(
			&log.ID, &log.EventType, &log.EntityType, &log.EntityID, &log.UserID,
			&log.FirebaseUID, &log.AnonymousID, &log.IPAddress, &log.UserAgent,
			&log.ActionDetails, &log.OldValues, &log.NewValues, &log.EventTime,
			&log.SessionID, &log.DeviceInfo, &log.LocationInfo)

		if err != nil {
			return nil, fmt.Errorf("failed to scan audit log: %v", err)
		}

		logs = append(logs, log)
	}

	return logs, nil
}

// GetAuditStats obtiene estadísticas de auditoría
func (ar *AuditRepository) GetAuditStats(ctx context.Context, days int) (map[string]interface{}, error) {
	query := `
		SELECT 
			event_type,
			COUNT(*) as count,
			COUNT(DISTINCT user_id) as unique_users,
			COUNT(DISTINCT anonymous_id) as unique_sessions
		FROM audit_logs
		WHERE event_time >= NOW() - INTERVAL '%d days'
		GROUP BY event_type
		ORDER BY count DESC`

	rows, err := ar.db.QueryContext(ctx, fmt.Sprintf(query, days))
	if err != nil {
		return nil, fmt.Errorf("failed to get audit stats: %v", err)
	}
	defer rows.Close()

	stats := make(map[string]interface{})
	eventStats := make([]map[string]interface{}, 0)

	for rows.Next() {
		var eventType string
		var count, uniqueUsers, uniqueSessions int

		err := rows.Scan(&eventType, &count, &uniqueUsers, &uniqueSessions)
		if err != nil {
			return nil, fmt.Errorf("failed to scan audit stats: %v", err)
		}

		eventStats = append(eventStats, map[string]interface{}{
			"event_type":      eventType,
			"count":           count,
			"unique_users":    uniqueUsers,
			"unique_sessions": uniqueSessions,
		})
	}

	stats["event_statistics"] = eventStats
	stats["period_days"] = days

	return stats, nil
}
