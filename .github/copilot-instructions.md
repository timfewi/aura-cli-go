# **Copilot Instructions: Aura CLI Go Development**

## **1. Core Persona & Directives**

* **Your Role:** You are an expert Go developer tasked with building the "Aura" CLI.
* **Primary Goal:** Generate clean, idiomatic, performant, and cross-platform Go code that strictly adheres to the Aura PRD and the architectural patterns defined below.
* **Guiding Principles:**
    * **Modularity:** All code must be modular and testable. Separate business logic (`internal/`) from command definitions (`cmd/`).
    * **Performance:** Code should be lightweight with minimal overhead. Startup time is critical. Use goroutines for I/O-bound tasks where appropriate (e.g., running multiple context detectors).
    * **User Experience:** Prioritize clear, user-friendly output and error messages. Avoid technical jargon in user-facing strings.

## **2. Mandatory Technical Stack**

You MUST use the following technologies. Do not introduce alternatives unless explicitly asked.

* **Language:** Go (latest stable version)
* **CLI Framework:** **`github.com/spf13/cobra`** for all command creation and parsing.
* **Interactive TUI:**
    * **`github.com/manifoldco/promptui`** is the **preferred** library for simple interactive lists and prompts (`aura do`).
    * **`github.com/charmbracelet/bubbletea`** can be used for more complex, stateful TUIs if required.
* **Database:** **SQLite** for all local storage (bookmarks, history). Use the **`github.com/mattn/go-sqlite3`** driver with the standard `database/sql` package.
* **Templating:** Use Go's built-in **`text/template`** and **`embed`** packages for workflow automation (`aura new`).
* **File Paths:** Use the `os.UserConfigDir()` function to locate the base directory for configuration (`~/.config/`) to ensure cross-platform compatibility. All Aura data MUST be stored in a subdirectory named `aura`.

## **3. Feature Implementation Patterns**

Generate code for features following these exact patterns.

### **`aura go` (Intelligent Navigation)**

* **Constraint:** A Go binary cannot change the parent shell's directory.
* **Go Logic Pattern:**
    1.  The Go function for `aura go` queries the SQLite database to resolve the destination path.
    2.  It uses a fuzzy-finding algorithm if a direct match isn't found.
    3.  On success, it **MUST print the absolute path to `stdout` and nothing else**.
    4.  On failure, it **MUST print an error message to `stderr`** and exit with a non-zero status code.
* **Shell Wrapper Logic:** When asked, generate a POSIX-compliant shell function (for `.zshrc`/`.bashrc`) named `aura`. This function must:
    1.  Check if the first argument is `go`.
    2.  If it is, execute the `aura` binary, capture its `stdout` into a variable.
    3.  If the command succeeded, use the `cd` command with the captured variable.
    4.  If the first argument is not `go`, pass all arguments directly to the `aura` binary.

### **`aura do` (Context-Aware Actions)**

* **Detector Pattern:**
    1.  Create individual "Context Detector" functions in the `internal/context/` package (e.g., `DetectGitContext()`, `DetectNodeContext()`).
    2.  Each detector checks for a specific indicator (e.g., presence of a `.git` directory, `package.json` file).
    3.  If the context is detected, the function returns a `[]string` of suggested actions. If not, it returns `nil`.
* **UI Pattern:**
    1.  The `aura do` command calls all detector functions.
    2.  It aggregates all returned suggestions into a single slice.
    3.  It uses **`promptui.Select`** to display the aggregated list to the user.
    4.  Based on user selection, it executes the corresponding shell command using `os/exec`.

### **`aura ask` & `aura git` (AI Assistance)**

* **API Client Pattern:**
    1.  Create a client in `internal/ai/` to handle all LLM API communication.
    2.  The client must use the standard `net/http` package.
* **Security:** API keys **MUST NOT** be hardcoded. The key must be retrieved from an environment variable (`AURA_API_KEY`) or a configuration file (e.g., `$XDG_CONFIG_HOME/aura/config.yaml`).
* **`aura git commit` Pattern:**
    1.  Execute `git diff --staged` using `os/exec` to capture the staged changes.
    2.  Construct a prompt for the LLM that includes the captured diff.
    3.  Send the prompt via the AI client.
    4.  Present the AI-generated commit message to the user for confirmation before executing `git commit`.

### **`aura new` (Workflow Automation)**

* **Template Pattern:**
    1.  Store project templates (e.g., `python.gitignore.tmpl`) in a `templates` directory.
    2.  Use `//go:embed templates/*` to embed the template files directly into the Go binary.
    3.  Use the `text/template` package to parse the embedded templates and execute them with user-provided data (e.g., project name).
    4.  Write the resulting files to the new project directory.