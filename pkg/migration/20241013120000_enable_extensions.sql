-- +goose Up
-- +goose StatementBegin

-- Habilitar la extensión PostGIS para geolocalización
CREATE EXTENSION IF NOT EXISTS postgis;

-- Habilitar UUID para identificadores únicos
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- +goose StatementEnd

-- +goose Down  
-- +goose StatementBegin

-- Remover extensiones si es necesario
DROP EXTENSION IF EXISTS postgis CASCADE;
DROP EXTENSION IF EXISTS "uuid-ossp";

-- +goose StatementEnd