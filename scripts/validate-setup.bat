@echo off
echo ===========================================
echo    🎯 Aura CLI - Setup Validation
echo ===========================================
echo.

echo ✅ Checking Go installation...
go version
if %errorlevel% neq 0 (
    echo ❌ Go is not installed
    exit /b 1
)
echo.

echo ✅ Checking project build...
make build
if %errorlevel% neq 0 (
    echo ❌ Build failed
    exit /b 1
)
echo.

echo ✅ Testing Aura CLI binary...
.\bin\aura.exe --version
if %errorlevel% neq 0 (
    echo ❌ Binary test failed
    exit /b 1
)
echo.

echo ✅ Testing database functionality...
.\bin\aura.exe bookmark add validation-test .
.\bin\aura.exe bookmark list
if %errorlevel% neq 0 (
    echo ❌ Database test failed
    exit /b 1
)
echo.

echo ✅ Running tests...
make test
echo.

echo 🎉 SUCCESS! Your Aura CLI development environment is ready!
echo.
echo Next steps:
echo   1. Run 'make dev-run' to start development with hot reload
echo   2. Run 'make docker-dev' for Docker development environment
echo   3. Check DEVELOPMENT.md for detailed instructions
echo.
pause
