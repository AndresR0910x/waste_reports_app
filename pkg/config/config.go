package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	// Database configuration
	DBHost string
	DBPort string
	DBUser string
	DBPass string
	DBName string

	// Server configuration
	ServerPort string

	// Firebase configuration
	FirebaseCredentialsPath string
	FirebaseProjectID       string
	FirebaseStorageBucket   string

	// External APIs
	CedulaAPIURL string

	// Environment
	Environment string
}

func LoadConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}

	return &Config{
		// Database configuration
		DBHost: getEnv("DB_HOST", "localhost"),
		DBPort: getEnv("DB_PORT", "5432"),
		DBUser: getEnv("DB_USER", "postgres"),
		DBPass: getEnv("DB_PASS", ""),
		DBName: getEnv("DB_NAME", "residuos_db"),

		// Server configuration
		ServerPort: getEnv("SERVER_PORT", "8080"),

		// Firebase configuration
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", ""),
		FirebaseProjectID:       getEnv("FIREBASE_PROJECT_ID", ""),
		FirebaseStorageBucket:   getEnv("FIREBASE_STORAGE_BUCKET", ""),

		// External APIs
		CedulaAPIURL: getEnv("CEDULA_API_URL", ""),

		// Environment
		Environment: getEnv("ENVIRONMENT", "development"),
	}
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
