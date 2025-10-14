package models

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"
)

// Role representa un rol en el sistema
type Role struct {
	ID          int         `json:"id" db:"id"`
	Name        string      `json:"name" db:"name"`
	Description string      `json:"description" db:"description"`
	Permissions StringArray `json:"permissions" db:"permissions"`
	CreatedAt   time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at" db:"updated_at"`
}

// User representa un usuario sincronizado con Firebase
type User struct {
	ID           int         `json:"id" db:"id"`
	FirebaseUID  string      `json:"firebase_uid" db:"firebase_uid"`
	Email        *string     `json:"email,omitempty" db:"email"`
	RoleID       int         `json:"role_id" db:"role_id"`
	Cedula       *string     `json:"cedula,omitempty" db:"cedula"`
	Phone        *string     `json:"phone,omitempty" db:"phone"`
	FullName     *string     `json:"full_name,omitempty" db:"full_name"`
	IsActive     bool        `json:"is_active" db:"is_active"`
	LastLogin    *time.Time  `json:"last_login,omitempty" db:"last_login"`
	DeviceTokens StringArray `json:"device_tokens" db:"device_tokens"`
	CreatedAt    time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time   `json:"updated_at" db:"updated_at"`
	AnonymousID  string      `json:"anonymous_id" db:"anonymous_id"`

	// Campos relacionales (no almacenados en DB directamente)
	Role *Role `json:"role,omitempty"`
}

// Report representa un reporte de problema urbano
type Report struct {
	ID                        int         `json:"id" db:"id"`
	UserID                    int         `json:"user_id" db:"user_id"`
	Title                     string      `json:"title" db:"title"`
	Description               string      `json:"description" db:"description"`
	Category                  string      `json:"category" db:"category"`
	Location                  Location    `json:"location" db:"location"`
	Address                   *string     `json:"address,omitempty" db:"address"`
	Status                    string      `json:"status" db:"status"`
	Priority                  string      `json:"priority" db:"priority"`
	PhotoURL                  *string     `json:"photo_url,omitempty" db:"photo_url"`
	EvidencePhotos            StringArray `json:"evidence_photos" db:"evidence_photos"`
	AssignedOperatorID        *int        `json:"assigned_operator_id,omitempty" db:"assigned_operator_id"`
	EstimatedResolution       *time.Time  `json:"estimated_resolution,omitempty" db:"estimated_resolution"`
	ActualResolution          *time.Time  `json:"actual_resolution,omitempty" db:"actual_resolution"`
	CitizenSatisfactionRating *int        `json:"citizen_satisfaction_rating,omitempty" db:"citizen_satisfaction_rating"`
	CitizenFeedback           *string     `json:"citizen_feedback,omitempty" db:"citizen_feedback"`
	SyncStatus                string      `json:"sync_status" db:"sync_status"`
	CreatedAt                 time.Time   `json:"created_at" db:"created_at"`
	UpdatedAt                 time.Time   `json:"updated_at" db:"updated_at"`

	// Campos relacionales
	User             *User `json:"user,omitempty"`
	AssignedOperator *User `json:"assigned_operator,omitempty"`
}

// ReportUpdate representa una actualización de reporte
type ReportUpdate struct {
	ID                  int         `json:"id" db:"id"`
	ReportID            int         `json:"report_id" db:"report_id"`
	OperatorID          int         `json:"operator_id" db:"operator_id"`
	PreviousStatus      *string     `json:"previous_status,omitempty" db:"previous_status"`
	NewStatus           string      `json:"new_status" db:"new_status"`
	UpdateType          string      `json:"update_type" db:"update_type"`
	Comments            *string     `json:"comments,omitempty" db:"comments"`
	EvidenceURL         *string     `json:"evidence_url,omitempty" db:"evidence_url"`
	EvidencePhotos      StringArray `json:"evidence_photos" db:"evidence_photos"`
	InternalNotes       *string     `json:"internal_notes,omitempty" db:"internal_notes"`
	EstimatedCompletion *time.Time  `json:"estimated_completion,omitempty" db:"estimated_completion"`
	CreatedAt           time.Time   `json:"created_at" db:"created_at"`

	// Campos relacionales
	Report   *Report `json:"report,omitempty"`
	Operator *User   `json:"operator,omitempty"`
}

// AuditLog representa un registro de auditoría
type AuditLog struct {
	ID            int       `json:"id" db:"id"`
	EventType     string    `json:"event_type" db:"event_type"`
	EntityType    *string   `json:"entity_type,omitempty" db:"entity_type"`
	EntityID      *int      `json:"entity_id,omitempty" db:"entity_id"`
	UserID        *int      `json:"user_id,omitempty" db:"user_id"`
	FirebaseUID   *string   `json:"firebase_uid,omitempty" db:"firebase_uid"`
	AnonymousID   string    `json:"anonymous_id" db:"anonymous_id"`
	IPAddress     *string   `json:"ip_address,omitempty" db:"ip_address"`
	UserAgent     *string   `json:"user_agent,omitempty" db:"user_agent"`
	ActionDetails JSONData  `json:"action_details,omitempty" db:"action_details"`
	OldValues     JSONData  `json:"old_values,omitempty" db:"old_values"`
	NewValues     JSONData  `json:"new_values,omitempty" db:"new_values"`
	EventTime     time.Time `json:"event_time" db:"event_time"`
	SessionID     *string   `json:"session_id,omitempty" db:"session_id"`
	DeviceInfo    JSONData  `json:"device_info,omitempty" db:"device_info"`
	LocationInfo  JSONData  `json:"location_info,omitempty" db:"location_info"`
}

