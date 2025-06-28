# Go Architecture Best Practices

## Project Structure

### Standard Go Project Layout

```txt
project/
├── cmd/
│   └── app/
│       └── main.go
├── internal/
│   ├── config/
│   ├── handlers/
│   ├── models/
│   ├── repositories/
│   └── services/
├── pkg/
│   └── utils/
├── api/
│   └── openapi/
├── web/
│   ├── static/
│   └── templates/
├── scripts/
├── deployments/
├── test/
├── docs/
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Directory Conventions

- **`cmd/`**: Main applications for this project
- **`internal/`**: Private application and library code
- **`pkg/`**: Library code that's ok to use by external applications
- **`api/`**: API contract files (OpenAPI/Swagger specs, protocol definition files)
- **`web/`**: Web application specific components
- **`scripts/`**: Scripts to perform various build, install, analysis, etc operations
- **`test/`**: Additional external test apps and test data

## Design Principles

### Go-Specific Principles

- **Simplicity**: Prefer simple, explicit code over clever abstractions
- **Composition over Inheritance**: Use interfaces and struct embedding
- **Explicit Error Handling**: Handle errors explicitly, don't ignore them
- **Concurrency**: Use goroutines and channels for concurrent operations
- **Minimal Dependencies**: Prefer standard library over external packages

### SOLID Principles in Go

```go
// Single Responsibility Principle
type UserRepository interface {
    GetUser(id string) (*User, error)
    SaveUser(user *User) error
}

type UserService struct {
    repo UserRepository
}

// Open/Closed Principle - extensible through interfaces
type PaymentProcessor interface {
    ProcessPayment(amount float64) error
}

type CreditCardProcessor struct{}
type PayPalProcessor struct{}

func (c *CreditCardProcessor) ProcessPayment(amount float64) error {
    // Implementation
    return nil
}

// Interface Segregation - small, focused interfaces
type Reader interface {
    Read([]byte) (int, error)
}

type Writer interface {
    Write([]byte) (int, error)
}

// Dependency Inversion - depend on interfaces
type OrderService struct {
    paymentProcessor PaymentProcessor
    userRepo        UserRepository
}
```

## Code Organization

### Package Design

```go
// Good: Clear package structure with focused responsibilities
package user

import (
    "context"
    "time"
)

// User represents a user entity
type User struct {
    ID        string    `json:"id"`
    Email     string    `json:"email"`
    CreatedAt time.Time `json:"created_at"`
}

// Repository defines user data access operations
type Repository interface {
    GetByID(ctx context.Context, id string) (*User, error)
    Create(ctx context.Context, user *User) error
    Update(ctx context.Context, user *User) error
    Delete(ctx context.Context, id string) error
}

// Service handles user business logic
type Service struct {
    repo Repository
}

func NewService(repo Repository) *Service {
    return &Service{repo: repo}
}
```

### Configuration Management

```go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DatabaseConfig struct {
    Host     string
    Port     int
    User     string
    Password string
    Name     string
}

func Load() (*Config, error) {
    cfg := &Config{
        Server: ServerConfig{
            Port:         getEnvOrDefault("SERVER_PORT", "8080"),
            ReadTimeout:  parseDurationOrDefault("SERVER_READ_TIMEOUT", 30*time.Second),
            WriteTimeout: parseDurationOrDefault("SERVER_WRITE_TIMEOUT", 30*time.Second),
        },
        Database: DatabaseConfig{
            Host:     getEnvOrDefault("DB_HOST", "localhost"),
            Port:     parseIntOrDefault("DB_PORT", 5432),
            User:     getEnvOrDefault("DB_USER", "postgres"),
            Password: os.Getenv("DB_PASSWORD"),
            Name:     getEnvOrDefault("DB_NAME", "myapp"),
        },
    }

    return cfg, nil
}

