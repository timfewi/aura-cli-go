#!/bin/bash

# Aura CLI - One-Click Installation Script
# Usage: curl -fsSL https://raw.githubusercontent.com/aura-cli/aura/main/install.sh | bash
# Or: git clone <repo> && cd aura-cli-go && ./install.sh

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
MAGENTA='\033[0;35m'
CYAN='\033[0;36m'
WHITE='\033[1;37m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

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

# Detect OS and architecture
detect_platform() {
    local os=$(uname -s | tr '[:upper:]' '[:lower:]')
    local arch=$(uname -m)
    
    case "$arch" in
        x86_64) arch="amd64" ;;
        arm64|aarch64) arch="arm64" ;;
        i686|i386) arch="386" ;;
        *) 
            print_error "Unsupported architecture: $arch"
            exit 1
            ;;
    esac
    
    echo "${os}_${arch}"
}

# Install Go if not available
ensure_go() {
    print_status "Checking Go installation..."
    
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go is already installed: go version $go_version"
        return 0
    fi
    
    print_warning "Go is not installed. Attempting to install..."
    
    local platform=$(detect_platform)
    local go_version="1.21.5"
    local go_archive="go${go_version}.${platform}.tar.gz"
    local go_url="https://golang.org/dl/${go_archive}"
    local temp_dir=$(mktemp -d)
    
    print_status "Downloading Go ${go_version} for ${platform}..."
    
    case "$(uname -s)" in
        Linux*)
            print_status "Downloading Go for Linux..."
            curl -L "$go_url" -o "$temp_dir/$go_archive"
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "$temp_dir/$go_archive"
            
            # Add to PATH for current session
            export PATH="/usr/local/go/bin:$PATH"
            
            # Add to shell profile
            local shell_rc=""
            if [[ -f "$HOME/.bashrc" ]]; then
                shell_rc="$HOME/.bashrc"
            elif [[ -f "$HOME/.zshrc" ]]; then
                shell_rc="$HOME/.zshrc"
            fi
            
            if [[ -n "$shell_rc" ]] && ! grep -q "/usr/local/go/bin" "$shell_rc"; then
                echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$shell_rc"
            fi
            ;;
        Darwin*)
            print_status "Downloading Go for macOS..."
            curl -L "$go_url" -o "$temp_dir/$go_archive"
            sudo rm -rf /usr/local/go
            sudo tar -C /usr/local -xzf "$temp_dir/$go_archive"
            
            # Add to PATH for current session
            export PATH="/usr/local/go/bin:$PATH"
            
            # Add to shell profile
            local shell_rc=""
            if [[ -f "$HOME/.bash_profile" ]]; then
                shell_rc="$HOME/.bash_profile"
            elif [[ -f "$HOME/.zshrc" ]]; then
                shell_rc="$HOME/.zshrc"
            fi
            
            if [[ -n "$shell_rc" ]] && ! grep -q "/usr/local/go/bin" "$shell_rc"; then
                echo 'export PATH="/usr/local/go/bin:$PATH"' >> "$shell_rc"
            fi
            ;;
        *)
            print_error "Automatic Go installation not supported on this platform"
            print_status "Please install Go manually from: https://golang.org/dl/"
            exit 1
            ;;
    esac
    
    # Cleanup
    rm -rf "$temp_dir"
    
    # Verify installation
    if command -v go &> /dev/null; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go installed successfully: go version $go_version"
    else
        print_error "Go installation failed"
        exit 1
    fi
}

