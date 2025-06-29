package db

import (
	"database/sql"
	"fmt"
	"os/exec"
	"strings"

	_ "modernc.org/sqlite"

	"github.com/timfewi/aura-cli-go/internal/config"
)

// DB represents the database connection.
type DB struct {
	conn          *sql.DB
	isDockerMode  bool
	containerName string
}

// New creates a new database connection and initializes tables.
func New() (*DB, error) {
	db := &DB{
		isDockerMode:  config.IsDockerMode(),
		containerName: "aura-db",
	}

	if db.isDockerMode {
		// Ensure the Docker container is running
		if err := config.EnsureAuraDbRunning(); err != nil {
			return nil, fmt.Errorf("failed to ensure Docker container is running: %w", err)
		}

		// For Docker mode, we don't maintain a persistent connection
		// Instead, we execute commands via docker exec
		db.conn = nil
	} else {
		// Traditional file-based SQLite connection
		conn, err := sql.Open("sqlite", config.DatabasePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open database: %w", err)
		}
		db.conn = conn
	}

	if err := db.initialize(); err != nil {
		if db.conn != nil {
			db.conn.Close()
		}
		return nil, fmt.Errorf("failed to initialize database: %w", err)
	}

	return db, nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// execSQL executes SQL in either Docker or local mode
func (db *DB) execSQL(query string, args ...any) error {
	if db.isDockerMode {
		return db.execDockerSQL(query, args...)
	}

	_, err := db.conn.Exec(query, args...)
	return err
}

// execDockerSQL executes SQL via docker exec
func (db *DB) execDockerSQL(query string, args ...any) error {
	// Build the SQL command with arguments
	sqlCmd := query
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		sqlCmd = strings.ReplaceAll(sqlCmd, placeholder, fmt.Sprintf("'%v'", arg))
	}

	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db", sqlCmd)
	return cmd.Run()
}

// queryDockerSQL executes a query via docker exec and returns parsed results
func (db *DB) queryDockerSQL(query string, args ...any) ([][]string, error) {
	// Build the SQL command with arguments
	sqlCmd := query
	for i, arg := range args {
		placeholder := fmt.Sprintf("$%d", i+1)
		sqlCmd = strings.ReplaceAll(sqlCmd, placeholder, fmt.Sprintf("'%v'", arg))
	}

	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db", sqlCmd)
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var results [][]string

	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.Split(line, "|")
		results = append(results, parts)
	}

	return results, nil
}

// initialize creates the necessary tables.
func (db *DB) initialize() error {
	createBookmarksTable := `
	CREATE TABLE IF NOT EXISTS bookmarks (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		alias TEXT UNIQUE NOT NULL,
		path TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	createHistoryTable := `
	CREATE TABLE IF NOT EXISTS navigation_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		accessed_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);`

	if err := db.execSQL(createBookmarksTable); err != nil {
		return fmt.Errorf("failed to create bookmarks table: %w", err)
	}

	if err := db.execSQL(createHistoryTable); err != nil {
		return fmt.Errorf("failed to create history table: %w", err)
	}

	return nil
}
