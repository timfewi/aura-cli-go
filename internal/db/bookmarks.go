package db

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

// Bookmark represents a directory bookmark.
type Bookmark struct {
	ID        int       `json:"id"`
	Alias     string    `json:"alias"`
	Path      string    `json:"path"`
	CreatedAt time.Time `json:"created_at"`
}

// AddBookmark adds a new bookmark to the database.
func (db *DB) AddBookmark(alias, path string) error {
	if db.isDockerMode {
		return db.addBookmarkDocker(alias, path)
	}

	query := `INSERT INTO bookmarks (alias, path) VALUES (?, ?)`
	_, err := db.conn.Exec(query, alias, path)
	if err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}
	return nil
}

func (db *DB) addBookmarkDocker(alias, path string) error {
	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		fmt.Sprintf("INSERT INTO bookmarks (alias, path) VALUES ('%s', '%s');",
			strings.ReplaceAll(alias, "'", "''"),
			strings.ReplaceAll(path, "'", "''")))

	return cmd.Run()
}

// GetBookmark retrieves a bookmark by alias.
func (db *DB) GetBookmark(alias string) (*Bookmark, error) {
	if db.isDockerMode {
		return db.getBookmarkDocker(alias)
	}

	query := `SELECT id, alias, path, created_at FROM bookmarks WHERE alias = ?`
	row := db.conn.QueryRow(query, alias)

	var bookmark Bookmark
	err := row.Scan(&bookmark.ID, &bookmark.Alias, &bookmark.Path, &bookmark.CreatedAt)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get bookmark: %w", err)
	}

	return &bookmark, nil
}

func (db *DB) getBookmarkDocker(alias string) (*Bookmark, error) {
	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		fmt.Sprintf("SELECT id, alias, path, created_at FROM bookmarks WHERE alias = '%s';",
			strings.ReplaceAll(alias, "'", "''")))

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(lines) == 0 || lines[0] == "" {
		return nil, nil
	}

	parts := strings.Split(lines[0], "|")
	if len(parts) < 4 {
		return nil, fmt.Errorf("invalid bookmark data")
	}

	id, _ := strconv.Atoi(parts[0])
	createdAt, _ := time.Parse("2006-01-02 15:04:05", parts[3])

	return &Bookmark{
		ID:        id,
		Alias:     parts[1],
		Path:      parts[2],
		CreatedAt: createdAt,
	}, nil
}

// ListBookmarks returns all bookmarks.
func (db *DB) ListBookmarks() ([]*Bookmark, error) {
	if db.isDockerMode {
		return db.listBookmarksDocker()
	}

	query := `SELECT id, alias, path, created_at FROM bookmarks ORDER BY alias`
	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var bookmark Bookmark
		err := rows.Scan(&bookmark.ID, &bookmark.Alias, &bookmark.Path, &bookmark.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, &bookmark)
	}

	return bookmarks, nil
}

func (db *DB) listBookmarksDocker() ([]*Bookmark, error) {
	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		"SELECT id, alias, path, created_at FROM bookmarks ORDER BY alias;")

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var bookmarks []*Bookmark

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		id, _ := strconv.Atoi(parts[0])
		createdAt, _ := time.Parse("2006-01-02 15:04:05", parts[3])

		bookmarks = append(bookmarks, &Bookmark{
			ID:        id,
			Alias:     parts[1],
			Path:      parts[2],
			CreatedAt: createdAt,
		})
	}

	return bookmarks, nil
}

// RemoveBookmark removes a bookmark by alias.
func (db *DB) RemoveBookmark(alias string) error {
	if db.isDockerMode {
		return db.removeBookmarkDocker(alias)
	}

	query := `DELETE FROM bookmarks WHERE alias = ?`
	result, err := db.conn.Exec(query, alias)
	if err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bookmark '%s' not found", alias)
	}

	return nil
}

func (db *DB) removeBookmarkDocker(alias string) error {
	// First check if bookmark exists
	existing, err := db.getBookmarkDocker(alias)
	if err != nil {
		return err
	}
	if existing == nil {
		return fmt.Errorf("bookmark '%s' not found", alias)
	}

	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		fmt.Sprintf("DELETE FROM bookmarks WHERE alias = '%s';",
			strings.ReplaceAll(alias, "'", "''")))

	return cmd.Run()
}

// AddNavigationHistory adds a path to navigation history.
func (db *DB) AddNavigationHistory(path string) error {
	if db.isDockerMode {
		return db.addNavigationHistoryDocker(path)
	}

	query := `INSERT INTO navigation_history (path) VALUES (?)`
	_, err := db.conn.Exec(query, path)
	if err != nil {
		return fmt.Errorf("failed to add navigation history: %w", err)
	}
	return nil
}

func (db *DB) addNavigationHistoryDocker(path string) error {
	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		fmt.Sprintf("INSERT INTO navigation_history (path) VALUES ('%s');",
			strings.ReplaceAll(path, "'", "''")))

	return cmd.Run()
}