# Check if Docker is available
check_docker() {
    if command -v docker &> /dev/null && docker ps &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Build binary from source
build_binary() {
    print_status "Building Aura CLI from source..."
    
    if [[ ! -f "go.mod" ]]; then
        print_error "go.mod not found. Please run this script from the aura-cli-go directory."
        exit 1
    fi
    
    # Create bin directory
    mkdir -p bin
    
    print_status "Downloading Go dependencies..."
    go mod download
    go mod tidy
    
    print_status "Compiling binary..."
    go build -ldflags="-s -w" -o bin/aura ./cmd/aura/main.go
    
    if [[ -f "bin/aura" ]]; then
        print_success "Binary built successfully: bin/aura"
    else
        print_error "Failed to build binary"
        exit 1
    fi
}

# Install binary to system
install_binary() {
    print_status "Installing Aura CLI binary..."
    
    local install_dir="$HOME/.local/bin"
    
    # Create installation directory
    mkdir -p "$install_dir"
    
    # Check for existing binary
    if [[ -f "./bin/aura" ]]; then
        print_status "Found existing binary, installing..."
        cp "./bin/aura" "$install_dir/aura"
    else
        print_status "No binary found, building from source..."
        build_binary
        cp "./bin/aura" "$install_dir/aura"
    fi
    
    # Make executable
    chmod +x "$install_dir/aura"
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$install_dir:"* ]]; then
        print_status "Adding $install_dir to PATH..."
        
        # Determine shell profile
        local shell_rc=""
        local shell_name=$(basename "$SHELL")
        
        case "$shell_name" in
            bash) shell_rc="$HOME/.bashrc" ;;
            zsh) shell_rc="$HOME/.zshrc" ;;
            *) shell_rc="$HOME/.profile" ;;
        esac
        
        if [[ -n "$shell_rc" ]] && ! grep -q "$install_dir" "$shell_rc" 2>/dev/null; then
            echo "export PATH=\"$install_dir:\$PATH\"" >> "$shell_rc"
            print_success "Added to PATH in $shell_rc"
        fi
        
        # Update current session
        export PATH="$install_dir:$PATH"
    else
        print_status "$install_dir is already in PATH"
    fi
    
    print_success "Aura binary installed to $install_dir/aura"
}

# Start database container
start_database() {
    print_status "Setting up database..."
    
    if ! check_docker; then
        print_warning "Docker not available. Using local SQLite database."
        return 0
    fi
    
    # Check if container already exists
    if docker ps -a -q -f name=aura-db | grep -q .; then
        print_status "Aura database container already exists"
        
        # Start if not running
        if ! docker ps -q -f name=aura-db | grep -q .; then
            print_status "Starting existing database container..."
            docker start aura-db > /dev/null
        fi
        
        print_success "Database container is running"
        return 0
    fi
    
    print_status "Starting Aura database container..."
    
    # Create and start container
    docker run -d \
        --name aura-db \
        --restart unless-stopped \
        -v aura-data:/data \
        alpine:latest \
        sh -c "
            apk add --no-cache sqlite &&
            if [ ! -f /data/aura.db ]; then
                sqlite3 /data/aura.db 'CREATE TABLE IF NOT EXISTS bookmarks (id INTEGER PRIMARY KEY AUTOINCREMENT, alias TEXT UNIQUE NOT NULL, path TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS navigation_history (id INTEGER PRIMARY KEY AUTOINCREMENT, path TEXT NOT NULL, accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP);'
            fi &&
            while true; do sleep 30; done
        " > /dev/null
    
    # Wait for container to be ready
    print_status "Waiting for database to be ready..."
    sleep 3
    
    if docker ps -q -f name=aura-db | grep -q .; then
        print_success "Aura database container started successfully"
    else
        print_warning "Failed to start database container, will use local SQLite"
    fi
}

