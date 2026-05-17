# Waka3x API Endpoints

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

Waka3x provides a REST API compatible with WakaTime clients. All endpoints return JSON. Authentication via API key (header) or session cookie.

## Base URL

- Development: `http://localhost:3000/api`
- Production: `https://your-domain.com/api`

## Authentication

### API Key (for editor plugins)
```
Authorization: Basic <base64(api_key)>
```

### Session Cookie (for web UI)
Automatic cookie-based authentication after login.

## Core API Endpoints

### Heartbeats

**POST /api/heartbeat**  
**POST /api/heartbeats**

Submit coding activity heartbeats.

**Request Body**:
```json
[{
  "entity": "/path/to/file.go",
  "type": "file",
  "category": "coding",
  "project": "waka3x",
  "branch": "main",
  "language": "Go",
  "is_write": true,
  "editor": "VS Code",
  "operating_system": "macOS",
  "machine": "MacBook-Pro",
  "time": 1621234567.123
}]
```

**Response**: `201 Created`

**Handler**: `routes/api/heartbeat.go`

### Summary

**GET /api/summary**

Get aggregated statistics for a time range.

**Query Parameters**:
- `from` (required): Start date (ISO 8601 or Unix timestamp)
- `to` (required): End date (ISO 8601 or Unix timestamp)
- `recompute` (optional): Force recalculation (boolean)

**Response**:
```json
{
  "user_id": "user-123",
  "from": "2026-05-01T00:00:00Z",
  "to": "2026-05-17T23:59:59Z",
  "projects": [
    {"key": "waka3x", "total": 3600000000000}
  ],
  "languages": [
    {"key": "Go", "total": 2400000000000}
  ],
  "editors": [...],
  "operating_systems": [...],
  "machines": [...]
}
```

**Handler**: `routes/api/summary.go`

### Authentication

**POST /api/login**

Login with username/password.

**Request Body**:
```json
{
  "username": "user@example.com",
  "password": "password123"
}
```

**Response**: `200 OK` + session cookie

**POST /api/signup**

Create new account.

**Request Body**:
```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "password123",
  "password_repeat": "password123",
  "location": "America/New_York"
}
```

**Response**: `201 Created`

**GET /api/logout**

Logout current session.

**Response**: `302 Redirect` to landing page

**Handler**: `routes/api/auth.go`

### User Settings

**GET /api/settings**

Get current user settings.

**Response**: User object with settings

**POST /api/settings**

Update user settings.

**Request Body**: Partial user object with fields to update

**Response**: `200 OK`

**Handler**: `routes/api/settings.go`

### Projects

**GET /api/projects**

List all projects for current user.

**Query Parameters**:
- `from` (optional): Filter by date range
- `to` (optional): Filter by date range

**Response**:
```json
[
  {"name": "waka3x", "total": 3600000000000},
  {"name": "other-project", "total": 1800000000000}
]
```

**Handler**: `routes/api/projects.go`

### Leaderboard

**GET /api/leaderboard**

Get public leaderboard rankings.

**Response**:
```json
[
  {
    "user": {"id": "user-1", "username": "alice"},
    "rank": 1,
    "total": 7200000000000
  },
  {
    "user": {"id": "user-2", "username": "bob"},
    "rank": 2,
    "total": 5400000000000
  }
]
```

**Handler**: `routes/api/leaderboard.go`

### Badges

**GET /api/badge/:user/interval/:interval**

Generate SVG badge for user statistics.

**Path Parameters**:
- `user`: Username or user ID
- `interval`: `today`, `week`, `month`, `year`, `all_time`

**Response**: SVG image

**Handler**: `routes/api/badge.go`

### Health Check

**GET /api/health**

Check API and database health.

**Response**:
```json
{
  "status": "ok",
  "database": "connected"
}
```

**Handler**: `routes/api/health.go`

## WakaTime Compatibility API

Base path: `/api/compat/wakatime/v1`

### Status Bar

