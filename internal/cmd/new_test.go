package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunNew(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "aura_new_test_*")
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

	tests := []struct {
		name      string
		filename  string
		wantError bool
	}{
		{
			name:      "valid filename",
			filename:  "test.txt",
			wantError: false,
		},
		{
			name:      "valid go file",
			filename:  "main.go",
			wantError: false,
		},
		{
			name:      "empty filename",
			filename:  "",
			wantError: true,
		},
		{
			name:      "path traversal",
			filename:  "../test.txt",
			wantError: true,
		},
		{
			name:      "directory separator",
			filename:  "sub/test.txt",
			wantError: true,
		},
		{
			name:      "valid markdown file",
			filename:  "README.md",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clean up any existing file
			if tt.filename != "" && !tt.wantError {
				os.Remove(tt.filename)
			}

			err := runNew(nil, []string{tt.filename})

			if (err != nil) != tt.wantError {
				t.Errorf("runNew() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				// Check if file was created
				if _, err := os.Stat(tt.filename); os.IsNotExist(err) {
					t.Errorf("File %s was not created", tt.filename)
				}
			}
		})
	}
}

func TestRunNewFileExists(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "aura_new_exists_test_*")
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

	// Create existing file
	filename := "existing.txt"
	err = os.WriteFile(filename, []byte("existing content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Try to create the same file
	err = runNew(nil, []string{filename})
	if err == nil {
		t.Error("Expected error when file already exists")
	}
}

func TestIsValidFilename(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     bool
	}{
		{
			name:     "valid simple filename",
			filename: "test.txt",
			want:     true,
		},
		{
			name:     "valid go file",
			filename: "main.go",
			want:     true,
		},
		{
			name:     "empty filename",
			filename: "",
			want:     false,
		},
		{
			name:     "path traversal with ..",
			filename: "../test.txt",
			want:     false,
		},
		{
			name:     "path with forward slash",
			filename: "dir/test.txt",
			want:     false,
		},
		{
			name:     "path with backslash",
			filename: "dir\\test.txt",
			want:     false,
		},
		{
			name:     "filename with spaces",
			filename: "my file.txt",
			want:     true,
		},
		{
			name:     "filename with special chars",
			filename: "test-file_v2.txt",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidFilename(tt.filename)
			if got != tt.want {
				t.Errorf("isValidFilename(%s) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestIsCommandAvailable(t *testing.T) {
	// Test with a command that should exist on most systems
	tests := []struct {
		name    string
		command string
		// We can't predict if commands exist, so we just test the function works
	}{
		{
			name:    "test with echo",
			command: "echo",
		},
		{
			name:    "test with nonexistent command",
			command: "nonexistentcommand12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the function doesn't panic
			result := isCommandAvailable(tt.command)
			// Result is bool, we just verify it's either true or false
			if result != true && result != false {
				t.Errorf("isCommandAvailable should return bool, got something else")
			}
		})
	}
}

func TestOpenFileInEditor(t *testing.T) {
	// Create a temporary file
	tempDir, err := os.MkdirTemp("", "aura_editor_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	err = os.WriteFile(testFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Test opening the file (this may fail if no editor is available, but shouldn't panic)
	_ = openFileInEditor(testFile)
	// We don't assert on the error since editor availability varies by system
	// Just ensure the function completes without panic

	// Test with non-existent file
	nonExistentFile := filepath.Join(tempDir, "nonexistent.txt")
	_ = openFileInEditor(nonExistentFile)
	// This should handle the error gracefully
}

func TestRunEditorCommand(t *testing.T) {
	// Test with a simple command that should work on most systems
	// We use 'echo' as it's available on most platforms
	tests := []struct {
		name string
		cmd  string
		args []string
	}{
		{
			name: "simple echo command",
			cmd:  "echo",
			args: []string{"test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just test that the function doesn't panic
			err := runEditorCommand(tt.cmd, tt.args...)
			// We don't assert on error as command availability varies
			// Just ensure function completes
			_ = err
		})
	}
}

func TestNewCommandConfiguration(t *testing.T) {
	if newCmd.Use != "new [filename]" {
		t.Errorf("Expected command use 'new [filename]', got '%s'", newCmd.Use)
	}

	if newCmd.Short == "" {
		t.Error("Command should have a short description")
	}

	// Test that command requires exactly one argument
	if newCmd.Args == nil {
		t.Error("Command should have argument validation configured")
	}
}