# Setup shell integration
setup_shell_integration() {
    print_status "Setting up shell integration..."
    
    local shell_rc=""
    local shell_name=$(basename "$SHELL")
    
    case "$shell_name" in
        bash)
            shell_rc="$HOME/.bashrc"
            ;;
        zsh)
            shell_rc="$HOME/.zshrc"
            ;;
        fish)
            print_warning "Fish shell detected. Manual setup required."
            print_status "Add this to your fish config:"
            echo "function aura"
            echo "    if test \$argv[1] = 'go'"
            echo "        set result (command aura \$argv)"
            echo "        if test \$status -eq 0 -a -n \"\$result\""
            echo "            cd \"\$result\""
            echo "        end"
            echo "    else"
            echo "        command aura \$argv"
            echo "    end"
            echo "end"
            return
            ;;
        *)
            print_warning "Unknown shell: $shell_name. Manual setup may be required."
            return
            ;;
    esac
    
    # Check if aura function already exists
    if grep -q "aura()" "$shell_rc" 2>/dev/null; then
        print_status "Shell integration already exists in $shell_rc"
        return
    fi
    
    # Add shell function
    cat >> "$shell_rc" << 'EOF'

# Aura CLI integration
aura() {
    if [[ "$1" == "go" ]]; then
        local result
        result=$(command aura "$@")
        if [[ $? -eq 0 && -n "$result" ]]; then
            cd "$result"
        else
            return 1
        fi
    else
        command aura "$@"
    fi
}
EOF
    
    print_success "Shell integration added to $shell_rc"
    print_warning "Please restart your shell or run: source $shell_rc"
}

# Create config
create_config() {
    local config_dir="$HOME/.config/aura"
    mkdir -p "$config_dir"
    
    # Determine database config based on Docker availability
    local db_config=""
    if check_docker > /dev/null 2>&1; then
        db_config='database:
  type: "docker"
  container: "aura-db"
  path: "/data/aura.db"'
    else
        db_config='database:
  type: "local"
  path: "'"$config_dir"'/aura.db"'
    fi
    
    cat > "$config_dir/config.yaml" << EOF
# Aura CLI Configuration
$db_config

ai:
  provider: "openai"
  model: "gpt-3.5-turbo"
  # Set AURA_API_KEY environment variable for AI features

navigation:
  auto_bookmark_cwd: true
  fuzzy_search: true
  max_history: 1000
EOF
    
    print_success "Configuration created at $config_dir/config.yaml"
}