**GET /api/compat/wakatime/v1/users/current/statusbar/today**

Get today's coding time for editor status bar.

**Response**:
```json
{
  "data": {
    "grand_total": {
      "digital": "2 hrs 30 mins",
      "hours": 2,
      "minutes": 30,
      "total_seconds": 9000
    }
  }
}
```

**Handler**: `routes/compat/wakatime/v1/statusbar.go`

### All Time Stats

**GET /api/compat/wakatime/v1/users/current/all_time_since_today**

Get total coding time since account creation.

**Handler**: `routes/compat/wakatime/v1/all.go`

### Summaries

**GET /api/compat/wakatime/v1/users/current/summaries**

Get daily summaries for date range.

**Query Parameters**:
- `start`: Start date (YYYY-MM-DD)
- `end`: End date (YYYY-MM-DD)

**Handler**: `routes/compat/wakatime/v1/summaries.go`

### Stats

**GET /api/compat/wakatime/v1/users/current/stats/:range**

Get statistics for predefined range.

**Path Parameters**:
- `range`: `last_7_days`, `last_30_days`, `last_6_months`, `last_year`

**Handler**: `routes/compat/wakatime/v1/stats.go`

### Heartbeats

**POST /api/compat/wakatime/v1/users/current/heartbeats**  
**POST /api/compat/wakatime/v1/users/current/heartbeats.bulk**

Submit heartbeats (WakaTime format).

**Handler**: `routes/compat/wakatime/v1/heartbeat.go`

## Shields.io Compatibility API

Base path: `/api/compat/shields/v1`

**GET /api/compat/shields/v1/:user/:interval**

Generate shields.io compatible badge JSON.

**Path Parameters**:
- `user`: Username
- `interval`: `today`, `week`, `month`, `year`, `all_time`

**Response**:
```json
{
  "schemaVersion": 1,
  "label": "coding time",
  "message": "2 hrs 30 mins",
  "color": "blue"
}
```

**Handler**: `routes/compat/shields/v1/badge.go`

## Metrics API

**GET /api/metrics**

Prometheus metrics endpoint (if enabled).

**Response**: Prometheus text format

**Handler**: `routes/api/metrics.go`

## Error Responses

All endpoints return standard error format:

```json
{
  "error": "Error message description"
}
```

**Common Status Codes**:
- `200 OK` - Success
- `201 Created` - Resource created
- `400 Bad Request` - Invalid input
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Rate Limiting

Not implemented in application. Use reverse proxy (nginx) for rate limiting.

## CORS

Configured via `rs/cors` middleware. Default: allow all origins in development, restrict in production.

## API Documentation

**Swagger UI**: Available at `/swagger-ui/` when running

**OpenAPI Spec**: Generated from code annotations using swaggo/swag

## Key Handler Files

- `routes/api/heartbeat.go` - Heartbeat ingestion
- `routes/api/summary.go` - Summary retrieval
- `routes/api/auth.go` - Authentication
- `routes/api/settings.go` - User settings
- `routes/api/badge.go` - Badge generation
- `routes/api/projects.go` - Project listing
- `routes/api/leaderboard.go` - Leaderboard
- `routes/api/health.go` - Health check
- `routes/api/metrics.go` - Prometheus metrics
- `routes/compat/wakatime/v1/*.go` - WakaTime compatibility
- `routes/compat/shields/v1/*.go` - Shields.io compatibility

## Testing API Endpoints

### Using curl

```bash
# Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"user","password":"pass"}' \
  -c cookies.txt

# Get summary (with session)
curl http://localhost:3000/api/summary?from=2026-05-01&to=2026-05-17 \
  -b cookies.txt

# Submit heartbeat (with API key)
curl -X POST http://localhost:3000/api/heartbeat \
  -H "Authorization: Basic $(echo -n 'your-api-key' | base64)" \
  -H "Content-Type: application/json" \
  -d '[{"entity":"/file.go","time":1621234567}]'
```

### Using Bruno

API collection available in `bruno/` directory.
