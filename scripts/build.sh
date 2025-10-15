#!/bin/bash

# Build script for Backend Residuos App
# Usage: ./scripts/build.sh [options]

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="development"
OUTPUT_DIR="build"
BINARY_NAME="backend-residuos-app"
VERSION=$(git describe --tags --always --dirty 2>/dev/null || echo "dev")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -o|--output)
            OUTPUT_DIR="$2"
            shift 2
            ;;
        -n|--name)
            BINARY_NAME="$2"
            shift 2
            ;;
        --clean)
            CLEAN=true
            shift
            ;;
        -h|--help)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  -e, --environment  Set environment (development, staging, production)"
            echo "  -o, --output       Set output directory (default: build)"
            echo "  -n, --name         Set binary name (default: backend-residuos-app)"
            echo "  --clean           Clean build directory before building"
            echo "  -h, --help        Show this help message"
            exit 0
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if Go is installed
check_go() {
    if ! command -v go &> /dev/null; then
        log_error "Go is not installed or not in PATH"
        exit 1
    fi
    
    GO_VERSION=$(go version | awk '{print $3}')
    log_info "Using Go version: $GO_VERSION"
}

# Clean build directory
clean_build() {
    if [[ "$CLEAN" == true ]]; then
        log_info "Cleaning build directory..."
        rm -rf "$OUTPUT_DIR"
    fi
    
    mkdir -p "$OUTPUT_DIR"
}

# Download dependencies
download_deps() {
    log_info "Downloading dependencies..."
    go mod download
    go mod tidy
}

# Run tests
run_tests() {
    log_info "Running tests..."
    if go test -v ./tests/auth/validation_service_test.go; then
        log_success "Tests passed!"
    else
        log_warning "Some tests failed, but continuing with build..."
    fi
}

# Build for different platforms
build_binary() {
    local os=$1
    local arch=$2
    local output_name=$3
    
    log_info "Building for $os/$arch..."
    
    # Build flags
    local ldflags="-w -s"
    ldflags="$ldflags -X main.version=$VERSION"
    ldflags="$ldflags -X main.buildTime=$BUILD_TIME"
    ldflags="$ldflags -X main.gitCommit=$GIT_COMMIT"
    ldflags="$ldflags -X main.environment=$ENVIRONMENT"
    
    CGO_ENABLED=0 GOOS=$os GOARCH=$arch go build \
        -ldflags="$ldflags" \
        -o "$OUTPUT_DIR/$output_name" \
        ./cmd/api
    
    if [[ $? -eq 0 ]]; then
        local size=$(du -h "$OUTPUT_DIR/$output_name" | cut -f1)
        log_success "Built $output_name (size: $size)"
    else
        log_error "Failed to build $output_name"
        exit 1
    fi
}

# Main build process
main() {
    log_info "Starting build process..."
    log_info "Environment: $ENVIRONMENT"
    log_info "Version: $VERSION"
    log_info "Build Time: $BUILD_TIME"
    log_info "Git Commit: $GIT_COMMIT"
    
    # Pre-build checks
    check_go
    clean_build
    download_deps
    run_tests
    
    # Build for different platforms
    case "$ENVIRONMENT" in
        "development")
            # Build only for current platform in development
            if [[ "$OSTYPE" == "linux-gnu"* ]]; then
                build_binary "linux" "amd64" "$BINARY_NAME-linux-amd64"
            elif [[ "$OSTYPE" == "darwin"* ]]; then
                build_binary "darwin" "amd64" "$BINARY_NAME-darwin-amd64"
                build_binary "darwin" "arm64" "$BINARY_NAME-darwin-arm64"
            elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "win32" ]]; then
                build_binary "windows" "amd64" "$BINARY_NAME-windows-amd64.exe"
            fi
            ;;
        "staging"|"production")
            # Build for all platforms in staging/production
            build_binary "linux" "amd64" "$BINARY_NAME-linux-amd64"
            build_binary "linux" "arm64" "$BINARY_NAME-linux-arm64"
            build_binary "darwin" "amd64" "$BINARY_NAME-darwin-amd64"
            build_binary "darwin" "arm64" "$BINARY_NAME-darwin-arm64"
            build_binary "windows" "amd64" "$BINARY_NAME-windows-amd64.exe"
            ;;
    esac
    
    # Create build info file
    cat > "$OUTPUT_DIR/build-info.json" << EOF
{
    "version": "$VERSION",
    "build_time": "$BUILD_TIME",
    "git_commit": "$GIT_COMMIT",
    "environment": "$ENVIRONMENT",
    "go_version": "$GO_VERSION"
}
EOF

    log_success "Build completed successfully!"
    log_info "Build artifacts are available in: $OUTPUT_DIR/"
    
    # List built files
    echo ""
    log_info "Built files:"
    ls -la "$OUTPUT_DIR/"
}

# Run main function
main "$@"