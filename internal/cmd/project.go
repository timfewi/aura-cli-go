package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"

	"github.com/timfewi/aura-cli-go/assets"
)

var projectCmd = &cobra.Command{
	Use:   "project [project-name]",
	Short: "Create a project from template",
	Long: `Create a project with predefined templates and structure.

Examples:
  aura project my-api --type python
  aura project my-app --type node
  aura project my-tool --type go`,
	Args: cobra.ExactArgs(1),
	RunE: runProject,
}

var (
	projectType string
	description string
	author      string
)

type ProjectData struct {
	ProjectName string
	Type        string
	Description string
	Author      string
	ModuleName  string
	GoVersion   string
	RepoURL     string
}

func runProject(cmd *cobra.Command, args []string) error {
	projectName := args[0]

	// Validate project name
	if !isValidProjectName(projectName) {
		return fmt.Errorf("invalid project name. Use only letters, numbers, hyphens, and underscores")
	}

	// Check if directory already exists
	if _, err := os.Stat(projectName); !os.IsNotExist(err) {
		return fmt.Errorf("directory '%s' already exists", projectName)
	}

	// If type not specified, prompt for it
	if projectType == "" {
		var err error
		projectType, err = promptForProjectType()
		if err != nil {
			return err
		}
	}

	// Validate project type
	validTypes := []string{"python", "node", "go"}
	if !contains(validTypes, projectType) {
		return fmt.Errorf("unsupported project type '%s'. Supported types: %s", projectType, strings.Join(validTypes, ", "))
	}

	// Get additional information
	if description == "" {
		description = fmt.Sprintf("A new %s project", projectType)
	}

	if author == "" {
		author = "Your Name"
	}

	// Create project data
	projectData := ProjectData{
		ProjectName: projectName,
		Type:        projectType,
		Description: description,
		Author:      author,
		ModuleName:  fmt.Sprintf("github.com/%s/%s", author, projectName),
		GoVersion:   "1.21",
		RepoURL:     fmt.Sprintf("https://github.com/%s/%s.git", author, projectName),
	}

	// Create project directory
	if err := os.MkdirAll(projectName, 0755); err != nil {
		return fmt.Errorf("failed to create project directory: %w", err)
	}

	// Generate project files
	if err := generateProjectFiles(projectName, projectData); err != nil {
		// Clean up on error
		os.RemoveAll(projectName)
		return fmt.Errorf("failed to generate project files: %w", err)
	}

	// Initialize git repository
	if err := initializeGitRepository(projectName); err != nil {
		fmt.Printf("Warning: Failed to initialize git repository: %v\n", err)
	}

	fmt.Printf("✓ Created %s project '%s'\n", projectType, projectName)
	fmt.Printf("📁 Directory: %s\n", projectName)
	fmt.Printf("🚀 Get started:\n")
	fmt.Printf("   cd %s\n", projectName)

	switch projectType {
	case "python":
		fmt.Printf("   python -m venv venv\n")
		fmt.Printf("   source venv/bin/activate  # On Windows: venv\\Scripts\\activate\n")
		fmt.Printf("   pip install -r requirements.txt\n")
		fmt.Printf("   python main.py\n")
	case "node":
		fmt.Printf("   npm install\n")
		fmt.Printf("   npm start\n")
	case "go":
		fmt.Printf("   go mod tidy\n")
		fmt.Printf("   go run .\n")
	}

	return nil
}

func promptForProjectType() (string, error) {
	prompt := promptui.Select{
		Label: "Select project type",
		Items: []string{"python", "node", "go"},
		Templates: &promptui.SelectTemplates{
			Label:    "{{ . }}?",
			Active:   "▸ {{ . | cyan }}",
			Inactive: "  {{ . | white }}",
			Selected: "✓ {{ . | green }}",
		},
	}

	_, result, err := prompt.Run()
	return result, err
}

func generateProjectFiles(projectDir string, data ProjectData) error {
	// Common files for all projects
	commonFiles := []string{
		"README.md.tmpl",
	}

	// Type-specific files
	var typeFiles []string
	var gitignoreTemplate string

	switch data.Type {
	case "python":
		typeFiles = []string{"main.py.tmpl"}
		gitignoreTemplate = "python.gitignore.tmpl"

		// Create requirements.txt
		reqFile := filepath.Join(projectDir, "requirements.txt")
		if err := writeStringToFile(reqFile, "# Add your Python dependencies here\n"); err != nil {
			return err
		}

	case "node":
		typeFiles = []string{"package.json.tmpl", "index.js.tmpl"}
		gitignoreTemplate = "node.gitignore.tmpl"

	case "go":
		typeFiles = []string{"go.mod.tmpl", "main.go.tmpl"}
		gitignoreTemplate = "go.gitignore.tmpl"
	}

	// Generate common files
	for _, templateFile := range commonFiles {
		if err := generateFromTemplate(projectDir, templateFile, data); err != nil {
			return err
		}
	}

	// Generate type-specific files
	for _, templateFile := range typeFiles {
		if err := generateFromTemplate(projectDir, templateFile, data); err != nil {
			return err
		}
	}

	// Generate .gitignore
	if gitignoreTemplate != "" {
		if err := generateGitignore(projectDir, gitignoreTemplate, data); err != nil {
			return err
		}
	}

	return nil
}

func generateFromTemplate(projectDir, templateFile string, data ProjectData) error {
	// Read template from embedded assets
	templateContent, err := assets.Templates.ReadFile("templates/" + templateFile)
	if err != nil {
		return fmt.Errorf("failed to read template %s: %w", templateFile, err)
	}

	// Parse template
	tmpl, err := template.New(templateFile).Parse(string(templateContent))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templateFile, err)
	}

	// Determine output filename (remove .tmpl extension)
	outputFile := strings.TrimSuffix(templateFile, ".tmpl")
	outputPath := filepath.Join(projectDir, outputFile)

	// Create output file
	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", outputPath, err)
	}
	defer file.Close()

	// Execute template
	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templateFile, err)
	}

	return nil
}

func generateGitignore(projectDir, templateFile string, data ProjectData) error {
	// Read gitignore template
	templateContent, err := assets.Templates.ReadFile("templates/" + templateFile)
	if err != nil {
		return fmt.Errorf("failed to read gitignore template: %w", err)
	}

	// Write .gitignore file
	gitignorePath := filepath.Join(projectDir, ".gitignore")
	return writeStringToFile(gitignorePath, string(templateContent))
}

func writeStringToFile(filePath, content string) error {
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

func initializeGitRepository(projectDir string) error {
	cmd := fmt.Sprintf("cd %s && git init", projectDir)
	return runShellCommand(cmd)
}

func runShellCommand(command string) error {
	var cmd *exec.Cmd
	if isWindows() {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	return cmd.Run()
}

func isValidProjectName(name string) bool {
	if name == "" {
		return false
	}

	for _, char := range name {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return false
		}
	}

	return true
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func init() {
	projectCmd.Flags().StringVar(&projectType, "type", "", "Project type (python, node, go)")
	projectCmd.Flags().StringVar(&description, "description", "", "Project description")
	projectCmd.Flags().StringVar(&author, "author", "", "Author name")

	rootCmd.AddCommand(projectCmd)
}
