package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestInitialize(t *testing.T) {
	// Save original environment
	originalEnv := os.Getenv("AURA_ENV")
	originalDir := ConfigDir
	originalPath := DatabasePath
	originalType := DatabaseType

	defer func() {
		os.Setenv("AURA_ENV", originalEnv)
		ConfigDir = originalDir
		DatabasePath = originalPath
		DatabaseType = originalType
	}()

	tests := []struct {
		name       string
		envVar     string
		expectEnv  string
		expectType string
	}{
		{
			name:       "default environment",
			envVar:     "",
			expectEnv:  "production",
			expectType: "file",
		},
		{
			name:       "development environment",
			envVar:     "development",
			expectEnv:  "development",
			expectType: "file",
		},
		{
			name:       "test environment",
			envVar:     "test",
			expectEnv:  "test",
			expectType: "file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment
			if tt.envVar == "" {
				os.Unsetenv("AURA_ENV")
			} else {
				os.Setenv("AURA_ENV", tt.envVar)
			}

			err := Initialize()
			if err != nil {
				t.Errorf("Initialize() error = %v", err)
			}

			if Environment != tt.expectEnv {
				t.Errorf("Environment = %v, want %v", Environment, tt.expectEnv)
			}

			if DatabaseType != tt.expectType {
				t.Errorf("DatabaseType = %v, want %v", DatabaseType, tt.expectType)
			}

			// Check that config directory is set
			if ConfigDir == "" {
				t.Error("ConfigDir is empty")
			}

			// Check that database path is set
			if DatabasePath == "" {
				t.Error("DatabasePath is empty")
			}
		})
	}
}

func TestIsDevelopment(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want bool
	}{
		{
			name: "development environment",
			env:  "development",
			want: true,
		},
		{
			name: "production environment",
			env:  "production",
			want: false,
		},
		{
			name: "test environment",
			env:  "test",
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Environment = tt.env
			got := IsDevelopment()
			if got != tt.want {
				t.Errorf("IsDevelopment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsDockerMode(t *testing.T) {
	tests := []struct {
		name   string
		dbType string
		want   bool
	}{
		{
			name:   "docker mode",
			dbType: "docker",
			want:   true,
		},
		{
			name:   "file mode",
			dbType: "file",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DatabaseType = tt.dbType
			got := IsDockerMode()
			if got != tt.want {
				t.Errorf("IsDockerMode() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want string
	}{
		{
			name: "development environment",
			env:  "development",
			want: "debug",
		},
		{
			name: "production environment",
			env:  "production",
			want: "info",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Environment = tt.env
			got := GetLogLevel()
			if got != tt.want {
				t.Errorf("GetLogLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestGetLogFile(t *testing.T) {
	// Set up test config directory
	tempDir, err := os.MkdirTemp("", "aura_config_test_*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	ConfigDir = tempDir

	logFile := GetLogFile()
	expectedPath := filepath.Join(tempDir, "aura.log")

	if logFile != expectedPath {
		t.Errorf("GetLogFile() = %v, want %v", logFile, expectedPath)
	}
}

func TestGetDatabaseConnection(t *testing.T) {
	tests := []struct {
		name     string
		dbType   string
		dbPath   string
		expected string
	}{
		{
			name:     "file mode",
			dbType:   "file",
			dbPath:   "/test/path/db.sqlite",
			expected: "/test/path/db.sqlite",
		},
		{
			name:     "docker mode",
			dbType:   "docker",
			dbPath:   "/data/aura.db",
			expected: "docker:aura-db:/data/aura.db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DatabaseType = tt.dbType
			DatabasePath = tt.dbPath

			got := GetDatabaseConnection()
			if got != tt.expected {
				t.Errorf("GetDatabaseConnection() = %v, want %v", got, tt.expected)
			}
		})
	}
}
