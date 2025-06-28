package config

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

var (
	// ConfigDir is the Aura configuration directory.
	ConfigDir string
	// DatabasePath is the path to the SQLite database.
	DatabasePath string
	// Environment indicates the current environment (development, production).
	Environment string
	// DatabaseType indicates if using local file or Docker container
	DatabaseType string
)

// Initialize sets up the configuration directories and paths.
func Initialize() error {
	// Set environment
	Environment = os.Getenv("AURA_ENV")
	if Environment == "" {
		Environment = "production"
	}

	// Check for environment-specific database path
	if dbPath := os.Getenv("AURA_DB_PATH"); dbPath != "" {
		DatabasePath = dbPath
		ConfigDir = filepath.Dir(dbPath)
		DatabaseType = "file"
	} else {
		// Check if running in Docker container mode (production)
		if isDockerAvailable() && isAuraDbRunning() {
			DatabaseType = "docker"
			DatabasePath = "/data/aura.db" // Path inside container

			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				return err
			}
			ConfigDir = filepath.Join(userConfigDir, "aura")
		} else {
			// Fallback to local file
			DatabaseType = "file"
			userConfigDir, err := os.UserConfigDir()
			if err != nil {
				return err
			}
			ConfigDir = filepath.Join(userConfigDir, "aura")
			DatabasePath = filepath.Join(ConfigDir, "aura.db")
		}
	}

	// Create the config directory if it doesn't exist (for file mode)
	if DatabaseType == "file" {
		if err := os.MkdirAll(ConfigDir, 0755); err != nil {
			return err
		}
	}

	return nil
}

// GetDatabaseConnection returns the appropriate database connection string
func GetDatabaseConnection() string {
	if DatabaseType == "docker" {
		// Use Docker exec to access the database
		return "docker:aura-db:" + DatabasePath
	}
	return DatabasePath
}

// isDockerAvailable checks if Docker is available
func isDockerAvailable() bool {
	_, err := exec.LookPath("docker")
	return err == nil
}

// isAuraDbRunning checks if the aura-db container is running
func isAuraDbRunning() bool {
	cmd := exec.Command("docker", "ps", "-q", "-f", "name=aura-db")
	output, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(output)) != ""
}

// IsDevelopment returns true if running in development mode.
func IsDevelopment() bool {
	return Environment == "development"
}

// IsDockerMode returns true if using Docker database
func IsDockerMode() bool {
	return DatabaseType == "docker"
}

// GetLogLevel returns the configured log level.
func GetLogLevel() string {
	if level := os.Getenv("AURA_LOG_LEVEL"); level != "" {
		return level
	}
	if IsDevelopment() {
		return "debug"
	}
	return "info"
}

// GetLogFile returns the configured log file path.
func GetLogFile() string {
	if file := os.Getenv("AURA_LOG_FILE"); file != "" {
		return file
	}
	return filepath.Join(ConfigDir, "aura.log")
}

// EnsureAuraDbRunning starts the Aura database container if not running
func EnsureAuraDbRunning() error {
	if !isDockerAvailable() {
		return fmt.Errorf("Docker is not available")
	}

	if isAuraDbRunning() {
		return nil // Already running
	}

	// Start the container
	cmd := exec.Command("docker", "run", "-d",
		"--name", "aura-db",
		"--restart", "unless-stopped",
		"-v", "aura-data:/data",
		"alpine:latest",
		"sh", "-c",
		`apk add --no-cache sqlite && 
		 if [ ! -f /data/aura.db ]; then 
		   sqlite3 /data/aura.db 'CREATE TABLE IF NOT EXISTS bookmarks (id INTEGER PRIMARY KEY AUTOINCREMENT, alias TEXT UNIQUE NOT NULL, path TEXT NOT NULL, created_at DATETIME DEFAULT CURRENT_TIMESTAMP); CREATE TABLE IF NOT EXISTS navigation_history (id INTEGER PRIMARY KEY AUTOINCREMENT, path TEXT NOT NULL, accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP);'
		 fi && 
		 while true; do sleep 30; done`)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to start aura-db container: %w", err)
	}

	return nil
}
