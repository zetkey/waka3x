# Waka3x Documentation Index

**AI Agent First Documentation**  
Last Updated: 2026-05-17

## Project Overview

**Waka3x** is a minimalist, self-hosted WakaTime-compatible backend for coding statistics tracking. It's a community fork of Wakapi with a modern Vue 3 frontend and enhanced features.

### Key Information
- **Repository**: https://github.com/zetkey/waka3x
- **License**: MIT
- **Backend**: Go 1.26+ with Chi router, GORM ORM
- **Frontend**: Vue 3 + Vite + TypeScript + Tailwind CSS + shadcn-vue
- **Database**: SQLite (default), MySQL, or PostgreSQL
- **Current Branch**: refactor/vue-frontend
- **Main Branch**: master

### Core Features
- ✅ WakaTime-compatible heartbeat tracking
- ✅ Statistics for projects, languages, editors, hosts, OS
- ✅ REST API with Swagger documentation
- ✅ Modern Vue 3 SPA frontend
- ✅ Badges generation (shields.io compatible)
- ✅ Weekly email reports
- ✅ Public leaderboards
- ✅ Prometheus metrics export
- ✅ Multiple authentication methods (local, OIDC, WebAuthn)
- ✅ WakaTime data import and relay
- ✅ Self-hosted with Docker support

## Documentation Structure

### Core Architecture
- **[01-ARCHITECTURE.md](./01-ARCHITECTURE.md)** - System architecture, design patterns, data flow
- **[02-BACKEND-STRUCTURE.md](./02-BACKEND-STRUCTURE.md)** - Go backend organization, patterns, conventions
- **[03-FRONTEND-STRUCTURE.md](./03-FRONTEND-STRUCTURE.md)** - Vue frontend structure, components, state management

### Data Layer
- **[04-DATABASE-MODELS.md](./04-DATABASE-MODELS.md)** - Database schema, models, relationships, migrations

### API & Services
- **[05-API-ENDPOINTS.md](./05-API-ENDPOINTS.md)** - REST API endpoints, request/response formats
- **[06-SERVICES-LAYER.md](./06-SERVICES-LAYER.md)** - Business logic services, dependencies, responsibilities

### Development & Operations
- **[07-DEVELOPMENT-GUIDE.md](./07-DEVELOPMENT-GUIDE.md)** - Setup, development workflow, testing, deployment
- **[08-CONFIGURATION.md](./08-CONFIGURATION.md)** - Configuration options, environment variables, settings
- **[09-AUTHENTICATION.md](./09-AUTHENTICATION.md)** - Authentication mechanisms, security, session management
- **[10-BACKGROUND-JOBS.md](./10-BACKGROUND-JOBS.md)** - Scheduled tasks, cron jobs, background workers

## Quick Start for AI Agents

### Understanding the Codebase
1. Start with **01-ARCHITECTURE.md** for high-level system design
2. Read **02-BACKEND-STRUCTURE.md** for Go code organization
3. Review **04-DATABASE-MODELS.md** for data structures
4. Check **06-SERVICES-LAYER.md** for business logic

### Making Changes
1. **Backend changes**: See 02-BACKEND-STRUCTURE.md for patterns
2. **Frontend changes**: See 03-FRONTEND-STRUCTURE.md for component structure
3. **API changes**: See 05-API-ENDPOINTS.md for endpoint conventions
4. **Database changes**: See 04-DATABASE-MODELS.md for migration process

### Common Tasks
- **Add new API endpoint**: 02-BACKEND-STRUCTURE.md → 05-API-ENDPOINTS.md
- **Add new model**: 04-DATABASE-MODELS.md
- **Add new service**: 06-SERVICES-LAYER.md
- **Add new frontend view**: 03-FRONTEND-STRUCTURE.md
- **Configure feature**: 08-CONFIGURATION.md
- **Add background job**: 10-BACKGROUND-JOBS.md

## Project Structure Overview

```
waka3x/
├── main.go                 # Application entry point
├── config/                 # Configuration management
├── models/                 # Data models and types
├── repositories/           # Data access layer
├── services/               # Business logic layer
├── routes/                 # HTTP handlers and routing
│   ├── api/               # Main API handlers
│   └── compat/            # WakaTime/Shields compatibility
├── middlewares/           # HTTP middlewares
├── migrations/            # Database migrations
├── frontend/              # Vue 3 SPA
│   ├── src/
│   │   ├── components/   # Vue components
│   │   ├── views/        # Page views
│   │   ├── stores/       # Pinia state stores
│   │   ├── router/       # Vue Router config
│   │   └── lib/          # Utilities and helpers
│   └── dist/             # Built frontend (embedded)
├── static/               # Static assets
├── docs/                 # This documentation
└── scripts/              # Utility scripts
```

## Key Technologies

### Backend Stack
- **Language**: Go 1.26
- **Web Framework**: Chi v5 (router)
- **ORM**: GORM v1.31
- **Database Drivers**: SQLite (glebarez), MySQL, PostgreSQL
- **Authentication**: Argon2id, OIDC (go-oidc), WebAuthn
- **Job Scheduling**: robfig/cron v3
- **Logging**: slog (structured logging)
- **Monitoring**: Sentry, Prometheus

### Frontend Stack
- **Framework**: Vue 3.5
- **Build Tool**: Vite 8
- **Language**: TypeScript 6
- **Styling**: Tailwind CSS 4
- **UI Components**: shadcn-vue, Reka UI
- **State Management**: Pinia (via @vueuse/core)
- **HTTP Client**: Axios
- **Charts**: Chart.js + vue-chartjs
- **Form Validation**: vee-validate + zod
- **Icons**: lucide-vue-next, vue3-simple-icons

### Infrastructure
- **Containerization**: Docker + Docker Compose
- **Reverse Proxy**: Nginx (recommended)
- **Process Management**: SystemD (for non-Docker)

## Development Workflow

1. **Backend**: `go run main.go` (port 3000)
2. **Frontend**: `cd frontend && bun dev` (port 5173, proxies to backend)
3. **Database**: Auto-migrates on startup (unless `skip_migrations: true`)
4. **API Docs**: Available at `/swagger-ui/`

## Important Conventions

### Code Style
- **Go**: Standard Go formatting (gofmt), repository pattern
- **Vue**: Composition API, TypeScript, script setup syntax
- **Naming**: camelCase (JS/TS), snake_case (DB), PascalCase (Go types)

### Git Workflow
- **Main branch**: `master`
- **Current work**: `refactor/vue-frontend`
- **Commit style**: Conventional commits (feat:, fix:, refactor:, etc.)

## External Resources

- **Original Wakapi**: https://github.com/muety/wakapi
- **WakaTime API**: https://wakatime.com/developers
- **Shields.io**: https://shields.io/
- **Vue 3 Docs**: https://vuejs.org/
- **GORM Docs**: https://gorm.io/

## Notes for AI Agents

- All file paths in documentation are relative to project root
- Line numbers reference current state (may change with updates)
- Configuration examples use YAML format (config.default.yml)
- API examples use curl/HTTP format
- Database examples use GORM syntax
- Frontend examples use Vue 3 Composition API

## Getting Help

- **Issues**: https://github.com/zetkey/waka3x/issues
- **Original Wakapi Wiki**: https://github.com/muety/wakapi/wiki
- **WakaTime Docs**: https://wakatime.com/help
