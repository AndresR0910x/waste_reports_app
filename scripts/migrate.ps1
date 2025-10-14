# Script de PowerShell para gestión de migraciones
# Uso: ./migrate.ps1 [acción] [opciones]

param(
    [Parameter(Position=0)]
    [ValidateSet("up", "down", "status", "reset", "version", "validate", "help")]
    [string]$Action = "help",
    
    [Parameter()]
    [int64]$Version = 0,
    
    [Parameter()]
    [switch]$Force
)

# Configuración
$ProjectRoot = $PSScriptRoot
$MigrateCmd = "go run cmd/migrate/main.go"

function Show-Help {
    Write-Host "=== Sistema de Migraciones Backend Residuos App ===" -ForegroundColor Green
    Write-Host ""
    Write-Host "Uso: ./migrate.ps1 [ACCIÓN] [OPCIONES]" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "ACCIONES:" -ForegroundColor Cyan
    Write-Host "  up       - Ejecutar todas las migraciones pendientes"
    Write-Host "  down     - Revertir la última migración"
    Write-Host "  status   - Mostrar estado actual de migraciones"
    Write-Host "  reset    - Reiniciar completamente la base de datos (solo desarrollo)"
    Write-Host "  version  - Mostrar versión actual o migrar a versión específica"
    Write-Host "  validate - Validar que las migraciones sean correctas"
    Write-Host ""
    Write-Host "OPCIONES:" -ForegroundColor Cyan
    Write-Host "  -Version [número]  - Versión específica para migrar (usar con 'version')"
    Write-Host "  -Force            - Forzar operación sin confirmación"
    Write-Host ""
    Write-Host "EJEMPLOS:" -ForegroundColor Yellow
    Write-Host "  ./migrate.ps1 up                    # Ejecutar migraciones"
    Write-Host "  ./migrate.ps1 status                # Ver estado"
    Write-Host "  ./migrate.ps1 version -Version 3    # Migrar a versión 3"
    Write-Host "  ./migrate.ps1 reset -Force          # Reiniciar DB sin confirmación"
    Write-Host ""
}

function Test-Environment {
    # Verificar que estamos en el directorio correcto
    if (-not (Test-Path "go.mod")) {
        Write-Error "Error: Debes ejecutar este script desde la raíz del proyecto (donde está go.mod)"
        exit 1
    }
    
    # Verificar que Go está instalado
    try {
        $goVersion = go version
        Write-Host "✓ Go encontrado: $goVersion" -ForegroundColor Green
    } catch {
        Write-Error "Error: Go no está instalado o no está en el PATH"
        exit 1
    }
    
    # Verificar archivo .env
    if (-not (Test-Path ".env")) {
        Write-Warning "Archivo .env no encontrado. Usando variables de entorno del sistema."
    } else {
        Write-Host "✓ Archivo .env encontrado" -ForegroundColor Green
    }
}

function Invoke-Migration {
    param([string]$MigrationAction, [int64]$TargetVersion = 0)
    
    Write-Host "Ejecutando migración: $MigrationAction" -ForegroundColor Yellow
    
    try {
        if ($TargetVersion -gt 0) {
            $command = "$MigrateCmd -action=$MigrationAction -version=$TargetVersion"
        } else {
            $command = "$MigrateCmd -action=$MigrationAction"
        }
        
        Write-Host "Comando: $command" -ForegroundColor Gray
        Invoke-Expression $command
        
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✓ Migración completada exitosamente" -ForegroundColor Green
        } else {
            Write-Error "✗ Error en la migración (código: $LASTEXITCODE)"
            exit $LASTEXITCODE
        }
    } catch {
        Write-Error "Error ejecutando migración: $_"
        exit 1
    }
}

function Confirm-Action {
    param([string]$Message)
    
    if ($Force) {
        return $true
    }
    
    $response = Read-Host "$Message (y/N)"
    return ($response -eq 'y' -or $response -eq 'Y' -or $response -eq 'yes')
}

# Script principal
switch ($Action) {
    "help" {
        Show-Help
        exit 0
    }
    
    "up" {
        Test-Environment
        Write-Host "=== Ejecutando Migraciones ===" -ForegroundColor Green
        Invoke-Migration "up"
    }
    
    "down" {
        Test-Environment
        if (Confirm-Action "¿Estás seguro de que quieres revertir la última migración?") {
            Write-Host "=== Revirtiendo Migración ===" -ForegroundColor Yellow
            Invoke-Migration "down"
        } else {
            Write-Host "Operación cancelada" -ForegroundColor Gray
        }
    }
    
    "status" {
        Test-Environment
        Write-Host "=== Estado de Migraciones ===" -ForegroundColor Cyan
        Invoke-Migration "status"
    }
    
    "reset" {
        Test-Environment
        
        # Verificar que no estamos en producción
        $env = [Environment]::GetEnvironmentVariable("ENVIRONMENT")
        if ($env -eq "production") {
            Write-Error "ERROR: No se puede reiniciar la base de datos en producción"
            exit 1
        }
        
        $warningMessage = @"
⚠️  ADVERTENCIA: Esta operación eliminará TODOS los datos de la base de datos.
   Esta acción NO se puede deshacer.
   ¿Estás completamente seguro de que quieres continuar?
"@
        
        if (Confirm-Action $warningMessage) {
            Write-Host "=== Reiniciando Base de Datos ===" -ForegroundColor Red
            Invoke-Migration "reset"
        } else {
            Write-Host "Operación cancelada - Base de datos intacta" -ForegroundColor Green
        }
    }
    
    "version" {
        Test-Environment
        if ($Version -gt 0) {
            Write-Host "=== Migrando a Versión $Version ===" -ForegroundColor Cyan
            Invoke-Migration "version" $Version
        } else {
            Write-Host "=== Versión Actual ===" -ForegroundColor Cyan
            Invoke-Migration "version"
        }
    }
    
    "validate" {
        Test-Environment
        Write-Host "=== Validando Migraciones ===" -ForegroundColor Cyan
        Invoke-Migration "validate"
    }
    
    default {
        Write-Error "Acción desconocida: $Action"
        Show-Help
        exit 1
    }
}

Write-Host ""
Write-Host "Operación completada." -ForegroundColor Green