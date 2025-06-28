package db

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/timfewi/aura-cli-go/internal/config"
)

func TestMain(m *testing.M) {
	// Setup test environment
	config.Environment = "test"
	config.DatabaseType = "file"

	// Create temp directory for test database
	tempDir, err := os.MkdirTemp("", "aura_test_*")
	if err != nil {
		panic(err)
	}

	config.ConfigDir = tempDir
	config.DatabasePath = filepath.Join(tempDir, "test_aura.db")

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(tempDir)
	os.Exit(code)
}

func TestNew(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	if db == nil {
		t.Fatal("Database instance is nil")
	}

	if db.isDockerMode {
		t.Error("Expected file mode in tests, got docker mode")
	}
}

func TestNewDockerMode(t *testing.T) {
	// Skip this test if Docker is not available or running
	if !isDockerAvailable() {
		t.Skip("Docker not available or not running, skipping Docker mode test")
	}

	// Save original config
	originalType := config.DatabaseType
	defer func() {
		config.DatabaseType = originalType
	}()

	// Set docker mode
	config.DatabaseType = "docker"

	db, err := New()
	if err != nil {
		t.Skipf("Failed to create database in docker mode (Docker not ready): %v", err)
	}
	defer db.Close()

	// Test that we can perform basic operations
	err = db.AddBookmark("docker-test", "/test/path")
	if err != nil {
		t.Errorf("Failed to add bookmark in docker mode: %v", err)
	}
}

func TestClose(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database: %v", err)
	}

	// Closing again should not cause error
	err = db.Close()
	if err != nil {
		t.Errorf("Failed to close database twice: %v", err)
	}
}

func TestInitialize(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Test that tables exist by trying to query them
	_, err = db.conn.Query("SELECT COUNT(*) FROM bookmarks")
	if err != nil {
		t.Errorf("Bookmarks table not initialized: %v", err)
	}

	_, err = db.conn.Query("SELECT COUNT(*) FROM navigation_history")
	if err != nil {
		t.Errorf("Navigation history table not initialized: %v", err)
	}
}

// Helper function to check if Docker is available and running
func isDockerAvailable() bool {
	// Check if docker command exists
	if _, err := exec.LookPath("docker"); err != nil {
		return false
	}

	// Check if docker daemon is running
	cmd := exec.Command("docker", "ps")
	return cmd.Run() == nil
}
