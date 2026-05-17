# Waka3x Backend Structure

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

The Waka3x backend is a Go application using a layered architecture with clear separation of concerns. This document details the organization, patterns, and conventions used throughout the codebase.

## Directory Structure

```
waka3x/
├── main.go                      # Application entry point (485 lines)
├── config/                      # Configuration management
│   ├── config.go               # Config loading and validation
│   ├── db.go                   # Database configuration
│   ├── oidc.go                 # OIDC provider setup
│   ├── sentry.go               # Error tracking setup
│   ├── jobqueue.go             # Background job queue
│   └── eventbus.go             # Event bus for pub/sub
├── models/                      # Data models and types (~6400 lines)
│   ├── user.go                 # User model and auth types
│   ├── heartbeat.go            # Heartbeat data structure
│   ├── summary.go              # Summary aggregation types
│   ├── alias.go                # Project/language aliases
│   ├── duration.go             # Time duration calculations
│   ├── filters.go              # Query filter types
│   ├── api_key.go              # API key model
│   ├── leaderboard.go          # Leaderboard types
│   └── shared.go               # Common types
├── repositories/                # Data access layer
│   ├── heartbeat.go            # Heartbeat CRUD
│   ├── summary.go              # Summary CRUD
│   ├── user.go                 # User CRUD
│   ├── alias.go                # Alias CRUD
│   ├── leaderboard.go          # Leaderboard queries
│   ├── metrics.go              # Metrics aggregation
│   └── ...
├── services/                    # Business logic layer
│   ├── heartbeat.go            # Heartbeat processing
│   ├── summary.go              # Summary generation
│   ├── aggregation.go          # Background aggregation
│   ├── user.go                 # User management
│   ├── report.go               # Email report generation
│   ├── leaderboard.go          # Leaderboard calculation
│   ├── housekeeping.go         # Data cleanup
│   └── ...
├── routes/                      # HTTP handlers
│   ├── api/                    # Main API handlers
│   │   ├── heartbeat.go        # POST /api/heartbeat
│   │   ├── summary.go          # GET /api/summary
│   │   ├── auth.go             # Authentication endpoints
│   │   ├── settings.go         # User settings
│   │   ├── badge.go            # Badge generation
│   │   └── ...
│   ├── compat/                 # Compatibility APIs
│   │   ├── wakatime/v1/        # WakaTime API compatibility
│   │   └── shields/v1/         # Shields.io badge API
│   └── subscription.go         # Subscription management
├── middlewares/                 # HTTP middlewares
│   ├── authenticate.go         # Authentication middleware
│   ├── logging.go              # Request logging
│   ├── security.go             # Security headers
│   ├── sentry.go               # Error tracking
│   └── custom/                 # Custom middlewares
├── migrations/                  # Database migrations
│   ├── migrations.go           # Migration runner
│   └── common/                 # Shared migration utilities
├── helpers/                     # Utility functions
├── utils/                       # General utilities
│   └── fs/                     # Filesystem utilities
└── mocks/                       # Test mocks
```

## Layer Responsibilities

### 1. Models Layer (`models/`)

**Purpose**: Define data structures and types used throughout the application.

**Key Files**:
- `user.go` - User model, authentication types (Login, Signup, etc.)
- `heartbeat.go` - Heartbeat data structure from editor plugins
- `summary.go` - Pre-computed summary statistics
- `alias.go` - User-defined aliases for projects/languages
- `duration.go` - Time duration calculations
- `shared.go` - Common types (KeyStringValue, Interval, etc.)

**Conventions**:
- Struct tags for JSON serialization and GORM mapping
- Custom types for time handling (CustomTime)
- Validation methods on models
- No business logic in models (pure data structures)

**Example**:
```go
type User struct {
    ID             string     `json:"id" gorm:"primary_key"`
    ApiKey         string     `json:"api_key" gorm:"unique"`
    Email          string     `json:"email" gorm:"uniqueIndex"`
    Password       string     `json:"-"`
    CreatedAt      CustomTime `swaggertype:"string"`
    IsAdmin        bool       `json:"-" gorm:"default:false"`
    // ... more fields
}
```

### 2. Repository Layer (`repositories/`)

**Purpose**: Abstract database operations, provide clean interface for data access.

**Pattern**: Interface + Implementation

**Key Interfaces**:
- `IHeartbeatRepository` - Heartbeat data access
- `ISummaryRepository` - Summary data access
- `IUserRepository` - User data access
- `IAliasRepository` - Alias data access

**Common Methods**:
- `Insert(model)` - Create new record
- `Update(model)` - Update existing record
- `Delete(id)` - Delete by ID
- `GetById(id)` - Fetch by ID
- `GetByUser(userId)` - Fetch user's records
- `Count(filters)` - Count with filters

