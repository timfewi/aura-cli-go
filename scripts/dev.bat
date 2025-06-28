@echo off
REM Quick development script for Aura CLI

if "%1"=="build" (
    echo Building Aura CLI...
    make build
) else if "%1"=="test" (
    echo Running tests...
    make test
) else if "%1"=="lint" (
    echo Running linter...
    make lint
) else if "%1"=="format" (
    echo Formatting code...
    make format
) else if "%1"=="dev" (
    echo Starting development server with hot reload...
    air
) else if "%1"=="docker" (
    echo Starting Docker development environment...
    docker-compose up -d aura-dev
    docker-compose exec aura-dev bash
) else (
    echo Usage: %0 {build^|test^|lint^|format^|dev^|docker}
    echo.
    echo Commands:
    echo   build   - Build the binary
    echo   test    - Run tests
    echo   lint    - Run linter
    echo   format  - Format code
    echo   dev     - Start development with hot reload
    echo   docker  - Start Docker development environment
)
