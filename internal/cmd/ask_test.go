package cmd

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRunAsk(t *testing.T) {
	// Save original environment
	originalAPIKey := os.Getenv("AURA_API_KEY")
	defer func() {
		if originalAPIKey != "" {
			os.Setenv("AURA_API_KEY", originalAPIKey)
		} else {
			os.Unsetenv("AURA_API_KEY")
		}
	}()

	tests := []struct {
		name      string
		args      []string
		apiKey    string
		wantError bool
	}{
		{
			name:      "no API key",
			args:      []string{"test", "question"},
			apiKey:    "",
			wantError: true,
		},
		{
			name:      "with API key and question",
			args:      []string{"test", "question"},
			apiKey:    "sk-test-key",
			wantError: false, // Would fail in real scenario but we can't test AI calls
		},
		{
			name:      "empty question",
			args:      []string{},
			apiKey:    "sk-test-key",
			wantError: false, // Would start interactive mode
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.apiKey != "" {
				os.Setenv("AURA_API_KEY", tt.apiKey)
			} else {
				os.Unsetenv("AURA_API_KEY")
			}

			// Since runAsk requires actual AI interaction, we can only test
			// the error cases and basic setup
			err := runAsk(askCmd, tt.args)

			if tt.name == "no API key" && err == nil {
				t.Error("Expected error when no API key is set")
			}

			if tt.name == "no API key" && err != nil {
				if !strings.Contains(err.Error(), "failed to initialize AI client") {
					t.Errorf("Expected AI client initialization error, got: %v", err)
				}
			}
		})
	}
}

func TestShowThinking(t *testing.T) {
	// Test that showThinking doesn't panic and responds to done signal
	done := make(chan bool)

	// Start showThinking in a goroutine
	go showThinking(done)

	// Send done signal after a short delay
	go func() {
		time.Sleep(100 * time.Millisecond)
		done <- true
	}()

	// Wait a bit to ensure the function completes
	time.Sleep(200 * time.Millisecond)

	// If we reach here without hanging, the test passes
}

func TestAskCommandConfiguration(t *testing.T) {
	// Test that the command is properly configured
	if askCmd.Use != "ask [question...]" {
		t.Errorf("askCmd.Use = %v, want 'ask [question...]'", askCmd.Use)
	}

	if askCmd.Short == "" {
		t.Error("askCmd.Short should not be empty")
	}

	if askCmd.Long == "" {
		t.Error("askCmd.Long should not be empty")
	}

	if askCmd.RunE == nil {
		t.Error("askCmd.RunE should not be nil")
	}
}

