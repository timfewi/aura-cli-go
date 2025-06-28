package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/internal/db"
)

var bookmarkCmd = &cobra.Command{
	Use:   "bookmark",
	Short: "Manage directory bookmarks",
	Long:  `Add, remove, and list directory bookmarks for quick navigation.`,
}

var bookmarkAddCmd = &cobra.Command{
	Use:   "add [alias] [path]",
	Short: "Add a bookmark",
	Long: `Add a bookmark for quick navigation.
	
Examples:
  aura bookmark add notes ~/Documents/notes
  aura bookmark add proj .                    # Bookmark current directory
  aura bookmark add this as notes             # Natural language syntax`,
	RunE: runBookmarkAdd,
}

var bookmarkListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all bookmarks",
	Long:  `List all saved bookmarks.`,
	RunE:  runBookmarkList,
}

var bookmarkRemoveCmd = &cobra.Command{
	Use:   "remove [alias]",
	Short: "Remove a bookmark",
	Long:  `Remove a bookmark by its alias.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runBookmarkRemove,
}

func runBookmarkAdd(cmd *cobra.Command, args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("alias is required")
	}

	var alias, path string

	// Handle natural language syntax: "aura bookmark add this as notes"
	if len(args) >= 3 && args[0] == "this" && args[1] == "as" {
		alias = args[2]
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else if len(args) == 1 {
		// If only alias provided, use current directory
		alias = args[0]
		var err error
		path, err = os.Getwd()
		if err != nil {
			return fmt.Errorf("failed to get current directory: %w", err)
		}
	} else if len(args) >= 2 {
		alias = args[0]
		path = strings.Join(args[1:], " ")
	} else {
		return fmt.Errorf("invalid arguments")
	}

	// Convert to absolute path
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("failed to resolve path: %w", err)
	}

	// Verify the path exists and is a directory
	stat, err := os.Stat(absPath)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path '%s' does not exist", absPath)
		}
		return fmt.Errorf("failed to check path: %w", err)
	}

	if !stat.IsDir() {
		return fmt.Errorf("'%s' is not a directory", absPath)
	}

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// Check if bookmark already exists
	existing, err := database.GetBookmark(alias)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if existing != nil {
		fmt.Printf("Bookmark '%s' already exists, pointing to: %s\n", alias, existing.Path)
		fmt.Printf("Updating to point to: %s\n", absPath)

		// Remove old bookmark and add new one
		if err := database.RemoveBookmark(alias); err != nil {
			return fmt.Errorf("failed to update bookmark: %w", err)
		}
	}

	if err := database.AddBookmark(alias, absPath); err != nil {
		return fmt.Errorf("failed to add bookmark: %w", err)
	}

	fmt.Printf("Bookmark '%s' added for: %s\n", alias, absPath)
	return nil
}

func runBookmarkList(cmd *cobra.Command, args []string) error {
	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	bookmarks, err := database.ListBookmarks()
	if err != nil {
		return fmt.Errorf("failed to list bookmarks: %w", err)
	}

	if len(bookmarks) == 0 {
		fmt.Println("No bookmarks found. Add one with: aura bookmark add <alias> <path>")
		return nil
	}

	fmt.Println("Saved bookmarks:")
	for _, bookmark := range bookmarks {
		fmt.Printf("  %s -> %s\n", bookmark.Alias, bookmark.Path)
	}

	return nil
}

func runBookmarkRemove(cmd *cobra.Command, args []string) error {
	alias := args[0]

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	if err := database.RemoveBookmark(alias); err != nil {
		return fmt.Errorf("failed to remove bookmark: %w", err)
	}

	fmt.Printf("Bookmark '%s' removed\n", alias)
	return nil
}

func init() {
	bookmarkCmd.AddCommand(bookmarkAddCmd)
	bookmarkCmd.AddCommand(bookmarkListCmd)
	bookmarkCmd.AddCommand(bookmarkRemoveCmd)
	rootCmd.AddCommand(bookmarkCmd)
}
