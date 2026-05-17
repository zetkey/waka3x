# Waka3x Services Layer

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

The services layer contains business logic, orchestrates repositories, and enforces business rules. Services are stateless and use dependency injection.

## Service Pattern

```go
type ServiceName struct {
    repository IRepository
    otherService IOtherService
}

func NewServiceName(repo IRepository, other IOtherService) IServiceName {
    return &ServiceName{
        repository: repo,
        otherService: other,
    }
}
```

## Core Services

### HeartbeatService (`services/heartbeat.go`)

**Purpose**: Process incoming heartbeats from editor plugins.

**Dependencies**:
- `IHeartbeatRepository` - Data persistence
- `ILanguageMappingService` - Language detection

**Key Methods**:
- `ProcessHeartbeat(*Heartbeat)` - Validate and store heartbeat
- `InsertBatch([]*Heartbeat)` - Bulk insert
- `GetByUser(userID, from, to)` - Retrieve user heartbeats
- `DeleteByUser(userID)` - Delete all user heartbeats

**Business Logic**:
- Validates heartbeat data (required fields, timestamps)
- Applies language mappings based on file extension
- Generates hash for deduplication
- Checks heartbeat age against `heartbeat_max_age` config

### SummaryService (`services/summary.go`)

**Purpose**: Generate and retrieve aggregated statistics.

**Dependencies**:
- `ISummaryRepository` - Summary persistence
- `IHeartbeatService` - Heartbeat retrieval
- `IDurationService` - Duration calculations
- `IAliasService` - Apply user aliases
- `IProjectLabelService` - Project labels

**Key Methods**:
- `Summarize(user, from, to)` - Generate summary for time range
- `Retrieve(user, from, to)` - Get existing or generate new summary
- `DeleteByUser(userID)` - Delete all user summaries

**Business Logic**:
- Fetches heartbeats for time range
- Calculates durations between heartbeats
- Groups by project, language, editor, OS, machine
- Applies user-defined aliases
- Stores pre-computed summary
- Returns cached summary if exists

### AggregationService (`services/aggregation.go`)

**Purpose**: Background job to generate daily summaries for all users.

**Dependencies**:
- `IUserService` - User retrieval
- `ISummaryService` - Summary generation
- `IHeartbeatService` - Heartbeat queries
- `IDurationService` - Duration calculations

**Key Methods**:
- `Run()` - Execute aggregation for all users
- `Schedule()` - Schedule daily cron job

**Business Logic**:
- Runs daily at configured time (default: 2:15 AM)
- Iterates all active users
- Generates summaries for yesterday
- Logs progress and errors

**Cron Schedule**: `app.aggregation_time` (default: `0 15 2 * * *`)

### UserService (`services/user.go`)

**Purpose**: User management and authentication.

**Dependencies**:
- `IUserRepository` - User persistence
- `IKeyValueService` - Settings storage
- `IMailService` - Email sending
- `IApiKeyService` - API key management

**Key Methods**:
- `Create(*User)` - Create new user
- `Update(*User)` - Update user
- `GetById(id)` - Get user by ID
- `GetByApiKey(key)` - Get user by API key
- `GetByEmail(email)` - Get user by email
- `Authenticate(username, password)` - Verify credentials
- `ResetPassword(email)` - Send password reset email
- `SetPassword(token, password)` - Set new password
- `Delete(user)` - Delete user and all data

**Business Logic**:
- Password hashing with Argon2id
- API key generation (UUID)
- Email validation
- Reset token generation (time-limited)

### DurationService (`services/duration.go`)

**Purpose**: Calculate time durations between heartbeats.

**Dependencies**:
- `IDurationRepository` - Duration persistence
- `IHeartbeatService` - Heartbeat retrieval
- `IUserService` - User settings (timeout)
- `ILanguageMappingService` - Language mappings

**Key Methods**:
- `Get(user, from, to)` - Get durations for time range
- `Calculate(heartbeats)` - Calculate durations from heartbeats

**Business Logic**:
- Groups consecutive heartbeats by project/file
- Calculates time between heartbeats
- Applies timeout threshold (default: 10 minutes)
- If gap > timeout, duration = timeout
- Stores calculated durations

### ReportService (`services/report.go`)

**Purpose**: Generate and send email reports.

**Dependencies**:
- `ISummaryService` - Summary retrieval
- `IUserService` - User retrieval
- `IMailService` - Email sending

**Key Methods**:
- `SendWeekly()` - Send weekly reports to all subscribed users
- `Schedule()` - Schedule weekly cron job

**Business Logic**:
- Runs weekly at configured time (default: Friday 6 PM)
- Iterates users with `reports_weekly = true`
- Generates summary for past week
- Sends HTML email with statistics
- Logs success/failure

**Cron Schedule**: `app.report_time_weekly` (default: `0 0 18 * * 5`)

### LeaderboardService (`services/leaderboard.go`)

**Purpose**: Calculate and cache leaderboard rankings.

**Dependencies**:
- `ILeaderboardRepository` - Leaderboard persistence
- `ISummaryService` - Summary retrieval
- `IUserService` - User retrieval

**Key Methods**:
- `Compute()` - Calculate current leaderboard
- `Get()` - Retrieve cached leaderboard
- `Schedule()` - Schedule periodic updates

**Business Logic**:
- Calculates total coding time for configured period
- Ranks users by total time
- Filters by `public_leaderboard = true`
- Caches results in database
- Updates twice daily (6 AM, 6 PM)

**Cron Schedule**: `app.leaderboard_generation_time` (default: `0 0 6 * * *,0 0 18 * * *`)

