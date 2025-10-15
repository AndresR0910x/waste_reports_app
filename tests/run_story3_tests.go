package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func main() {
	fmt.Println("🧪 Ejecutando tests para Story 3 (REQ-03): Creation of Online Reports")
	fmt.Println("=" * 70)

	// Tests a ejecutar
	testFiles := []string{
		"./tests/reports/cedula_validation_service_test.go",
		"./tests/reports/photo_upload_service_test.go",
		"./tests/reports/report_service_test.go",
		"./tests/reports/report_handler_test.go",
	}

	allPassed := true
	results := make(map[string]bool)

	for _, testFile := range testFiles {
		fmt.Printf("\n📋 Ejecutando tests de: %s\n", testFile)
		fmt.Println("-" * 50)

		// Ejecutar test
		cmd := exec.Command("go", "test", "-v", "-count=1", testFile)
		cmd.Dir = "."

		output, err := cmd.CombinedOutput()

		if err != nil {
			fmt.Printf("❌ FALLÓ: %s\n", testFile)
			fmt.Printf("Error: %v\n", err)
			fmt.Printf("Output:\n%s\n", string(output))
			allPassed = false
			results[testFile] = false
		} else {
			fmt.Printf("✅ PASÓ: %s\n", testFile)
			results[testFile] = true

			// Mostrar estadísticas de tests
			lines := strings.Split(string(output), "\n")
			for _, line := range lines {
				if strings.Contains(line, "PASS:") || strings.Contains(line, "FAIL:") {
					fmt.Printf("   %s\n", line)
				}
			}
		}
	}

	// Resumen final
	fmt.Println("\n" + "="*70)
	fmt.Println("📊 RESUMEN DE TESTS")
	fmt.Println("=" * 70)

	passed := 0
	total := len(testFiles)

	for testFile, result := range results {
		status := "❌ FALLÓ"
		if result {
			status = "✅ PASÓ"
			passed++
		}
		fmt.Printf("%s - %s\n", status, testFile)
	}

	fmt.Printf("\nResultado: %d/%d tests pasaron\n", passed, total)

	if allPassed {
		fmt.Println("🎉 ¡TODOS LOS TESTS PASARON!")
		fmt.Println("\n✅ Story 3 (REQ-03): Creation of Online Reports - Tests completados exitosamente")
		fmt.Println("✅ Componentes probados:")
		fmt.Println("   - Validación de cédulas ecuatorianas")
		fmt.Println("   - Subida de fotos a Firebase Storage")
		fmt.Println("   - Servicio de reportes con CRUD completo")
		fmt.Println("   - Handlers HTTP con validaciones de seguridad")
		os.Exit(0)
	} else {
		fmt.Println("💥 ALGUNOS TESTS FALLARON")
		fmt.Printf("❌ %d/%d tests fallaron\n", total-passed, total)
		fmt.Println("\n🔧 Por favor revisa los errores arriba y corrige los problemas")
		os.Exit(1)
	}
}
