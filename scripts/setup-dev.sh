#!/bin/bash

# Aura CLI Development Setup Script
set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if running on Windows (Git Bash/WSL)
is_windows() {
    [[ "$OSTYPE" == "msys" || "$OSTYPE" == "cygwin" || -n "$WSL_DISTRO_NAME" ]]
}

# Check prerequisites
check_prerequisites() {
    print_status "Checking prerequisites..."
    
    # Check if Go is installed
    if ! command -v go &> /dev/null; then
        print_error "Go is not installed. Please install Go 1.21 or later."
        print_status "Visit: https://golang.org/dl/"
        exit 1
    fi
    
    # Check Go version
    GO_VERSION=$(go version | grep -oP 'go\K[0-9]+\.[0-9]+' || echo "0.0")
    REQUIRED_VERSION="1.21"
    if [[ "$(printf '%s\n' "$REQUIRED_VERSION" "$GO_VERSION" | sort -V | head -n1)" != "$REQUIRED_VERSION" ]]; then
        print_error "Go version $GO_VERSION is too old. Please install Go $REQUIRED_VERSION or later."
        exit 1
    fi
    
    print_success "Go version $GO_VERSION is compatible"
    
    # Check if Docker is installed
    if ! command -v docker &> /dev/null; then
        print_warning "Docker is not installed. Docker is recommended for development."
        print_status "Visit: https://docs.docker.com/get-docker/"
        DOCKER_AVAILABLE=false
    else
        print_success "Docker is available"
        DOCKER_AVAILABLE=true
    fi
    
    # Check if Docker Compose is available
    if [[ "$DOCKER_AVAILABLE" == "true" ]]; then
        if ! command -v docker-compose &> /dev/null && ! docker compose version &> /dev/null; then
            print_warning "Docker Compose is not available"
            DOCKER_COMPOSE_AVAILABLE=false
        else
            print_success "Docker Compose is available"
            DOCKER_COMPOSE_AVAILABLE=true
        fi
    fi
}

# Setup directories
setup_directories() {
    print_status "Setting up directories..."
    
    mkdir -p data/sqlite
    mkdir -p logs
    mkdir -p bin
    mkdir -p config
    
    print_success "Directories created"
}

# Install development dependencies
install_dev_deps() {
    print_status "Installing development dependencies..."
    
    # Install golangci-lint
    if ! command -v golangci-lint &> /dev/null; then
        print_status "Installing golangci-lint..."
        go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
    fi
    
    # Install goimports
    if ! command -v goimports &> /dev/null; then
        print_status "Installing goimports..."
        go install golang.org/x/tools/cmd/goimports@latest
    fi
    
    # Install air for hot reloading
    if ! command -v air &> /dev/null; then
        print_status "Installing air for hot reloading..."
        go install github.com/air-verse/air@latest
    fi
    
    print_success "Development tools installed"
}

# Download Go dependencies
install_go_deps() {
    print_status "Installing Go dependencies..."
    
    go mod download
    go mod tidy
    
    print_success "Go dependencies installed"
}

# Initialize SQLite database
init_database() {
    print_status "Initializing SQLite database..."
    
    # Create the database file if it doesn't exist
    if [[ ! -f "data/sqlite/aura.db" ]]; then
        touch data/sqlite/aura.db
        print_status "Created database file"
    fi
    
    # Initialize schema
    if command -v sqlite3 &> /dev/null; then
        sqlite3 data/sqlite/aura.db < scripts/init-db.sql
        print_success "Database schema initialized"
    else
        print_warning "sqlite3 command not found. Database will be initialized on first run."
    fi
}

