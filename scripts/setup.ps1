#!/usr/bin/env pwsh

# Setup script for Clean City Waste Management Backend
# This script helps initialize the development environment

param(
    [switch]$SkipDeps,
    [switch]$SkipDB,
    [switch]$SkipFirebase,
    [string]$Environment = "development"
)

$ErrorActionPreference = "Stop"
$ProjectRoot = Split-Path -Parent $PSScriptRoot

Write-Host "🚀 Setting up Clean City Backend Environment..." -ForegroundColor Green
Write-Host "Project Root: $ProjectRoot" -ForegroundColor Blue

# Function to check if command exists
function Test-Command($cmdname) {
    return [bool](Get-Command -Name $cmdname -ErrorAction SilentlyContinue)
}

# Check prerequisites
Write-Host "`n📋 Checking prerequisites..." -ForegroundColor Yellow

$prerequisites = @(
    @{ Name = "Go"; Command = "go"; Version = "go version" },
    @{ Name = "PostgreSQL"; Command = "psql"; Version = "psql --version" },
    @{ Name = "Git"; Command = "git"; Version = "git --version" }
)

foreach ($prereq in $prerequisites) {
    if (Test-Command $prereq.Command) {
        $version = & $prereq.Command version 2>$null | Select-Object -First 1
        Write-Host "  ✅ $($prereq.Name): $version" -ForegroundColor Green
    } else {
        Write-Host "  ❌ $($prereq.Name) is not installed or not in PATH" -ForegroundColor Red
        Write-Host "     Please install $($prereq.Name) and try again" -ForegroundColor Red
        exit 1
    }
}

# Set up project directory
Set-Location $ProjectRoot

# Install Go dependencies
if (-not $SkipDeps) {
    Write-Host "`n📦 Installing Go dependencies..." -ForegroundColor Yellow
    
    try {
        go mod download
        go mod tidy
        Write-Host "  ✅ Go dependencies installed successfully" -ForegroundColor Green
    } catch {
        Write-Host "  ❌ Failed to install Go dependencies: $_" -ForegroundColor Red
        exit 1
    }
}

# Create .env file if it doesn't exist
Write-Host "`n⚙️  Setting up environment configuration..." -ForegroundColor Yellow

