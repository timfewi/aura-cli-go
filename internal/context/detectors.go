package context

import (
	"os"
	"path/filepath"
)

// Action represents a suggested action with a display name and command.
type Action struct {
	Name    string
	Command string
}

// DetectGitContext checks for Git repository and returns relevant actions.
func DetectGitContext() []Action {
	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		return nil
	}

	return []Action{
		{Name: "View status", Command: "git status"},
		{Name: "View changes", Command: "git diff"},
		{Name: "View staged changes", Command: "git diff --staged"},
		{Name: "Add all changes", Command: "git add ."},
		{Name: "Commit changes", Command: "git commit"},
		{Name: "Push changes", Command: "git push"},
		{Name: "Pull latest changes", Command: "git pull"},
		{Name: "View commit history", Command: "git log --oneline -10"},
		{Name: "Create new branch", Command: "git checkout -b"},
		{Name: "Stash changes", Command: "git stash"},
		{Name: "Pop stash", Command: "git stash pop"},
	}
}

// DetectNodeContext checks for Node.js project and returns relevant actions.
func DetectNodeContext() []Action {
	if _, err := os.Stat("package.json"); os.IsNotExist(err) {
		return nil
	}

	actions := []Action{
		{Name: "Install dependencies", Command: "npm install"},
		{Name: "Update dependencies", Command: "npm update"},
		{Name: "Run dev server", Command: "npm run dev"},
		{Name: "Run build", Command: "npm run build"},
		{Name: "Run tests", Command: "npm test"},
		{Name: "Check for vulnerabilities", Command: "npm audit"},
		{Name: "View package info", Command: "npm list --depth=0"},
	}

	// Check for common scripts
	if _, err := os.Stat("yarn.lock"); err == nil {
		// Yarn project
		yarnActions := []Action{
			{Name: "Install dependencies (Yarn)", Command: "yarn install"},
			{Name: "Run dev server (Yarn)", Command: "yarn dev"},
			{Name: "Run build (Yarn)", Command: "yarn build"},
			{Name: "Run tests (Yarn)", Command: "yarn test"},
		}
		actions = append(yarnActions, actions...)
	}

	return actions
}

// DetectPythonContext checks for Python project and returns relevant actions.
func DetectPythonContext() []Action {
	hasPyProject := false
	hasRequirements := false
	hasPipfile := false

	if _, err := os.Stat("pyproject.toml"); err == nil {
		hasPyProject = true
	}
	if _, err := os.Stat("requirements.txt"); err == nil {
		hasRequirements = true
	}
	if _, err := os.Stat("Pipfile"); err == nil {
		hasPipfile = true
	}

	if !hasPyProject && !hasRequirements && !hasPipfile {
		// Check for .py files in current directory
		files, err := filepath.Glob("*.py")
		if err != nil || len(files) == 0 {
			return nil
		}
	}

	actions := []Action{
		{Name: "Run Python REPL", Command: "python"},
		{Name: "List Python files", Command: "find . -name '*.py' -type f"},
		{Name: "Check Python version", Command: "python --version"},
	}

	if hasRequirements {
		actions = append(actions,
			Action{Name: "Install requirements", Command: "pip install -r requirements.txt"},
			Action{Name: "Generate requirements", Command: "pip freeze > requirements.txt"},
		)
	}

	if hasPipfile {
		actions = append(actions,
			Action{Name: "Install dependencies (Pipenv)", Command: "pipenv install"},
			Action{Name: "Activate virtual env", Command: "pipenv shell"},
			Action{Name: "Run with Pipenv", Command: "pipenv run python"},
		)
	}

	if hasPyProject {
		actions = append(actions,
			Action{Name: "Install project", Command: "pip install -e ."},
		)
	}

	// Check for common Python tools
	if _, err := os.Stat("setup.py"); err == nil {
		actions = append(actions,
			Action{Name: "Install package", Command: "python setup.py install"},
		)
	}

	if _, err := os.Stat("pytest.ini"); err == nil || hasFilePattern("test_*.py") || hasFilePattern("*_test.py") {
		actions = append(actions,
			Action{Name: "Run tests", Command: "pytest"},
			Action{Name: "Run tests with coverage", Command: "pytest --cov"},
		)
	}

	return actions
}

// DetectGoContext checks for Go project and returns relevant actions.
func DetectGoContext() []Action {
	if _, err := os.Stat("go.mod"); os.IsNotExist(err) {
		// Check for .go files
		files, err := filepath.Glob("*.go")
		if err != nil || len(files) == 0 {
			return nil
		}
	}

	return []Action{
		{Name: "Build project", Command: "go build"},
		{Name: "Run project", Command: "go run ."},
		{Name: "Test project", Command: "go test ./..."},
		{Name: "Get dependencies", Command: "go mod tidy"},
		{Name: "Format code", Command: "go fmt ./..."},
		{Name: "Lint code", Command: "golangci-lint run"},
		{Name: "View dependencies", Command: "go list -m all"},
		{Name: "Check for updates", Command: "go list -u -m all"},
		{Name: "Clean module cache", Command: "go clean -modcache"},
	}
}

// DetectDockerContext checks for Docker project and returns relevant actions.
func DetectDockerContext() []Action {
	hasDockerfile := false
	hasDockerCompose := false

	if _, err := os.Stat("Dockerfile"); err == nil {
		hasDockerfile = true
	}
	if _, err := os.Stat("docker-compose.yml"); err == nil {
		hasDockerCompose = true
	}
	if !hasDockerCompose {
		if _, err := os.Stat("docker-compose.yaml"); err == nil {
			hasDockerCompose = true
		}
	}

	if !hasDockerfile && !hasDockerCompose {
		return nil
	}

	var actions []Action

	if hasDockerfile {
		actions = append(actions,
			Action{Name: "Build Docker image", Command: "docker build -t $(basename $(pwd)) ."},
			Action{Name: "Run Docker container", Command: "docker run -it $(basename $(pwd))"},
		)
	}

	if hasDockerCompose {
		actions = append(actions,
			Action{Name: "Start services", Command: "docker-compose up"},
			Action{Name: "Start services (detached)", Command: "docker-compose up -d"},
			Action{Name: "Stop services", Command: "docker-compose down"},
			Action{Name: "View logs", Command: "docker-compose logs"},
			Action{Name: "Rebuild and start", Command: "docker-compose up --build"},
		)
	}

	return actions
}

// DetectMakeContext checks for Makefile and returns relevant actions.
func DetectMakeContext() []Action {
	if _, err := os.Stat("Makefile"); os.IsNotExist(err) {
		if _, err := os.Stat("makefile"); os.IsNotExist(err) {
			return nil
		}
	}

	return []Action{
		{Name: "Show available targets", Command: "make help"},
		{Name: "Build (default target)", Command: "make"},
		{Name: "Clean build artifacts", Command: "make clean"},
		{Name: "Install", Command: "make install"},
		{Name: "Run tests", Command: "make test"},
	}
}

// hasFilePattern checks if any files match the given pattern.
func hasFilePattern(pattern string) bool {
	files, err := filepath.Glob(pattern)
	return err == nil && len(files) > 0
}
