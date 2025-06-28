@echo off
setlocal EnableDelayedExpansion

REM Aura CLI - One-Click Installation Script for Windows
REM Usage: install.bat

echo.
echo 🚀 Aura CLI - One-Click Installation
echo ====================================
echo.

echo [INFO] Installing Aura CLI - Intelligent Navigation Assistant
echo.

REM Determine installation directory based on privileges
net session >nul 2>&1
if %errorlevel% equ 0 (
    echo [INFO] Running as Administrator - Installing system-wide
    set "INSTALL_DIR=%PROGRAMFILES%\Aura"
    set "PATH_SCOPE=MACHINE"
) else (
    echo [INFO] Installing to user directory
    set "INSTALL_DIR=%USERPROFILE%\bin"
    set "PATH_SCOPE=USER"
)

REM Check if Go is available and install if needed
call :ensure_go
if %errorlevel% neq 0 exit /b 1

REM Build and install binary
call :install_binary
if %errorlevel% neq 0 exit /b 1

REM Create config
call :create_config
if %errorlevel% neq 0 exit /b 1

REM Setup PowerShell integration
call :setup_shell_integration
if %errorlevel% neq 0 exit /b 1

REM Setup API key (optional)
call :setup_api_key
if %errorlevel% neq 0 exit /b 1

REM Test installation
call :test_installation
if %errorlevel% neq 0 (
    echo [WARNING] Installation test had issues, but binary should be working
    echo [INFO] Try restarting Command Prompt and running: aura --version
)

echo.
echo [SUCCESS] 🎉 Aura CLI installed successfully!
echo.
echo [INFO] Quick start:
echo   1. Restart Command Prompt/PowerShell or refresh PATH

REM Check if API key was set
if defined AURA_API_KEY (
    echo   2. ✅ AI features are ready!
) else (
    echo   2. Set AI API key: set AURA_API_KEY=your-openai-api-key
)

echo   3. Try: aura --version
echo   4. Try: aura --help
echo   5. Try: aura bookmark add home %USERPROFILE%
echo   6. Try: aura go home
echo   7. Try: aura do

if defined AURA_API_KEY (
    echo   8. Try: aura ask "how to list files"
) else (
    echo   8. Try: aura ask "how to list files" ^(after setting API key^)
)

echo.
echo [INFO] Navigation made simple:
echo   - Add bookmark: aura bookmark add proj C:\projects
echo   - Navigate: aura go proj
echo   - Smart actions: aura do

if defined AURA_API_KEY (
    echo   - AI help: aura ask "git commands"
) else (
    echo   - AI help: aura ask "git commands" ^(after setting API key^)
)

echo.
echo [SUCCESS] Happy navigating! 🎯
echo.
echo [INFO] Need help? Check out the README.md or run: aura --help
echo.
pause
exit /b 0

:ensure_go
where go >nul 2>&1
if %errorlevel% equ 0 (
    for /f "tokens=*" %%i in ('go version') do set GO_VERSION=%%i
    echo [SUCCESS] Go is already installed: !GO_VERSION!
    exit /b 0
)

echo [WARNING] Go is not installed. Please install Go manually.
echo [INFO] Download from: https://golang.org/dl/
echo [INFO] After installing Go, re-run this script.
echo [ERROR] Go installation required to continue.
pause
exit /b 1

:install_binary
echo [INFO] Installing Aura CLI binary...

REM Create installation directory
if not exist "%INSTALL_DIR%" (
    mkdir "%INSTALL_DIR%"
    echo [SUCCESS] Created directory: %INSTALL_DIR%
)

REM Check for existing binary
if exist ".\bin\aura.exe" (
    echo [INFO] Found existing binary, installing...
    copy ".\bin\aura.exe" "%INSTALL_DIR%\aura.exe" >nul
) else (
    echo [INFO] No binary found, building from source...
    call :build_binary
    if %errorlevel% neq 0 exit /b 1
    copy ".\bin\aura.exe" "%INSTALL_DIR%\aura.exe" >nul
)

echo [SUCCESS] Binary copied to: %INSTALL_DIR%\aura.exe

REM Add to PATH
call :add_to_path "%INSTALL_DIR%"
exit /b 0

:build_binary
echo [INFO] Building Aura CLI from source...

if not exist "go.mod" (
    echo [ERROR] go.mod not found. Please run this script from the aura-cli-go directory.
    exit /b 1
)

if not exist "bin" mkdir bin

echo [INFO] Downloading Go dependencies...
go mod download
go mod tidy

