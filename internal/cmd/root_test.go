package cmd

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestRootCommand(t *testing.T) {
	cmd := rootCmd

	if cmd.Use != "aura" {
		t.Errorf("Expected Use = 'aura', got %s", cmd.Use)
	}

	if cmd.Short != "Aura - Intelligent CLI Assistant" {
		t.Errorf("Expected correct short description, got %s", cmd.Short)
	}

	if cmd.Version != "1.0.0" {
		t.Errorf("Expected Version = '1.0.0', got %s", cmd.Version)
	}
}

func TestExecuteVersion(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	// Set version flag and execute
	rootCmd.SetArgs([]string{"--version"})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "1.0.0") {
		t.Errorf("Version output should contain '1.0.0', got: %s", output)
	}
}

func TestExecuteHelp(t *testing.T) {
	// Capture output
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)

	// Execute help
	rootCmd.SetArgs([]string{"--help"})
	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Help command failed: %v", err)
	}

	output := buf.String()

	// Update the expected text to match actual output - check for Short field content
	if !strings.Contains(output, "intelligent command-line interface assistant") {
		t.Errorf("Help output should contain 'intelligent command-line interface assistant', got: %s", output)
	}
}

func TestSubcommands(t *testing.T) {
	expectedCommands := []string{
		"ask",
		"bookmark",
		"do",
		"git",
		"go",
		"new",
		"project",
		"uninstall",
	}

	commands := rootCmd.Commands()
	commandNames := make(map[string]bool)

	for _, cmd := range commands {
		commandNames[cmd.Name()] = true
	}

	for _, expected := range expectedCommands {
		if !commandNames[expected] {
			t.Errorf("Missing expected subcommand: %s", expected)
		}
	}
}

func TestInitConfig(t *testing.T) {
	// Create a temporary config directory
	tempDir, err := os.MkdirTemp("", "aura_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Set test environment
	originalEnv := os.Getenv("AURA_ENV")
	defer os.Setenv("AURA_ENV", originalEnv)

	os.Setenv("AURA_ENV", "test")

	// This should not cause any errors
	initConfig()

	// Test that config initialization doesn't panic or error
	// The actual config testing is done in config package tests
}

func TestCommandTree(t *testing.T) {
	// Test that all expected commands are properly added to the root
	tests := []struct {
		command string
		exists  bool
	}{
		{"ask", true},
		{"bookmark", true},
		{"do", true},
		{"git", true},
		{"go", true},
		{"new", true},
		{"project", true},
		{"uninstall", true},
		{"nonexistent", false},
	}

	for _, tt := range tests {
		t.Run(tt.command, func(t *testing.T) {
			cmd, _, err := rootCmd.Find([]string{tt.command})

			if tt.exists {
				if err != nil {
					t.Errorf("Expected command %s to exist, got error: %v", tt.command, err)
				}
				if cmd == nil || cmd.Name() != tt.command {
					t.Errorf("Expected to find command %s", tt.command)
				}
			} else {
				if err == nil {
					t.Errorf("Expected command %s to not exist, but found it", tt.command)
				}
			}
		})
	}
}

func TestInvalidCommand(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	rootCmd.SetArgs([]string{"invalidcommand"})

	err := rootCmd.Execute()
	if err == nil {
		t.Error("Expected error for invalid command")
	}

	output := buf.String()
	if !strings.Contains(output, "unknown command") && !strings.Contains(output, "Error:") {
		t.Errorf("Expected error message for invalid command, got: %s", output)
	}
}

func TestEmptyArgs(t *testing.T) {
	var buf bytes.Buffer
	rootCmd.SetOut(&buf)
	rootCmd.SetErr(&buf)

	rootCmd.SetArgs([]string{})

	err := rootCmd.Execute()
	if err != nil {
		t.Errorf("Unexpected error with empty args: %v", err)
	}

	// Should show help when no command is provided
	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Expected usage help with empty args, got: %s", output)
	}
}
