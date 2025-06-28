package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/internal/context"
)

var doCmd = &cobra.Command{
	Use:   "do",
	Short: "Show context-aware action suggestions",
	Long: `Analyze the current directory and suggest relevant actions based on the detected context.
	
This command detects various project types (Git, Node.js, Python, Go, Docker, etc.)
and presents an interactive list of common actions you might want to perform.`,
	RunE: runDo,
}

func runDo(cmd *cobra.Command, args []string) error {
	// Collect actions from all detectors
	var allActions []context.Action

	// Run detectors in parallel for better performance
	detectors := []func() []context.Action{
		context.DetectGitContext,
		context.DetectNodeContext,
		context.DetectPythonContext,
		context.DetectGoContext,
		context.DetectDockerContext,
		context.DetectMakeContext,
	}

	for _, detector := range detectors {
		if actions := detector(); actions != nil {
			allActions = append(allActions, actions...)
		}
	}

	if len(allActions) == 0 {
		fmt.Println("No specific context detected in this directory.")
		fmt.Println("Try running 'aura do' in a project directory (Git repo, Node.js project, etc.)")
		return nil
	}

	// Add general actions that are always available
	generalActions := []context.Action{
		{Name: "Open current directory", Command: getOpenCommand()},
		{Name: "List directory contents", Command: getListCommand()},
		{Name: "Show disk usage", Command: getDiskUsageCommand()},
		{Name: "Find large files", Command: getFindLargeFilesCommand()},
	}
	allActions = append(allActions, generalActions...)

	// Create display items for the prompt
	items := make([]string, len(allActions))
	for i, action := range allActions {
		items[i] = action.Name
	}

	// Create interactive prompt
	prompt := promptui.Select{
		Label: "Select an action",
		Items: items,
		Size:  10,
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
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

	selectedAction := allActions[selectedIndex]

	// Show the command that will be executed
	fmt.Printf("Executing: %s\n", selectedAction.Command)

	// Execute the selected command
	return executeCommand(selectedAction.Command)
}

func executeCommand(command string) error {
	// Parse the command into parts
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	// Handle special cases for interactive commands
	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// getOpenCommand returns the appropriate command to open the current directory
// based on the operating system.
func getOpenCommand() string {
	switch {
	case isWindows():
		return "explorer ."
	case isMacOS():
		return "open ."
	default:
		return "xdg-open ."
	}
}

// getListCommand returns the appropriate command to list directory contents
// based on the operating system.
func getListCommand() string {
	if isWindows() {
		return "dir"
	}
	return "ls -la"
}

// getDiskUsageCommand returns the appropriate command to show disk usage
// based on the operating system.
func getDiskUsageCommand() string {
	if isWindows() {
		return "powershell -Command \"Get-ChildItem | Measure-Object -Property Length -Sum | Select-Object @{Name='Size(MB)';Expression={[math]::Round($_.Sum/1MB,2)}}\""
	}
	return "du -sh *"
}

// getFindLargeFilesCommand returns the appropriate command to find large files
// based on the operating system.
func getFindLargeFilesCommand() string {
	if isWindows() {
		return "powershell -Command \"Get-ChildItem -Recurse -File | Where-Object {$_.Length -gt 10MB} | Select-Object Name, @{Name='Size(MB)';Expression={[math]::Round($_.Length/1MB,2)}}, FullName | Sort-Object 'Size(MB)' -Descending\""
	}
	return "find . -type f -size +10M -exec ls -lh {} \\;"
}

func isWindows() bool {
	// Check environment variable first for testing purposes
	if os := os.Getenv("OS"); os != "" {
		return strings.Contains(strings.ToLower(os), "windows")
	}
	return runtime.GOOS == "windows"
}

func isMacOS() bool {
	// Check environment variable first for testing purposes
	if ostype := os.Getenv("OSTYPE"); ostype != "" {
		return strings.Contains(strings.ToLower(ostype), "darwin")
	}
	return runtime.GOOS == "darwin"
}

func init() {
	rootCmd.AddCommand(doCmd)
}
