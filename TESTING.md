# Aura CLI - Test Configuration

## Test Coverage Requirements

- **Minimum Coverage**: 80% per package
- **Overall Coverage**: 85%
- **Critical Packages**: 95% (db, ai, cmd)

## Test Types

### Unit Tests
- **Location**: `*_test.go` files alongside source
- **Naming**: `Test*` functions
- **Scope**: Individual functions and methods

### Integration Tests  
- **Location**: `tests/` directory
- **Naming**: `TestIntegration*` functions
- **Scope**: Cross-package functionality

### Benchmark Tests
- **Location**: `*_test.go` files
- **Naming**: `Benchmark*` functions
- **Scope**: Performance critical paths

## Running Tests

### All Tests
```bash
# Linux/macOS
./test-all.sh

# Windows
test-all.bat

# Manual
go test ./...
```

### Specific Package
```bash
go test -v ./internal/db
go test -v ./internal/ai
go test -v ./internal/cmd
```

### With Coverage
```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Race Detection
```bash
go test -race ./...
```

### Benchmarks
```bash
go test -bench=. ./...
go test -bench=. -benchmem ./...
```

## Test Structure

### Package Structure
```
internal/
├── ai/
│   ├── client.go
│   └── client_test.go
├── cmd/
│   ├── root.go
│   ├── root_test.go
│   ├── ask.go
│   ├── ask_test.go
│   └── ...
├── db/
│   ├── db.go
│   ├── db_test.go
│   ├── bookmarks.go
│   └── bookmarks_test.go
└── ...
```

### Test Function Naming
```go
// Unit tests
func TestFunctionName(t *testing.T) {}
func TestStructName_MethodName(t *testing.T) {}

// Table-driven tests
func TestFunctionName_Cases(t *testing.T) {
    tests := []struct {
        name string
        // test fields
    }{
        // test cases
    }
}

// Benchmarks
func BenchmarkFunctionName(b *testing.B) {}

// Examples
func ExampleFunctionName() {}
```

## Mock and Test Utilities

### Database Testing
- Use temporary databases for isolation
- Clean up after each test
- Test both Docker and file modes

### AI Client Testing
- Use `httptest.Server` for mocking API calls
- Test error conditions and timeouts
- Verify request formatting

### Command Testing
- Capture stdout/stderr for verification
- Test with various argument combinations
- Verify exit codes

### File System Testing
- Use temporary directories
- Clean up test files
- Test cross-platform paths

## Coverage Targets by Package

| Package | Target | Critical |
|---------|--------|----------|
| `cmd/aura` | 70% | No |
| `internal/ai` | 95% | Yes |
| `internal/cmd` | 85% | Yes |
| `internal/config` | 90% | Yes |
| `internal/context` | 80% | No |
| `internal/db` | 95% | Yes |
| `assets` | 75% | No |

## Test Environment

### Environment Variables
```bash
export AURA_ENV=test
export AURA_API_KEY=sk-test-key-for-testing
```

### Test Database
- Use SQLite in-memory or temporary file
- Initialize with test schema
- Isolated per test

### Mock Services
- AI API responses
- File system operations
- External commands

## Continuous Integration

### GitHub Actions
```yaml
- name: Run Tests
  run: |
    go test -v -race -coverprofile=coverage.out ./...
    go tool cover -func=coverage.out
```

### Coverage Reporting
- Upload to codecov.io
- Fail if coverage drops below threshold
- Generate HTML reports

## Test Data

### Fixtures
```
tests/
├── fixtures/
│   ├── git-repos/
│   ├── node-projects/
│   ├── python-projects/
│   └── templates/
└── testdata/
    ├── bookmarks.sql
    ├── api-responses.json
    └── config-files/
```

### Mock Data
- Sample AI responses
- Test bookmarks
- Example project structures

## Performance Testing

### Critical Paths
- Database operations
- File system navigation
- AI API calls
- Template rendering

### Benchmarks
```go
func BenchmarkAddBookmark(b *testing.B) {
    // Setup
    for i := 0; i < b.N; i++ {
        // Benchmark code
    }
}
```

### Memory Profiling
```bash
go test -memprofile=mem.prof -bench=.
go tool pprof mem.prof
```

## Error Testing

### Expected Errors
- Invalid input validation
- Network failures
- File system errors
- Database constraints

### Error Scenarios
```go
tests := []struct {
    name      string
    input     string
    wantError bool
    errorMsg  string
}{
    {
        name:      "invalid input",
        input:     "",
        wantError: true,
        errorMsg:  "input cannot be empty",
    },
}
```

## Integration Testing

### Cross-Package Tests
- End-to-end command execution
- Database + AI integration
- File operations + templates

### System Testing
- Full CLI workflows
- Real file system operations
- External tool integration

## Test Maintenance

### Regular Tasks
- Update test data
- Verify mock accuracy
- Performance regression testing
- Coverage analysis

### Best Practices
- Keep tests independent
- Use descriptive test names
- Test edge cases
- Mock external dependencies
- Clean up resources
