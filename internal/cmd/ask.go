package cmd

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/internal/ai"
)

var askCmd = &cobra.Command{
	Use:   "ask [question...]",
	Short: "Ask AI assistant for help",
	Long: `Ask the AI assistant questions about commands, code, or general development tasks.
	
Examples:
  aura ask "how to find large files"
  aura ask "explain this bash script"
  cat script.py | aura ask "explain this code"
  aura ask "best practices for git workflow"`,
	RunE: runAsk,
}

func runAsk(cmd *cobra.Command, args []string) error {
	client, err := ai.NewClient()
	if err != nil {
		return fmt.Errorf("failed to initialize AI client: %w", err)
	}

	var question string

	// Check if there's input from stdin (piped content)
	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		// There's piped input
		stdinBytes, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("failed to read from stdin: %w", err)
		}
		stdinContent := strings.TrimSpace(string(stdinBytes))

		if len(args) == 0 {
			// If no question provided, use default
			question = fmt.Sprintf("Explain this:\n\n%s", stdinContent)
		} else {
			// Combine question with piped content
			userQuestion := strings.Join(args, " ")
			question = fmt.Sprintf("%s\n\nContent:\n%s", userQuestion, stdinContent)
		}
	} else {
		// No piped input, use command line arguments
		if len(args) == 0 {
			// Interactive mode
			return runInteractiveAsk(client)
		}
		question = strings.Join(args, " ")
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Show thinking indicator
	done := make(chan bool)
	go showThinking(done)

	// Get response from AI
	response, err := client.Ask(ctx, question)
	done <- true

	if err != nil {
		return fmt.Errorf("AI request failed: %w", err)
	}

	// Print the response
	fmt.Printf("\n%s\n", response)
	return nil
}

func runInteractiveAsk(client *ai.Client) error {
	fmt.Println("Aura AI Assistant - Interactive Mode")
	fmt.Println("Type your questions or 'exit' to quit.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("❯ ")
		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		if input == "exit" || input == "quit" {
			fmt.Println("Goodbye!")
			break
		}

		// Create context with timeout
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

		// Show thinking indicator
		done := make(chan bool)
		go showThinking(done)

		// Get response from AI
		response, err := client.Ask(ctx, input)
		done <- true
		cancel()

		if err != nil {
			fmt.Printf("Error: %v\n\n", err)
			continue
		}

		// Print the response
		fmt.Printf("\n%s\n\n", response)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading input: %w", err)
	}

	return nil
}

func showThinking(done chan bool) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}
	i := 0

	fmt.Print("Thinking ")
	for {
		select {
		case <-done:
			fmt.Print("\r" + strings.Repeat(" ", 20) + "\r") // Clear the line
			return
		case <-ticker.C:
			fmt.Printf("\rThinking %s", chars[i%len(chars)])
			i++
		}
	}
}

func init() {
	rootCmd.AddCommand(askCmd)
}