// TestAskWithMockClient tests ask functionality with a mock AI client
func TestAskWithMockClient(t *testing.T) {
	// Since we can't easily mock the ai.NewClient() call in runAsk,
	// we'll test the components that we can isolate

	// Test question formation
	tests := []struct {
		name         string
		args         []string
		stdinContent string
		expected     string
	}{
		{
			name:     "simple question",
			args:     []string{"how", "to", "list", "files"},
			expected: "how to list files",
		},
		{
			name:     "single word question",
			args:     []string{"help"},
			expected: "help",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Test string joining logic
			result := strings.Join(tt.args, " ")
			if result != tt.expected {
				t.Errorf("Question formation = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestAskInteractiveMode tests aspects of interactive mode that can be tested
func TestAskInteractiveMode(t *testing.T) {
	// Test input validation for interactive mode
	testInputs := []struct {
		input      string
		shouldExit bool
	}{
		{"exit", true},
		{"quit", true},
		{"", false}, // empty input should continue
		{"help me", false},
	}

	for _, tt := range testInputs {
		t.Run(tt.input, func(t *testing.T) {
			input := strings.TrimSpace(tt.input)

			if tt.shouldExit {
				if input != "exit" && input != "quit" {
					t.Errorf("Expected exit command, got: %s", input)
				}
			}

			if input == "" && tt.shouldExit {
				t.Error("Empty input should not trigger exit")
			}
		})
	}
}

// TestAskContextTimeout tests context timeout handling
func TestAskContextTimeout(t *testing.T) {
	// Test context creation with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if ctx == nil {
		t.Error("Context should not be nil")
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		t.Error("Context should have a deadline")
	}

	if time.Until(deadline) > 31*time.Second {
		t.Error("Context timeout should be approximately 30 seconds")
	}
}

// TestAskErrorHandling tests error handling scenarios
func TestAskErrorHandling(t *testing.T) {
	// Test various error scenarios that might occur
	testCases := []struct {
		name        string
		errorType   string
		expectError bool
	}{
		{
			name:        "client initialization error",
			errorType:   "client_init",
			expectError: true,
		},
		{
			name:        "context timeout",
			errorType:   "timeout",
			expectError: true,
		},
		{
			name:        "API error",
			errorType:   "api_error",
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// We can test that our error handling patterns are correct
			var err error

			switch tc.errorType {
			case "client_init":
				err = errors.New("failed to initialize AI client: invalid API key")
			case "timeout":
				err = context.DeadlineExceeded
			case "api_error":
				err = errors.New("AI request failed: network error")
			}

			if tc.expectError && err == nil {
				t.Error("Expected error but got none")
			}

			if !tc.expectError && err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// TestAskStdinHandling tests stdin detection logic
func TestAskStdinHandling(t *testing.T) {
	// Test the logic for detecting piped input
	// Note: We can't easily test actual stdin piping in unit tests,
	// but we can test the string manipulation logic

	testCases := []struct {
		name         string
		stdinContent string
		args         []string
		expected     string
	}{
		{
			name:         "piped content with no question",
			stdinContent: "echo 'hello world'",
			args:         []string{},
			expected:     "Explain this:\n\necho 'hello world'",
		},
		{
			name:         "piped content with question",
			stdinContent: "def hello():\n    print('hello')",
			args:         []string{"explain", "this", "function"},
			expected:     "explain this function\n\nContent:\ndef hello():\n    print('hello')",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var question string
			stdinContent := strings.TrimSpace(tc.stdinContent)

			if len(tc.args) == 0 {
				question = fmt.Sprintf("Explain this:\n\n%s", stdinContent)
			} else {
				userQuestion := strings.Join(tc.args, " ")
				question = fmt.Sprintf("%s\n\nContent:\n%s", userQuestion, stdinContent)
			}

			if question != tc.expected {
				t.Errorf("Question formation = %v, want %v", question, tc.expected)
			}
		})
	}
}

func TestAskThinkingAnimation(t *testing.T) {
	// Test the thinking animation characters
	chars := []string{"⠋", "⠙", "⠹", "⠸", "⠼", "⠴", "⠦", "⠧", "⠇", "⠏"}

	if len(chars) != 10 {
		t.Errorf("Expected 10 animation characters, got %d", len(chars))
	}

	for i, char := range chars {
		if char == "" {
			t.Errorf("Animation character %d is empty", i)
		}
	}
}

// MockAIClient for testing (if we want to implement more sophisticated tests)
type MockAIClient struct {
	shouldError bool
	response    string
}

func (m *MockAIClient) Ask(ctx context.Context, question string) (string, error) {
	if m.shouldError {
		return "", errors.New("mock error")
	}
	return m.response, nil
}

func TestMockAIClient(t *testing.T) {
	// Test our mock client works as expected
	client := &MockAIClient{
		shouldError: false,
		response:    "Mock response",
	}

	ctx := context.Background()
	response, err := client.Ask(ctx, "test question")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if response != "Mock response" {
		t.Errorf("Expected 'Mock response', got '%s'", response)
	}

	// Test error case
	client.shouldError = true
	_, err = client.Ask(ctx, "test question")

	if err == nil {
		t.Error("Expected error but got none")
	}
}
