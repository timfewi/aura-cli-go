package cmd

import (
	"os"
	"strings"
	"testing"

	"github.com/timfewi/aura-cli-go/internal/context"
)

func TestRunDo(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "aura_do_test_*")
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

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test with no context (empty directory)
	err = runDo(doCmd, []string{})
	if err != nil {
		t.Errorf("runDo() error = %v", err)
	}

	// Test with Git context
	err = os.Mkdir(".git", 0755)
	if err != nil {
		t.Fatalf("Failed to create .git directory: %v", err)
	}

	// Note: Since runDo uses interactive prompts, we can't fully test it
	// without mocking the promptui.Select, but we can test that it doesn't panic
	// and that the detector logic works correctly
}

func TestExecuteCommand(t *testing.T) {
	tests := []struct {
		name      string
		command   string
		wantError bool
	}{
		{
			name:      "empty command",
			command:   "",
			wantError: true,
		},
		{
			name:      "valid command - cross-platform list",
			command:   getWorkingListCommand(), // Use a working command
			wantError: false,
		},
		{
			name:      "invalid command",
			command:   "nonexistentcommand12345",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := executeCommand(tt.command)
			if (err != nil) != tt.wantError {
				t.Errorf("executeCommand() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// getWorkingListCommand returns a command that will work on the current platform
func getWorkingListCommand() string {
	if isWindows() {
		return "cmd /c dir" // Use cmd /c to execute built-in commands
	}
	return "ls"
}

func TestGetOpenCommand(t *testing.T) {
	cmd := getOpenCommand()
	if cmd == "" {
		t.Error("getOpenCommand() returned empty string")
	}

	// Should return appropriate command based on OS
	validCommands := []string{"explorer .", "open .", "xdg-open ."}
	found := false
	for _, validCmd := range validCommands {
		if cmd == validCmd {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("getOpenCommand() returned unexpected command: %s", cmd)
	}
}

func TestGetListCommand(t *testing.T) {
	cmd := getListCommand()
	if cmd == "" {
		t.Error("getListCommand() returned empty string")
	}

	// Should return either "dir" or "ls -la"
	if cmd != "dir" && cmd != "ls -la" {
		t.Errorf("getListCommand() returned unexpected command: %s", cmd)
	}
}

func TestGetDiskUsageCommand(t *testing.T) {
	cmd := getDiskUsageCommand()
	if cmd == "" {
		t.Error("getDiskUsageCommand() returned empty string")
	}

	// Should return appropriate command based on OS
	if isWindows() {
		if !strings.Contains(cmd, "powershell") {
			t.Errorf("getDiskUsageCommand() should contain 'powershell' on Windows, got: %s", cmd)
		}
	} else {
		if cmd != "du -sh *" {
			t.Errorf("getDiskUsageCommand() should return 'du -sh *' on non-Windows, got: %s", cmd)
		}
	}
}

func TestGetFindLargeFilesCommand(t *testing.T) {
	cmd := getFindLargeFilesCommand()
	if cmd == "" {
		t.Error("getFindLargeFilesCommand() returned empty string")
	}

	// Should return appropriate command based on OS
	if isWindows() {
		if !strings.Contains(cmd, "powershell") {
			t.Errorf("getFindLargeFilesCommand() should contain 'powershell' on Windows, got: %s", cmd)
		}
	} else {
		if !strings.Contains(cmd, "find") {
			t.Errorf("getFindLargeFilesCommand() should contain 'find' on non-Windows, got: %s", cmd)
		}
	}
}

func TestIsWindows(t *testing.T) {
	// Save original OS env
	originalOS := os.Getenv("OS")
	defer os.Setenv("OS", originalOS)

	// Test Windows detection
	os.Setenv("OS", "Windows_NT")
	if !isWindows() {
		t.Error("isWindows() should return true when OS=Windows_NT")
	}

	// Test non-Windows
	os.Setenv("OS", "Linux")
	if isWindows() {
		t.Error("isWindows() should return false when OS=Linux")
	}
}

func TestIsMacOS(t *testing.T) {
	// Save original OSTYPE env
	originalOSType := os.Getenv("OSTYPE")
	defer os.Setenv("OSTYPE", originalOSType)

	// Test macOS detection
	os.Setenv("OSTYPE", "darwin21")
	if !isMacOS() {
		t.Error("isMacOS() should return true when OSTYPE contains darwin")
	}

	// Test non-macOS
	os.Setenv("OSTYPE", "linux-gnu")
	if isMacOS() {
		t.Error("isMacOS() should return false when OSTYPE=linux-gnu")
	}
}

func TestDoCommandConfiguration(t *testing.T) {
	// Test that the command is properly configured
	if doCmd.Use != "do" {
		t.Errorf("doCmd.Use = %v, want 'do'", doCmd.Use)
	}

	if doCmd.Short == "" {
		t.Error("doCmd.Short should not be empty")
	}

	if doCmd.Long == "" {
		t.Error("doCmd.Long should not be empty")
	}

	if doCmd.RunE == nil {
		t.Error("doCmd.RunE should not be nil")
	}
}

// TestDoWithContexts tests the do command with various project contexts
func TestDoWithContexts(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "aura_do_context_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Save current directory
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer func () {
		if err := os.Chdir(originalDir); err != nil {
			t.Errorf("Failed to restore original directory: %v", err)
		}
	}()

	// Change to temp directory
	err = os.Chdir(tempDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// Test the detector functions directly since runDo requires interactive input
	detectors := []func() []context.Action{
		context.DetectGitContext,
		context.DetectNodeContext,
		context.DetectPythonContext,
		context.DetectGoContext,
		context.DetectDockerContext,
		context.DetectMakeContext,
	}

	// Initially should return no actions
	for i, detector := range detectors {
		actions := detector()
		if len(actions) != 0 {
			t.Errorf("Detector %d should return no actions in empty directory, got %d", i, len(actions))
		}
	}

	// Create various project files and test detection
	testCases := []struct {
		name     string
		files    []string
		detector func() []context.Action
		expected bool
	}{
		{
			name:     "Git context",
			files:    []string{".git/"},
			detector: context.DetectGitContext,
			expected: true,
		},
		{
			name:     "Node.js context",
			files:    []string{"package.json"},
			detector: context.DetectNodeContext,
			expected: true,
		},
		{
			name:     "Python context",
			files:    []string{"requirements.txt"},
			detector: context.DetectPythonContext,
			expected: true,
		},
		{
			name:     "Go context",
			files:    []string{"go.mod"},
			detector: context.DetectGoContext,
			expected: true,
		},
		{
			name:     "Docker context",
			files:    []string{"Dockerfile"},
			detector: context.DetectDockerContext,
			expected: true,
		},
		{
			name:     "Make context",
			files:    []string{"Makefile"},
			detector: context.DetectMakeContext,
			expected: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Clean up from previous test
			for _, file := range tc.files {
				if strings.HasSuffix(file, "/") {
					os.RemoveAll(file)
				} else {
					os.Remove(file)
				}
			}

			// Create test files
			for _, file := range tc.files {
				if strings.HasSuffix(file, "/") {
					err = os.Mkdir(file, 0755)
				} else {
					err = os.WriteFile(file, []byte("test content"), 0644)
				}
				if err != nil {
					t.Fatalf("Failed to create test file %s: %v", file, err)
				}
			}

			// Test detector
			actions := tc.detector()
			hasActions := len(actions) > 0

			if hasActions != tc.expected {
				t.Errorf("Expected %v actions for %s, got %d actions", tc.expected, tc.name, len(actions))
			}

			// Clean up
			for _, file := range tc.files {
				if strings.HasSuffix(file, "/") {
					os.RemoveAll(file)
				} else {
					os.Remove(file)
				}
			}
		})
	}
}

func TestGeneralActions(t *testing.T) {
	// Test that general actions are always available
	generalActions := []context.Action{
		{Name: "Open current directory", Command: getOpenCommand()},
		{Name: "List directory contents", Command: getListCommand()},
		{Name: "Show disk usage", Command: getDiskUsageCommand()},
		{Name: "Find large files", Command: getFindLargeFilesCommand()},
	}

	if len(generalActions) != 4 {
		t.Errorf("Expected 4 general actions, got %d", len(generalActions))
	}

	for i, action := range generalActions {
		if action.Name == "" {
			t.Errorf("General action %d has empty name", i)
		}
		if action.Command == "" {
			t.Errorf("General action %d has empty command", i)
		}
	}
}
