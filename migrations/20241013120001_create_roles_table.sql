-- +goose Up
-- +goose StatementBegin

-- Tabla Roles - Define los tipos de usuarios del sistema
CREATE TABLE roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,
    description TEXT,
    permissions TEXT[], -- Array de permisos para flexibilidad futura
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Insertar roles iniciales del sistema
INSERT INTO roles (name, description, permissions) VALUES 
    ('ciudadano', 'Usuario ciudadano que puede reportar problemas', ARRAY['create_report', 'view_own_reports']),
    ('operador', 'Operador que gestiona y actualiza reportes', ARRAY['view_reports', 'update_reports', 'view_analytics']),
    ('administrador', 'Administrador con acceso completo al sistema', ARRAY['all']);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS roles CASCADE;

-- +goose StatementEnd