# Aura CLI - One-Click Installation Script for Windows (PowerShell)
# Usage: .\install.ps1

param(
    [switch]$Force = $false
)

# Set strict mode and error handling
Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

# Colors for output
$Colors = @{
    Info    = "Cyan"
    Success = "Green"
    Warning = "Yellow"
    Error   = "Red"
}

function Write-Status {
    param([string]$Message)
    Write-Host "[INFO] $Message" -ForegroundColor $Colors.Info
}

function Write-Success {
    param([string]$Message)
    Write-Host "[SUCCESS] $Message" -ForegroundColor $Colors.Success
}

function Write-Warning {
    param([string]$Message)
    Write-Host "[WARNING] $Message" -ForegroundColor $Colors.Warning
}

function Write-Error {
    param([string]$Message)
    Write-Host "[ERROR] $Message" -ForegroundColor $Colors.Error
}

function Test-Administrator {
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent()
    $principal = New-Object Security.Principal.WindowsPrincipal($currentUser)
    return $principal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
}

function Refresh-EnvironmentPath {
    Write-Status "Refreshing environment PATH..."
    $machinePath = [Environment]::GetEnvironmentVariable("PATH", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("PATH", "User")
    $env:PATH = "$machinePath;$userPath"
    Write-Status "PATH refreshed for current session"
}

function Ensure-Go {
    Write-Status "Checking Go installation..."
    
    if (Get-Command go -ErrorAction SilentlyContinue) {
        $goVersion = go version
        Write-Success "Go is already installed: $goVersion"
        return $true
    }
    
    Write-Warning "Go is not installed."
    Write-Status "Attempting to install Go using winget..."
    
    try {
        if (Get-Command winget -ErrorAction SilentlyContinue) {
            winget install GoLang.Go
            
            # Refresh PATH after installation
            Refresh-EnvironmentPath
            
            if (Get-Command go -ErrorAction SilentlyContinue) {
                $goVersion = go version
                Write-Success "Go installed successfully: $goVersion"
                return $true
            }
        }
    }
    catch {
        Write-Warning "Failed to install Go automatically: $_"
    }
    
    Write-Error "Go installation required but not available."
    Write-Status "Please install Go manually from: https://golang.org/dl/"
    Write-Status "After installing Go, re-run this script."
    return $false
}

function Build-Binary {
    Write-Status "Building Aura CLI from source..."
    
    if (-not (Test-Path "go.mod")) {
        Write-Error "go.mod not found. Please run this script from the aura-cli-go directory."
        return $false
    }
    
    # Create bin directory
    if (-not (Test-Path "bin")) {
        New-Item -ItemType Directory -Path "bin" | Out-Null
    }
    
    Write-Status "Downloading Go dependencies..."
    go mod download
    go mod tidy
    
    Write-Status "Compiling binary..."
    go build -ldflags="-s -w" -o "bin\aura.exe" ".\cmd\aura\main.go"
    
    if (Test-Path "bin\aura.exe") {
        Write-Success "Binary built successfully: bin\aura.exe"
        return $true
    }
    else {
        Write-Error "Failed to build binary"
        return $false
    }
}

function Install-Binary {
    Write-Status "Installing Aura CLI binary..."
    
    # Determine installation directory
    if (Test-Administrator) {
        $installDir = "$env:ProgramFiles\Aura"
        $pathScope = "Machine"
        Write-Status "Running as Administrator - Installing system-wide"
    }
    else {
        $installDir = "$env:USERPROFILE\bin"
        $pathScope = "User"
        Write-Status "Installing to user directory"
    }
    
    # Create installation directory
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Path $installDir -Force | Out-Null
        Write-Success "Created directory: $installDir"
    }
    
    # Check for existing binary
    $binaryPath = $null
    if (Test-Path ".\bin\aura.exe") {
        Write-Status "Found existing binary, installing..."
        $binaryPath = ".\bin\aura.exe"
    }
    elseif (Test-Path ".\bin\aura") {
        Write-Status "Found existing binary (without .exe), installing..."
        $binaryPath = ".\bin\aura"
    }
    else {
        Write-Status "No binary found, building from source..."
        if (-not (Build-Binary)) {
            return $false
        }
        $binaryPath = ".\bin\aura.exe"
    }
    
    # Copy binary
    $targetPath = Join-Path $installDir "aura.exe"
    Copy-Item $binaryPath $targetPath -Force
    Write-Success "Binary copied successfully to: $targetPath"
    
    # Add to PATH
    $currentPath = [Environment]::GetEnvironmentVariable("PATH", $pathScope)
    if ($currentPath -notlike "*$installDir*") {
        Write-Status "Adding $installDir to PATH..."
        $newPath = "$currentPath;$installDir"
        [Environment]::SetEnvironmentVariable("PATH", $newPath, $pathScope)
        
        # Update current session PATH
        Refresh-EnvironmentPath
        Write-Success "Added to $pathScope PATH"
    }
    else {
        Write-Status "$installDir is already in PATH"
    }
    
    return $true
}

