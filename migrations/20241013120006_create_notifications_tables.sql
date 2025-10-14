-- +goose Up
-- +goose StatementBegin

-- Tabla Notifications - Gestión de notificaciones push
CREATE TABLE notifications (
    id SERIAL PRIMARY KEY,
    user_id INT,
    firebase_uid VARCHAR(255), -- Para usuarios con Firebase
    notification_type VARCHAR(50) NOT NULL 
        CHECK (notification_type IN ('report_update', 'assignment', 'reminder', 'system', 'marketing')),
    title VARCHAR(255) NOT NULL,
    body TEXT NOT NULL,
    data JSONB, -- Datos adicionales para la notificación
    fcm_token TEXT, -- Token específico si se envía a un dispositivo
    topic VARCHAR(100), -- Tema si se envía a un grupo
    status VARCHAR(20) DEFAULT 'pending' 
        CHECK (status IN ('pending', 'sent', 'delivered', 'failed')),
    sent_at TIMESTAMP,
    delivered_at TIMESTAMP,
    error_message TEXT, -- Mensaje de error si falló
    retry_count INT DEFAULT 0,
    scheduled_for TIMESTAMP, -- Para notificaciones programadas
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Índices para notificaciones
CREATE INDEX idx_notifications_user_id ON notifications(user_id);
CREATE INDEX idx_notifications_firebase_uid ON notifications(firebase_uid);
CREATE INDEX idx_notifications_type ON notifications(notification_type);
CREATE INDEX idx_notifications_status ON notifications(status);
CREATE INDEX idx_notifications_scheduled_for ON notifications(scheduled_for);
CREATE INDEX idx_notifications_created_at ON notifications(created_at);

-- Tabla Firebase_Tokens - Gestión de tokens FCM por usuario
CREATE TABLE firebase_tokens (
    id SERIAL PRIMARY KEY,
    user_id INT,
    firebase_uid VARCHAR(255),
    token TEXT NOT NULL UNIQUE,
    device_type VARCHAR(20) CHECK (device_type IN ('android', 'ios', 'web')),
    device_id VARCHAR(255), -- ID único del dispositivo
    app_version VARCHAR(50), -- Versión de la app
    is_active BOOLEAN DEFAULT TRUE,
    last_used TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Índices para tokens
CREATE INDEX idx_firebase_tokens_user_id ON firebase_tokens(user_id);
CREATE INDEX idx_firebase_tokens_firebase_uid ON firebase_tokens(firebase_uid);
CREATE INDEX idx_firebase_tokens_token ON firebase_tokens(token);
CREATE INDEX idx_firebase_tokens_device_id ON firebase_tokens(device_id);
CREATE INDEX idx_firebase_tokens_is_active ON firebase_tokens(is_active);

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin

DROP TABLE IF EXISTS firebase_tokens CASCADE;
DROP TABLE IF EXISTS notifications CASCADE;

-- +goose StatementEnd