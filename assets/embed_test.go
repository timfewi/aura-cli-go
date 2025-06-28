package assets

import (
	"testing"
)

func TestTemplatesEmbed(t *testing.T) {
	// Test reading the templates directory instead of checking Open function
	entries, err := Templates.ReadDir("templates")
	if err != nil {
		t.Errorf("Failed to read templates directory: %v", err)
	}

	if len(entries) == 0 {
		t.Error("No template files found")
	}

	// Test for expected template files
	expectedTemplates := []string{
		"go.gitignore.tmpl",
		"go.mod.tmpl",
		"index.js.tmpl",
		"main.go.tmpl",
		"main.py.tmpl",
		"node.gitignore.tmpl",
		"package.json.tmpl",
		"python.gitignore.tmpl",
		"README.md.tmpl",
	}

	templateNames := make(map[string]bool)
	for _, entry := range entries {
		templateNames[entry.Name()] = true
	}

	for _, expected := range expectedTemplates {
		if !templateNames[expected] {
			t.Errorf("Missing expected template file: %s", expected)
		}
	}
}

func TestTemplateContent(t *testing.T) {
	// Test reading specific template content
	templates := []struct {
		name          string
		shouldContain []string
	}{
		{
			name: "templates/main.go.tmpl",
			shouldContain: []string{
				"package main",
				"import",
				"func main",
				"{{.ProjectName}}",
			},
		},
		{
			name: "templates/go.mod.tmpl",
			shouldContain: []string{
				"module",
				"go ",
				"{{.ModuleName}}",
			},
		},
		{
			name: "templates/package.json.tmpl",
			shouldContain: []string{
				"\"name\":",
				"\"version\":",
				"{{.ProjectName}}",
			},
		},
		{
			name: "templates/README.md.tmpl",
			shouldContain: []string{
				"#",
				"{{.ProjectName}}",
			},
		},
	}

	for _, tt := range templates {
		t.Run(tt.name, func(t *testing.T) {
			content, err := Templates.ReadFile(tt.name)
			if err != nil {
				t.Errorf("Failed to read template %s: %v", tt.name, err)
				return
			}

			contentStr := string(content)
			if contentStr == "" {
				t.Errorf("Template %s is empty", tt.name)
				return
			}

			for _, expected := range tt.shouldContain {
				if !contains(contentStr, expected) {
					t.Errorf("Template %s should contain '%s'", tt.name, expected)
				}
			}
		})
	}
}

func TestGitignoreTemplates(t *testing.T) {
	gitignoreTemplates := []struct {
		name          string
		shouldContain []string
	}{
		{
			name: "templates/go.gitignore.tmpl",
			shouldContain: []string{
				"*.exe",
				"*.dll",
				"*.so",
				"*.dylib",
			},
		},
		{
			name: "templates/node.gitignore.tmpl",
			shouldContain: []string{
				"node_modules/",
				"npm-debug.log*",
				"*.tgz",
			},
		},
		{
			name: "templates/python.gitignore.tmpl",
			shouldContain: []string{
				"__pycache__/",
				"*.py[cod]", // This matches the actual template content
				"*.egg-info/",
			},
		},
	}

	for _, tt := range gitignoreTemplates {
		t.Run(tt.name, func(t *testing.T) {
			content, err := Templates.ReadFile(tt.name)
			if err != nil {
				t.Errorf("Failed to read gitignore template %s: %v", tt.name, err)
				return
			}

			contentStr := string(content)
			if contentStr == "" {
				t.Errorf("Gitignore template %s is empty", tt.name)
				return
			}

			for _, expected := range tt.shouldContain {
				if !contains(contentStr, expected) {
					t.Errorf("Gitignore template %s should contain '%s'", tt.name, expected)
				}
			}
		})
	}
}

func TestTemplateFileAccess(t *testing.T) {
	// Test that we can open and read template files
	testFiles := []string{
		"templates/main.go.tmpl",
		"templates/main.py.tmpl",
		"templates/index.js.tmpl",
	}

	for _, filename := range testFiles {
		t.Run(filename, func(t *testing.T) {
			// Test opening file
			file, err := Templates.Open(filename)
			if err != nil {
				t.Errorf("Failed to open template file %s: %v", filename, err)
				return
			}
			defer file.Close()

			// Test reading file info
			info, err := file.Stat()
			if err != nil {
				t.Errorf("Failed to get file info for %s: %v", filename, err)
				return
			}

			if info.Size() == 0 {
				t.Errorf("Template file %s is empty", filename)
			}

			if info.IsDir() {
				t.Errorf("Template file %s should not be a directory", filename)
			}
		})
	}
}

func TestTemplatesDirectoryStructure(t *testing.T) {
	// Test that templates directory exists and has correct structure
	entries, err := Templates.ReadDir("templates")
	if err != nil {
		t.Fatalf("Failed to read templates directory: %v", err)
	}

	// Count different types of templates
	var goTemplates, nodeTemplates, pythonTemplates, gitignoreTemplates int

	for _, entry := range entries {
		name := entry.Name()
		switch {
		case contains(name, "go") && !contains(name, "gitignore"):
			goTemplates++
		case contains(name, "node") || contains(name, "js") || contains(name, "package.json"):
			nodeTemplates++
		case contains(name, "python") || contains(name, "py"):
			pythonTemplates++
		case contains(name, "gitignore"):
			gitignoreTemplates++
		}
	}

	if goTemplates == 0 {
		t.Error("No Go templates found")
	}

	if nodeTemplates == 0 {
		t.Error("No Node.js templates found")
	}

	if pythonTemplates == 0 {
		t.Error("No Python templates found")
	}

	if gitignoreTemplates == 0 {
		t.Error("No gitignore templates found")
	}
}

// Helper function since strings.Contains might not be available
func contains(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
