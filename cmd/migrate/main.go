package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"backend-residuos-app/pkg/config"
	"backend-residuos-app/pkg/database"
	"backend-residuos-app/pkg/migration"
)

func main() {
	var (
		action  = flag.String("action", "up", "Migration action: up, down, status, reset, version")
		version = flag.Int64("version", 0, "Target version for migration")
	)
	flag.Parse()

	// Cargar configuración
	cfg := config.LoadConfig()

	// Conectar a la base de datos
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Crear manager de migraciones
	migrationManager := migration.NewMigrationManager(db, cfg)

	// Ejecutar acción solicitada
	switch *action {
	case "up":
		log.Println("Running migrations...")
		if err := migrationManager.RunMigrations(); err != nil {
			log.Fatalf("Migration failed: %v", err)
		}
		log.Println("Migrations completed successfully")

	case "down":
		log.Println("Rolling back last migration...")
		if err := migrationManager.RollbackMigration(); err != nil {
			log.Fatalf("Rollback failed: %v", err)
		}
		log.Println("Rollback completed successfully")

	case "status":
		log.Println("Migration status:")
		if err := migrationManager.GetMigrationStatus(); err != nil {
			log.Fatalf("Failed to get migration status: %v", err)
		}

	case "reset":
		if cfg.Environment == "production" {
			log.Fatal("Database reset is not allowed in production environment")
		}

		fmt.Print("WARNING: This will delete all data. Are you sure? (yes/no): ")
		var confirm string
		fmt.Scanln(&confirm)

		if confirm != "yes" {
			log.Println("Reset cancelled")
			os.Exit(0)
		}

		log.Println("Resetting database...")
		if err := migrationManager.ResetDatabase(); err != nil {
			log.Fatalf("Database reset failed: %v", err)
		}
		log.Println("Database reset completed")

	case "version":
		currentVersion, err := migrationManager.GetVersion()
		if err != nil {
			log.Fatalf("Failed to get current version: %v", err)
		}
		log.Printf("Current database version: %d", currentVersion)

		if *version > 0 {
			log.Printf("Migrating to version: %d", *version)
			if err := migrationManager.MigrateToVersion(*version); err != nil {
				log.Fatalf("Migration to version %d failed: %v", *version, err)
			}
			log.Printf("Migration to version %d completed", *version)
		}

	case "validate":
		log.Println("Validating migrations...")
		if err := migrationManager.ValidateMigrations(); err != nil {
			log.Fatalf("Migration validation failed: %v", err)
		}
		log.Println("Migrations are valid")

	default:
		log.Printf("Unknown action: %s", *action)
		flag.Usage()
		os.Exit(1)
	}
}
