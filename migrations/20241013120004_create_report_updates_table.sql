-- +goose Up
-- +goose StatementBegin

-- Tabla Report_Updates - Historial de actualizaciones de reportes
CREATE TABLE report_updates (
    id SERIAL PRIMARY KEY,
    report_id INT NOT NULL,
    operator_id INT NOT NULL,
    previous_status VARCHAR(20),
    new_status VARCHAR(20) NOT NULL 
        CHECK (new_status IN ('pending', 'in_progress', 'resolved', 'rejected')),
    update_type VARCHAR(30) NOT NULL 
        CHECK (update_type IN ('status_change', 'assignment', 'comment', 'evidence')),
    comments TEXT, -- Comentarios del operador
    evidence_url VARCHAR(255), -- URL de evidencia en Firebase Storage
    evidence_photos TEXT[], -- Array de URLs de fotos de evidencia
    internal_notes TEXT, -- Notas internas no visibles para ciudadanos
    estimated_completion TIMESTAMP, -- Nueva estimación de finalización
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (report_id) REFERENCES reports(id) ON DELETE CASCADE,
    FOREIGN KEY (operator_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Índices para optimización
CREATE INDEX idx_report_updates_report_id ON report_updates(report_id);
CREATE INDEX idx_report_updates_operator_id ON report_updates(operator_id);
CREATE INDEX idx_report_updates_created_at ON report_updates(created_at);
CREATE INDEX idx_report_updates_update_type ON report_updates(update_type);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS report_updates CASCADE;

-- +goose StatementEnd