# Setup API key
setup_api_key() {
    print_status "Setting up AI features..."
    
    # Check if API key is already set
    if [[ -n "$AURA_API_KEY" ]] || [[ -n "$OPENAI_API_KEY" ]]; then
        if [[ -n "$AURA_API_KEY" ]]; then
            print_success "AURA_API_KEY is already configured"
        else
            print_success "OPENAI_API_KEY is already configured"
        fi
        return 0
    fi
    
    # Check if API key is set in shell profile
    local shell_rc=""
    local shell_name=$(basename "$SHELL")
    
    case "$shell_name" in
        bash) shell_rc="$HOME/.bashrc" ;;
        zsh) shell_rc="$HOME/.zshrc" ;;
        *) shell_rc="$HOME/.profile" ;;
    esac
    
    if [[ -f "$shell_rc" ]] && grep -q "AURA_API_KEY\|OPENAI_API_KEY" "$shell_rc"; then
        print_success "API key is already configured in $shell_rc"
        return 0
    fi
    
    echo ""
    echo -e "${MAGENTA} AI Features Setup (Optional)${NC}"
    echo -e "${MAGENTA}================================${NC}"
    echo ""
    echo -e "${WHITE}Aura CLI can provide AI-powered assistance with:${NC}"
    echo -e "${GRAY}  • Smart git commit messages (\`aura git commit\`)${NC}"
    echo -e "${GRAY}  • Command line help (\`aura ask 'how to...'\`)${NC}"
    echo -e "${GRAY}  • Code explanations and suggestions${NC}"
    echo ""
    echo -e "${WHITE}To enable AI features, you need an OpenAI API key.${NC}"
    echo -e "${BLUE}Get one free at: https://platform.openai.com/api-keys${NC}"
    echo ""
    
    read -p "Would you like to set up AI features now? (y/N): " -n 1 -r
    echo
    
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo ""
        echo -e "${YELLOW}Please enter your OpenAI API key:${NC}"
        echo -e "${GRAY}(It should start with 'sk-' and be about 51 characters long)${NC}"
        echo ""
        
        read -p "API Key: " -s api_key
        echo
        
        if [[ -n "$api_key" && ${#api_key} -gt 10 ]]; then
            if [[ "$api_key" == sk-* ]]; then
                # Add to shell profile
                if [[ -n "$shell_rc" ]]; then
                    echo "" >> "$shell_rc"
                    echo "# Aura CLI - AI Features" >> "$shell_rc"
                    echo "export AURA_API_KEY=\"$api_key\"" >> "$shell_rc"
                    print_success "API key saved to $shell_rc"
                    print_status "AI features are now available. Try: aura ask 'Hello!'"
                    
                    # Set for current session
                    export AURA_API_KEY="$api_key"
                else
                    print_warning "Could not determine shell profile. Please set manually:"
                    echo "  export AURA_API_KEY=\"$api_key\""
                fi
            else
                print_warning "API key doesn't look correct (should start with 'sk-')"
                print_status "You can set it manually later with:"
                echo "  export AURA_API_KEY=\"your-api-key\""
            fi
        else
            print_warning "API key seems too short or empty"
            print_status "You can set it manually later with:"
            echo "  export AURA_API_KEY=\"your-api-key\""
        fi
    else
        print_status "Skipping AI setup. You can enable it later by setting:"
        echo "  export AURA_API_KEY=\"your-openai-api-key\""
        echo "  Or add it to your shell profile (~/.bashrc or ~/.zshrc)"
    fi
    
    return 0
}

# Test installation
test_installation() {
    print_status "Testing installation..."
    
    if command -v aura &> /dev/null; then
        local version=$(aura --version 2>/dev/null || echo "unknown")
        print_success "Aura CLI is working! Version: $version"
        
        # Test basic functionality
        print_status "Running basic functionality test..."
        aura bookmark add test-install "$PWD" > /dev/null 2>&1 || true
        aura bookmark list > /dev/null 2>&1 && print_success "Database connectivity: OK"
        aura bookmark remove test-install > /dev/null 2>&1 || true
        
        return 0
    else
        print_error "Installation test failed: 'aura' command not found"
        return 1
    fi
}

# Main installation function
main() {
    echo "Aura CLI - One-Click Installation"
    echo "===================================="
    echo ""
    
    print_status "Installing Aura CLI - Intelligent Navigation Assistant"
    echo ""
    
    # Ensure Go is available
    ensure_go
    
    # Check Docker (optional)
    check_docker > /dev/null 2>&1 || true
    
    # Build and install
    install_binary
    
    # Setup database
    start_database
    
    # Create config
    create_config
    
    # Setup shell integration
    setup_shell_integration
    
    # Setup API key (optional)
    setup_api_key
    
    # Test installation
    test_installation
    
    echo ""
    print_success "Aura CLI installed successfully!"
    echo ""
    print_status "Quick start:"
    echo "  1. Restart your shell or run: source ~/.bashrc (or ~/.zshrc)"
    echo "  2. Set AI API key: export AURA_API_KEY='your-openai-api-key'"
    echo "  3. Try: aura --help"
    echo "  4. Try: aura bookmark add home ~"
    echo "  5. Try: aura go home"
    echo "  6. Try: aura do"
    echo "  7. Try: aura ask 'how to list files'"
    echo ""
    print_status "Navigation made simple:"
    echo "  - Add bookmark: aura bookmark add proj ~/projects"
    echo "  - Navigate: aura go proj"
    echo "  - Smart actions: aura do"
    echo "  - AI help: aura ask 'git commands'"
    echo ""
    print_success "Happy navigating! 🎯"
    echo ""
    print_status "Need help? Check out the README.md or run: aura --help"
}

# Run main function
main "$@"
