package cmd

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/internal/ai"
)

var gitCmd = &cobra.Command{
	Use:   "git",
	Short: "AI-powered Git operations",
	Long:  `AI-powered Git operations including commit message generation.`,
}

var gitCommitCmd = &cobra.Command{
	Use:   "commit",
	Short: "Generate AI commit message and commit",
	Long: `Generate an AI-powered commit message based on staged changes and commit.
	
This command will:
1. Check for staged changes using 'git diff --staged'
2. Send the diff to AI for commit message generation
3. Present the suggested commit message for approval
4. Commit with the approved message`,
	RunE: runGitCommit,
}

func runGitCommit(cmd *cobra.Command, args []string) error {
	// Check if we're in a git repository
	if !isGitRepository() {
		return fmt.Errorf("not a git repository")
	}

	// Get staged changes
	diff, err := getStagedDiff()
	if err != nil {
		return fmt.Errorf("failed to get staged changes: %w", err)
	}

	if strings.TrimSpace(diff) == "" {
		fmt.Println("No staged changes found. Stage some changes first with 'git add'.")
		return nil
	}

	// Initialize AI client
	client, err := ai.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize AI client: %w", err)
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Show thinking indicator
	done := make(chan bool)
	go showThinking(done)

	// Generate commit message
	commitMessage, err := client.GenerateCommitMessage(ctx, diff)
	done <- true

	if err != nil {
		return fmt.Errorf("failed to generate commit message: %w", err)
	}

	// Clean up the commit message
	commitMessage = strings.TrimSpace(commitMessage)

	// Remove any markdown formatting or quotes that might be added
	commitMessage = strings.Trim(commitMessage, "`\"'")

	// Present the commit message for approval
	fmt.Printf("\nSuggested commit message:\n")
	fmt.Printf("─────────────────────────────────────\n")
	fmt.Printf("%s\n", commitMessage)
	fmt.Printf("─────────────────────────────────────\n")

	// Ask for approval
	prompt := promptui.Select{
		Label: "Do you want to use this commit message?",
		Items: []string{"Yes, commit with this message", "No, let me edit it", "Cancel"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	selectedIndex, _, err := prompt.Run()
	if err != nil {
		if err == promptui.ErrInterrupt {
			fmt.Println("Cancelled.")
			return nil
		}
		return fmt.Errorf("prompt failed: %w", err)
	}

	switch selectedIndex {
	case 0: // Yes, commit
		return commitWithMessage(commitMessage)
	case 1: // Edit
		return editAndCommit(commitMessage)
	case 2: // Cancel
		fmt.Println("Cancelled.")
		return nil
	}

	return nil
}

func isGitRepository() bool {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func getStagedDiff() (string, error) {
	cmd := exec.Command("git", "diff", "--staged")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return string(output), nil
}

func commitWithMessage(message string) error {
	cmd := exec.Command("git", "commit", "-m", message)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	fmt.Println("✓ Committed successfully!")
	return nil
}

func editAndCommit(originalMessage string) error {
	// Create a temporary file with the message
	tempFile, err := os.CreateTemp("", "aura-commit-*.txt")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tempFile.Name())

	if _, err := tempFile.WriteString(originalMessage); err != nil {
		return fmt.Errorf("failed to write to temp file: %w", err)
	}
	tempFile.Close()

	// Get editor from environment or use default
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = getDefaultEditor()
	}

	// Open editor
	cmd := exec.Command(editor, tempFile.Name())
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("editor failed: %w", err)
	}

	// Read the edited message
	editedMessage, err := os.ReadFile(tempFile.Name())
	if err != nil {
		return fmt.Errorf("failed to read edited message: %w", err)
	}

	message := strings.TrimSpace(string(editedMessage))
	if message == "" {
		fmt.Println("Empty commit message. Aborting.")
		return nil
	}

	return commitWithMessage(message)
}

func getDefaultEditor() string {
	// Check common environment variables
	for _, env := range []string{"VISUAL", "EDITOR"} {
		if editor := os.Getenv(env); editor != "" {
			return editor
		}
	}

	// Platform-specific defaults
	switch {
	case isWindows():
		return "notepad"
	case isMacOS():
		return "nano"
	default:
		return "nano"
	}
}

func init() {
	gitCmd.AddCommand(gitCommitCmd)
	rootCmd.AddCommand(gitCmd)
}
