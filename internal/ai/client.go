package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"
)

// Client represents an AI client for making requests to an LLM API.
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// Message represents a chat message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatRequest represents a request to the chat API.
type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float64   `json:"temperature,omitempty"`
	MaxTokens   int       `json:"max_tokens,omitempty"`
}

// ChatResponse represents a response from the chat API.
type ChatResponse struct {
	Choices []struct {
		Message Message `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

// NewClient creates a new AI client.
func NewClient() (*Client, error) {
	apiKey := os.Getenv("AURA_API_KEY")
	if apiKey == "" {
		// Try OpenAI API key as fallback
		apiKey = os.Getenv("OPENAI_API_KEY")
		if apiKey == "" {
			return nil, fmt.Errorf("AURA_API_KEY or OPENAI_API_KEY environment variable is required")
		}
	}

	baseURL := os.Getenv("AURA_API_URL")
	if baseURL == "" {
		baseURL = "https://api.openai.com/v1"
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}, nil
}

// Ask sends a question to the AI and returns the response.
func (c *Client) Ask(ctx context.Context, question string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are Aura, an intelligent CLI assistant that helps developers and system administrators work more efficiently.

CORE CAPABILITIES:
- Command-line operations and shell scripting
- Programming languages (Go, Python, JavaScript, etc.)
- System administration and DevOps tasks
- Development workflow optimization
- Git operations and version control
- File and directory navigation
- Database operations (especially SQLite)
- AI/LLM integration guidance

SYSTEM CONTEXT:
- Operating System: %s
- Architecture: %s
- Shell: PowerShell (Windows) / Bash/Zsh (Unix-like)

RESPONSE GUIDELINES:
1. Provide actionable, practical solutions
2. Include specific commands when appropriate
3. Explain why a solution works, not just how
4. Consider security implications
5. Offer alternatives when possible
6. Use platform-appropriate commands for %s
7. Format code blocks with proper syntax highlighting
8. Keep responses concise but comprehensive

COMMAND FORMAT:
- For Windows PowerShell: Use PowerShell syntax
- For Unix/Linux: Use bash/shell syntax
- Always specify which shell/platform when ambiguous

Remember: You're part of the Aura ecosystem - a CLI tool focused on intelligent navigation and context-aware actions.`, runtime.GOOS, runtime.GOARCH, runtime.GOOS)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: question},
	}

	return c.chat(ctx, messages)
}

// GenerateCommitMessage generates a Git commit message based on the diff.
func (c *Client) GenerateCommitMessage(ctx context.Context, diff string) (string, error) {
	if diff == "" {
		return "", fmt.Errorf("no staged changes found")
	}

	systemPrompt := `You are an expert Git commit message generator that follows industry best practices and conventional commit standards.

COMMIT MESSAGE RULES:
1. Format: <type>(<scope>): <description>
2. Keep subject line under 50 characters
3. Use present tense, imperative mood ("add" not "added" or "adds")
4. No period at the end of subject line
5. Capitalize first letter of description
6. Be specific and descriptive

CONVENTIONAL COMMIT TYPES:
- feat: New feature for the user
- fix: Bug fix for the user  
- docs: Documentation changes
- style: Code formatting (no logic change)
- refactor: Code restructuring without behavior change
- test: Adding or updating tests
- chore: Maintenance tasks, dependencies, tooling
- perf: Performance improvements
- ci: CI/CD related changes
- build: Build system or dependencies

SCOPE EXAMPLES:
- Component/module name: auth, api, ui, cli, db
- Area of change: parser, config, utils, cmd
- Specific feature: navigation, bookmarks, ai

ANALYSIS APPROACH:
1. Identify the primary change from the diff
2. Determine if it's a breaking change (use ! after type/scope)
3. Choose the most appropriate type and scope
4. Focus on user impact, not implementation details
5. If multiple changes, prioritize the most significant

EXAMPLES:
- feat(bookmarks): add fuzzy search functionality
- fix(db): handle SQLite connection timeout gracefully  
- docs(readme): update installation instructions for Windows
- refactor(cmd): extract common validation logic
- chore(deps): update Go modules to latest versions

Generate ONE concise commit message. Do not include body or footer unless it's a breaking change.`

	prompt := fmt.Sprintf("Generate a commit message for these changes:\n\n%s", diff)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	return c.chat(ctx, messages)
}

