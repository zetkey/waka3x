# Waka3x Architecture

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## System Overview

Waka3x is a full-stack web application that tracks coding activity through heartbeat events sent by WakaTime-compatible editor plugins. It processes, aggregates, and visualizes coding statistics.

### High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                    Editor Plugins                            │
│         (VS Code, JetBrains, Vim, etc.)                     │
└────────────────────┬────────────────────────────────────────┘
                     │ Heartbeats (HTTP POST)
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                   Waka3x Backend (Go)                        │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              HTTP Layer (Chi Router)                  │  │
│  │  - API Routes (/api/*)                               │  │
│  │  - Compat Routes (WakaTime, Shields.io)             │  │
│  │  - Frontend SPA serving                              │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               ▼                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Middleware Layer                         │  │
│  │  - Authentication (API Key, Session, OIDC)          │  │
│  │  - Logging & Monitoring                              │  │
│  │  - Security Headers                                  │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               ▼                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Services Layer                           │  │
│  │  - HeartbeatService (process incoming data)          │  │
│  │  - SummaryService (aggregate statistics)             │  │
│  │  - UserService (user management)                     │  │
│  │  - AggregationService (background processing)        │  │
│  │  - ReportService (email reports)                     │  │
│  │  - LeaderboardService (rankings)                     │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               ▼                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │            Repository Layer                           │  │
│  │  - HeartbeatRepository (CRUD operations)             │  │
│  │  - SummaryRepository                                 │  │
│  │  - UserRepository                                    │  │
│  │  - (GORM-based data access)                          │  │
│  └────────────┬─────────────────────────────────────────┘  │
│               ▼                                              │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Database (GORM)                          │  │
│  │  - SQLite / MySQL / PostgreSQL                       │  │
│  └──────────────────────────────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                     │ JSON API
                     ▼
┌─────────────────────────────────────────────────────────────┐
│                 Frontend (Vue 3 SPA)                         │
│  - Dashboard views                                           │
│  - Statistics visualization (Chart.js)                       │
│  - Settings management                                       │
│  - Authentication UI                                         │
└─────────────────────────────────────────────────────────────┘
```

## Core Design Patterns

### 1. Layered Architecture

**Layers (top to bottom):**
1. **HTTP Layer** (`routes/`, `routes/api/`) - Request handling, routing
2. **Middleware Layer** (`middlewares/`) - Cross-cutting concerns
3. **Service Layer** (`services/`) - Business logic
4. **Repository Layer** (`repositories/`) - Data access
5. **Model Layer** (`models/`) - Data structures

**Key Principle**: Each layer only depends on layers below it. No circular dependencies.

### 2. Repository Pattern

**Purpose**: Abstracts database operations from business logic.

**Structure**:
```go
// Interface definition
type IHeartbeatRepository interface {
    Insert(*Heartbeat) error
    GetByUser(string, time.Time, time.Time) ([]*Heartbeat, error)
    DeleteByUser(string) error
    // ...
}

// Implementation
type HeartbeatRepository struct {
    db *gorm.DB
}
```

**Location**: `repositories/` directory  
**Usage**: Services depend on repository interfaces, not implementations

### 3. Service Pattern

**Purpose**: Encapsulates business logic and orchestrates repositories.

**Structure**:
```go
type HeartbeatService struct {
    repository IHeartbeatRepository
    languageMappingService ILanguageMappingService
}

func (srv *HeartbeatService) ProcessHeartbeat(hb *Heartbeat) error {
    // Business logic here
    // Validation, transformation, persistence
}
```

**Location**: `services/` directory  
**Characteristics**: Stateless, testable, composable

### 4. Dependency Injection

**Implementation**: Constructor-based DI in `main.go`

```go
// Repositories initialized first
heartbeatRepository := repositories.NewHeartbeatRepository(db)

// Services receive dependencies via constructors
heartbeatService := services.NewHeartbeatService(
    heartbeatRepository,
    languageMappingService,
)

// Handlers receive services
heartbeatHandler := api.NewHeartbeatApiHandler(
    userService,
    heartbeatService,
    languageMappingService,
)
```

**Benefits**: Testability, loose coupling, clear dependencies

## Data Flow

### Heartbeat Processing Flow

```
1. Editor Plugin → POST /api/heartbeat
                    ↓
2. Chi Router → HeartbeatApiHandler.Post()
                    ↓
3. Authentication Middleware (validates API key)
                    ↓
4. HeartbeatService.ProcessHeartbeat()
   - Validates heartbeat data
   - Applies language mappings
   - Checks for duplicates
                    ↓
5. HeartbeatRepository.Insert()
   - GORM inserts to database
                    ↓
6. Response: 201 Created
```

### Summary Generation Flow

```
1. Cron Job triggers (daily at 2:15 AM)
                    ↓
2. AggregationService.Run()
   - Iterates all users
   - For each user:
                    ↓
3. SummaryService.Summarize(user, from, to)
   - Fetches heartbeats for time range
   - Groups by project, language, editor, etc.
   - Calculates durations
   - Applies aliases
                    ↓
4. SummaryRepository.Insert()
   - Stores pre-computed summary
                    ↓
5. Cache invalidation (if applicable)
```

### API Request Flow

```
1. Frontend → GET /api/summary?from=X&to=Y
                    ↓
2. Chi Router → SummaryApiHandler.Get()
                    ↓
3. Authentication Middleware
   - Validates session cookie OR API key
                    ↓
4. SummaryService.Retrieve(user, from, to)
   - Checks for existing summary
   - If missing, generates on-demand
                    ↓
5. Response: JSON summary data
                    ↓
6. Frontend renders charts/tables
```

## Key Components

### main.go (Application Entry Point)

**Responsibilities**:
- Parse command-line flags
- Load configuration
- Initialize database connection
- Run migrations
- Initialize all repositories
- Initialize all services
- Initialize all handlers
- Set up routing
- Start background jobs
- Start HTTP server

**Location**: `/main.go` (485 lines)

### Configuration System

**Files**:
- `config/config.go` - Configuration struct and loading logic
- `config.default.yml` - Default configuration template
- User provides: `waka3x.yml` or environment variables

**Key Features**:
- YAML-based configuration
- Environment variable overrides
- Validation on startup
- Hot-reload not supported (requires restart)

**Location**: `config/` directory

### Database Layer (GORM)

**ORM**: GORM v1.31  
**Supported Databases**: SQLite, MySQL, PostgreSQL

**Key Features**:
- Auto-migrations on startup
- Connection pooling
- Transaction support
- Soft deletes (where applicable)

**Migration System**: `migrations/` directory
- Runs automatically unless `skip_migrations: true`
- Version-based migrations
- Idempotent (safe to re-run)

### Background Jobs

**Scheduler**: robfig/cron v3

**Jobs**:
1. **Aggregation** - Daily summary generation (2:15 AM)
2. **Reports** - Weekly email reports (Friday 6 PM)
3. **Housekeeping** - Data cleanup (Sunday 6 AM)
4. **Leaderboard** - Ranking updates (6 AM, 6 PM)
5. **Database Optimization** - Vacuum/optimize (1st of month, 8 AM)

**Implementation**: Each service schedules its own jobs in `Schedule()` method

## Authentication & Authorization

### Authentication Methods

1. **API Key** (for editor plugins)
   - Header: `Authorization: Basic <base64(api_key)>`
   - Stored in `users.api_key` (unique)

2. **Session Cookie** (for web UI)
   - Cookie-based sessions (gorilla/sessions)
   - Stored in encrypted cookie
   - Configurable max age

3. **OIDC** (OpenID Connect)
   - Supports multiple providers
   - Configuration in `config.yml`
   - User linking via `sub` claim

4. **WebAuthn** (Security keys, biometrics)
   - FIDO2/WebAuthn standard
   - Stored credentials in `webauthn_credentials` table

### Authorization

**Model**: Simple role-based
- `is_admin` flag on User model
- Admin-only endpoints check this flag
- Most endpoints are user-scoped (users can only access their own data)

## Frontend Architecture

### Technology Stack

- **Framework**: Vue 3 (Composition API)
- **Build Tool**: Vite 8
- **Language**: TypeScript 6
- **Styling**: Tailwind CSS 4
- **UI Components**: shadcn-vue (Reka UI)
- **State Management**: Pinia stores (via @vueuse/core)
- **Routing**: Vue Router
- **HTTP**: Axios

### Build & Deployment

**Development**: `bun dev` (port 5173, proxies API to :3000)  
**Production**: Built to `frontend/dist/`, embedded in Go binary via `//go:embed`

**Serving**: Go serves SPA with fallback to `index.html` for client-side routing

## Performance Considerations

### Caching Strategy

1. **In-Memory Cache** (patrickmn/go-cache)
   - User objects
   - Configuration values
   - Frequently accessed data

2. **Pre-computed Summaries**
   - Daily aggregation generates summaries
   - Stored in `summaries` table
   - Avoids real-time aggregation on every request

3. **Database Indexes**
   - Composite indexes on frequently queried columns
   - See `migrations/` for index definitions

### Scalability

**Current Design**: Single-instance, monolithic
- Suitable for small to medium deployments (< 1000 users)
- Database is the bottleneck for large deployments

**Potential Improvements**:
- Horizontal scaling requires session store externalization
- Background jobs could be moved to separate workers
- Read replicas for database

## Security Architecture

### Input Validation

- **API Layer**: Schema validation (gorilla/schema)
- **Service Layer**: Business rule validation
- **Database Layer**: GORM constraints

### Password Security

- **Hashing**: Argon2id (alexedwards/argon2id)
- **Salt**: Configured in `security.password_salt`
- **Reset Tokens**: Time-limited, single-use

### CSRF Protection

- **Session-based requests**: CSRF tokens (gorilla/csrf)
- **API key requests**: Not applicable (stateless)

### Rate Limiting

**Not implemented** - Recommended to use reverse proxy (nginx) for rate limiting

## Monitoring & Observability

### Logging

- **Library**: slog (structured logging)
- **Format**: Text or JSON (configurable)
- **Levels**: Debug, Info, Warn, Error

### Error Tracking

- **Sentry Integration**: Optional (configure `sentry.dsn`)
- **Captures**: Panics, errors, performance data

### Metrics

- **Prometheus Exporter**: `/api/metrics` endpoint
- **Metrics**: Request counts, durations, database stats

## Testing Strategy

**Current State**: Limited test coverage

**Existing Tests**:
- `models/*_test.go` - Model validation tests
- `services/*_test.go` - Service unit tests
- `routes/api/*_test.go` - Handler tests

**Test Framework**: testify/assert

## Deployment Architecture

### Docker Deployment

**Image**: Built from `Dockerfile`
- Multi-stage build (Go build + frontend build)
- Final image: Alpine-based, ~50MB
- Embedded frontend in binary

**Docker Compose**: `compose.yml`
- Single container deployment
- Volume for SQLite data
- Secrets for sensitive config

### Standalone Deployment

**SystemD Service**: `etc/waka3x.service`
- Runs as dedicated user
- Auto-restart on failure
- Logs to journald

**Reverse Proxy**: Nginx recommended
- TLS termination
- Rate limiting
- Static asset caching

## File Organization

```
waka3x/
├── main.go                    # Entry point (485 lines)
├── config/                    # Configuration management
│   ├── config.go             # Config struct and loading
│   ├── db.go                 # Database configuration
│   ├── oidc.go               # OIDC provider config
│   └── sentry.go             # Sentry integration
├── models/                    # Data models (~6400 lines total)
│   ├── user.go               # User model
│   ├── heartbeat.go          # Heartbeat model
│   ├── summary.go            # Summary model
│   └── ...
├── repositories/              # Data access layer
│   ├── heartbeat.go
│   ├── summary.go
│   └── ...
├── services/                  # Business logic layer
│   ├── heartbeat.go
│   ├── summary.go
│   ├── aggregation.go
│   └── ...
├── routes/                    # HTTP handlers
│   ├── api/                  # Main API
│   │   ├── heartbeat.go
│   │   ├── summary.go
│   │   └── ...
│   └── compat/               # Compatibility APIs
│       ├── wakatime/v1/
│       └── shields/v1/
├── middlewares/               # HTTP middlewares
│   ├── authenticate.go
│   ├── logging.go
│   └── ...
├── migrations/                # Database migrations
├── frontend/                  # Vue 3 SPA
│   ├── src/
│   └── dist/                 # Built assets (embedded)
└── static/                    # Static files
```

## Key Design Decisions

1. **Embedded Frontend**: Frontend built into Go binary for easy deployment
2. **GORM ORM**: Abstracts database differences, supports multiple DBs
3. **Repository Pattern**: Enables testing, swappable implementations
4. **Pre-computed Summaries**: Trade storage for query performance
5. **WakaTime Compatibility**: Allows using existing editor plugins
6. **Single Binary**: Simplifies deployment, no external dependencies (except DB)

## Future Architecture Considerations

- **Microservices**: Could split into API, worker, and frontend services
- **Message Queue**: For async heartbeat processing (RabbitMQ, Redis)
- **Caching Layer**: Redis for distributed caching
- **CDN**: For static assets and badges
- **Read Replicas**: For scaling read-heavy workloads
