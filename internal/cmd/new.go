package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

var newCmd = &cobra.Command{
	Use:   "new [filename]",
	Short: "Create a new file and open it in your default editor",
	Long: `Create a new file in the current directory and open it in VS Code (if available) or the system's default editor.

Examples:
  aura new hello.txt
  aura new hello.cs
  aura new README.md`,
	Args: cobra.ExactArgs(1),
	RunE: runNew,
}

func runNew(cmd *cobra.Command, args []string) error {
	filename := args[0]

	// Validate filename (no path traversal, no empty)
	if !isValidFilename(filename) {
		return fmt.Errorf("invalid filename: %s", filename)
	}

	// Check if file already exists
	if _, err := os.Stat(filename); err == nil {
		return fmt.Errorf("file '%s' already exists", filename)
	}

	// Create the file
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	file.Close()

	fmt.Printf("✓ Created file '%s'\n", filename)

	// Open the file in VS Code or default editor
	if err := openFileInEditor(filename); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: Could not open file in editor: %v\n", err)
	}

	return nil
}

func isValidFilename(name string) bool {
	if name == "" {
		return false
	}
	// Prevent path traversal
	if strings.Contains(name, "..") || strings.ContainsAny(name, `/\`) {
		return false
	}
	return true
}

func openFileInEditor(filename string) error {
	absPath, err := filepath.Abs(filename)
	if err != nil {
		absPath = filename
	}

	// Try VS Code first
	if isCommandAvailable("code") {
		return runEditorCommand("code", absPath)
	}

	// Fallback to system default editor
	switch runtime.GOOS {
	case "windows":
		return runEditorCommand("cmd", "/c", "start", "", absPath)
	case "darwin":
		return runEditorCommand("open", absPath)
	default: // Linux and others
		return runEditorCommand("xdg-open", absPath)
	}
}

func isCommandAvailable(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func runEditorCommand(name string, args ...string) error {
	cmd := exec.Command(name, args...)

	// For Windows "start" command, we should use Start() to avoid blocking
	// For other commands, we should use Run() to wait for completion
	if runtime.GOOS == "windows" && name == "cmd" && len(args) > 0 && args[0] == "/c" && len(args) > 1 && args[1] == "start" {
		return cmd.Start()
	}

	// For VS Code and other editors, start in background
	return cmd.Start()
}

func init() {
	rootCmd.AddCommand(newCmd)
}
