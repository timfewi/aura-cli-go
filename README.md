# 🚀 Aura CLI - Intelligent Navigation Assistant

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org/dl/)
[![Docker](https://img.shields.io/badge/Docker-Required-blue.svg)](https://docs.docker.com/get-docker/)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE-2.0.txt)

> **Navigate your filesystem like never before** - Aura replaces `cd` with intelligent bookmarks, AI assistance, and context-aware suggestions.

## ⚡ Installation

```bash
git clone https://github.com/timfewi/aura-cli-go.git
cd aura-cli-go
./install.sh
```

That's it! The installer will:
- ✅ Install Go if needed
- ✅ Build the binary
- ✅ Set up shell integration
- ✅ Configure database
- ✅ Test the installation

Restart your shell and run `aura --help` to get started!

---

## 🎯 Core Features

| Feature | Command | Description |
|---------|---------|-------------|
| **Smart Navigation** | `aura go project` | Jump to bookmarked directories instantly |
| **AI Assistant** | `aura ask "find large files"` | Get intelligent command suggestions |
| **Context Actions** | `aura do` | See relevant actions for your current project |
| **Auto Bookmarks** | `aura bookmark add proj .` | Save locations with memorable names |

---

## 🚀 Quick Start

After installation, restart your shell and try these commands:

### 1. Create Your First Bookmark
```bash
cd ~/Documents/projects
aura bookmark add projects .
```

### 2. Navigate Anywhere
```bash
aura go projects    # Jump to ~/Documents/projects instantly
aura go proj        # Fuzzy matching works too!
```

### 3. Get Smart Suggestions
```bash
cd your-git-repo
aura do             # Shows: git status, git push, npm install, etc.
```

### 4. Ask AI for Help
```bash
aura ask "how to find files larger than 100MB"
aura ask "git commands for beginners"
echo "console.log('hello')" | aura ask "explain this code"
```

---

## 🔧 Configuration

### AI Features (Optional)
Set your OpenAI API key to enable AI assistance:

```bash
# Add to your shell profile (.bashrc, .zshrc, etc.)
export AURA_API_KEY="your-openai-api-key"
```

### Database Location
Aura automatically uses a Docker container for the database. If Docker isn't available, it falls back to a local SQLite file.

**For AI Model Access**: The database runs in a Docker container with volume `aura-data:/data`, making it easily accessible to other AI systems.

---

## 📱 Usage Examples

### Navigation
```bash
# Add bookmarks
aura bookmark add code ~/code
aura bookmark add docs ~/Documents
aura bookmark add this .                    # Current directory

# Navigate
aura go code                               # Jump to ~/code
aura go doc                                # Fuzzy search finds ~/Documents
aura go                                    # Interactive selection

# List bookmarks
aura bookmark list
```

### Context-Aware Actions
```bash
# In a Git repository
aura do
# Shows: git status, git commit, git push, etc.

# In a Node.js project  
aura do
# Shows: npm install, npm start, npm test, etc.

# In any directory
aura do
# Shows: list files, find large files, disk usage, etc.
```

### AI Assistance
```bash
# Get command help
aura ask "compress a folder with tar"
aura ask "docker commands cheat sheet"

# Explain code
cat script.py | aura ask "what does this do"

# Generate git commits (in a git repo with staged changes)
aura git commit                            # AI generates commit message
```

---

## 🏗️ Architecture

Aura is designed for **simplicity** and **AI integration**:

```
┌─────────────────┐    ┌──────────────────┐    ┌─────────────────┐
│   Aura CLI      │────│  Docker SQLite   │────│  AI Models      │
│   (Go Binary)   │    │   (Container)    │    │  (External)     │
└─────────────────┘    └──────────────────┘    └─────────────────┘
        │                       │                       │
        ▼                       ▼                       ▼
   Navigation              Bookmarks              Context Data
   Commands                History                for AI
```

**Key Benefits:**
- **Zero Dependencies**: Pure Go binary with no external requirements
- **Containerized Data**: Database accessible to AI models via Docker
- **Cross-Platform**: Works on Windows, macOS, and Linux
- **Offline First**: Core navigation works without internet

---

## 🛠️ Development

### Building from Source
```bash
git clone https://github.com/aura-cli/aura.git
cd aura
make build
```

### Available Commands
```bash
make help           # Show all available commands
make build          # Build binary
make build-all      # Build for all platforms  
make test           # Run tests
make lint           # Run linter
make clean          # Clean build artifacts
make start-db       # Start database container
make stop-db        # Stop database container
```

---

## 🌟 Why Choose Aura?

| Traditional Navigation | Aura Navigation |
|----------------------|-----------------|
| `cd ~/long/path/to/project` | `aura go project` |
| `ls`, `find`, `grep` | `aura do` → smart suggestions |
| Google for commands | `aura ask "command syntax"` |
| Remember complex paths | `aura bookmark add name path` |

**Result**: Navigate faster, work smarter, learn continuously.

---

## 🚢 Installation Details

### What the installer does:
1. ✅ Downloads/builds the Aura binary
2. ✅ Starts a Docker container for the database  
3. ✅ Sets up shell integration (bash/zsh/PowerShell)
4. ✅ Creates configuration files
5. ✅ Adds Aura to your system PATH

### System Requirements:
- **Docker**: Required for database container
- **Go 1.21+**: Only needed if building from source
- **Bash/Zsh/PowerShell**: For shell integration

---

## 🤝 Contributing

We welcome contributions! Here's how to get started:

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes and add tests
4. Run tests: `make test`
5. Submit a pull request

---

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE-2.0.txt) file for details.

---

## 🎯 Mission

**Make command-line navigation as intuitive as web browsing.**

Aura transforms your terminal experience by combining intelligent bookmarks, AI assistance, and context-aware suggestions into a single, powerful tool that learns and adapts to your workflow.

**Happy navigating!** 🚀
