package context

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectGitContext(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "aura_git_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	// Test without .git directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	actions := DetectGitContext()
	if len(actions) != 0 {
		t.Errorf("Expected no Git actions without .git directory, got %d", len(actions))
	}

	// Create .git directory
	gitDir := filepath.Join(tempDir, ".git")
	err = os.Mkdir(gitDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Test with .git directory
	actions = DetectGitContext()
	if len(actions) == 0 {
		t.Error("Expected Git actions with .git directory, got none")
	}

	// Check for expected actions
	expectedActions := []string{
		"View status",
		"View changes",
		"Add all changes",
		"Commit changes",
		"Push changes",
		"Pull latest changes",
	}

	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name] = true
	}

	for _, expected := range expectedActions {
		if !actionNames[expected] {
			t.Errorf("Missing expected Git action: %s", expected)
		}
	}
}

func TestDetectNodeContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_node_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test without package.json
	actions := DetectNodeContext()
	if len(actions) != 0 {
		t.Errorf("Expected no Node actions without package.json, got %d", len(actions))
	}

	// Create package.json
	packageJSON := `{
		"name": "test-project",
		"scripts": {
			"start": "node index.js",
			"test": "jest",
			"build": "webpack"
		}
	}`

	err = os.WriteFile("package.json", []byte(packageJSON), 0644)
	if err != nil {
		t.Fatalf("Failed to create package.json: %v", err)
	}

	// Test with package.json
	actions = DetectNodeContext()
	if len(actions) == 0 {
		t.Error("Expected Node actions with package.json, got none")
	}

	// Check for expected actions - update to match actual implementation
	expectedActions := []string{
		"Install dependencies",
		"Run dev server", // Not "Run start script"
		"Run tests",      // Not "Run test script"
		"Run build",      // Not "Run build script"
	}

	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name] = true
	}

	for _, expected := range expectedActions {
		if !actionNames[expected] {
			t.Errorf("Missing expected Node action: %s", expected)
		}
	}
}

func TestDetectPythonContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_python_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test without Python files
	actions := DetectPythonContext()
	if len(actions) != 0 {
		t.Errorf("Expected no Python actions without Python files, got %d", len(actions))
	}

	// Create requirements.txt
	err = os.WriteFile("requirements.txt", []byte("flask==2.0.1\nrequests==2.25.1"), 0644)
	if err != nil {
		t.Fatalf("Failed to create requirements.txt: %v", err)
	}

	// Test with requirements.txt
	actions = DetectPythonContext()
	if len(actions) == 0 {
		t.Error("Expected Python actions with requirements.txt, got none")
	}

	// Check for install requirements action
	found := false
	for _, action := range actions {
		if action.Name == "Install requirements" {
			found = true
			break
		}
	}
	if !found {
		t.Error("Missing 'Install requirements' action")
	}
}

func TestDetectGoContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_go_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test without go.mod
	actions := DetectGoContext()
	if len(actions) != 0 {
		t.Errorf("Expected no Go actions without go.mod, got %d", len(actions))
	}

	// Create go.mod
	goMod := `module test-project

go 1.21

require (
	github.com/spf13/cobra v1.8.0
)`

	err = os.WriteFile("go.mod", []byte(goMod), 0644)
	if err != nil {
		t.Fatalf("Failed to create go.mod: %v", err)
	}

	// Test with go.mod
	actions = DetectGoContext()
	if len(actions) == 0 {
		t.Error("Expected Go actions with go.mod, got none")
	}

	// Check for expected actions - update to match actual implementation
	expectedActions := []string{
		"Build project",
		"Test project",     // Not "Run tests"
		"Get dependencies", // Not "Download dependencies"
	}

	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name] = true
	}

	for _, expected := range expectedActions {
		if !actionNames[expected] {
			t.Errorf("Missing expected Go action: %s", expected)
		}
	}
}

func TestDetectDockerContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_docker_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test without Dockerfile
	actions := DetectDockerContext()
	if len(actions) != 0 {
		t.Errorf("Expected no Docker actions without Dockerfile, got %d", len(actions))
	}

	// Create Dockerfile
	dockerfile := `FROM alpine:latest
RUN apk add --no-cache nodejs npm
COPY . /app
WORKDIR /app
EXPOSE 3000
CMD ["node", "index.js"]`

	err = os.WriteFile("Dockerfile", []byte(dockerfile), 0644)
	if err != nil {
		t.Fatalf("Failed to create Dockerfile: %v", err)
	}

	// Test with Dockerfile
	actions = DetectDockerContext()
	if len(actions) == 0 {
		t.Error("Expected Docker actions with Dockerfile, got none")
	}

	// Check for expected actions - update to match actual implementation
	expectedActions := []string{
		"Build Docker image",
		"Run Docker container", // Not "Run container"
	}

	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name] = true
	}

	for _, expected := range expectedActions {
		if !actionNames[expected] {
			t.Errorf("Missing expected Docker action: %s", expected)
		}
	}
}

func TestDetectMakeContext(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_make_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test without Makefile
	actions := DetectMakeContext()
	if len(actions) != 0 {
		t.Errorf("Expected no Make actions without Makefile, got %d", len(actions))
	}

	// Create Makefile
	makefile := `build:
	go build -o bin/app ./cmd/main.go

test:
	go test ./...

clean:
	rm -rf bin/

.PHONY: build test clean`

	err = os.WriteFile("Makefile", []byte(makefile), 0644)
	if err != nil {
		t.Fatalf("Failed to create Makefile: %v", err)
	}

	// Test with Makefile
	actions = DetectMakeContext()
	if len(actions) == 0 {
		t.Error("Expected Make actions with Makefile, got none")
	}

	// Check for expected actions - update to match actual implementation
	expectedActions := []string{
		"Show available targets", // Not "View available targets"
		"Build (default target)", // Not "Run default target"
	}

	actionNames := make(map[string]bool)
	for _, action := range actions {
		actionNames[action.Name] = true
	}

	for _, expected := range expectedActions {
		if !actionNames[expected] {
			t.Errorf("Missing expected Make action: %s", expected)
		}
	}
}

func TestHasFilePattern(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_pattern_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test with no matching files
	if hasFilePattern("*.go") {
		t.Error("Expected no Go files, but hasFilePattern returned true")
	}

	// Create test files
	testFiles := []string{
		"main.go",
		"handler.go",
		"package.json",
		"README.md",
	}

	for _, file := range testFiles {
		err = os.WriteFile(file, []byte("test content"), 0644)
		if err != nil {
			t.Fatalf("Failed to create test file %s: %v", file, err)
		}
	}

	// Test patterns
	tests := []struct {
		pattern string
		want    bool
	}{
		{"*.go", true},
		{"*.js", false},
		{"*.json", true},
		{"*.md", true},
		{"*.txt", false},
		{"main.*", true},
		{"nonexistent.*", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			got := hasFilePattern(tt.pattern)
			if got != tt.want {
				t.Errorf("hasFilePattern(%s) = %v, want %v", tt.pattern, got, tt.want)
			}
		})
	}
}
