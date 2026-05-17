# Waka3x Database Models

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

Waka3x uses GORM as its ORM, supporting SQLite (default), MySQL, and PostgreSQL. Models are defined in the `models/` directory with GORM struct tags for schema definition.

## Core Models

### User (`models/user.go`)

**Primary Model**: Represents a user account.

```go
type User struct {
    ID                     string     `gorm:"primary_key"`
    ApiKey                 string     `gorm:"unique"`
    Email                  string     `gorm:"uniqueIndex:idx_user_email"`
    Password               string     `json:"-"`
    CreatedAt              CustomTime
    LastLoggedInAt         CustomTime
    Location               string
    StartOfWeek            int        `gorm:"default:1"`
    IsAdmin                bool       `gorm:"default:false"`
    HasData                bool       `gorm:"default:false"`
    
    // Sharing settings
    ShareDataMaxDays       int
    ShareEditors           bool       `gorm:"default:false"`
    ShareLanguages         bool       `gorm:"default:false"`
    ShareProjects          bool       `gorm:"default:false"`
    ShareOSs               bool       `gorm:"default:false"`
    ShareMachines          bool       `gorm:"default:false"`
    
    // Authentication
    AuthType               string     `gorm:"default:local;uniqueIndex:idx_oidc"`
    Sub                    string     `gorm:"uniqueIndex:idx_oidc"` // OIDC subject
    WebauthnID             string
    Credentials            []*WebAuthnCredential
    
    // Integrations
    WakatimeApiKey         string
    WakatimeApiUrl         string
    
    // Subscription
    SubscribedUntil        *CustomTime
    StripeCustomerId       string
    
    // Settings
    ReportsWeekly          bool       `gorm:"default:false"`
    PublicLeaderboard      bool       `gorm:"default:false"`
    HeartbeatsTimeoutSec   int        `gorm:"default:600"`
}
```

**Table**: `users`  
**Key Indexes**: `idx_user_email`, `idx_oidc` (composite on auth_type + sub)

### Heartbeat (`models/heartbeat.go`)

**Primary Model**: Raw coding activity event from editor plugins.

