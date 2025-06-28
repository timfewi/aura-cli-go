# 🚀 Aura CLI - Complete Setup Guide

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/dl/)
[![Docker](https://img.shields.io/badge/Docker-Required-blue.svg)](https://docs.docker.com/get-docker/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE-2.0.txt)

> **Step-by-step guide to set up Aura CLI**  
> Transform your command-line navigation with intelligent bookmarks, AI assistance, and context-aware suggestions.

---

## 📋 Prerequisites

- **Docker Desktop** (for database container)
- **PowerShell** (Windows) or **Bash/Zsh** (Linux/macOS)
- **Git** (to clone the repository)
- **Go** (if building from source)
- **OpenAI API Key** (optional, for AI features - can be set during installation)

**Check prerequisites:**
```bash
docker --version
git --version
go version
```

---

## 🛠️ Installation Methods

### Method 1: Quick Manual Setup (Recommended)

#### 1. Clone the Repository
```bash
git clone https://github.com/timfewi/aura-cli-go.git
cd aura-cli-go
```

#### 2. Build the Binary
```bash
go build -o bin/aura.exe cmd/aura/main.go
# Or use the Makefile
make build
```

#### 3. Install System-Wide

**Windows (PowerShell):**
```powershell
$localBin = "$env:USERPROFILE\bin"
New-Item -ItemType Directory -Path $localBin -Force
Copy-Item ".\bin\aura.exe" "$localBin\aura.exe"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($currentPath -notlike "*$localBin*") {
    $newPath = "$currentPath;$localBin"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    Write-Host "✅ Aura added to PATH"
}
# Restart PowerShell
```

**Linux/macOS (Bash/Zsh):**
```bash
mkdir -p ~/.local/bin
cp bin/aura ~/.local/bin/
echo 'export PATH="$HOME/.local/bin:$PATH"' >> ~/.bashrc
source ~/.bashrc
```

#### 4. Start Database Container
```bash
make start-db
# Or manually:
docker run -d --name aura-db --restart unless-stopped \
  -v aura-data:/data alpine:latest \
  sh -c "apk add --no-cache sqlite && \
  if [ ! -f /data/aura.db ]; then \
    sqlite3 /data/aura.db 'CREATE TABLE IF NOT EXISTS bookmarks (id INTEGER PRIMARY KEY AUTOINCREMENT, alias TEXT UNIQUE NOT NULL, path TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS navigation_history (id INTEGER PRIMARY KEY AUTOINCREMENT, path TEXT NOT NULL, accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP);'; \
  fi && \
  while true; do sleep 30; done"
```

#### 5. Verify Installation
```bash
aura --help
aura --version
aura bookmark list
```

---

### Method 2: Using Installation Script ⭐ **NEW: Interactive API Key Setup**

**Windows (PowerShell):**
```powershell
.\install.ps1
```

**Windows (Command Prompt):**
```cmd
install.bat
```

**Linux/macOS:**
```bash
chmod +x install.sh
./install.sh
```

> **🎉 New Feature**: All installation scripts now **automatically prompt you to set up your OpenAI API key** during installation! No more manual configuration required.

---

## 🗂️ Which Installation Script Should I Use?

Aura CLI supports all major operating systems. Choose the installation script that matches your environment:

- **Windows (PowerShell):**  
  Use [`install.ps1`](install.ps1) for the best experience in PowerShell.  
  ```powershell
  .\install.ps1
  ```

- **Windows (Command Prompt):**  
  Use [`install.bat`](install.bat) if you prefer the classic Command Prompt.  
  ```cmd
  REM Run in Command Prompt
  install.bat
  ```

- **Linux/macOS (Bash/Zsh):**  
  Use [`install.sh`](install.sh) for all Unix-like systems (Linux, macOS).  
  ```bash
  chmod +x install.sh
  ./install.sh
  ```

> **✨ Interactive Setup:**  
> All scripts now feature an **interactive AI setup wizard** that:
> - Checks if you already have an API key configured
> - Prompts you to enter your OpenAI API key during installation
> - Validates the key format
> - Saves it permanently to your environment
> - Shows different completion messages based on setup status

---

## 🔑 AI Features Setup

### Automatic Setup (Recommended)

When you run any installation script, you'll see:

```
🤖 AI Features Setup (Optional)
================================

Aura CLI can provide AI-powered assistance with:
  • Smart git commit messages (aura git commit)
  • Command line help (aura ask 'how to...')
  • Code explanations and suggestions

To enable AI features, you need an OpenAI API key.
Get one free at: https://platform.openai.com/api-keys

Would you like to set up AI features now? (y/N):
```

Simply:
1. Type `y` and press Enter
2. Get your API key from [platform.openai.com/api-keys](https://platform.openai.com/api-keys)
3. Paste it when prompted
4. Done! AI features are ready to use

### Manual Setup

If you skipped the automatic setup, you can still configure it manually:

**Windows:**
```powershell
[Environment]::SetEnvironmentVariable("AURA_API_KEY", "sk-your-actual-key", "User")
# Restart PowerShell
```

**Linux/macOS:**
```bash
echo 'export AURA_API_KEY="sk-your-actual-key"' >> ~/.bashrc
source ~/.bashrc
# Or for Zsh
echo 'export AURA_API_KEY="sk-your-actual-key"' >> ~/.zshrc
source ~/.zshrc
```

### Test AI Features
```bash
aura ask "Hello, can you help me?"
cd your-git-repo
git add .
aura git commit
```

---

## 🚀 Post-Installation: First Steps

### 1. Create Bookmarks
```bash
aura bookmark add home ~
aura bookmark add projects ~/projects
aura bookmark add docs ~/Documents
aura bookmark add this .
aura bookmark list
```

### 2. Test Navigation
```bash
aura go home
aura go proj
aura go
```

### 3. Try Context Actions
```bash
cd ~/projects/some-git-repo
aura do
cd ~/projects/node-app
aura do
```

### 4. Try AI Features (if configured)
```bash
aura ask "how to compress files with tar"
cd ~/your-git-repo
git add .
aura git commit  # AI will suggest a commit message!
```

---

## 🔧 Configuration

### Database Configuration

Aura uses a hybrid database approach:
- **Docker mode** (default): Database in container for AI access
- **Local mode** (fallback): SQLite file if Docker unavailable

**Verify Database:**
```bash
make db-status
docker ps | grep aura-db
```

**Advanced: `~/.config/aura/config.yaml`**
```yaml
database:
  mode: "docker"  # or "file"
  path: "/data/aura.db"
ai:
  provider: "openai"
  model: "gpt-4.1-nano"
```

---

## 🌟 Shell Integration

### Bash/Zsh Function
Add to `.bashrc` or `.zshrc`:
```bash
aura() {
    if [ "$1" = "go" ]; then
        local result
        result=$(command aura "$@")
        if [ $? -eq 0 ] && [ -d "$result" ]; then
            cd "$result"
        else
            echo "$result"
        fi
    else
        command aura "$@"
    fi
}
```

### PowerShell Function
Add to your PowerShell profile:
```powershell
function aura {
    if ($args[0] -eq "go") {
        $result = & "aura.exe" $args
        if ($LASTEXITCODE -eq 0 -and (Test-Path $result -PathType Container)) {
            Set-Location $result
        } else {
            Write-Output $result
        }
    } else {
        & "aura.exe" $args
    }
}
```

---

## 📊 System Requirements

- **RAM:** 100MB+ (recommended 200MB)
- **Disk:** 50MB+ (recommended 100MB)
- **OS:** Windows 10+, macOS 10.15+, Linux
- **Docker:** Required for database container

---

## 🔒 Security Considerations

- Store API keys in environment variables only
- Never commit secrets to version control
- Database runs in isolated Docker container
- No embedded credentials in binary
- Installation scripts use secure input for API keys

---

## 🔍 Troubleshooting

### "aura: command not found"
```bash
ls -la ~/.local/bin/aura
ls $env:USERPROFILE\bin\
echo $PATH
echo $env:PATH
# Restart terminal after PATH changes
```

### Database Issues
```bash
docker ps
make stop-db
make start-db
docker logs aura-db
```

### AI Features Not Working
```bash
echo $AURA_API_KEY
echo $env:AURA_API_KEY
aura ask "test" --verbose
```

### Permission Errors
```bash
chmod +x ~/.local/bin/aura
sudo cp bin/aura /usr/local/bin/
```

### API Key Issues

**Check if API key is set:**
```bash
# Linux/macOS
echo $AURA_API_KEY

# Windows PowerShell
echo $env:AURA_API_KEY

# Windows Command Prompt
echo %AURA_API_KEY%
```

**Reset API key:**
```bash
# Linux/macOS
unset AURA_API_KEY
export AURA_API_KEY="sk-your-new-key"

# Windows PowerShell
$env:AURA_API_KEY = $null
$env:AURA_API_KEY = "sk-your-new-key"

# Windows Command Prompt
set AURA_API_KEY=
set AURA_API_KEY=sk-your-new-key
```

---

## 🏗️ Development Setup

```bash
git clone https://github.com/aura-cli/aura-cli-go.git
cd aura-cli-go
go mod download
make build
make test
make build-all
```

**Development Commands:**
```bash
make dev
go run cmd/aura/main.go --help
make format
make lint
make clean
```

---

## 🎯 Usage Examples

**Navigation:**
```bash
aura bookmark add code ~/code
aura bookmark add downloads ~/Downloads
aura go code
aura go down
aura bookmark list
aura bookmark remove code
```

**Context-Aware Actions:**
```bash
cd ~/projects/react-app
aura do
cd ~/projects/python-app
aura do
cd ~/projects/git-repo
aura do
```

**AI Assistance:**
```bash
aura ask "how to compress files with tar"
cat script.py | aura ask "explain this code"
aura ask "what does this bash script do" < script.sh
git add .
aura git commit
```

---

## 🤝 Getting Help

- **README:** Basic usage and features
- **GitHub Issues:** Bug reports and feature requests
- **GitHub Discussions:** Community support

**Support:**
- [GitHub Issues](https://github.com/timfewi/aura-cli-go/issues)
- [GitHub Discussions](https://github.com/timfewi/aura-cli-go/discussions)

**Contributing:** See [CONTRIBUTING.md](.github\CONTRIBUTING.md)

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE-2.0.txt) file for details.

---

## 🎉 Congratulations!

You now have:
- ✅ Smart navigation with bookmarks
- ✅ AI-powered assistance (if configured)
- ✅ Context-aware suggestions
- ✅ Containerized database
- ✅ Cross-platform compatibility
- ✅ **Interactive setup experience**

*Happy navigating with Aura CLI!* ✨