### HousekeepingService (`services/housekeeping.go`)

**Purpose**: Data cleanup and maintenance tasks.

**Dependencies**:
- `IUserService` - User management
- `IHeartbeatService` - Heartbeat deletion
- `ISummaryService` - Summary deletion
- Repository (for database operations)

**Key Methods**:
- `CleanUp()` - Execute cleanup tasks
- `Schedule()` - Schedule weekly cron job

**Business Logic**:
- Deletes old heartbeats (if `data_retention_months` > 0)
- Deletes inactive users (if `max_inactive_months` > 0)
- Optimizes database (VACUUM for SQLite/Postgres, OPTIMIZE for MySQL)
- Runs weekly (Sunday 6 AM)

**Cron Schedule**: `app.data_cleanup_time` (default: `0 0 6 * * 0`)

### AliasService (`services/alias.go`)

**Purpose**: Manage user-defined aliases for projects/languages.

**Dependencies**:
- `IAliasRepository` - Alias persistence

**Key Methods**:
- `Create(*Alias)` - Create new alias
- `GetByUser(userID)` - Get all user aliases
- `Delete(id)` - Delete alias
- `Apply(items, aliases)` - Apply aliases to summary items

**Business Logic**:
- Normalizes project/language names
- Applies aliases during summary generation
- Supports wildcards in alias keys

### LanguageMappingService (`services/language_mapping.go`)

**Purpose**: Map file extensions to programming languages.

**Dependencies**:
- `ILanguageMappingRepository` - Mapping persistence

**Key Methods**:
- `GetByUser(userID)` - Get user mappings
- `GetLanguage(extension)` - Detect language from extension
- `Create(*LanguageMapping)` - Create custom mapping

**Business Logic**:
- Uses system defaults from config (`app.custom_languages`)
- Allows user overrides
- Falls back to WakaTime detection

### ProjectLabelService (`services/project_label.go`)

**Purpose**: Manage project labels/tags.

**Dependencies**:
- `IProjectLabelRepository` - Label persistence

**Key Methods**:
- `GetByUser(userID)` - Get all user labels
- `Create(*ProjectLabel)` - Create label
- `Delete(id)` - Delete label

### ActivityService (`services/activity.go`)

**Purpose**: Generate activity heatmap data.

**Dependencies**:
- `ISummaryService` - Summary retrieval

**Key Methods**:
- `GetActivity(user, from, to)` - Get daily activity for date range

**Business Logic**:
- Returns daily coding time for heatmap visualization
- Used by frontend dashboard

### DiagnosticsService (`services/diagnostics.go`)

**Purpose**: Store and retrieve diagnostic information.

**Dependencies**:
- `IDiagnosticsRepository` - Diagnostics persistence

**Key Methods**:
- `Create(*Diagnostics)` - Store diagnostic data
- `GetByUser(userID)` - Retrieve user diagnostics

### MiscService (`services/misc.go`)

**Purpose**: Miscellaneous utility operations.

**Dependencies**:
- `IUserService` - User operations
- `IHeartbeatService` - Heartbeat operations
- `ISummaryService` - Summary operations
- `IKeyValueService` - Settings storage
- `IMailService` - Email sending

**Key Methods**:
- `ComputeOldestHeartbeat(user)` - Find oldest heartbeat timestamp
- `CountHeartbeats(user)` - Count total heartbeats
- `Schedule()` - Schedule periodic tasks

### ApiKeyService (`services/api_key.go`)

**Purpose**: Manage additional API keys for users.

**Dependencies**:
- `IApiKeyRepository` - API key persistence

**Key Methods**:
- `Create(userID, name)` - Generate new API key
- `GetByUser(userID)` - List user's API keys
- `Delete(id)` - Revoke API key

### WebAuthnService (`services/webauthn.go`)

**Purpose**: Handle WebAuthn/FIDO2 authentication.

**Dependencies**:
- `IWebAuthnRepository` - Credential persistence

**Key Methods**:
- `BeginRegistration(user)` - Start credential registration
- `FinishRegistration(user, response)` - Complete registration
- `BeginLogin(user)` - Start authentication
- `FinishLogin(user, response)` - Complete authentication

## Service Initialization

**Location**: `main.go` (lines 186-203)

**Order**:
1. Initialize repositories (depend on `*gorm.DB`)
2. Initialize services (depend on repositories and other services)
3. Initialize handlers (depend on services)

**Example**:
```go
// Repositories
heartbeatRepository := repositories.NewHeartbeatRepository(db)
summaryRepository := repositories.NewSummaryRepository(db)

// Services
heartbeatService := services.NewHeartbeatService(heartbeatRepository, languageMappingService)
summaryService := services.NewSummaryService(summaryRepository, heartbeatService, durationService, aliasService, projectLabelService)
```

## Background Job Scheduling

Services with scheduled tasks implement `Schedule()` method:

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

**Invocation**: Called in `main.go` after service initialization:
```go
go aggregationService.Schedule()
go reportService.Schedule()
go housekeepingService.Schedule()
go leaderboardService.Schedule()
```

## Testing Services

**Pattern**: Mock dependencies using interfaces

```go
type MockHeartbeatRepository struct {
    mock.Mock
}

func TestHeartbeatService_ProcessHeartbeat(t *testing.T) {
    mockRepo := new(MockHeartbeatRepository)
    mockRepo.On("Insert", mock.Anything).Return(nil)
    
    service := NewHeartbeatService(mockRepo, nil)
    err := service.ProcessHeartbeat(&Heartbeat{})
    
    assert.NoError(t, err)
    mockRepo.AssertExpectations(t)
}
```
