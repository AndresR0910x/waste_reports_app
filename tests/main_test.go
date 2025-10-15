package tests

import (
	"os"
	"testing"
)

// TestMain ejecuta configuración antes de todos los tests
func TestMain(m *testing.M) {
	// Configurar variables de entorno para testing
	os.Setenv("GIN_MODE", "test")
	os.Setenv("DB_HOST", "localhost")
	os.Setenv("DB_PORT", "5432")
	os.Setenv("DB_NAME", "test_residuos_db")
	os.Setenv("DB_USER", "test_user")
	os.Setenv("DB_PASSWORD", "test_password")
	os.Setenv("FIREBASE_PROJECT_ID", "test-project")

	// Ejecutar tests
	code := m.Run()

	// Limpiar después de los tests
	cleanup()

	// Salir con el código de resultado
	os.Exit(code)
}

// cleanup realiza limpieza después de los tests
func cleanup() {
	// Aquí puedes agregar lógica de limpieza si es necesaria
	// Por ejemplo, limpiar base de datos de test, archivos temporales, etc.
}
