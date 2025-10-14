package database

import (
	"database/sql"
	"fmt"
	"time"

	"backend-residuos-app/pkg/config"

	_ "github.com/lib/pq" // Driver PostgreSQL
)

type DB struct {
	*sql.DB
}

// Connect establece una conexión con la base de datos PostgreSQL
func Connect(cfg *config.Config) (*DB, error) {
	// Construir DSN (Data Source Name)
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=require",
		cfg.DBHost,
		cfg.DBPort,
		cfg.DBUser,
		cfg.DBPass,
		cfg.DBName,
	)

	// Abrir conexión
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, fmt.Errorf("error opening database connection: %v", err)
	}

	// Configurar pool de conexiones
	db.SetMaxOpenConns(25)                 // Máximo 25 conexiones abiertas
	db.SetMaxIdleConns(25)                 // Máximo 25 conexiones idle
	db.SetConnMaxLifetime(5 * time.Minute) // Tiempo de vida máximo de una conexión

	// Verificar conectividad
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("error pinging database: %v", err)
	}

	return &DB{db}, nil
}

// Close cierra la conexión a la base de datos
func (db *DB) Close() error {
	return db.DB.Close()
}

// Ping verifica si la conexión a la base de datos sigue activa
func (db *DB) Ping() error {
	return db.DB.Ping()
}

// GetStats retorna estadísticas de la conexión a la base de datos
func (db *DB) GetStats() sql.DBStats {
	return db.DB.Stats()
}