// Notification representa una notificación push
type Notification struct {
	ID               int        `json:"id" db:"id"`
	UserID           *int       `json:"user_id,omitempty" db:"user_id"`
	FirebaseUID      *string    `json:"firebase_uid,omitempty" db:"firebase_uid"`
	NotificationType string     `json:"notification_type" db:"notification_type"`
	Title            string     `json:"title" db:"title"`
	Body             string     `json:"body" db:"body"`
	Data             JSONData   `json:"data,omitempty" db:"data"`
	FCMToken         *string    `json:"fcm_token,omitempty" db:"fcm_token"`
	Topic            *string    `json:"topic,omitempty" db:"topic"`
	Status           string     `json:"status" db:"status"`
	SentAt           *time.Time `json:"sent_at,omitempty" db:"sent_at"`
	DeliveredAt      *time.Time `json:"delivered_at,omitempty" db:"delivered_at"`
	ErrorMessage     *string    `json:"error_message,omitempty" db:"error_message"`
	RetryCount       int        `json:"retry_count" db:"retry_count"`
	ScheduledFor     *time.Time `json:"scheduled_for,omitempty" db:"scheduled_for"`
	CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}

// FirebaseToken representa un token FCM de usuario
type FirebaseToken struct {
	ID          int       `json:"id" db:"id"`
	UserID      *int      `json:"user_id,omitempty" db:"user_id"`
	FirebaseUID *string   `json:"firebase_uid,omitempty" db:"firebase_uid"`
	Token       string    `json:"token" db:"token"`
	DeviceType  *string   `json:"device_type,omitempty" db:"device_type"`
	DeviceID    *string   `json:"device_id,omitempty" db:"device_id"`
	AppVersion  *string   `json:"app_version,omitempty" db:"app_version"`
	IsActive    bool      `json:"is_active" db:"is_active"`
	LastUsed    time.Time `json:"last_used" db:"last_used"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}

// Location representa un punto geográfico
type Location struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// Scan implementa sql.Scanner para Location
func (l *Location) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	// PostGIS devuelve POINT como texto en formato "POINT(lng lat)"
	if str, ok := value.(string); ok {
		var lat, lng float64
		n, err := fmt.Sscanf(str, "POINT(%f %f)", &lng, &lat)
		if err != nil || n != 2 {
			return fmt.Errorf("cannot parse location: %s", str)
		}
		l.Latitude = lat
		l.Longitude = lng
		return nil
	}

	return fmt.Errorf("cannot scan %T into Location", value)
}

// Value implementa driver.Valuer para Location
func (l Location) Value() (driver.Value, error) {
	return fmt.Sprintf("POINT(%f %f)", l.Longitude, l.Latitude), nil
}

// StringArray representa un array de strings para PostgreSQL
type StringArray []string

// Scan implementa sql.Scanner para StringArray
func (sa *StringArray) Scan(value interface{}) error {
	if value == nil {
		*sa = nil
		return nil
	}

	if b, ok := value.([]byte); ok {
		return json.Unmarshal(b, sa)
	}

	if s, ok := value.(string); ok {
		return json.Unmarshal([]byte(s), sa)
	}

	return fmt.Errorf("cannot scan %T into StringArray", value)
}

// Value implementa driver.Valuer para StringArray
func (sa StringArray) Value() (driver.Value, error) {
	if sa == nil {
		return nil, nil
	}
	return json.Marshal(sa)
}

// JSONData representa datos JSON para PostgreSQL
type JSONData map[string]interface{}

// Scan implementa sql.Scanner para JSONData
func (jd *JSONData) Scan(value interface{}) error {
	if value == nil {
		*jd = nil
		return nil
	}

	if b, ok := value.([]byte); ok {
		return json.Unmarshal(b, jd)
	}

	if s, ok := value.(string); ok {
		return json.Unmarshal([]byte(s), jd)
	}

	return fmt.Errorf("cannot scan %T into JSONData", value)
}

// Value implementa driver.Valuer para JSONData
func (jd JSONData) Value() (driver.Value, error) {
	if jd == nil {
		return nil, nil
	}
	return json.Marshal(jd)
}

// Constants para roles
const (
	RoleCiudadano     = "ciudadano"
	RoleOperador      = "operador"
	RoleAdministrador = "administrador"
)

// Constants para status de reportes
const (
	ReportStatusPending    = "pending"
	ReportStatusInProgress = "in_progress"
	ReportStatusResolved   = "resolved"
	ReportStatusRejected   = "rejected"
)

// Constants para prioridades
const (
	PriorityLow    = "low"
	PriorityMedium = "medium"
	PriorityHigh   = "high"
	PriorityUrgent = "urgent"
)

// Constants para tipos de actualización
const (
	UpdateTypeStatusChange = "status_change"
	UpdateTypeAssignment   = "assignment"
	UpdateTypeComment      = "comment"
	UpdateTypeEvidence     = "evidence"
)
