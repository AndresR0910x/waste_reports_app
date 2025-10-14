-- +goose Up
-- +goose StatementBegin

-- Tabla Audit_Logs - Registro de auditoría para trazabilidad
CREATE TABLE audit_logs (
    id SERIAL PRIMARY KEY,
    event_type VARCHAR(50) NOT NULL, -- Tipo de evento (login, create_report, update_status, etc.)
    entity_type VARCHAR(50), -- Tipo de entidad afectada (user, report, etc.)
    entity_id INT, -- ID de la entidad afectada
    user_id INT, -- Usuario que realizó la acción
    firebase_uid VARCHAR(255), -- UID de Firebase del usuario
    anonymous_id UUID NOT NULL, -- ID anónimo para auditoría
    ip_address INET, -- Dirección IP del usuario
    user_agent TEXT, -- User agent del navegador/aplicación
    action_details JSONB, -- Detalles de la acción en formato JSON
    old_values JSONB, -- Valores anteriores (para updates)
    new_values JSONB, -- Valores nuevos (para updates)
    event_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    session_id VARCHAR(255), -- ID de sesión
    device_info JSONB, -- Información del dispositivo
    location_info JSONB -- Información de ubicación si está disponible
);

-- Índices para optimización de consultas de auditoría
CREATE INDEX idx_audit_logs_event_type ON audit_logs(event_type);
CREATE INDEX idx_audit_logs_entity_type ON audit_logs(entity_type);
CREATE INDEX idx_audit_logs_entity_id ON audit_logs(entity_id);
CREATE INDEX idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX idx_audit_logs_firebase_uid ON audit_logs(firebase_uid);
CREATE INDEX idx_audit_logs_anonymous_id ON audit_logs(anonymous_id);
CREATE INDEX idx_audit_logs_event_time ON audit_logs(event_time);
CREATE INDEX idx_audit_logs_session_id ON audit_logs(session_id);

-- Índice compuesto para consultas frecuentes
CREATE INDEX idx_audit_logs_composite ON audit_logs(event_type, entity_type, event_time);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS audit_logs CASCADE;

-- +goose StatementEnd