echo [INFO] Compiling binary...
go build -ldflags="-s -w" -o "bin\aura.exe" ".\cmd\aura\main.go"

if exist "bin\aura.exe" (
    echo [SUCCESS] Binary built successfully: bin\aura.exe
    exit /b 0
) else (
    echo [ERROR] Failed to build binary
    exit /b 1
)

:add_to_path
set "DIR_TO_ADD=%~1"

REM Check if already in PATH
echo %PATH% | findstr /i "%DIR_TO_ADD%" >nul
if %errorlevel% equ 0 (
    echo [INFO] %DIR_TO_ADD% is already in PATH
    exit /b 0
) else (
    echo [INFO] Adding %DIR_TO_ADD% to PATH...
    
    if "%PATH_SCOPE%"=="MACHINE" (
        REM System-wide installation (requires admin)
        for /f "tokens=2*" %%A in ('reg query "HKLM\SYSTEM\CurrentControlSet\Control\Session Manager\Environment" /v PATH 2^>nul') do set "SYSTEM_PATH=%%B"
        setx PATH "%SYSTEM_PATH%;%DIR_TO_ADD%" /M >nul 2>&1
        if %errorlevel% equ 0 (
            echo [SUCCESS] Added to system PATH
        ) else (
            echo [WARNING] Failed to add to system PATH, trying user PATH...
            goto USER_PATH
        )
    ) else (
        :USER_PATH
        REM User installation
        for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v PATH 2^>nul') do set "USER_PATH=%%B"
        if "%USER_PATH%"=="" (
            setx PATH "%DIR_TO_ADD%" >nul 2>&1
        ) else (
            setx PATH "%USER_PATH%;%DIR_TO_ADD%" >nul 2>&1
        )
        
        if %errorlevel% equ 0 (
            echo [SUCCESS] Added to user PATH
        ) else (
            echo [WARNING] Failed to modify PATH automatically
            echo [INFO] Please manually add %DIR_TO_ADD% to your PATH
        )
    )
)

echo [SUCCESS] Aura binary installed to %INSTALL_DIR%\aura.exe
exit /b 0

:create_config
echo [INFO] Creating configuration...

set "CONFIG_DIR=%USERPROFILE%\.config\aura"
if not exist "%CONFIG_DIR%" (
    mkdir "%CONFIG_DIR%"
    echo [SUCCESS] Created config directory: %CONFIG_DIR%
)

if not exist "%CONFIG_DIR%\config.yaml" (
    (
        echo # Aura CLI Configuration
        echo database:
        echo   type: "local"
        echo   path: "%CONFIG_DIR%\aura.db"
        echo.
        echo ai:
        echo   provider: "openai"
        echo   model: "gpt-4.1-nano"
        echo   # Set AURA_API_KEY environment variable for AI features
        echo.
        echo navigation:
        echo   auto_bookmark_cwd: true
        echo   fuzzy_search: true
        echo   max_history: 1000
    ) > "%CONFIG_DIR%\config.yaml"
    echo [SUCCESS] Created config file: %CONFIG_DIR%\config.yaml
)

exit /b 0

:setup_shell_integration
echo [INFO] Setting up PowerShell integration...

set "PROFILE_PATH=%USERPROFILE%\Documents\PowerShell\Microsoft.PowerShell_profile.ps1"
if not exist "%USERPROFILE%\Documents\PowerShell" (
    mkdir "%USERPROFILE%\Documents\PowerShell"
)

REM Check if aura function already exists
if exist "%PROFILE_PATH%" (
    findstr /c:"function aura" "%PROFILE_PATH%" >nul 2>&1
    if %errorlevel% equ 0 (
        echo [INFO] PowerShell integration already exists
        exit /b 0
    )
)

(
    echo.
    echo # Aura CLI integration
    echo function aura {
    echo     if ($args[0] -eq "go"^) {
    echo         $result = ^& aura.exe $args
    echo         if ($LASTEXITCODE -eq 0 -and $result^) {
    echo             Set-Location $result
    echo         } else {
    echo             return $LASTEXITCODE
    echo         }
    echo     } else {
    echo         ^& aura.exe $args
    echo     }
    echo }
) >> "%PROFILE_PATH%"

echo [SUCCESS] PowerShell integration added to profile
echo [WARNING] Please restart PowerShell or run: . $PROFILE
exit /b 0

:setup_api_key
echo [INFO] Setting up AI features...

