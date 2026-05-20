# Test Helpers for Waka3x

This package provides common test utilities and helpers for writing tests across the Waka3x codebase.

## Usage

### Test Database Setup

```go
import "github.com/zetkey/waka3x/testing/helpers"

func TestSomething(t *testing.T) {
    db := helpers.SetupTestDB(t)
    defer helpers.CleanupTestDB(t, db)
    
    // Your test code here
}
```

### Mock Factories

```go
// Create a test user
user := helpers.NewTestUser("testuser", "test@example.com")

// Create test heartbeats
heartbeats := helpers.NewTestHeartbeats(user.ID, 10)
```

### Assertions

```go
// Assert no error
helpers.AssertNoError(t, err)

// Assert error contains message
helpers.AssertErrorContains(t, err, "expected message")
```

## Test Structure

Tests should follow this structure:

1. **Arrange**: Set up test data and mocks
2. **Act**: Execute the code under test
3. **Assert**: Verify the results

## Running Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run specific package tests
go test ./services/...

# Run specific test
go test -run TestUserService_GetByEmail ./services/...
```

## Coverage Goals

- **Critical paths**: 80%+ coverage (auth, heartbeat processing, summary generation)
- **Business logic**: 70%+ coverage (services layer)
- **Data access**: 60%+ coverage (repositories)
- **Overall**: 60%+ coverage

## Best Practices

1. **Use table-driven tests** for multiple test cases
2. **Mock external dependencies** (database, HTTP clients, etc.)
3. **Test error cases** as well as happy paths
4. **Keep tests fast** - use in-memory databases
5. **Make tests independent** - no shared state between tests
6. **Use descriptive test names** - TestServiceName_MethodName_Scenario
