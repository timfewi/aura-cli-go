package ai

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// Save original environment
	originalKey := os.Getenv("AURA_API_KEY")
	originalOpenAIKey := os.Getenv("OPENAI_API_KEY")
	defer func() {
		os.Setenv("AURA_API_KEY", originalKey)
		os.Setenv("OPENAI_API_KEY", originalOpenAIKey)
	}()

	// Test without API key
	os.Unsetenv("AURA_API_KEY")
	os.Unsetenv("OPENAI_API_KEY")

	_, err := NewClient()
	if err == nil {
		t.Error("Expected API key error, got nil")
	}

	// Check that error message matches expected format
	expectedMsg := "AURA_API_KEY or OPENAI_API_KEY environment variable is required"
	if err.Error() != expectedMsg {
		t.Errorf("Expected API key error, got: %s", err.Error())
	}

	// Test with AURA_API_KEY
	os.Setenv("AURA_API_KEY", "sk-test-key-123")

	client, err := NewClient()
	if err != nil {
		t.Errorf("Unexpected error with AURA_API_KEY: %v", err)
	}
	if client == nil {
		t.Error("Client is nil")
		return
	}
	if client.apiKey != "sk-test-key-123" {
		t.Errorf("API key = %v, want sk-test-key-123", client.apiKey)
	}

	// Test with OPENAI_API_KEY (fallback)
	os.Unsetenv("AURA_API_KEY")
	os.Setenv("OPENAI_API_KEY", "sk-openai-key-456")

	client, err = NewClient()
	if err != nil {
		t.Errorf("Unexpected error with OPENAI_API_KEY: %v", err)
	}
	if client.apiKey != "sk-openai-key-456" {
		t.Errorf("API key = %v, want sk-openai-key-456", client.apiKey)
	}
}

func TestClientAsk(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and headers
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		auth := r.Header.Get("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			t.Errorf("Expected Bearer token, got %s", auth)
		}

		contentType := r.Header.Get("Content-Type")
		if contentType != "application/json" {
			t.Errorf("Expected application/json content type, got %s", contentType)
		}

		// Return mock response
		response := `{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "This is a test response from the AI."
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
		
	}))
	defer server.Close()

	// Create client with mock server
	client := &Client{
		apiKey:  "sk-test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	response, err := client.Ask(ctx, "Hello, how are you?")

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	expected := "This is a test response from the AI."
	if response != expected {
		t.Errorf("Response = %v, want %v", response, expected)
	}
}

func TestClientAskWithError(t *testing.T) {
	// Create a mock server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"error": {
				"message": "Invalid API key",
				"type": "authentication_error"
			}
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "sk-invalid-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	_, err := client.Ask(ctx, "Hello")

	if err == nil {
		t.Error("Expected error for invalid API key")
	}

	if !strings.Contains(err.Error(), "Invalid API key") {
		t.Errorf("Expected authentication error, got: %v", err)
	}
}

func TestClientGenerateCommitMessage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "feat: add user authentication system\n\n- Implement login/logout functionality\n- Add password hashing\n- Create user session management"
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "sk-test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	testDiff := `diff --git a/auth.go b/auth.go
new file mode 100644
index 0000000..1234567
--- /dev/null
+++ b/auth.go
@@ -0,0 +1,10 @@
+package main
+
+func login(username, password string) bool {
+    // TODO: implement login logic
+    return false
+}`

	ctx := context.Background()
	message, err := client.GenerateCommitMessage(ctx, testDiff)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(message, "feat:") {
		t.Errorf("Expected conventional commit format, got: %v", message)
	}

	if !strings.Contains(message, "authentication") {
		t.Errorf("Expected commit message about authentication, got: %v", message)
	}
}

func TestClientExplainCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "This Go function implements a simple HTTP server that listens on port 8080 and responds with 'Hello, World!' to all requests."
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "sk-test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	testCode := `package main
import (
    "fmt"
    "net/http"
)

func main() {
    http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
        fmt.Fprintf(w, "Hello, World!")
    })
    http.ListenAndServe(":8080", nil)
}`

	ctx := context.Background()
	explanation, err := client.ExplainCode(ctx, testCode)

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(explanation, "HTTP server") {
		t.Errorf("Expected explanation about HTTP server, got: %v", explanation)
	}
}

func TestClientSuggestCommands(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := `{
			"choices": [
				{
					"message": {
						"role": "assistant",
						"content": "Here are some commands to find large files:\n\n1. find . -type f -size +100M -exec ls -lh {} \\;\n2. du -h . | sort -hr | head -20\n3. ncdu ."
					}
				}
			]
		}`

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(response)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "sk-test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	suggestions, err := client.SuggestCommands(ctx, "find large files", "/home/user", map[string]interface{}{
		"os": "linux",
	})

	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}

	if !strings.Contains(suggestions, "find") || !strings.Contains(suggestions, "du") {
		t.Errorf("Expected file finding commands, got: %v", suggestions)
	}
}

func TestClientWithTimeout(t *testing.T) {
	// Create a server that delays response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(2 * time.Second) // Delay longer than timeout
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`{"choices":[{"message":{"content":"response"}}]}`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "sk-test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	// Create context with short timeout
	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	_, err := client.Ask(ctx, "Hello")

	if err == nil {
		t.Error("Expected timeout error")
	}

	if !strings.Contains(err.Error(), "context deadline exceeded") {
		t.Errorf("Expected timeout error, got: %v", err)
	}
}

func TestClientWithInvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte(`invalid json response`)); err != nil {
			t.Errorf("Failed to write response: %v", err)
		}
	}))
	defer server.Close()

	client := &Client{
		apiKey:  "sk-test-key",
		baseURL: server.URL,
		client:  &http.Client{Timeout: 30 * time.Second},
	}

	ctx := context.Background()
	_, err := client.Ask(ctx, "Hello")

	if err == nil {
		t.Error("Expected JSON parsing error")
	}

	if !strings.Contains(err.Error(), "json") && !strings.Contains(err.Error(), "unmarshal") {
		t.Errorf("Expected JSON error, got: %v", err)
	}
}
