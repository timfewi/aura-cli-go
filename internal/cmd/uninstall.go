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

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Uninstall Aura CLI and remove all related files",
	Long: `Remove the Aura CLI binary and all configuration/data files from your system.

This will delete:
- The Aura binary from your PATH (if found)
- The Aura config and data directory (~/.config/aura or %APPDATA%\aura)
- The Aura database (if present)
`,
	RunE: runUninstall,
}

func runUninstall(cmd *cobra.Command, args []string) error {
	binaryName := "aura"
	var binaryPaths []string

	// Try to find the binary in PATH
	if path, err := exec.LookPath(binaryName); err == nil {
		binaryPaths = append(binaryPaths, path)
	}

	// Also check ./bin/aura(.exe)
	binDir := filepath.Join(".", "bin")
	files, _ := os.ReadDir(binDir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), binaryName) {
			binaryPaths = append(binaryPaths, filepath.Join(binDir, f.Name()))
		}
	}

	// Remove binaries
	for _, bin := range binaryPaths {
		if err := os.Remove(bin); err == nil {
			fmt.Printf("Removed binary: %s\n", bin)
		}
	}

	// Remove config/data directory
	var configDir string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData != "" {
			configDir = filepath.Join(appData, "aura")
		}
	} else {
		home, _ := os.UserHomeDir()
		configDir = filepath.Join(home, ".config", "aura")
	}
	if configDir != "" {
		if err := os.RemoveAll(configDir); err == nil {
			fmt.Printf("Removed config/data directory: %s\n", configDir)
		}
	}

	// Remove database if present
	dbPath := filepath.Join("data", "sqlite", "aura.db")
	if err := os.Remove(dbPath); err == nil {
		fmt.Printf("Removed database: %s\n", dbPath)
	}

	fmt.Println("✓ Aura CLI has been uninstalled.")
	return nil
}

func init() {
	rootCmd.AddCommand(uninstallCmd)
}
