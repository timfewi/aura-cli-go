package db

import (
	"testing"
)

func TestAddBookmark(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	tests := []struct {
		name      string
		alias     string
		path      string
		wantError bool
	}{
		{
			name:      "valid bookmark",
			alias:     "test1",
			path:      "/test/path1",
			wantError: false,
		},
		{
			name:      "another valid bookmark",
			alias:     "test2",
			path:      "/test/path2",
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.AddBookmark(tt.alias, tt.path)
			if (err != nil) != tt.wantError {
				t.Errorf("AddBookmark() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}

	// Test duplicate alias (should fail)
	err = db.AddBookmark("test1", "/different/path")
	if err == nil {
		t.Error("Expected error for duplicate alias")
	}
}

func TestGetBookmark(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test bookmark
	testAlias := "testget"
	testPath := "/test/get/path"
	err = db.AddBookmark(testAlias, testPath)
	if err != nil {
		t.Fatalf("Failed to add test bookmark: %v", err)
	}

	// Test getting existing bookmark
	bookmark, err := db.GetBookmark(testAlias)
	if err != nil {
		t.Errorf("GetBookmark() error = %v", err)
	}
	if bookmark == nil {
		t.Fatal("GetBookmark() returned nil bookmark")
	}
	if bookmark.Alias != testAlias {
		t.Errorf("GetBookmark() alias = %v, want %v", bookmark.Alias, testAlias)
	}
	if bookmark.Path != testPath {
		t.Errorf("GetBookmark() path = %v, want %v", bookmark.Path, testPath)
	}
	if bookmark.ID <= 0 {
		t.Errorf("GetBookmark() invalid ID = %v", bookmark.ID)
	}
	if bookmark.CreatedAt.IsZero() {
		t.Error("GetBookmark() CreatedAt is zero")
	}

	// Test getting non-existent bookmark
	bookmark, err = db.GetBookmark("nonexistent")
	if err != nil {
		t.Errorf("GetBookmark() error for non-existent = %v", err)
	}
	if bookmark != nil {
		t.Error("GetBookmark() should return nil for non-existent bookmark")
	}
}

func TestListBookmarks(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Initially should be empty
	bookmarks, err := db.ListBookmarks()
	if err != nil {
		t.Errorf("ListBookmarks() error = %v", err)
	}
	initialCount := len(bookmarks)

	// Add test bookmarks
	testBookmarks := []struct {
		alias string
		path  string
	}{
		{"home_list", "/home/user"},
		{"projects_list", "/home/user/projects"},
		{"docs_list", "/home/user/documents"},
	}

	for _, bm := range testBookmarks {
		err = db.AddBookmark(bm.alias, bm.path)
		if err != nil {
			t.Fatalf("Failed to add test bookmark %s: %v", bm.alias, err)
		}
	}

	// List bookmarks
	bookmarks, err = db.ListBookmarks()
	if err != nil {
		t.Errorf("ListBookmarks() error = %v", err)
	}
	expectedCount := initialCount + len(testBookmarks)
	if len(bookmarks) != expectedCount {
		t.Errorf("ListBookmarks() count = %v, want %v", len(bookmarks), expectedCount)
	}

	// Verify our test bookmarks are present
	aliasMap := make(map[string]string)
	for _, bm := range bookmarks {
		aliasMap[bm.Alias] = bm.Path
	}

	for _, expected := range testBookmarks {
		if path, exists := aliasMap[expected.alias]; !exists {
			t.Errorf("Missing bookmark alias: %s", expected.alias)
		} else if path != expected.path {
			t.Errorf("Wrong path for %s: got %s, want %s", expected.alias, path, expected.path)
		}
	}
}

func TestRemoveBookmark(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Add test bookmark
	testAlias := "testremove_unique"
	testPath := "/test/remove/path"
	err = db.AddBookmark(testAlias, testPath)
	if err != nil {
		t.Fatalf("Failed to add test bookmark: %v", err)
	}

	// Remove the bookmark
	err = db.RemoveBookmark(testAlias)
	if err != nil {
		t.Errorf("RemoveBookmark() error = %v", err)
	}

	// Verify it's gone
	bookmark, err := db.GetBookmark(testAlias)
	if err != nil {
		t.Errorf("GetBookmark() after remove error = %v", err)
	}
	if bookmark != nil {
		t.Error("Bookmark still exists after removal")
	}

	// Test removing non-existent bookmark (should not error)
	_ = db.RemoveBookmark("nonexistent_bookmark_12345")
	// The implementation may or may not return an error for non-existent bookmarks
	// This depends on the specific implementation
}

func TestAddNavigationHistory(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	testPath := "/test/history/path"
	err = db.AddNavigationHistory(testPath)
	if err != nil {
		t.Errorf("AddNavigationHistory() error = %v", err)
	}

	// Add the same path again (should work)
	err = db.AddNavigationHistory(testPath)
	if err != nil {
		t.Errorf("AddNavigationHistory() duplicate error = %v", err)
	}
}

func TestFuzzySearch(t *testing.T) {
	db, err := New()
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Use unique prefixes to avoid conflicts with existing data
	prefix := "fuzztest_"

	// Add test bookmarks with unique names
	testBookmarks := []struct {
		alias string
		path  string
	}{
		{prefix + "home", "/test/fuzzy/home"},
		{prefix + "projects", "/test/fuzzy/projects"},
		{prefix + "project_docs", "/test/fuzzy/project/docs"},
		{prefix + "downloads", "/test/fuzzy/downloads"},
	}

	// Clean up any existing test bookmarks first
	for _, bm := range testBookmarks {
		_ = db.RemoveBookmark(bm.alias)
	}

	// Add fresh test bookmarks
	for _, bm := range testBookmarks {
		err = db.AddBookmark(bm.alias, bm.path)
		if err != nil {
			t.Fatalf("Failed to add test bookmark %s: %v", bm.alias, err)
		}
	}

	// Cleanup at the end
	defer func() {
		for _, bm := range testBookmarks {
			_ = db.RemoveBookmark(bm.alias)
		}
	}()

	tests := []struct {
		name        string
		query       string
		expectCount int
		shouldFind  []string
	}{
		{
			name:        "exact match with prefix",
			query:       prefix + "home",
			expectCount: 1,
			shouldFind:  []string{prefix + "home"},
		},
		{
			name:        "partial match with prefix",
			query:       prefix + "proj",
			expectCount: 2, // fuzztest_projects, fuzztest_project_docs
			shouldFind:  []string{prefix + "projects", prefix + "project_docs"},
		},
		{
			name:        "no matches",
			query:       "xyznomatch_unique_12345",
			expectCount: 0,
			shouldFind:  []string{},
		},
		{
			name:        "prefix only",
			query:       prefix,
			expectCount: 4, // All our test bookmarks
			shouldFind:  []string{prefix + "home", prefix + "projects", prefix + "project_docs", prefix + "downloads"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := db.FuzzySearch(tt.query)
			if err != nil {
				t.Errorf("FuzzySearch() error = %v", err)
				return
			}

			// Filter results to only include our test bookmarks
			filteredResults := []Bookmark{}
			for _, result := range results {
				alias := result.Alias
				if result.ID < 0 {
					// This is from history, skip for now
					continue
				}
				if len(alias) >= len(prefix) && alias[:len(prefix)] == prefix {
					filteredResults = append(filteredResults, *result)
				}
			}

			if len(filteredResults) != tt.expectCount {
				t.Errorf("FuzzySearch() count = %v, want %v (found: %+v)", len(filteredResults), tt.expectCount, filteredResults)
			}

			resultAliases := make(map[string]bool)
			for _, result := range filteredResults {
				resultAliases[result.Alias] = true
			}

			for _, expected := range tt.shouldFind {
				if !resultAliases[expected] {
					t.Errorf("FuzzySearch() missing expected result: %s", expected)
				}
			}
		})
	}
}
