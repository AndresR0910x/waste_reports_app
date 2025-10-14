-- +goose Up
-- +goose StatementBegin

-- Tabla Reports - Reportes de ciudadanos sobre problemas urbanos
CREATE TABLE reports (
    id SERIAL PRIMARY KEY,
    user_id INT NOT NULL,
    title VARCHAR(255) NOT NULL, -- Título del reporte
    description TEXT NOT NULL, -- Descripción detallada del problema
    category VARCHAR(50) NOT NULL, -- Categoría: basura, alumbrado, baches, etc.
    location GEOGRAPHY(POINT) NOT NULL, -- Geolocalización usando PostGIS
    address TEXT, -- Dirección legible para humanos
    status VARCHAR(20) DEFAULT 'pending' 
        CHECK (status IN ('pending', 'in_progress', 'resolved', 'rejected')),
    priority VARCHAR(10) DEFAULT 'medium'
        CHECK (priority IN ('low', 'medium', 'high', 'urgent')),
    photo_url VARCHAR(255), -- URL de foto en Firebase Storage
    evidence_photos TEXT[], -- Array de URLs de fotos adicionales
    assigned_operator_id INT, -- Operador asignado
    estimated_resolution TIMESTAMP, -- Tiempo estimado de resolución
    actual_resolution TIMESTAMP, -- Tiempo real de resolución
    citizen_satisfaction_rating INT CHECK (citizen_satisfaction_rating BETWEEN 1 AND 5),
    citizen_feedback TEXT, -- Comentario del ciudadano sobre la resolución
    sync_status VARCHAR(20) DEFAULT 'pending' 
        CHECK (sync_status IN ('pending', 'synced', 'failed')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (assigned_operator_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Índices para optimización de consultas
CREATE INDEX idx_reports_user_id ON reports(user_id);
CREATE INDEX idx_reports_status ON reports(status);
CREATE INDEX idx_reports_category ON reports(category);
CREATE INDEX idx_reports_priority ON reports(priority);
CREATE INDEX idx_reports_assigned_operator ON reports(assigned_operator_id);
CREATE INDEX idx_reports_created_at ON reports(created_at);
CREATE INDEX idx_reports_location ON reports USING GIST (location);
CREATE INDEX idx_reports_sync_status ON reports(sync_status);

-- Trigger para actualizar updated_at
CREATE TRIGGER update_reports_updated_at BEFORE UPDATE ON reports
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_reports_updated_at ON reports;
DROP TABLE IF EXISTS reports CASCADE;

-- +goose StatementEnd