# Contributing to Aura CLI

Thank you for your interest in contributing to Aura CLI! 🎉

## Table of Contents
- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Submitting Changes](#submitting-changes)
- [Code Style](#code-style)
- [Testing](#testing)

## Getting Started

### Prerequisites
- Go 1.21 or later
- Docker (for development database)
- Git

### Development Setup

1. **Fork and clone the repository**
   ```bash
   git clone https://github.com/your-username/aura-cli-go.git
   cd aura-cli-go
   ```

2. **Set up development environment**
   ```bash
   # Windows
   .\scripts\setup-dev.bat
   
   # Unix/Linux/macOS
   ./scripts/setup-dev.sh
   ```

3. **Verify setup**
   ```bash
   make test
   make build
   ./bin/aura --help
   ```

## Making Changes

### Branch Naming
- `feature/your-feature-name` - New features
- `fix/issue-description` - Bug fixes
- `docs/what-you-changed` - Documentation updates
- `chore/maintenance-task` - Maintenance tasks

### Commit Messages
Follow [Conventional Commits](https://conventionalcommits.org/):

```
<type>(<scope>): <description>

[optional body]

[optional footer]
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code formatting
- `refactor`: Code restructuring
- `test`: Adding tests
- `chore`: Maintenance

**Examples:**
```bash
feat(navigation): add fuzzy search for bookmarks
fix(database): handle connection timeout gracefully
docs(readme): update installation instructions
```

## Submitting Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Write clear, focused commits
   - Add tests for new functionality
   - Update documentation if needed

3. **Test your changes**
   ```bash
   make test
   make lint
   make build
   ```

4. **Push and create a pull request**
   ```bash
   git push origin feature/your-feature-name
   ```

### Pull Request Guidelines
- **Clear title and description**
- **Link related issues** using "Fixes #123"
- **Include screenshots** for UI changes
- **Test instructions** for reviewers
- **Breaking changes** noted in description

## Code Style

### Go Code Standards
We follow the [Go Code Review Guidelines](https://github.com/golang/go/wiki/CodeReviewComments) and our internal [Go conventions](.github/rules/go.instructions.md).

**Key points:**
- Use `gofmt` and `goimports`
- Write clear, descriptive variable names
- Add comments for exported functions
- Handle errors properly
- Use meaningful test names

### Project Structure
```
aura-cli-go/
├── cmd/                # Main applications
├── internal/           # Private application code
├── assets/            # Embedded resources
├── scripts/           # Development scripts
├── docs/              # Documentation
└── tests/             # Test files
```

## Testing

### Running