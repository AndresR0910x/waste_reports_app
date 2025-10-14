package migration

import (
	"embed"
	"fmt"
	"log"

	"backend-residuos-app/pkg/config"
	"backend-residuos-app/pkg/database"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var embedMigrations embed.FS

// MigrationManager gestiona las migraciones de la base de datos
type MigrationManager struct {
	db  *database.DB
	cfg *config.Config
}

// NewMigrationManager crea una nueva instancia del gestor de migraciones
func NewMigrationManager(db *database.DB, cfg *config.Config) *MigrationManager {
	return &MigrationManager{
		db:  db,
		cfg: cfg,
	}
}

// RunMigrations ejecuta todas las migraciones pendientes
func (mm *MigrationManager) RunMigrations() error {
	// Configurar goose para usar PostgreSQL
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	log.Println("Running database migrations...")

	// Ejecutar migraciones
	if err := goose.Up(mm.db.DB, "."); err != nil {
		return fmt.Errorf("failed to run migrations: %v", err)
	}

	log.Println("Migrations completed successfully")
	return nil
}

// RollbackMigration revierte la última migración
func (mm *MigrationManager) RollbackMigration() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	log.Println("Rolling back last migration...")

	if err := goose.Down(mm.db.DB, "."); err != nil {
		return fmt.Errorf("failed to rollback migration: %v", err)
	}

	log.Println("Migration rollback completed")
	return nil
}

// GetMigrationStatus obtiene el estado actual de las migraciones
func (mm *MigrationManager) GetMigrationStatus() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	return goose.Status(mm.db.DB, ".")
}

// CreateMigration crea una nueva migración (para desarrollo)
func (mm *MigrationManager) CreateMigration(name string) error {
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	return goose.Create(mm.db.DB, "./migrations", name, "sql")
}

// ValidateMigrations verifica que todas las migraciones sean válidas
func (mm *MigrationManager) ValidateMigrations() error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	// Verificar el estado de las migraciones
	return mm.GetMigrationStatus()
}

// ResetDatabase reinicia completamente la base de datos (solo para desarrollo)
func (mm *MigrationManager) ResetDatabase() error {
	if mm.cfg.Environment == "production" {
		return fmt.Errorf("database reset is not allowed in production")
	}

	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	log.Println("WARNING: Resetting database - all data will be lost!")

	// Revertir todas las migraciones
	if err := goose.Reset(mm.db.DB, "."); err != nil {
		return fmt.Errorf("failed to reset database: %v", err)
	}

	// Ejecutar todas las migraciones nuevamente
	if err := goose.Up(mm.db.DB, "."); err != nil {
		return fmt.Errorf("failed to run migrations after reset: %v", err)
	}

	log.Println("Database reset and migrations completed")
	return nil
}

// GetVersion obtiene la versión actual de la base de datos
func (mm *MigrationManager) GetVersion() (int64, error) {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return 0, fmt.Errorf("failed to set dialect: %v", err)
	}

	return goose.GetDBVersion(mm.db.DB)
}

// MigrateToVersion migra a una versión específica
func (mm *MigrationManager) MigrateToVersion(version int64) error {
	goose.SetBaseFS(embedMigrations)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("failed to set dialect: %v", err)
	}

	currentVersion, err := mm.GetVersion()
	if err != nil {
		return fmt.Errorf("failed to get current version: %v", err)
	}

	log.Printf("Migrating from version %d to version %d", currentVersion, version)

	if err := goose.UpTo(mm.db.DB, ".", version); err != nil {
		return fmt.Errorf("failed to migrate to version %d: %v", version, err)
	}

	log.Printf("Migration to version %d completed", version)
	return nil
}