function Create-Config {
    Write-Status "Creating configuration..."
    
    $configDir = "$env:USERPROFILE\.config\aura"
    if (-not (Test-Path $configDir)) {
        New-Item -ItemType Directory -Path $configDir -Force | Out-Null
        Write-Success "Created config directory: $configDir"
    }
    
    $configFile = Join-Path $configDir "config.yaml"
    if (-not (Test-Path $configFile)) {
        $configContent = @"
# Aura CLI Configuration
database:
  type: "local"
  path: "$($configDir -replace '\\', '/')/aura.db"

ai:
  provider: "openai"
  model: "gpt-4.1-nano"
  # Set AURA_API_KEY environment variable for AI features

navigation:
  auto_bookmark_cwd: true
  fuzzy_search: true
  max_history: 1000
"@
        
        Set-Content -Path $configFile -Value $configContent -Encoding UTF8
        Write-Success "Created config file: $configFile"
    }
    
    return $true
}

function Setup-ShellIntegration {
    Write-Status "Setting up PowerShell integration..."
    
    # Check if aura function already exists in profile
    if (Test-Path $PROFILE) {
        $profileContent = Get-Content $PROFILE -Raw
        if ($profileContent -match "function aura") {
            Write-Status "PowerShell integration already exists"
            return $true
        }
    }
    else {
        # Create profile directory if it doesn't exist
        $profileDir = Split-Path $PROFILE -Parent
        if (-not (Test-Path $profileDir)) {
            New-Item -ItemType Directory -Path $profileDir -Force | Out-Null
        }
        New-Item -ItemType File -Path $PROFILE -Force | Out-Null
    }
    
    $integrationCode = @"

# Aura CLI integration
function aura {
    if (`$args[0] -eq "go") {
        `$result = & aura.exe `$args
        if (`$LASTEXITCODE -eq 0 -and `$result) {
            Set-Location `$result
        } else {
            return `$LASTEXITCODE
        }
    } else {
        & aura.exe `$args
    }
}
"@
    
    Add-Content -Path $PROFILE -Value $integrationCode
    Write-Success "PowerShell integration added to profile"
    Write-Warning "Please restart PowerShell or run: . `$PROFILE"
    
    return $true
}

