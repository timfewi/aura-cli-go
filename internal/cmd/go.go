package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/internal/db"
)

var goCmd = &cobra.Command{
	Use:   "go [destination]",
	Short: "Navigate to bookmarked directories",
	Long: `Navigate to bookmarked directories using aliases or fuzzy search.
	
Examples:
  aura go my-project     # Navigate to bookmarked 'my-project'
  aura go notes          # Navigate to bookmarked 'notes'
  aura go proj           # Fuzzy search for directories matching 'proj'`,
	Args: cobra.MinimumNArgs(1),
	RunE: runGo,
}

func runGo(cmd *cobra.Command, args []string) error {
	query := strings.Join(args, " ")

	database, err := db.New()
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer database.Close()

	// First try exact bookmark match
	bookmark, err := database.GetBookmark(query)
	if err != nil {
		return fmt.Errorf("database error: %w", err)
	}

	if bookmark != nil {
		// Add to navigation history
		if err := database.AddNavigationHistory(bookmark.Path); err != nil {
			// Log warning but don't fail navigation
			fmt.Fprintf(os.Stderr, "Warning: failed to add to navigation history: %v\n", err)
		}

		// Verify the path exists
		if _, err := os.Stat(bookmark.Path); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Bookmarked path '%s' no longer exists\n", bookmark.Path)
			return fmt.Errorf("path not found")
		}

		// Print the absolute path to stdout for shell wrapper to use
		fmt.Print(bookmark.Path)
		return nil
	}

	// If no exact match, try fuzzy search
	results, err := database.FuzzySearch(query)
	if err != nil {
		return fmt.Errorf("search error: %w", err)
	}

	if len(results) == 0 {
		fmt.Fprintf(os.Stderr, "No bookmarks found matching '%s'\n", query)
		return fmt.Errorf("no matches found")
	}

	// If only one result, use it
	if len(results) == 1 {
		result := results[0]
		if err := database.AddNavigationHistory(result.Path); err != nil {
			// Log warning but don't fail navigation
			fmt.Fprintf(os.Stderr, "Warning: failed to add to navigation history: %v\n", err)
		}

		// Verify the path exists
		if _, err := os.Stat(result.Path); os.IsNotExist(err) {
			fmt.Fprintf(os.Stderr, "Error: Path '%s' no longer exists\n", result.Path)
			return fmt.Errorf("path not found")
		}

		fmt.Print(result.Path)
		return nil
	}

	// Multiple results - display them and ask user to be more specific
	fmt.Fprintf(os.Stderr, "Multiple matches found for '%s':\n", query)
	for _, result := range results {
		if strings.HasPrefix(result.Alias, "history:") {
			fmt.Fprintf(os.Stderr, "  %s\n", result.Path)
		} else {
			fmt.Fprintf(os.Stderr, "  %s -> %s\n", result.Alias, result.Path)
		}
	}
	fmt.Fprintf(os.Stderr, "Please be more specific.\n")
	return fmt.Errorf("ambiguous query")
}

func init() {
	rootCmd.AddCommand(goCmd)
}