# Create configuration files
create_config() {
    print_status "Creating configuration files..."
    
    # Create .env.example if it doesn't exist
    if [[ ! -f ".env.example" ]]; then
        cat > .env.example << 'EOF'
# Aura CLI Environment Configuration

# Database Configuration
AURA_DB_PATH=./data/sqlite/aura.db
AURA_ENV=development

# AI Configuration (optional)
AURA_API_KEY=your-openai-api-key-here
AURA_API_URL=https://api.openai.com/v1
AURA_MODEL=gpt-4.1-nano

# Development Configuration
AURA_LOG_LEVEL=debug
AURA_LOG_FILE=./logs/aura.log

# Docker Configuration
COMPOSE_PROJECT_NAME=aura-cli
EOF
        print_success "Created .env.example"
    fi
    
    # Create local .env if it doesn't exist
    if [[ ! -f ".env" ]]; then
        cp .env.example .env
        print_success "Created .env from template"
        print_warning "Please update .env with your actual configuration"
    fi
    
    # Create air configuration for hot reloading
    if [[ ! -f ".air.toml" ]]; then
        cat > .air.toml << 'EOF'
root = "."
testdata_dir = "testdata"
tmp_dir = "tmp"

[build]
  args_bin = []
  bin = "./tmp/main"
  cmd = "go build -o ./tmp/main ./cmd/aura"
  delay = 1000
  exclude_dir = ["assets", "tmp", "vendor", "testdata", "data", "logs", "bin"]
  exclude_file = []
  exclude_regex = ["_test.go"]
  exclude_unchanged = false
  follow_symlink = false
  full_bin = ""
  include_dir = []
  include_ext = ["go", "tpl", "tmpl", "html"]
  include_file = []
  kill_delay = "0s"
  log = "build-errors.log"
  poll = false
  poll_interval = 0
  rerun = false
  rerun_delay = 500
  send_interrupt = false
  stop_on_root = false

[color]
  app = ""
  build = "yellow"
  main = "magenta"
  runner = "green"
  watcher = "cyan"

[log]
  main_only = false
  time = false

[misc]
  clean_on_exit = false

[screen]
  clear_on_rebuild = false
  keep_scroll = true
EOF
        print_success "Created .air.toml for hot reloading"
    fi
}

# Setup Docker environment
setup_docker() {
    if [[ "$DOCKER_COMPOSE_AVAILABLE" == "true" ]]; then
        print_status "Setting up Docker development environment..."
        
        # Build development image
        docker-compose build aura-dev
        
        print_success "Docker development environment ready"
        print_status "Use 'make docker-dev' to start development environment"
    fi
}

# Create helpful scripts
create_scripts() {
    print_status "Creating development scripts..."
    
    # Create a simple development script
    cat > scripts/dev.sh << 'EOF'
#!/bin/bash
# Quick development script for Aura CLI

set -e

case "$1" in
    "build")
        echo "Building Aura CLI..."
        make build
        ;;
    "test")
        echo "Running tests..."
        make test
        ;;
    "lint")
        echo "Running linter..."
        make lint
        ;;
    "format")
        echo "Formatting code..."
        make format
        ;;
    "dev")
        echo "Starting development server with hot reload..."
        air
        ;;
    "docker")
        echo "Starting Docker development environment..."
        docker-compose up -d aura-dev
        docker-compose exec aura-dev bash
        ;;
    *)
        echo "Usage: $0 {build|test|lint|format|dev|docker}"
        echo ""
        echo "Commands:"
        echo "  build   - Build the binary"
        echo "  test    - Run tests"
        echo "  lint    - Run linter"
        echo "  format  - Format code"
        echo "  dev     - Start development with hot reload"
        echo "  docker  - Start Docker development environment"
        exit 1
        ;;
esac
EOF
    
    chmod +x scripts/dev.sh
    print_success "Created development script"
}

# Build the project
build_project() {
    print_status "Building project..."
    
    make build
    
    if [[ -f "bin/aura.exe" ]] || [[ -f "bin/aura" ]]; then
        print_success "Project built successfully"
    else
        print_error "Build failed"
        exit 1
    fi
}

# Main setup function
main() {
    echo "🚀 Aura CLI Development Setup"
    echo "=============================="
    echo ""
    
    check_prerequisites
    setup_directories
    install_dev_deps
    install_go_deps
    init_database
    create_config
    create_scripts
    
    if [[ "$DOCKER_COMPOSE_AVAILABLE" == "true" ]]; then
        setup_docker
    fi
    
    build_project
    
    echo ""
    echo "🎉 Setup completed successfully!"
    echo ""
    print_success "Next steps:"
    echo "  1. Update .env with your configuration"
    echo "  2. Run 'make dev-run' to start development"
    echo "  3. Run 'make docker-dev' for Docker development (if available)"
    echo "  4. Run 'make test' to run tests"
    echo ""
    print_status "Happy coding! 🎯"
}

# Run main function
main "$@"