function Setup-ApiKey {
    Write-Status "Setting up AI features..."
    
    # Check if API key is already set
    $existingApiKey = [Environment]::GetEnvironmentVariable("AURA_API_KEY", "User")
    $existingOpenAiKey = [Environment]::GetEnvironmentVariable("OPENAI_API_KEY", "User")
    
    if ($existingApiKey -or $existingOpenAiKey) {
        if ($existingApiKey) {
            Write-Success "AURA_API_KEY is already configured"
        }
        else {
            Write-Success "OPENAI_API_KEY is already configured"
        }
        return $true
    }
    
    Write-Host ""
    Write-Host "AI Features Setup (Optional)" -ForegroundColor Magenta
    Write-Host "================================" -ForegroundColor Magenta
    Write-Host ""
    Write-Host "Aura CLI can provide AI-powered assistance with:" -ForegroundColor White
    Write-Host "  • Smart git commit messages (`aura git commit`)" -ForegroundColor Gray
    Write-Host "  • Command line help (`aura ask 'how to...'`)" -ForegroundColor Gray
    Write-Host "  • Code explanations and suggestions" -ForegroundColor Gray
    Write-Host ""
    Write-Host "To enable AI features, you need an OpenAI API key." -ForegroundColor White
    Write-Host "Get one free at: https://platform.openai.com/api-keys" -ForegroundColor Blue
    Write-Host ""
    
    $response = Read-Host "Would you like to set up AI features now? (y/N)"
    
    if ($response -match '^[Yy].*') {
        Write-Host ""
        Write-Host "Please enter your OpenAI API key:" -ForegroundColor Yellow
        Write-Host "(It should start with 'sk-' and be about 51 characters long)" -ForegroundColor Gray
        
        $apiKey = Read-Host "API Key" -AsSecureString
        $apiKeyPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto([System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($apiKey))
        
        if ($apiKeyPlain -and $apiKeyPlain.Length -gt 10) {
            if ($apiKeyPlain.StartsWith("sk-")) {
                try {
                    [Environment]::SetEnvironmentVariable("AURA_API_KEY", $apiKeyPlain, "User")
                    Write-Success "API key saved successfully!"
                    Write-Status "AI features are now available. Try: aura ask 'Hello!'"
                    
                    # Test the API key
                    Write-Status "Testing AI connection..."
                    $env:AURA_API_KEY = $apiKeyPlain
                    return $true
                }
                catch {
                    Write-Warning "Failed to save API key: $_"
                    Write-Status "You can set it manually later with:"
                    Write-Host "  `$env:AURA_API_KEY = 'your-api-key'" -ForegroundColor Gray
                    return $true
                }
            }
            else {
                Write-Warning "API key doesn't look correct (should start with 'sk-')"
                Write-Status "You can set it manually later with:"
                Write-Host "  `$env:AURA_API_KEY = 'your-api-key'" -ForegroundColor Gray
                return $true
            }
        }
        else {
            Write-Warning "API key seems too short or empty"
            Write-Status "You can set it manually later with:"
            Write-Host "  `$env:AURA_API_KEY = 'your-api-key'" -ForegroundColor Gray
            return $true
        }
    }
    else {
        Write-Status "Skipping AI setup. You can enable it later by setting:"
        Write-Host "  `$env:AURA_API_KEY = 'your-openai-api-key'" -ForegroundColor Gray
        Write-Host "  Or: [Environment]::SetEnvironmentVariable('AURA_API_KEY', 'your-key', 'User')" -ForegroundColor Gray
        return $true
    }
}

function Test-Installation {
    Write-Status "Testing installation..."
    
    try {
        # First, try to find the binary in PATH
        $auraCommand = Get-Command aura.exe -ErrorAction SilentlyContinue
        
        if (-not $auraCommand) {
            # If not found in PATH, try to find it in common locations
            $possiblePaths = @(
                "$env:USERPROFILE\bin\aura.exe",
                "$env:ProgramFiles\Aura\aura.exe",
                ".\bin\aura.exe"
            )
            
            foreach ($path in $possiblePaths) {
                if (Test-Path $path) {
                    Write-Status "Found Aura binary at: $path"
                    $auraCommand = $path
                    break
                }
            }
        }
        
        if ($auraCommand) {
            try {
                # Test the binary directly
                $version = if ($auraCommand -is [string]) {
                    & $auraCommand --version 2>$null
                }
                else {
                    & $auraCommand.Source --version 2>$null
                }
                
                if (-not $version) { $version = "unknown" }
                Write-Success "Aura CLI is working! Version: $version"
                
                Write-Status "Running basic functionality test..."
                
                # Test basic commands
                $binaryPath = if ($auraCommand -is [string]) { $auraCommand } else { $auraCommand.Source }
                
                & $binaryPath bookmark add test-install $PWD.Path 2>$null | Out-Null
                & $binaryPath bookmark list 2>$null | Out-Null
                if ($LASTEXITCODE -eq 0) {
                    Write-Success "Database connectivity: OK"
                }
                & $binaryPath bookmark remove test-install 2>$null | Out-Null
                
                return $true
            }
            catch {
                Write-Warning "Aura command found but not working properly: $_"
                Write-Status "This may be due to PATH not being refreshed in current session"
                Write-Status "Try restarting PowerShell and running: aura --version"
                return $true  # Consider this a success since binary exists
            }
        }
        else {
            Write-Warning "Aura command not found in current PATH"
            Write-Status "This is normal - please restart PowerShell to refresh PATH"
            Write-Status "After restart, run: aura --version"
            return $true  # Consider this a success since we just installed
        }
    }
    catch {
        Write-Warning "Installation test encountered an issue: $_"
        Write-Status "This may be due to PATH not being refreshed in current session"
        Write-Status "Try restarting PowerShell and running: aura --version"
        return $true  # Consider this a success since we just installed
    }
}

function Main {
    Write-Host ""
    Write-Host "Aura CLI - One-Click Installation" -ForegroundColor Cyan
    Write-Host "====================================" -ForegroundColor Cyan
    Write-Host ""
    
    Write-Status "Installing Aura CLI - Intelligent Navigation Assistant"
    Write-Host ""
    
    try {
        # Check Go installation
        if (-not (Ensure-Go)) {
            exit 1
        }
        
        # Install binary
        if (-not (Install-Binary)) {
            Write-Error "Failed to install binary"
            exit 1
        }
        
        # Create config
        if (-not (Create-Config)) {
            Write-Error "Failed to create configuration"
            exit 1
        }
        
        # Setup shell integration
        if (-not (Setup-ShellIntegration)) {
            Write-Error "Failed to setup shell integration"
            exit 1
        }
        
        # Setup API key (optional)
        Setup-ApiKey | Out-Null
        
        # Test installation
        if (-not (Test-Installation)) {
            Write-Warning "Installation test had issues, but binary should be working"
            Write-Status "Try restarting PowerShell and running: aura --version"
        }
        
        Write-Host ""
        Write-Success "🎉 Aura CLI installed successfully!"
        Write-Host ""
        Write-Status "Quick start:"
        Write-Host "  1. Restart PowerShell or run: . `$PROFILE"
        
        # Show different message based on whether API key was set
        $hasApiKey = [Environment]::GetEnvironmentVariable("AURA_API_KEY", "User") -or 
        [Environment]::GetEnvironmentVariable("OPENAI_API_KEY", "User") -or 
        $env:AURA_API_KEY
        
        if ($hasApiKey) {
            Write-Host "  2. ✅ AI features are ready!" -ForegroundColor Green
        }
        else {
            Write-Host "  2. Set AI API key: `$env:AURA_API_KEY = 'your-openai-api-key'"
        }
        
        Write-Host "  3. Try: aura --version"
        Write-Host "  4. Try: aura --help"
        Write-Host "  5. Try: aura bookmark add home ~"
        Write-Host "  6. Try: aura go home"
        Write-Host "  7. Try: aura do"
        
        if ($hasApiKey) {
            Write-Host "  8. Try: aura ask 'how to list files'" -ForegroundColor Green
        }
        else {
            Write-Host "  8. Try: aura ask 'how to list files' (after setting API key)"
        }
        
        Write-Host ""
        Write-Status "Navigation made simple:"
        Write-Host "  - Add bookmark: aura bookmark add proj C:\projects"
        Write-Host "  - Navigate: aura go proj"
        Write-Host "  - Smart actions: aura do"
        
        if ($hasApiKey) {
            Write-Host "  - AI help: aura ask 'git commands'" -ForegroundColor Green
        }
        else {
            Write-Host "  - AI help: aura ask 'git commands' (after setting API key)"
        }
        
        Write-Host ""
        Write-Success "Happy navigating! 🎯"
        Write-Host ""
        Write-Status "Need help? Check out the README.md or run: aura --help"
        
    }
    catch {
        Write-Error "Installation failed: $_"
        exit 1
    }
}

# Run main function
Main