# Waka3x Development Guide

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Prerequisites

- **Go**: 1.26 or later
- **Bun**: Latest version (for frontend)
- **Database**: SQLite (default), MySQL, or PostgreSQL
- **Git**: For version control

## Initial Setup

### Clone Repository

```bash
git clone https://github.com/zetkey/waka3x.git
cd waka3x
```

### Backend Setup

```bash
# Install Go dependencies
go mod download

# Copy default config
cp config.default.yml waka3x.yml

# Edit configuration (optional)
vim waka3x.yml

# Run backend
go run main.go
```

Backend runs on `http://localhost:3000`

### Frontend Setup

```bash
cd frontend

# Install dependencies
bun install

# Run development server
bun dev
```

Frontend dev server runs on `http://localhost:5173` (proxies API to :3000)

### Database Setup

**SQLite (default)**: No setup needed, auto-creates `waka3x.db`

**MySQL**:
```yaml
db:
  dialect: mysql
  host: localhost
  port: 3306
  user: waka3x
  password: password
  name: waka3x
```

**PostgreSQL**:
```yaml
db:
  dialect: postgres
  host: localhost
  port: 5432
  user: waka3x
  password: password
  name: waka3x
```

## Development Workflow

### Running Backend

```bash
# Development mode (with hot reload using air)
air

# Or standard run
go run main.go

# With custom config
go run main.go -config custom.yml
```

### Running Frontend

```bash
cd frontend
bun dev
```

**Hot Module Replacement**: Changes auto-reload in browser

### Running Both

**Terminal 1**: `go run main.go`  
**Terminal 2**: `cd frontend && bun dev`

Access app at `http://localhost:5173`

## Building

### Backend Build

```bash
# Build binary
go build -o waka3x

# Run binary
./waka3x -config waka3x.yml
```

### Frontend Build

```bash
cd frontend
bun build
```

Output: `frontend/dist/` (embedded in Go binary)

### Full Build

```bash
# Build frontend
cd frontend && bun build && cd ..

# Build backend (embeds frontend)
go build -o waka3x

# Run
./waka3x
```

## Testing

### Backend Tests

```bash
# Run all tests
go test ./...

# Run specific package
go test ./services

# Run with coverage
go test -cover ./...

# Verbose output
go test -v ./...
```

### Frontend Tests

Currently no automated tests. Manual testing via browser.

## Code Quality

### Go Formatting

```bash
# Format all Go files
go fmt ./...

# Or use gofmt directly
gofmt -w .
```

### Frontend Formatting

```bash
cd frontend

# Format with Prettier
bun format

# Lint with ESLint
bun lint
```

## Database Migrations

**Auto-migration**: Runs on startup (GORM AutoMigrate)

**Disable migrations**:
```yaml
skip_migrations: true
```

**Manual migration**:
```bash
# Migrations run automatically on startup
go run main.go
```

**Migration files**: `migrations/` directory

## Configuration

### Development Config

```yaml
env: development
server:
  port: 3000
  listen_ipv4: 127.0.0.1
db:
  dialect: sqlite3
  name: waka3x.db
security:
  insecure_cookies: true
  allow_signup: true
```

### Environment Variables

Override config with env vars:
```bash
export WAKA3X_DB_DIALECT=postgres
export WAKA3X_DB_HOST=localhost
export WAKA3X_PASSWORD_SALT=your-secret-salt
```

## Debugging

### Backend Debugging

**Enable debug logging**:
```yaml
env: development
```

**pprof profiling**:
```yaml
enable_pprof: true
```

Access at `http://localhost:6060/debug/pprof`

**Sentry error tracking**:
```yaml
sentry:
  dsn: your-sentry-dsn
```

### Frontend Debugging

**Vue DevTools**: Install browser extension

**Console logging**: Use `console.log()` in components

**Network inspection**: Browser DevTools Network tab

## Docker Development

### Build Docker Image

```bash
docker build -t waka3x:dev .
```

