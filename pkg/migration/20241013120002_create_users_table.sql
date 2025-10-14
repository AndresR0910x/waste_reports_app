-- +goose Up
-- +goose StatementBegin

-- Tabla Users - Sincronizada con Firebase Authentication
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    firebase_uid VARCHAR(255) UNIQUE NOT NULL, -- UID único de Firebase Authentication
    email VARCHAR(255), -- Opcional para usuarios anónimos/ciudadanos
    role_id INT NOT NULL,
    cedula VARCHAR(20) UNIQUE, -- Cédula para validación de ciudadanos
    phone VARCHAR(20), -- Teléfono opcional
    full_name VARCHAR(255), -- Nombre completo
    is_active BOOLEAN DEFAULT TRUE, -- Estado del usuario
    last_login TIMESTAMP, -- Último login
    device_tokens TEXT[], -- Tokens FCM para notificaciones push
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    anonymous_id UUID DEFAULT gen_random_uuid(), -- Para auditoría anónima
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE RESTRICT
);

-- Índices para optimización de consultas
CREATE INDEX idx_users_firebase_uid ON users(firebase_uid);
CREATE INDEX idx_users_email ON users(email) WHERE email IS NOT NULL;
CREATE INDEX idx_users_cedula ON users(cedula) WHERE cedula IS NOT NULL;
CREATE INDEX idx_users_role_id ON users(role_id);
CREATE INDEX idx_users_is_active ON users(is_active);

-- Trigger para actualizar updated_at automáticamente
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_users_updated_at BEFORE UPDATE ON users
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS users CASCADE;

-- +goose StatementEnd