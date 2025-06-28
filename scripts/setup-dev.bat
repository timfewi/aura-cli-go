@echo off
REM Aura CLI Development Setup Script for Windows
setlocal EnableDelayedExpansion

echo 🚀 Aura CLI Development Setup (Windows)
echo ======================================
echo.

REM Check if Go is installed
where go >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] Go is not installed. Please install Go 1.21 or later.
    echo Visit: https://golang.org/dl/
    pause
    exit /b 1
)

echo [SUCCESS] Go is installed

REM Check if Docker is installed
where docker >nul 2>&1
if %errorlevel% neq 0 (
    echo [WARNING] Docker is not installed. Docker is recommended for development.
    echo Visit: https://docs.docker.com/get-docker/
    set DOCKER_AVAILABLE=false
) else (
    echo [SUCCESS] Docker is available
    set DOCKER_AVAILABLE=true
)

REM Setup directories
echo [INFO] Setting up directories...
if not exist "data\sqlite" mkdir data\sqlite
if not exist "logs" mkdir logs
if not exist "bin" mkdir bin
if not exist "config" mkdir config
echo [SUCCESS] Directories created

REM Install development dependencies
echo [INFO] Installing development dependencies...

where golangci-lint >nul 2>&1
if %errorlevel% neq 0 (
    echo [INFO] Installing golangci-lint...
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
)

where goimports >nul 2>&1
if %errorlevel% neq 0 (
    echo [INFO] Installing goimports...
    go install golang.org/x/tools/cmd/goimports@latest
)

where air >nul 2>&1
if %errorlevel% neq 0 (
    echo [INFO] Installing air for hot reloading...
    go install github.com/air-verse/air@latest
)

echo [SUCCESS] Development tools installed

REM Install Go dependencies
echo [INFO] Installing Go dependencies...
go mod download
go mod tidy
echo [SUCCESS] Go dependencies installed

REM Initialize SQLite database
echo [INFO] Initializing SQLite database...
if not exist "data\sqlite\aura.db" (
    type nul > data\sqlite\aura.db
    echo [INFO] Created database file
)

where sqlite3 >nul 2>&1
if %errorlevel% equ 0 (
    sqlite3 data\sqlite\aura.db < scripts\init-db.sql
    echo [SUCCESS] Database schema initialized
) else (
    echo [WARNING] sqlite3 command not found. Database will be initialized on first run.
)

REM Create configuration files
echo [INFO] Creating configuration files...

if not exist ".env.example" (
    (
        echo # Aura CLI Environment Configuration
        echo.
        echo # Database Configuration
        echo AURA_DB_PATH=./data/sqlite/aura.db
        echo AURA_ENV=development
        echo.
        echo # AI Configuration ^(optional^)
        echo AURA_API_KEY=your-openai-api-key-here
        echo AURA_API_URL=https://api.openai.com/v1
        echo AURA_MODEL=gpt-4
        echo.
        echo # Development Configuration
        echo AURA_LOG_LEVEL=debug
        echo AURA_LOG_FILE=./logs/aura.log
        echo.
        echo # Docker Configuration
        echo COMPOSE_PROJECT_NAME=aura-cli
    ) > .env.example
    echo [SUCCESS] Created .env.example
)

if not exist ".env" (
    copy .env.example .env >nul
    echo [SUCCESS] Created .env from template
    echo [WARNING] Please update .env with your actual configuration
)

REM Create air configuration
if not exist ".air.toml" (
    (
        echo root = "."
        echo testdata_dir = "testdata"
        echo tmp_dir = "tmp"
        echo.
        echo [build]
        echo   args_bin = []
        echo   bin = "./tmp/main.exe"
        echo   cmd = "go build -o ./tmp/main.exe ./cmd/aura"
        echo   delay = 1000
        echo   exclude_dir = ["assets", "tmp", "vendor", "testdata", "data", "logs", "bin"]
        echo   exclude_file = []
        echo   exclude_regex = ["_test.go"]
        echo   exclude_unchanged = false
        echo   follow_symlink = false
        echo   full_bin = ""
        echo   include_dir = []
        echo   include_ext = ["go", "tpl", "tmpl", "html"]
        echo   include_file = []
        echo   kill_delay = "0s"
        echo   log = "build-errors.log"
        echo   poll = false
        echo   poll_interval = 0
        echo   rerun = false
        echo   rerun_delay = 500
        echo   send_interrupt = false
        echo   stop_on_root = false
        echo.
        echo [color]
        echo   app = ""
        echo   build = "yellow"
        echo   main = "magenta"
        echo   runner = "green"
        echo   watcher = "cyan"
        echo.
        echo [log]
        echo   main_only = false
        echo   time = false
        echo.
        echo [misc]
        echo   clean_on_exit = false
        echo.
        echo [screen]
        echo   clear_on_rebuild = false
        echo   keep_scroll = true
    ) > .air.toml
    echo [SUCCESS] Created .air.toml for hot reloading
)

REM Create development script
if not exist "scripts\dev.bat" (
    (
        echo @echo off
        echo REM Quick development script for Aura CLI
        echo.
        echo if "%%1"=="build" ^(
        echo     echo Building Aura CLI...
        echo     make build
        echo ^) else if "%%1"=="test" ^(
        echo     echo Running tests...
        echo     make test
        echo ^) else if "%%1"=="lint" ^(
        echo     echo Running linter...
        echo     make lint
        echo ^) else if "%%1"=="format" ^(
        echo     echo Formatting code...
        echo     make format
        echo ^) else if "%%1"=="dev" ^(
        echo     echo Starting development server with hot reload...
        echo     air
        echo ^) else if "%%1"=="docker" ^(
        echo     echo Starting Docker development environment...
        echo     docker-compose up -d aura-dev
        echo     docker-compose exec aura-dev bash
        echo ^) else ^(
        echo     echo Usage: %%0 {build^|test^|lint^|format^|dev^|docker}
        echo     echo.
        echo     echo Commands:
        echo     echo   build   - Build the binary
        echo     echo   test    - Run tests
        echo     echo   lint    - Run linter
        echo     echo   format  - Format code
        echo     echo   dev     - Start development with hot reload
        echo     echo   docker  - Start Docker development environment
        echo ^)
    ) > scripts\dev.bat
    echo [SUCCESS] Created development script
)

REM Build the project
echo [INFO] Building project...
make build

if exist "bin\aura.exe" (
    echo [SUCCESS] Project built successfully
) else (
    echo [ERROR] Build failed
    pause
    exit /b 1
)

echo.
echo 🎉 Setup completed successfully!
echo.
echo [SUCCESS] Next steps:
echo   1. Update .env with your configuration
echo   2. Run 'make dev-run' to start development
echo   3. Run 'make docker-dev' for Docker development (if available)
echo   4. Run 'make test' to run tests
echo.
echo Happy coding! 🎯
pause