**Example**:
```go
type IHeartbeatRepository interface {
    Insert(*Heartbeat) error
    InsertBatch([]*Heartbeat) error
    GetByUser(string, time.Time, time.Time) ([]*Heartbeat, error)
    GetLatestByUser(string) (*Heartbeat, error)
    DeleteByUser(string) error
    Count(*Filters) (int64, error)
}

type HeartbeatRepository struct {
    db *gorm.DB
}

func NewHeartbeatRepository(db *gorm.DB) IHeartbeatRepository {
    return &HeartbeatRepository{db: db}
}
```

**Location**: `repositories/*.go`

### 3. Service Layer (`services/`)

**Purpose**: Implement business logic, orchestrate repositories, enforce business rules.

**Key Services**:
- `HeartbeatService` - Process incoming heartbeats, validate, deduplicate
- `SummaryService` - Generate and retrieve summaries
- `AggregationService` - Background aggregation jobs
- `UserService` - User management, authentication
- `ReportService` - Email report generation
- `LeaderboardService` - Leaderboard calculation
- `HousekeepingService` - Data cleanup and maintenance

**Service Pattern**:
```go
type HeartbeatService struct {
    repository             IHeartbeatRepository
    languageMappingService ILanguageMappingService
}

func NewHeartbeatService(
    repo IHeartbeatRepository,
    langSvc ILanguageMappingService,
) IHeartbeatService {
    return &HeartbeatService{
        repository:             repo,
        languageMappingService: langSvc,
    }
}

func (srv *HeartbeatService) ProcessHeartbeat(hb *Heartbeat) error {
    // Validation
    if err := srv.validate(hb); err != nil {
        return err
    }
    
    // Business logic
    srv.applyLanguageMapping(hb)
    
    // Persistence
    return srv.repository.Insert(hb)
}
```

**Conventions**:
- Services are stateless
- Dependencies injected via constructor
- Return errors, don't panic
- Use context.Context for cancellation

### 4. Routes/Handlers Layer (`routes/`)

**Purpose**: Handle HTTP requests, parse input, call services, format responses.

**Structure**:
- `routes/api/` - Main API endpoints
- `routes/compat/wakatime/v1/` - WakaTime API compatibility
- `routes/compat/shields/v1/` - Shields.io badge API

**Handler Pattern**:
```go
type HeartbeatApiHandler struct {
    userService      IUserService
    heartbeatService IHeartbeatService
    langMappingService ILanguageMappingService
}

func NewHeartbeatApiHandler(
    userSvc IUserService,
    hbSvc IHeartbeatService,
    langSvc ILanguageMappingService,
) *HeartbeatApiHandler {
    return &HeartbeatApiHandler{
        userService:      userSvc,
        heartbeatService: hbSvc,
        langMappingService: langSvc,
    }
}

func (h *HeartbeatApiHandler) RegisterRoutes(router chi.Router) {
    router.Post("/heartbeat", h.Post)
    router.Post("/heartbeats", h.Post)
}

func (h *HeartbeatApiHandler) Post(w http.ResponseWriter, r *http.Request) {
    user := middlewares.GetPrincipal(r)
    
    // Parse request
    var heartbeats []*Heartbeat
    if err := json.NewDecoder(r.Body).Decode(&heartbeats); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        return
    }
    
    // Process via service
    for _, hb := range heartbeats {
        hb.UserID = user.ID
        if err := h.heartbeatService.ProcessHeartbeat(hb); err != nil {
            // Handle error
        }
    }
    
    w.WriteHeader(http.StatusCreated)
}
```

**Key Files**:
- `routes/api/heartbeat.go` - Heartbeat ingestion
- `routes/api/summary.go` - Summary retrieval
- `routes/api/auth.go` - Authentication (login, signup, OIDC)
- `routes/api/settings.go` - User settings management
- `routes/api/badge.go` - Badge generation
- `routes/api/metrics.go` - Prometheus metrics

### 5. Middleware Layer (`middlewares/`)

**Purpose**: Cross-cutting concerns applied to HTTP requests.

**Key Middlewares**:
- `authenticate.go` - Authentication (API key, session, OIDC)
- `logging.go` - Request/response logging
- `security.go` - Security headers (CSP, HSTS, etc.)
- `sentry.go` - Error tracking integration

**Middleware Pattern**:
```go
func NewAuthenticateMiddleware(userService IUserService) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract credentials (API key or session)
            user, err := authenticateRequest(r, userService)
            if err != nil {
                w.WriteHeader(http.StatusUnauthorized)
                return
            }
            
            // Store user in context
            ctx := context.WithValue(r.Context(), "principal", user)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

## Key Patterns and Conventions

### Error Handling

**Pattern**: Return errors, don't panic
```go
func (srv *Service) DoSomething() error {
    if err := srv.validate(); err != nil {
        return fmt.Errorf("validation failed: %w", err)
    }
    return nil
}
```

**Logging Errors**:
```go
if err := srv.DoSomething(); err != nil {
    conf.Log().Error("operation failed", "error", err)
    return err
}
```

### Dependency Injection

**Location**: `main.go` (lines 172-248)

**Order**:
1. Initialize repositories (depend on `*gorm.DB`)
2. Initialize services (depend on repositories and other services)
3. Initialize handlers (depend on services)
4. Register routes

### Configuration Access

**Global Config**: `conf.Get()` returns `*conf.Config`

**Usage**:
```go
import conf "github.com/zetkey/waka3x/config"