// ExplainCode explains a piece of code.
func (c *Client) ExplainCode(ctx context.Context, code string) (string, error) {
	systemPrompt := `You are an expert code analysis assistant specializing in clear, educational explanations for developers of all skill levels.

EXPLANATION STRUCTURE:
1. **Overview**: What does this code do? (1-2 sentences)
2. **Key Components**: Break down important parts/functions/patterns
3. **Flow**: How does execution proceed?
4. **Important Details**: Gotchas, edge cases, or notable patterns
5. **Context**: Where/how this might be used
6. **Improvements**: Suggestions for optimization, readability, or best practices (if applicable)

EXPLANATION GUIDELINES:
- Start with the big picture, then dive into details
- Use clear, accessible language while maintaining technical accuracy
- Highlight design patterns and programming concepts
- Explain the "why" behind implementation choices
- Point out potential issues or limitations
- Use bullet points or numbered lists for clarity
- Include relevant terminology definitions when helpful
- Consider performance, security, and maintainability aspects

LANGUAGE-SPECIFIC FOCUS:
- Go: Goroutines, interfaces, error handling, packages, memory management
- JavaScript: Async/await, closures, prototypes, event handling
- Python: List comprehensions, decorators, context managers, generators
- SQL: Query optimization, indexes, joins, transactions
- Shell: Pipes, redirects, variable expansion, error handling

FORMAT:
Use markdown formatting with headers, code blocks, and emphasis where appropriate.`

	prompt := fmt.Sprintf("Explain this code:\n\n%s", code)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	return c.chat(ctx, messages)
}

// SuggestCommands suggests shell commands based on user intent and current context.
func (c *Client) SuggestCommands(ctx context.Context, intent string, workingDir string, contextInfo map[string]interface{}) (string, error) {
	systemPrompt := fmt.Sprintf(`You are Aura's command suggestion engine. Generate practical, safe shell commands based on user intent and current context.

SYSTEM INFO:
- OS: %s
- Architecture: %s
- Working Directory: %s

CONTEXT ANALYSIS:
Consider the provided context information to suggest relevant commands:
- Git repository status
- File types present
- Project structure
- Available tools and dependencies

COMMAND SUGGESTION RULES:
1. Prioritize safety - avoid destructive operations without warnings
2. Use platform-appropriate commands for %s
3. Provide alternatives when multiple approaches exist
4. Include brief explanations for complex commands
5. Consider the current working directory
6. Suggest incremental steps for complex tasks

OUTPUT FORMAT:
Return 3-5 relevant commands in this format:
1. [command] - [brief description]
2. [command] - [brief description]
...

SAFETY GUIDELINES:
- Warn about destructive operations (rm, del, format, etc.)
- Suggest dry-run options when available
- Recommend backups for risky operations
- Use relative paths when appropriate
- Include error checking in scripts`, runtime.GOOS, runtime.GOARCH, workingDir, runtime.GOOS)

	var contextStr string
	if len(contextInfo) > 0 {
		contextBytes, _ := json.MarshalIndent(contextInfo, "", "  ")
		contextStr = fmt.Sprintf("\n\nCurrent Context:\n%s", string(contextBytes))
	}

	prompt := fmt.Sprintf("User intent: %s%s", intent, contextStr)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	return c.chat(ctx, messages)
}

// chat sends a chat request to the API and returns the response.
func (c *Client) chat(ctx context.Context, messages []Message) (string, error) {
	model := os.Getenv("AURA_MODEL")
	if model == "" {
		model = "gpt-3.5-turbo" // Use a more standard, widely available model
	}

	request := ChatRequest{
		Model:       model,
		Messages:    messages,
		Temperature: 0.7,
		MaxTokens:   1000,
	}

	requestBody, err := json.Marshal(request)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/chat/completions", bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var response ChatResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return "", fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if response.Error != nil {
		return "", fmt.Errorf("API error: %s", response.Error.Message)
	}

	if len(response.Choices) == 0 {
		return "", fmt.Errorf("no response from API")
	}

	return response.Choices[0].Message.Content, nil
}

// DebugIssue helps debug errors and issues with context-aware suggestions.
func (c *Client) DebugIssue(ctx context.Context, errorMsg string, commandRun string, environment map[string]string) (string, error) {
	systemPrompt := fmt.Sprintf(`You are Aura's debugging assistant. Help users understand and resolve technical issues with actionable solutions.

SYSTEM INFO:
- OS: %s
- Architecture: %s

DEBUGGING APPROACH:
1. Analyze the error message for root cause
2. Consider the command that was run
3. Review environment context
4. Provide step-by-step resolution
5. Suggest prevention strategies

RESPONSE STRUCTURE:
## Problem Analysis
- Root cause explanation
- Why this error occurred

## Immediate Solution
- Step-by-step fix instructions
- Platform-specific commands for %s

## Alternative Approaches
- Different ways to achieve the same goal
- Workarounds if main solution fails

## Prevention
- How to avoid this issue in the future
- Best practices

GUIDELINES:
- Prioritize the most likely solutions first
- Include verification steps
- Explain why each solution works
- Consider security implications
- Provide context for beginners`, runtime.GOOS, runtime.GOARCH, runtime.GOOS)

	var envStr string
	if len(environment) > 0 {
		envBytes, _ := json.MarshalIndent(environment, "", "  ")
		envStr = fmt.Sprintf("\n\nEnvironment:\n%s", string(envBytes))
	}

	prompt := fmt.Sprintf(`Error occurred while running: %s

Error message: %s%s

Please help me understand and fix this issue.`, commandRun, errorMsg, envStr)

	messages := []Message{
		{Role: "system", Content: systemPrompt},
		{Role: "user", Content: prompt},
	}

	return c.chat(ctx, messages)
}