```go
type Heartbeat struct {
    ID             uint64     `gorm:"primary_key"`
    User           *User      `json:"-" gorm:"not null; constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    UserID         string     `json:"-" gorm:"not null; index:idx_time,idx_user_time"`
    Entity         string     `json:"entity" gorm:"not null; index:idx_entity"`
    Type           string     `json:"type"`
    Category       string     `json:"category"`
    Project        string     `json:"project" gorm:"index:idx_project"`
    Branch         string     `json:"branch"`
    Language       string     `json:"language" gorm:"index:idx_language"`
    IsWrite        bool       `json:"is_write"`
    Editor         string     `json:"editor" gorm:"index:idx_editor"`
    OperatingSystem string    `json:"operating_system" gorm:"index:idx_operating_system"`
    Machine        string     `json:"machine" gorm:"index:idx_machine"`
    UserAgent      string     `json:"user_agent"`
    Time           CustomTime `json:"time" gorm:"type:timestamp; default:CURRENT_TIMESTAMP; index:idx_time,idx_user_time"`
    Hash           string     `json:"-" gorm:"uniqueIndex"`
    CreatedAt      CustomTime `gorm:"type:timestamp"`
}
```

**Table**: `heartbeats`  
**Key Indexes**: 
- `idx_user_time` (composite: user_id + time) - Most queries
- `idx_time` - Time-based queries
- `idx_entity`, `idx_project`, `idx_language`, `idx_editor` - Filtering

**Hash Field**: MD5 hash of (user_id + time + entity) for deduplication

### Summary (`models/summary.go`)

**Primary Model**: Pre-computed aggregated statistics for a time range.

```go
type Summary struct {
    ID                   uint       `json:"-" gorm:"primary_key"`
    User                 *User      `json:"-" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    UserID               string     `json:"user_id" gorm:"not null; uniqueIndex:idx_user_summary"`
    FromTime             CustomTime `json:"from" gorm:"not null; type:timestamp; uniqueIndex:idx_user_summary"`
    ToTime               CustomTime `json:"to" gorm:"not null; type:timestamp; uniqueIndex:idx_user_summary"`
    Projects             []*SummaryItem `json:"projects" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Languages            []*SummaryItem `json:"languages" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Editors              []*SummaryItem `json:"editors" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    OperatingSystems     []*SummaryItem `json:"operating_systems" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Machines             []*SummaryItem `json:"machines" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    Labels               []*SummaryItem `json:"labels" gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    NumHeartbeats        int        `json:"-"`
}

type SummaryItem struct {
    ID         uint       `json:"-" gorm:"primary_key"`
    SummaryID  uint       `json:"-"`
    Type       uint8      `json:"-"`
    Key        string     `json:"key"`
    Total      time.Duration `json:"total"`
}
```

**Table**: `summaries`, `summary_items`  
**Key Index**: `idx_user_summary` (composite: user_id + from_time + to_time) - Unique constraint

### Alias (`models/alias.go`)

**Purpose**: User-defined mappings to normalize project/language names.

```go
type Alias struct {
    ID     uint   `gorm:"primary_key"`
    Type   uint8  `gorm:"index:idx_alias_type_key"`
    User   *User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    UserID string `gorm:"not null; index:idx_alias_user"`
    Key    string `gorm:"not null; index:idx_alias_type_key"`
    Value  string `gorm:"not null"`
}
```

**Types**: 
- `1` - Project alias
- `2` - Language alias
- `3` - Editor alias
- `4` - OS alias
- `5` - Machine alias

**Table**: `aliases`

### Duration (`models/duration.go`)

**Purpose**: Calculated time durations between heartbeats.

```go
type Duration struct {
    ID                uint       `gorm:"primary_key"`
    UserID            string     `gorm:"not null; index:idx_duration_user_time"`
    Time              CustomTime `gorm:"type:timestamp; index:idx_duration_user_time"`
    Duration          time.Duration `gorm:"type:bigint"`
    Project           string     `gorm:"index:idx_duration_project"`
    Language          string     `gorm:"index:idx_duration_language"`
    Editor            string
    OperatingSystem   string
    Machine           string
    Branch            string
    Entity            string
    NumHeartbeats     int
}
```

**Table**: `durations`  
**Key Index**: `idx_duration_user_time` (composite: user_id + time)

### ApiKey (`models/api_key.go`)

**Purpose**: Additional API keys for a user (beyond the primary one).

```go
type ApiKey struct {
    ID        string     `gorm:"primary_key"`
    UserID    string     `gorm:"not null; index"`
    Name      string     `gorm:"not null"`
    CreatedAt CustomTime `gorm:"type:timestamp"`
    ExpiresAt *CustomTime `gorm:"type:timestamp"`
}
```

**Table**: `api_keys`

### LanguageMapping (`models/language_mapping.go`)

**Purpose**: System-wide language file extension mappings.

```go
type LanguageMapping struct {
    ID        uint   `gorm:"primary_key"`
    User      *User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    UserID    string `gorm:"not null; index:idx_language_mapping_user"`
    Extension string `gorm:"not null"`
    Language  string `gorm:"not null"`
}
```

**Table**: `language_mappings`

### ProjectLabel (`models/project_label.go`)

**Purpose**: User-defined labels/tags for projects.

```go
type ProjectLabel struct {
    ID        uint   `gorm:"primary_key"`
    User      *User  `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    UserID    string `gorm:"not null; index:idx_project_label_user"`
    ProjectKey string `gorm:"not null"`
    Label     string `gorm:"not null"`
}
```

**Table**: `project_labels`

### LeaderboardItem (`models/leaderboard.go`)

**Purpose**: Cached leaderboard rankings.

```go
type LeaderboardItem struct {
    ID            uint       `gorm:"primary_key"`
    User          *User      `gorm:"constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
    UserID        string     `gorm:"not null; index:idx_leaderboard_user"`
    Total         time.Duration
    Key           string     `gorm:"index:idx_leaderboard_key"`
    By            uint8
    CreatedAt     CustomTime `gorm:"type:timestamp"`
}
```

**Table**: `leaderboard`

### WebAuthnCredential (`models/webauthn.go`)

**Purpose**: WebAuthn/FIDO2 credentials for passwordless auth.

```go
type WebAuthnCredential struct {
    ID              string `gorm:"primary_key"`
    UserID          string `gorm:"not null; index"`
    PublicKey       []byte
    AttestationType string
    AAGUID          []byte
    SignCount       uint32
    CreatedAt       CustomTime
}
```

**Table**: `webauthn_credentials`

## Database Schema

### Relationships

```
users (1) ----< (N) heartbeats
users (1) ----< (N) summaries
users (1) ----< (N) aliases
users (1) ----< (N) api_keys
users (1) ----< (N) language_mappings
users (1) ----< (N) project_labels
users (1) ----< (N) webauthn_credentials

summaries (1) ----< (N) summary_items
```

### Cascade Behavior

**ON DELETE CASCADE**: When user is deleted, all related records are deleted.  
**ON UPDATE CASCADE**: When user ID changes, related records are updated.

## Migrations

**Location**: `migrations/` directory  
**Runner**: `migrations/migrations.go`

**Execution**: Automatic on startup (unless `skip_migrations: true`)

**Process**:
1. GORM AutoMigrate creates/updates tables
2. Custom migrations run for data transformations
3. Indexes created/updated

**Key Migrations**:
- Initial schema creation
- Add new columns with defaults
- Create indexes for performance
- Data migrations (rare)

## Custom Types

### CustomTime

**Purpose**: Wrapper around `time.Time` for custom JSON/DB serialization.

```go
type CustomTime struct {
    time.Time
}

func (j CustomTime) MarshalJSON() ([]byte, error) {
    // Custom format
}
```

**Usage**: All timestamp fields use `CustomTime` instead of `time.Time`

## Query Patterns

### Common Queries

**Get user's heartbeats for date range**:
```go
db.Where("user_id = ? AND time >= ? AND time < ?", userID, from, to).
   Order("time ASC").
   Find(&heartbeats)
```

**Get or create summary**:
```go
db.Where("user_id = ? AND from_time = ? AND to_time = ?", userID, from, to).
   FirstOrCreate(&summary)
```

**Get user by API key**:
```go
db.Where("api_key = ?", apiKey).First(&user)
```

### Performance Considerations

- Use composite indexes for multi-column WHERE clauses
- Avoid N+1 queries with `Preload()`
- Use `Select()` to fetch only needed columns
- Batch inserts with `CreateInBatches()`

## Database Configuration

**File**: `config.default.yml` (db section)

**Options**:
- `dialect`: sqlite3, mysql, postgres
- `host`, `port`: Database server (not for SQLite)
- `name`: Database name or file path
- `max_conn`: Connection pool size
- `ssl`: Enable TLS (Postgres)

**Connection String Examples**:
- SQLite: `waka3x.db` (file path)
- MySQL: `user:pass@tcp(host:3306)/dbname`
- Postgres: `host=localhost user=user password=pass dbname=waka3x`

## Data Retention

**Configuration**: `app.data_retention_months`

**Behavior**: 
- `-1` = Keep forever (default)
- `N` = Delete heartbeats older than N months

**Cleanup Job**: Runs weekly (Sunday 6 AM) via `HousekeepingService`