func someFunction() {
    config := conf.Get()
    if config.App.LeaderboardEnabled {
        // ...
    }
}
```

### Database Transactions

**Pattern**: Use GORM transactions for multi-step operations
```go
func (repo *Repository) ComplexOperation() error {
    return repo.db.Transaction(func(tx *gorm.DB) error {
        if err := tx.Create(&model1).Error; err != nil {
            return err
        }
        if err := tx.Create(&model2).Error; err != nil {
            return err
        }
        return nil
    })
}
```

### Background Jobs

**Pattern**: Services implement `Schedule()` method
```go
func (srv *AggregationService) Schedule() {
    config := conf.Get()
    _, err := conf.Cron().AddFunc(config.App.AggregationTime, func() {
        srv.Run()
    })
    if err != nil {
        conf.Log().Error("failed to schedule aggregation", "error", err)
    }
}
```

**Invocation**: Called in `main.go` after service initialization
```go
go aggregationService.Schedule()
go reportService.Schedule()
go housekeepingService.Schedule()
```

## Testing Conventions

**Test Files**: `*_test.go` alongside source files

**Framework**: testify/assert

**Example**:
```go
func TestUserValidation(t *testing.T) {
    user := &User{Email: "invalid"}
    err := user.Validate()
    assert.Error(t, err)
}
```

**Mocking**: Interfaces enable easy mocking
```go
type MockHeartbeatRepository struct {
    mock.Mock
}

func (m *MockHeartbeatRepository) Insert(hb *Heartbeat) error {
    args := m.Called(hb)
    return args.Error(0)
}
```

## Common Utilities

### Configuration (`config/`)
- `config.go` - Load and validate configuration
- `db.go` - Database connection setup
- `oidc.go` - OIDC provider configuration

### Helpers (`helpers/`)
- Common helper functions used across the application

### Utils (`utils/`)
- `utils/fs/` - Filesystem utilities
- General utility functions

## API Documentation

**Swagger**: Auto-generated from code comments

**Annotations**:
```go
// @Summary Get user summary
// @Description Retrieve coding statistics summary for a time range
// @Tags summary
// @Accept json
// @Produce json
// @Param from query string true "Start date (ISO 8601)"
// @Param to query string true "End date (ISO 8601)"
// @Success 200 {object} Summary
// @Failure 401 {string} string "Unauthorized"
// @Router /summary [get]
// @Security ApiKeyAuth
func (h *SummaryApiHandler) Get(w http.ResponseWriter, r *http.Request) {
    // Implementation
}
```

**Access**: `/swagger-ui/` when running

## File Naming Conventions

- **Models**: Singular noun (e.g., `user.go`, `heartbeat.go`)
- **Repositories**: Singular noun (e.g., `user.go`, `heartbeat.go`)
- **Services**: Singular noun (e.g., `user.go`, `heartbeat.go`)
- **Handlers**: Descriptive name (e.g., `heartbeat.go`, `summary.go`)
- **Tests**: `*_test.go` suffix

## Import Conventions

**Standard Library First**, then **External**, then **Internal**:
```go
import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-chi/chi/v5"
    "gorm.io/gorm"
    
    conf "github.com/zetkey/waka3x/config"
    "github.com/zetkey/waka3x/models"
    "github.com/zetkey/waka3x/services"
)
```

## Code Style

- **Formatting**: `gofmt` (standard Go formatting)
- **Linting**: Go standard linters
- **Naming**: 
  - Exported: PascalCase
  - Unexported: camelCase
  - Acronyms: All caps (e.g., `APIKey`, `HTTPServer`)
- **Comments**: Exported items should have doc comments

## Key Entry Points for AI Agents

**To understand the application flow**:
1. Start with `main.go` - see initialization order
2. Look at `routes/api/*.go` - see HTTP endpoints
3. Examine `services/*.go` - see business logic
4. Check `repositories/*.go` - see data access

**To add a new feature**:
1. Define model in `models/`
2. Create repository interface and implementation in `repositories/`
3. Create service in `services/`
4. Create handler in `routes/api/`
5. Register routes in handler's `RegisterRoutes()` method
6. Update `main.go` to initialize new components

**To modify existing feature**:
1. Identify the relevant handler in `routes/api/`
2. Trace to the service being called
3. Modify service logic
4. Update repository if data access changes
5. Update model if data structure changes
