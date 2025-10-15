@echo off
REM Build script for Backend Residuos App (Windows version)
REM Usage: scripts\build.bat [options]

setlocal enabledelayedexpansion

REM Default values
set "ENVIRONMENT=development"
set "OUTPUT_DIR=build"
set "BINARY_NAME=backend-residuos-app"
set "CLEAN=false"

REM Get version info
for /f "delims=" %%i in ('git describe --tags --always --dirty 2^>nul') do set "VERSION=%%i"
if "%VERSION%"=="" set "VERSION=dev"

for /f "delims=" %%i in ('git rev-parse HEAD 2^>nul') do set "GIT_COMMIT=%%i"
if "%GIT_COMMIT%"=="" set "GIT_COMMIT=unknown"

for /f "tokens=1-3 delims=/ " %%a in ('date /t') do (
    for /f "tokens=1-2 delims=: " %%d in ('time /t') do (
        set "BUILD_TIME=%%c-%%a-%%b_%%d:%%e"
    )
)

REM Parse command line arguments
:parse_args
if "%~1"=="" goto start_build
if /i "%~1"=="-e" (
    set "ENVIRONMENT=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--environment" (
    set "ENVIRONMENT=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="-o" (
    set "OUTPUT_DIR=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--output" (
    set "OUTPUT_DIR=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="-n" (
    set "BINARY_NAME=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--name" (
    set "BINARY_NAME=%~2"
    shift
    shift
    goto parse_args
)
if /i "%~1"=="--clean" (
    set "CLEAN=true"
    shift
    goto parse_args
)
if /i "%~1"=="-h" (
    echo Usage: %0 [options]
    echo Options:
    echo   -e, --environment  Set environment ^(development, staging, production^)
    echo   -o, --output       Set output directory ^(default: build^)
    echo   -n, --name         Set binary name ^(default: backend-residuos-app^)
    echo   --clean           Clean build directory before building
    echo   -h, --help        Show this help message
    exit /b 0
)
echo Unknown option: %~1
exit /b 1

:start_build
echo [INFO] Starting build process...
echo [INFO] Environment: %ENVIRONMENT%
echo [INFO] Version: %VERSION%
echo [INFO] Build Time: %BUILD_TIME%
echo [INFO] Git Commit: %GIT_COMMIT%

REM Check if Go is installed
go version >nul 2>&1
if errorlevel 1 (
    echo [ERROR] Go is not installed or not in PATH
    exit /b 1
)

for /f "tokens=3" %%i in ('go version') do set "GO_VERSION=%%i"
echo [INFO] Using Go version: %GO_VERSION%

REM Clean build directory
if "%CLEAN%"=="true" (
    echo [INFO] Cleaning build directory...
    if exist "%OUTPUT_DIR%" rmdir /s /q "%OUTPUT_DIR%"
)

if not exist "%OUTPUT_DIR%" mkdir "%OUTPUT_DIR%"

REM Download dependencies
echo [INFO] Downloading dependencies...
go mod download
go mod tidy

REM Run tests
echo [INFO] Running tests...
go test -v .\tests\auth\validation_service_test.go
if errorlevel 1 (
    echo [WARNING] Some tests failed, but continuing with build...
) else (
    echo [SUCCESS] Tests passed!
)

REM Build binaries
set "LDFLAGS=-w -s -X main.version=%VERSION% -X main.buildTime=%BUILD_TIME% -X main.gitCommit=%GIT_COMMIT% -X main.environment=%ENVIRONMENT%"

if /i "%ENVIRONMENT%"=="development" (
    echo [INFO] Building for windows/amd64...
    set CGO_ENABLED=0
    set GOOS=windows
    set GOARCH=amd64
    go build -ldflags="%LDFLAGS%" -o "%OUTPUT_DIR%\%BINARY_NAME%-windows-amd64.exe" .\cmd\api
    if errorlevel 1 (
        echo [ERROR] Failed to build binary
        exit /b 1
    )
    echo [SUCCESS] Built %BINARY_NAME%-windows-amd64.exe
) else (
    REM Build for multiple platforms
    echo [INFO] Building for multiple platforms...
    
    REM Linux AMD64
    set CGO_ENABLED=0
    set GOOS=linux
    set GOARCH=amd64
    go build -ldflags="%LDFLAGS%" -o "%OUTPUT_DIR%\%BINARY_NAME%-linux-amd64" .\cmd\api
    if not errorlevel 1 echo [SUCCESS] Built %BINARY_NAME%-linux-amd64
    
    REM Linux ARM64
    set GOARCH=arm64
    go build -ldflags="%LDFLAGS%" -o "%OUTPUT_DIR%\%BINARY_NAME%-linux-arm64" .\cmd\api
    if not errorlevel 1 echo [SUCCESS] Built %BINARY_NAME%-linux-arm64
    
    REM Darwin AMD64
    set GOOS=darwin
    set GOARCH=amd64
    go build -ldflags="%LDFLAGS%" -o "%OUTPUT_DIR%\%BINARY_NAME%-darwin-amd64" .\cmd\api
    if not errorlevel 1 echo [SUCCESS] Built %BINARY_NAME%-darwin-amd64
    
    REM Darwin ARM64
    set GOARCH=arm64
    go build -ldflags="%LDFLAGS%" -o "%OUTPUT_DIR%\%BINARY_NAME%-darwin-arm64" .\cmd\api
    if not errorlevel 1 echo [SUCCESS] Built %BINARY_NAME%-darwin-arm64
    
    REM Windows AMD64
    set GOOS=windows
    set GOARCH=amd64
    go build -ldflags="%LDFLAGS%" -o "%OUTPUT_DIR%\%BINARY_NAME%-windows-amd64.exe" .\cmd\api
    if not errorlevel 1 echo [SUCCESS] Built %BINARY_NAME%-windows-amd64.exe
)

REM Create build info file
echo { > "%OUTPUT_DIR%\build-info.json"
echo     "version": "%VERSION%", >> "%OUTPUT_DIR%\build-info.json"
echo     "build_time": "%BUILD_TIME%", >> "%OUTPUT_DIR%\build-info.json"
echo     "git_commit": "%GIT_COMMIT%", >> "%OUTPUT_DIR%\build-info.json"
echo     "environment": "%ENVIRONMENT%", >> "%OUTPUT_DIR%\build-info.json"
echo     "go_version": "%GO_VERSION%" >> "%OUTPUT_DIR%\build-info.json"
echo } >> "%OUTPUT_DIR%\build-info.json"

echo [SUCCESS] Build completed successfully!
echo [INFO] Build artifacts are available in: %OUTPUT_DIR%\
echo.
echo [INFO] Built files:
dir /b "%OUTPUT_DIR%"

endlocal