REM Check if API key is already set in environment
if defined AURA_API_KEY (
    echo [SUCCESS] AURA_API_KEY is already configured
    exit /b 0
)
if defined OPENAI_API_KEY (
    echo [SUCCESS] OPENAI_API_KEY is already configured
    exit /b 0
)

REM Check if API key is set in registry
for /f "tokens=2*" %%A in ('reg query "HKCU\Environment" /v AURA_API_KEY 2^>nul') do (
    echo [SUCCESS] AURA_API_KEY is already configured in registry
    exit /b 0
)

echo.
echo 🤖 AI Features Setup ^(Optional^)
echo ================================
echo.
echo Aura CLI can provide AI-powered assistance with:
echo   • Smart git commit messages ^(aura git commit^)
echo   • Command line help ^(aura ask "how to..."^)
echo   • Code explanations and suggestions
echo.
echo To enable AI features, you need an OpenAI API key.
echo Get one free at: https://platform.openai.com/api-keys
echo.

set /p "SETUP_AI=Would you like to set up AI features now? (y/N): "

if /i "%SETUP_AI%"=="y" (
    echo.
    echo Please enter your OpenAI API key:
    echo ^(It should start with 'sk-' and be about 51 characters long^)
    echo.
    
    set /p "API_KEY=API Key: "
    
    if defined API_KEY (
        if "!API_KEY:~0,3!"=="sk-" (
            REM Set the API key permanently
            setx AURA_API_KEY "!API_KEY!" >nul 2>&1
            if %errorlevel% equ 0 (
                echo [SUCCESS] API key saved successfully!
                echo [INFO] AI features are now available. Try: aura ask "Hello!"
                set "AURA_API_KEY=!API_KEY!"
            ) else (
                echo [WARNING] Failed to save API key permanently
                echo [INFO] You can set it manually later with:
                echo   setx AURA_API_KEY "your-api-key"
            )
        ) else (
            echo [WARNING] API key doesn't look correct ^(should start with 'sk-'^)
            echo [INFO] You can set it manually later with:
            echo   setx AURA_API_KEY "your-api-key"
        )
    ) else (
        echo [WARNING] API key is empty
        echo [INFO] You can set it manually later with:
        echo   setx AURA_API_KEY "your-api-key"
    )
) else (
    echo [INFO] Skipping AI setup. You can enable it later by setting:
    echo   setx AURA_API_KEY "your-openai-api-key"
    echo   Or set it temporarily: set AURA_API_KEY=your-api-key
)

exit /b 0

:test_installation
echo [INFO] Testing installation...

REM First try to find aura.exe in PATH
where aura.exe >nul 2>&1
if %errorlevel% equ 0 (
    REM Found in PATH
    echo [INFO] Found Aura binary in PATH
    for /f "tokens=*" %%i in ('aura.exe --version 2^>nul') do set AURA_VERSION=%%i
    if "!AURA_VERSION!"=="" set AURA_VERSION=unknown
    echo [SUCCESS] Aura CLI is working! Version: !AURA_VERSION!
    
    echo [INFO] Running basic functionality test...
    aura.exe bookmark add test-install "%CD%" >nul 2>&1
    aura.exe bookmark list >nul 2>&1 && echo [SUCCESS] Database connectivity: OK
    aura.exe bookmark remove test-install >nul 2>&1
    
    exit /b 0
) else (
    REM Not found in PATH, try direct path
    if exist "%INSTALL_DIR%\aura.exe" (
        echo [INFO] Found Aura binary at: %INSTALL_DIR%\aura.exe
        for /f "tokens=*" %%i in ('"%INSTALL_DIR%\aura.exe" --version 2^>nul') do set AURA_VERSION=%%i
        if "!AURA_VERSION!"=="" set AURA_VERSION=unknown
        echo [SUCCESS] Aura CLI is working! Version: !AURA_VERSION!
        
        echo [INFO] Running basic functionality test...
        "%INSTALL_DIR%\aura.exe" bookmark add test-install "%CD%" >nul 2>&1
        "%INSTALL_DIR%\aura.exe" bookmark list >nul 2>&1 && echo [SUCCESS] Database connectivity: OK
        "%INSTALL_DIR%\aura.exe" bookmark remove test-install >nul 2>&1
        
        echo [WARNING] Aura command not found in current PATH
        echo [INFO] Please restart Command Prompt to refresh PATH
        exit /b 0
    ) else (
        echo [ERROR] Installation test failed: 'aura' command not found
        exit /b 1
    )
)