if (-not (Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "  ✅ Created .env from .env.example" -ForegroundColor Green
    Write-Host "  ⚠️  Please update .env with your actual configuration values" -ForegroundColor Yellow
} else {
    Write-Host "  ℹ️  .env file already exists" -ForegroundColor Blue
}

# Firebase setup
if (-not $SkipFirebase) {
    Write-Host "`n🔥 Firebase configuration..." -ForegroundColor Yellow
    
    if (-not (Test-Path "firebase-service-account-key.json")) {
        Write-Host "  ⚠️  Firebase service account key not found" -ForegroundColor Yellow
        Write-Host "     Please download your Firebase service account key and save it as:" -ForegroundColor Yellow
        Write-Host "     firebase-service-account-key.json" -ForegroundColor Yellow
        Write-Host "     You can download it from: https://console.firebase.google.com/project/YOUR_PROJECT/settings/serviceaccounts/adminsdk" -ForegroundColor Blue
    } else {
        Write-Host "  ✅ Firebase service account key found" -ForegroundColor Green
    }
}

# Database setup
if (-not $SkipDB) {
    Write-Host "`n🗄️  Database setup..." -ForegroundColor Yellow
    
    # Load environment variables from .env
    if (Test-Path ".env") {
        Get-Content ".env" | ForEach-Object {
            if ($_ -match "^([^#][^=]*?)=(.*)$") {
                $name = $matches[1].Trim()
                $value = $matches[2].Trim()
                [Environment]::SetEnvironmentVariable($name, $value, "Process")
            }
        }
    }
    
    $dbHost = $env:DB_HOST ?? "localhost"
    $dbPort = $env:DB_PORT ?? "5432"
    $dbUser = $env:DB_USER ?? "postgres"
    $dbName = $env:DB_NAME ?? "residuos_db"
    
    Write-Host "  📡 Checking PostgreSQL connection..." -ForegroundColor Blue
    
    # Test PostgreSQL connection
    try {
        $env:PGPASSWORD = $env:DB_PASS
        $null = psql -h $dbHost -p $dbPort -U $dbUser -d postgres -c "\q" 2>$null
        Write-Host "  ✅ PostgreSQL connection successful" -ForegroundColor Green
        
        # Check if database exists
        $dbExists = psql -h $dbHost -p $dbPort -U $dbUser -d postgres -t -c "SELECT 1 FROM pg_database WHERE datname='$dbName';" 2>$null
        
        if ($dbExists -match "1") {
            Write-Host "  ✅ Database '$dbName' already exists" -ForegroundColor Green
        } else {
            Write-Host "  📝 Creating database '$dbName'..." -ForegroundColor Blue
            psql -h $dbHost -p $dbPort -U $dbUser -d postgres -c "CREATE DATABASE $dbName;" 2>$null
            Write-Host "  ✅ Database '$dbName' created successfully" -ForegroundColor Green
        }
        
        # Check PostGIS extension
        Write-Host "  🌍 Checking PostGIS extension..." -ForegroundColor Blue
        $postGISExists = psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -t -c "SELECT 1 FROM pg_extension WHERE extname='postgis';" 2>$null
        
        if ($postGISExists -match "1") {
            Write-Host "  ✅ PostGIS extension already installed" -ForegroundColor Green
        } else {
            Write-Host "  🔧 Installing PostGIS extension..." -ForegroundColor Blue
            psql -h $dbHost -p $dbPort -U $dbUser -d $dbName -c "CREATE EXTENSION IF NOT EXISTS postgis;" 2>$null
            Write-Host "  ✅ PostGIS extension installed successfully" -ForegroundColor Green
        }
        
    } catch {
        Write-Host "  ❌ PostgreSQL connection failed: $_" -ForegroundColor Red
        Write-Host "     Please ensure PostgreSQL is running and credentials are correct in .env" -ForegroundColor Yellow
    } finally {
        Remove-Item Env:\PGPASSWORD -ErrorAction SilentlyContinue
    }
}

# Run migrations
Write-Host "`n🔄 Running database migrations..." -ForegroundColor Yellow

try {
    go run cmd/migrate/main.go -action=up
    Write-Host "  ✅ Database migrations completed successfully" -ForegroundColor Green
} catch {
    Write-Host "  ❌ Migration failed: $_" -ForegroundColor Red
    Write-Host "     You can run migrations manually later with: go run cmd/migrate/main.go -action=up" -ForegroundColor Yellow
}

# Build the project
Write-Host "`n🔨 Building the project..." -ForegroundColor Yellow

try {
    go build -o bin/api cmd/api/main.go
    Write-Host "  ✅ Project built successfully" -ForegroundColor Green
    Write-Host "  📁 Executable created at: bin/api" -ForegroundColor Blue
} catch {
    Write-Host "  ❌ Build failed: $_" -ForegroundColor Red
}

# Create necessary directories
Write-Host "`n📁 Creating necessary directories..." -ForegroundColor Yellow

$directories = @("logs", "temp", "uploads", "bin")
foreach ($dir in $directories) {
    if (-not (Test-Path $dir)) {
        New-Item -ItemType Directory -Path $dir -Force | Out-Null
        Write-Host "  ✅ Created directory: $dir" -ForegroundColor Green
    } else {
        Write-Host "  ℹ️  Directory already exists: $dir" -ForegroundColor Blue
    }
}

# Setup complete
Write-Host "`n🎉 Setup completed successfully!" -ForegroundColor Green
Write-Host "`nNext steps:" -ForegroundColor Yellow
Write-Host "1. Update your .env file with actual configuration values" -ForegroundColor White
Write-Host "2. Add your Firebase service account key as firebase-service-account-key.json" -ForegroundColor White
Write-Host "3. Run the application with: go run cmd/api/main.go" -ForegroundColor White
Write-Host "4. Or use the built executable: ./bin/api" -ForegroundColor White
Write-Host "`nUseful commands:" -ForegroundColor Yellow
Write-Host "  Migration up:    go run cmd/migrate/main.go -action=up" -ForegroundColor White
Write-Host "  Migration down:  go run cmd/migrate/main.go -action=down" -ForegroundColor White
Write-Host "  Migration status: go run cmd/migrate/main.go -action=status" -ForegroundColor White
Write-Host "  Run tests:       go test ./..." -ForegroundColor White

Write-Host "`n🌟 Happy coding!" -ForegroundColor Green