# Aura CLI Makefile - Simplified

# Variables
BINARY_NAME=aura
MAIN_PATH=cmd/aura/main.go
BUILD_DIR=bin
VERSION?=1.0.0
LDFLAGS=-ldflags "-X main.version=$(VERSION)"

# Default target
.PHONY: help
help: ## Show this help message
	@echo "Aura CLI Build Commands:"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2}'

.PHONY: build
build: ## Build the binary
	@echo "Building Aura CLI..."
	@mkdir -p $(BUILD_DIR) 2>/dev/null || md $(BUILD_DIR) 2>nul || echo "Directory exists"
	go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME).exe $(MAIN_PATH)
	@echo "✓ Built $(BUILD_DIR)/$(BINARY_NAME).exe"

.PHONY: build-all
build-all: ## Build for all platforms
	@echo "Building for all platforms..."
	@mkdir -p $(BUILD_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)
	
	# macOS ARM64 (Apple Silicon)
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)
	
	@echo "✓ Built binaries for all platforms"

.PHONY: install
install: build ## Install the binary to system PATH
	@echo "Installing Aura CLI..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME).exe /usr/local/bin/$(BINARY_NAME)
	@echo "✓ Installed to /usr/local/bin/$(BINARY_NAME)"

.PHONY: test
test: ## Run tests
	@echo "Running tests..."
	go test -v ./...

.PHONY: lint
lint: ## Run linter
	@echo "Running linter..."
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	else \
		echo "golangci-lint not found. Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
	fi

.PHONY: format
format: ## Format code
	@echo "Formatting code..."
	gofmt -s -w .
	@if command -v goimports >/dev/null 2>&1; then \
		goimports -w .; \
	fi

.PHONY: clean
clean: ## Clean build artifacts
	@echo "Cleaning..."
	rm -rf $(BUILD_DIR)
	@echo "✓ Cleaned build directory"

.PHONY: deps
deps: ## Download dependencies
	@echo "Downloading dependencies..."
	go mod download
	go mod tidy
	@echo "✓ Dependencies updated"

.PHONY: quick-install
quick-install: ## Quick installation (runs install script)
	@echo "Running quick installation..."
	@if [ -f "install.sh" ]; then \
		chmod +x install.sh && ./install.sh; \
	elif [ -f "install.bat" ]; then \
		./install.bat; \
	else \
		echo "No installation script found"; \
		exit 1; \
	fi

.PHONY: start-db
start-db: ## Start the Aura database container
	@echo "Starting Aura database container..."
	docker run -d --name aura-db --restart unless-stopped -v aura-data:/data alpine:latest sh -c "apk add --no-cache sqlite && if [ ! -f /data/aura.db ]; then sqlite3 /data/aura.db 'CREATE TABLE IF NOT EXISTS bookmarks (id INTEGER PRIMARY KEY AUTOINCREMENT, alias TEXT UNIQUE NOT NULL, path TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS navigation_history (id INTEGER PRIMARY KEY AUTOINCREMENT, path TEXT NOT NULL, accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP);'; fi && while true; do sleep 30; done"
	@echo "✓ Database container started"

.PHONY: stop-db
stop-db: ## Stop the Aura database container
	@echo "Stopping Aura database container..."
	docker stop aura-db || true
	docker rm aura-db || true
	@echo "✓ Database container stopped"

.PHONY: db-status
db-status: ## Check database container status
	@echo "Aura database container status:"
	@docker ps -f name=aura-db --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}" || echo "Container not running"