### Run with Docker Compose

```bash
# Copy environment file
cp .env.sample .env

# Edit .env with your values
vim .env

# Start services
docker compose up -d

# View logs
docker compose logs -f

# Stop services
docker compose down
```

## API Testing

### Using Bruno

API collection in `bruno/` directory:

```bash
# Open Bruno and import collection
# Collections: Heartbeats, Summary, Auth, Settings
```

### Using curl

```bash
# Login
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}' \
  -c cookies.txt

# Get summary
curl http://localhost:3000/api/summary?from=2026-05-01&to=2026-05-17 \
  -b cookies.txt
```

## Common Development Tasks

### Add New API Endpoint

1. Create handler in `routes/api/`
2. Implement handler methods
3. Register routes in `RegisterRoutes()`
4. Update `main.go` to initialize handler
5. Add Swagger annotations
6. Test endpoint

### Add New Model

1. Define struct in `models/`
2. Add GORM tags
3. Create repository interface in `repositories/`
4. Implement repository
5. Update migrations if needed
6. Test CRUD operations

### Add New Service

1. Define interface in `services/`
2. Implement service struct
3. Add constructor with dependencies
4. Implement business logic methods
5. Update `main.go` to initialize service
6. Write unit tests

### Add New Frontend View

1. Create component in `src/views/`
2. Add route in `src/router/index.ts`
3. Add navigation link in layout
4. Create API calls in `src/lib/api.ts`
5. Test in browser

## Troubleshooting

### Backend won't start

- Check config file syntax (YAML)
- Verify database connection
- Check port 3000 not in use
- Review logs for errors

### Frontend won't build

- Delete `node_modules` and reinstall: `bun install`
- Clear Vite cache: `rm -rf frontend/.vite`
- Check for TypeScript errors: `bun build`

### Database errors

- Check database credentials
- Verify database exists
- Check migrations ran successfully
- Review GORM logs

### API returns 401 Unauthorized

- Check API key is correct
- Verify session cookie is set
- Check authentication middleware

## Performance Optimization

### Backend

- Enable database connection pooling
- Use indexes on frequently queried columns
- Cache frequently accessed data
- Use batch inserts for heartbeats

### Frontend

- Lazy load routes
- Optimize bundle size
- Use production build
- Enable compression in reverse proxy

## Deployment

### Production Build

```bash
# Build frontend
cd frontend && bun build && cd ..

# Build backend
go build -o waka3x

# Copy files to server
scp waka3x config.yml user@server:/opt/waka3x/
```

### SystemD Service

```bash
# Copy service file
sudo cp etc/waka3x.service /etc/systemd/system/

# Edit service file
sudo vim /etc/systemd/system/waka3x.service

# Enable and start
sudo systemctl enable waka3x
sudo systemctl start waka3x

# Check status
sudo systemctl status waka3x
```

### Docker Deployment

```bash
# Build image
docker build -t waka3x:latest .

# Run container
docker run -d \
  -p 3000:3000 \
  -v waka3x-data:/data \
  -e WAKA3X_PASSWORD_SALT=your-salt \
  --name waka3x \
  waka3x:latest
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name waka3x.example.com;
    
    location / {
        proxy_pass http://localhost:3000;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## Git Workflow

### Branch Strategy

- `master` - Main branch (stable)
- `refactor/vue-frontend` - Current development branch
- Feature branches: `feature/name`
- Bug fixes: `fix/name`

### Commit Messages

Use conventional commits:
- `feat:` - New feature
- `fix:` - Bug fix
- `refactor:` - Code refactoring
- `docs:` - Documentation
- `test:` - Tests
- `build:` - Build system

Example: `feat: add weekly report email feature`

## Resources

- **Go Documentation**: https://golang.org/doc/
- **Vue 3 Documentation**: https://vuejs.org/
- **GORM Documentation**: https://gorm.io/
- **Vite Documentation**: https://vitejs.dev/
- **WakaTime API**: https://wakatime.com/developers