func getEnvOrDefault(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## Error Handling

### Go Error Patterns

```go
package errors

import (
    "errors"
    "fmt"
)

// Custom error types
type ValidationError struct {
    Field   string
    Message string
}

func (e *ValidationError) Error() string {
    return fmt.Sprintf("validation error: %s - %s", e.Field, e.Message)
}

type NotFoundError struct {
    Resource string
    ID       string
}

func (e *NotFoundError) Error() string {
    return fmt.Sprintf("%s with ID %s not found", e.Resource, e.ID)
}

// Error wrapping for context
func ProcessUser(id string) error {
    user, err := getUserFromDB(id)
    if err != nil {
        return fmt.Errorf("failed to get user %s: %w", id, err)
    }

    if err := validateUser(user); err != nil {
        return fmt.Errorf("user validation failed: %w", err)
    }

    return nil
}

// Sentinel errors for specific conditions
var (
    ErrUserNotFound     = errors.New("user not found")
    ErrInvalidCredentials = errors.New("invalid credentials")
    ErrDuplicateEmail   = errors.New("email already exists")
)
```

### Error Handling Best Practices

```go
// Good: Explicit error handling
func CreateUser(ctx context.Context, user *User) error {
    if err := validateUser(user); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    if err := s.repo.Create(ctx, user); err != nil {
        if errors.Is(err, ErrDuplicateEmail) {
            return &ValidationError{
                Field:   "email",
                Message: "email already exists",
            }
        }
        return fmt.Errorf("failed to create user: %w", err)
    }

    return nil
}

// Error checking with specific handling
func HandleUserCreation(w http.ResponseWriter, r *http.Request) {
    var user User
    if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
        http.Error(w, "invalid JSON", http.StatusBadRequest)
        return
    }

    if err := userService.CreateUser(r.Context(), &user); err != nil {
        var validationErr *ValidationError
        if errors.As(err, &validationErr) {
            http.Error(w, validationErr.Error(), http.StatusBadRequest)
            return
        }
        http.Error(w, "internal server error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
}
```

## Testing Strategy

### Unit Testing

```go
package user_test

import (
    "context"
    "errors"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"

    "myproject/internal/user"
)

// Mock repository
type MockRepository struct {
    mock.Mock
}

func (m *MockRepository) GetByID(ctx context.Context, id string) (*user.User, error) {
    args := m.Called(ctx, id)
    return args.Get(0).(*user.User), args.Error(1)
}

func (m *MockRepository) Create(ctx context.Context, u *user.User) error {
    args := m.Called(ctx, u)
    return args.Error(0)
}

func TestUserService_CreateUser(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    service := user.NewService(mockRepo)

    testUser := &user.User{
        Email: "test@example.com",
    }

    mockRepo.On("Create", mock.Anything, testUser).Return(nil)

    // Act
    err := service.CreateUser(context.Background(), testUser)

    // Assert
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}

func TestUserService_CreateUser_ValidationError(t *testing.T) {
    // Arrange
    mockRepo := new(MockRepository)
    service := user.NewService(mockRepo)

    invalidUser := &user.User{
        Email: "", // Invalid email
    }

    // Act
    err := service.CreateUser(context.Background(), invalidUser)

    // Assert
    assert.Error(t, err)
    var validationErr *user.ValidationError
    assert.True(t, errors.As(err, &validationErr))
}
```

### Integration Testing

```go
package integration_test

import (
    "context"
    "database/sql"
    "testing"

    _ "github.com/lib/pq"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/suite"

    "myproject/internal/user"
)

type UserIntegrationSuite struct {
    suite.Suite
    db          *sql.DB
    userRepo    user.Repository
    userService *user.Service
}

func (suite *UserIntegrationSuite) SetupSuite() {
    // Setup test database
    db, err := sql.Open("postgres", "postgres://test:test@localhost/testdb?sslmode=disable")
    suite.Require().NoError(err)

    suite.db = db
    suite.userRepo = user.NewPostgresRepository(db)
    suite.userService = user.NewService(suite.userRepo)
}

func (suite *UserIntegrationSuite) TearDownSuite() {
    suite.db.Close()
}

func (suite *UserIntegrationSuite) SetupTest() {
    // Clean up test data
    _, err := suite.db.Exec("DELETE FROM users")
    suite.Require().NoError(err)
}

func (suite *UserIntegrationSuite) TestCreateAndGetUser() {
    // Arrange
    ctx := context.Background()
    testUser := &user.User{
        Email: "integration@test.com",
    }

    // Act
    err := suite.userService.CreateUser(ctx, testUser)
    suite.Require().NoError(err)

    retrievedUser, err := suite.userRepo.GetByID(ctx, testUser.ID)

    // Assert
    suite.NoError(err)
    suite.Equal(testUser.Email, retrievedUser.Email)
}

func TestUserIntegrationSuite(t *testing.T) {
    suite.Run(t, new(UserIntegrationSuite))
}
```

## Concurrency Patterns

### Worker Pool Pattern

```go
package worker

import (
    "context"
    "sync"
)

type Job interface {
    Execute(ctx context.Context) error
}

type Pool struct {
    workers    int
    jobQueue   chan Job
    resultChan chan error
    wg         sync.WaitGroup
}

func NewPool(workers int, queueSize int) *Pool {
    return &Pool{
        workers:    workers,
        jobQueue:   make(chan Job, queueSize),
        resultChan: make(chan error, queueSize),
    }
}

func (p *Pool) Start(ctx context.Context) {
    for i := 0; i < p.workers; i++ {
        p.wg.Add(1)
        go p.worker(ctx)
    }
}

func (p *Pool) worker(ctx context.Context) {
    defer p.wg.Done()

    for {
        select {
        case job, ok := <-p.jobQueue:
            if !ok {
                return
            }
            err := job.Execute(ctx)
            p.resultChan <- err
        case <-ctx.Done():
            return
        }
    }
}

func (p *Pool) Submit(job Job) {
    p.jobQueue <- job
}

func (p *Pool) Stop() {
    close(p.jobQueue)
    p.wg.Wait()
    close(p.resultChan)
}
```

### Fan-out/Fan-in Pattern

```go
func ProcessDataConcurrently(ctx context.Context, data []string) ([]Result, error) {
    // Fan-out: distribute work to multiple goroutines
    resultChan := make(chan Result, len(data))
    errorChan := make(chan error, len(data))

    var wg sync.WaitGroup
    for _, item := range data {
        wg.Add(1)
        go func(item string) {
            defer wg.Done()
            result, err := processItem(ctx, item)
            if err != nil {
                errorChan <- err
                return
            }
            resultChan <- result
        }(item)
    }

    // Close channels when all goroutines complete
    go func() {
        wg.Wait()
        close(resultChan)
        close(errorChan)
    }()

    // Fan-in: collect results
    var results []Result
    var errors []error

    for {
        select {
        case result, ok := <-resultChan:
            if !ok {
                resultChan = nil
            } else {
                results = append(results, result)
            }
        case err, ok := <-errorChan:
            if !ok {
                errorChan = nil
            } else {
                errors = append(errors, err)
            }
        }

        if resultChan == nil && errorChan == nil {
            break
        }
    }

    if len(errors) > 0 {
        return nil, fmt.Errorf("processing errors: %v", errors)
    }

    return results, nil
}
```

## Performance Considerations

### Memory Management

```go
// Use object pools for frequently allocated objects
var bufferPool = sync.Pool{
    New: func() interface{} {
        return make([]byte, 0, 1024)
    },
}

func ProcessData(data []byte) []byte {
    // Get buffer from pool
    buf := bufferPool.Get().([]byte)
    defer bufferPool.Put(buf[:0]) // Reset and return to pool

    // Use buffer for processing
    buf = append(buf, data...)
    // ... process data
    
    return buf
}

// Preallocate slices when size is known
func ProcessItems(items []string) []Result {
    results := make([]Result, 0, len(items)) // Preallocate capacity
    for _, item := range items {
        result := processItem(item)
        results = append(results, result)
    }
    return results
}
```

### Profiling and Monitoring

```go
package main

import (
    "context"
    "log"
    "net/http"
    _ "net/http/pprof" // Enable pprof endpoints
    "time"

    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
    "github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
    requestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "Duration of HTTP requests.",
        },
        []string{"method", "endpoint"},
    )

    requestCounter = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests.",
        },
        []string{"method", "endpoint", "status"},
    )
)

func instrumentHandler(handler http.HandlerFunc, endpoint string) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        
        // Wrap ResponseWriter to capture status code
        ww := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
        
        handler(ww, r)
        
        duration := time.Since(start).Seconds()
        requestDuration.WithLabelValues(r.Method, endpoint).Observe(duration)
        requestCounter.WithLabelValues(r.Method, endpoint, http.StatusText(ww.statusCode)).Inc()
    }
}
```

## Documentation

### Code Documentation

```go
// Package user provides user management functionality.
// It includes user creation, retrieval, and validation operations.
package user

import (
    "context"
    "fmt"
    "regexp"
    "time"
)

// User represents a user in the system.
// It contains basic user information and timestamps.
type User struct {
    // ID is the unique identifier for the user
    ID string `json:"id" db:"id"`
    
    // Email is the user's email address and must be unique
    Email string `json:"email" db:"email"`
    
    // CreatedAt is the timestamp when the user was created
    CreatedAt time.Time `json:"created_at" db:"created_at"`
    
    // UpdatedAt is the timestamp when the user was last updated
    UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// Repository defines the interface for user data persistence.
// Implementations should handle database operations and error conditions.
type Repository interface {
    // GetByID retrieves a user by their unique identifier.
    // Returns ErrUserNotFound if the user doesn't exist.
    GetByID(ctx context.Context, id string) (*User, error)
    
    // Create persists a new user to the database.
    // Returns ErrDuplicateEmail if email already exists.
    Create(ctx context.Context, user *User) error
    
    // Update modifies an existing user in the database.
    // Returns ErrUserNotFound if the user doesn't exist.
    Update(ctx context.Context, user *User) error
    
    // Delete removes a user from the database.
    // Returns ErrUserNotFound if the user doesn't exist.
    Delete(ctx context.Context, id string) error
}

// Service handles user business logic and validation.
// It coordinates between the repository layer and external requests.
type Service struct {
    repo Repository
}

// NewService creates a new user service with the provided repository.
// The repository must not be nil.
func NewService(repo Repository) *Service {
    if repo == nil {
        panic("repository cannot be nil")
    }
    return &Service{repo: repo}
}

// CreateUser validates and creates a new user.
// It performs email validation and checks for duplicates.
//
// Example:
//   user := &User{Email: "john@example.com"}
//   err := service.CreateUser(ctx, user)
//   if err != nil {
//       // handle error
//   }
func (s *Service) CreateUser(ctx context.Context, user *User) error {
    if user == nil {
        return fmt.Errorf("user cannot be nil")
    }

    if err := s.validateUser(user); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }

    user.CreatedAt = time.Now()
    user.UpdatedAt = time.Now()

    if err := s.repo.Create(ctx, user); err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }

    return nil
}
```

### API Documentation

```go
// Package handlers provides HTTP handlers for the user API.
//
// The API follows RESTful conventions:
//   GET    /users/{id}     - Get user by ID
//   POST   /users          - Create new user
//   PUT    /users/{id}     - Update user
//   DELETE /users/{id}     - Delete user
//
// All endpoints return JSON responses and follow standard HTTP status codes:
//   200 - Success
//   201 - Created
//   400 - Bad Request (validation errors)
//   404 - Not Found
//   500 - Internal Server Error
package handlers
```

## Dependency Management

### Go Modules Best Practices

```bash
# Initialize module
go mod init github.com/username/project

# Add dependencies
go get github.com/gorilla/mux@v1.8.0

# Update dependencies
go get -u ./...

# Tidy dependencies
go mod tidy

# Vendor dependencies (optional)
go mod vendor
```

### Dependency Injection

```go
package main

import (
    "database/sql"
    "log"
    "net/http"

    "github.com/gorilla/mux"
    _ "github.com/lib/pq"

    "myproject/internal/config"
    "myproject/internal/handlers"
    "myproject/internal/user"
)

type Application struct {
    config      *config.Config
    userService *user.Service
    handlers    *handlers.UserHandler
}

func NewApplication() (*Application, error) {
    // Load configuration
    cfg, err := config.Load()
    if err != nil {
        return nil, err
    }

    // Initialize database
    db, err := sql.Open("postgres", cfg.Database.ConnectionString())
    if err != nil {
        return nil, err
    }

    // Wire dependencies
    userRepo := user.NewPostgresRepository(db)
    userService := user.NewService(userRepo)
    userHandler := handlers.NewUserHandler(userService)

    return &Application{
        config:      cfg,
        userService: userService,
        handlers:    userHandler,
    }, nil
}

func (app *Application) setupRoutes() *mux.Router {
    r := mux.NewRouter()
    
    // User routes
    r.HandleFunc("/users", app.handlers.CreateUser).Methods("POST")
    r.HandleFunc("/users/{id}", app.handlers.GetUser).Methods("GET")
    r.HandleFunc("/users/{id}", app.handlers.UpdateUser).Methods("PUT")
    r.HandleFunc("/users/{id}", app.handlers.DeleteUser).Methods("DELETE")
    
    return r
}

func main() {
    app, err := NewApplication()
    if err != nil {
        log.Fatal(err)
    }

    router := app.setupRoutes()
    
    log.Printf("Server starting on port %s", app.config.Server.Port)
    log.Fatal(http.ListenAndServe(":"+app.config.Server.Port, router))
}
```

## Security Best Practices

### Input Validation

```go
package validation

import (
    "fmt"
    "regexp"
    "strings"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)

func ValidateEmail(email string) error {
    email = strings.TrimSpace(email)
    if email == "" {
        return fmt.Errorf("email is required")
    }
    if len(email) > 254 {
        return fmt.Errorf("email is too long")
    }
    if !emailRegex.MatchString(email) {
        return fmt.Errorf("invalid email format")
    }
    return nil
}

func SanitizeString(input string) string {
    // Remove potential XSS characters
    input = strings.ReplaceAll(input, "<", "&lt;")
    input = strings.ReplaceAll(input, ">", "&gt;")
    input = strings.ReplaceAll(input, "\"", "&quot;")
    input = strings.ReplaceAll(input, "'", "&#x27;")
    return strings.TrimSpace(input)
}
```

### Secure Configuration

```go
package config

import (
    "crypto/rand"
    "encoding/base64"
    "os"
)

type SecurityConfig struct {
    JWTSecret     string
    EncryptionKey []byte
}

func loadSecurityConfig() (*SecurityConfig, error) {
    jwtSecret := os.Getenv("JWT_SECRET")
    if jwtSecret == "" {
        return nil, fmt.Errorf("JWT_SECRET environment variable is required")
    }

    encryptionKey := os.Getenv("ENCRYPTION_KEY")
    if encryptionKey == "" {
        // Generate a new key (for development only)
        key := make([]byte, 32)
        if _, err := rand.Read(key); err != nil {
            return nil, err
        }
        encryptionKey = base64.StdEncoding.EncodeToString(key)
    }

    keyBytes, err := base64.StdEncoding.DecodeString(encryptionKey)
    if err != nil {
        return nil, fmt.Errorf("invalid encryption key: %w", err)
    }

    return &SecurityConfig{
        JWTSecret:     jwtSecret,
        EncryptionKey: keyBytes,
    }, nil
}
```

## Build and Deployment

### Makefile

```makefile
# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
BINARY_NAME=myapp
BINARY_UNIX=$(BINARY_NAME)_unix

# Build the application
build:
    $(GOBUILD) -o $(BINARY_NAME) -v ./cmd/app

# Clean build artifacts
clean:
    $(GOCLEAN)
    rm -f $(BINARY_NAME)
    rm -f $(BINARY_UNIX)

# Run tests
test:
    $(GOTEST) -v ./...

# Run tests with coverage
test-coverage:
    $(GOTEST) -v -coverprofile=coverage.out ./...
    $(GOCMD) tool cover -html=coverage.out

# Build for Linux
build-linux:
    CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BINARY_UNIX) -v ./cmd/app

# Run the application
run:
    $(GOBUILD) -o $(BINARY_NAME) -v ./cmd/app
    ./$(BINARY_NAME)

# Install dependencies
deps:
    $(GOMOD) download
    $(GOMOD) tidy

# Lint code
lint:
    golangci-lint run

# Format code
fmt:
    $(GOCMD) fmt ./...

# Security scan
security:
    gosec ./...

.PHONY: build clean test test-coverage build-linux run deps lint fmt security
```

## Conclusion

Following these Go architecture best practices will result in:

- **Maintainable Code**: Clear structure and separation of concerns
- **Scalable Applications**: Proper dependency management and interface design
- **Reliable Software**: Comprehensive testing and error handling
- **High Performance**: Efficient concurrency patterns and memory management
- **Security**: Input validation and secure configuration management

Remember the Go philosophy: "Don't communicate by sharing memory; share memory by communicating." Use channels and goroutines effectively, keep interfaces small, and always handle errors explicitly.