// FuzzySearch performs fuzzy searching on both bookmarks and history.
func (db *DB) FuzzySearch(query string) ([]*Bookmark, error) {
	// Normalize the query
	query = strings.ToLower(strings.TrimSpace(query))
	if query == "" {
		return db.ListBookmarks()
	}

	// First try to find bookmarks
	bookmarks, err := db.searchBookmarks(query)
	if err != nil {
		return nil, err
	}

	// If we have matches, return them
	if len(bookmarks) > 0 {
		return bookmarks, nil
	}

	// If no bookmarks found, search in navigation history
	return db.searchHistory(query)
}

func (db *DB) searchBookmarks(query string) ([]*Bookmark, error) {
	if db.isDockerMode {
		return db.searchBookmarksDocker(query)
	}

	searchQuery := `
		SELECT id, alias, path, created_at 
		FROM bookmarks 
		WHERE LOWER(alias) LIKE ? OR LOWER(path) LIKE ?
		ORDER BY 
			CASE 
				WHEN LOWER(alias) = ? THEN 1
				WHEN LOWER(alias) LIKE ? THEN 2
				WHEN LOWER(path) LIKE ? THEN 3
				ELSE 4
			END`

	queryPattern := "%" + strings.ToLower(query) + "%"
	exactMatch := strings.ToLower(query)
	prefixMatch := strings.ToLower(query) + "%"

	rows, err := db.conn.Query(searchQuery, queryPattern, queryPattern, exactMatch, prefixMatch, prefixMatch)
	if err != nil {
		return nil, fmt.Errorf("failed to search bookmarks: %w", err)
	}
	defer rows.Close()

	var bookmarks []*Bookmark
	for rows.Next() {
		var bookmark Bookmark
		err := rows.Scan(&bookmark.ID, &bookmark.Alias, &bookmark.Path, &bookmark.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan bookmark: %w", err)
		}
		bookmarks = append(bookmarks, &bookmark)
	}

	return bookmarks, nil
}

func (db *DB) searchBookmarksDocker(query string) ([]*Bookmark, error) {
	queryPattern := "%" + strings.ToLower(query) + "%"

	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		fmt.Sprintf(`SELECT id, alias, path, created_at FROM bookmarks 
		WHERE LOWER(alias) LIKE '%s' OR LOWER(path) LIKE '%s' 
		ORDER BY alias;`, queryPattern, queryPattern))

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var bookmarks []*Bookmark

	for _, line := range lines {
		if line == "" {
			continue
		}

		parts := strings.Split(line, "|")
		if len(parts) < 4 {
			continue
		}

		id, _ := strconv.Atoi(parts[0])
		createdAt, _ := time.Parse("2006-01-02 15:04:05", parts[3])

		bookmarks = append(bookmarks, &Bookmark{
			ID:        id,
			Alias:     parts[1],
			Path:      parts[2],
			CreatedAt: createdAt,
		})
	}

	return bookmarks, nil
}

func (db *DB) searchHistory(query string) ([]*Bookmark, error) {
	if db.isDockerMode {
		return db.searchHistoryDocker(query)
	}

	historyQuery := `
		SELECT DISTINCT path 
		FROM navigation_history 
		WHERE LOWER(path) LIKE ? 
		ORDER BY accessed_at DESC 
		LIMIT 10`

	queryPattern := "%" + strings.ToLower(query) + "%"
	rows, err := db.conn.Query(historyQuery, queryPattern)
	if err != nil {
		return nil, fmt.Errorf("failed to search history: %w", err)
	}
	defer rows.Close()

	var historyResults []*Bookmark
	id := -1 // Use negative IDs to distinguish from real bookmarks
	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			continue
		}
		historyResults = append(historyResults, &Bookmark{
			ID:    id,
			Alias: fmt.Sprintf("history:%s", path),
			Path:  path,
		})
		id--
	}

	return historyResults, nil
}

func (db *DB) searchHistoryDocker(query string) ([]*Bookmark, error) {
	queryPattern := "%" + strings.ToLower(query) + "%"

	cmd := exec.Command("docker", "exec", db.containerName, "sqlite3", "/data/aura.db",
		fmt.Sprintf(`SELECT DISTINCT path FROM navigation_history 
		WHERE LOWER(path) LIKE '%s' 
		ORDER BY accessed_at DESC LIMIT 10;`, queryPattern))

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	lines := strings.Split(strings.TrimSpace(string(output)), "\n")
	var historyResults []*Bookmark
	id := -1

	for _, line := range lines {
		if line == "" {
			continue
		}

		historyResults = append(historyResults, &Bookmark{
			ID:    id,
			Alias: fmt.Sprintf("history:%s", line),
			Path:  line,
		})
		id--
	}

	return historyResults, nil
}
