# Waka3x Configuration

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

Waka3x uses YAML configuration files with environment variable overrides. Default config: `config.default.yml`. User config: `waka3x.yml`.

## Configuration File

**Location**: `waka3x.yml` (or specify with `-config` flag)

**Format**: YAML

**Loading Order**:
1. Load `config.default.yml` defaults
2. Override with `waka3x.yml` if exists
3. Override with environment variables

## Core Settings

### Environment

```yaml
env: production  # production, development
quick_start: false  # skip initial tasks on startup
skip_migrations: false  # skip database migrations
enable_pprof: false  # enable profiling endpoint
```

### Server

```yaml
server:
  listen_ipv4: 127.0.0.1  # IPv4 address (or '-' to disable)
  listen_ipv6: ::1  # IPv6 address (or '-' to disable)
  listen_socket: ""  # Unix socket path (or '-' to disable)
  listen_socket_mode: 0666  # Socket permissions
  port: 3000
  base_path: /  # Base URL path
  public_url: http://localhost:3000  # Public URL for emails
  timeout_sec: 30  # Request timeout
  tls_cert_path: ""  # TLS certificate (leave blank for HTTP)
  tls_key_path: ""  # TLS key (leave blank for HTTP)
  log_format: text  # text or json
```

### Database

```yaml
db:
  dialect: sqlite3  # sqlite3, mysql, postgres
  host: ""  # Database host (not for SQLite)
  port: 0  # Database port (not for SQLite)
  socket: ""  # Unix socket (alternative to host)
  user: ""  # Database user (not for SQLite)
  password: ""  # Database password (not for SQLite)
  name: waka3x.db  # Database name or file path
  charset: utf8mb4  # MySQL charset
  max_conn: 10  # Max concurrent connections
  ssl: false  # Enable TLS (Postgres)
  compress: false  # Enable compression (MySQL)
  automigrate_fail_silently: false  # Ignore migration errors
```

**SQLite Example**:
```yaml
db:
  dialect: sqlite3
  name: /data/waka3x.db
```

**MySQL Example**:
```yaml
db:
  dialect: mysql
  host: localhost
  port: 3306
  user: waka3x
  password: secret
  name: waka3x
  charset: utf8mb4
```

**PostgreSQL Example**:
```yaml
db:
  dialect: postgres
  host: localhost
  port: 5432
  user: waka3x
  password: secret
  name: waka3x
  ssl: true
```

### Application

```yaml
app:
  # Leaderboard
  leaderboard_enabled: true
  leaderboard_scope: 7_days  # Time range for rankings
  leaderboard_generation_time: '0 0 6 * * *,0 0 18 * * *'  # Cron schedule
  leaderboard_require_auth: false  # Require login to view
  
  # Background Jobs
  aggregation_time: '0 15 2 * * *'  # Daily summary generation
  report_time_weekly: '0 0 18 * * 5'  # Weekly email reports (Friday 6 PM)
  data_cleanup_time: '0 0 6 * * 0'  # Data cleanup (Sunday 6 AM)
  optimize_database_time: '0 0 8 1 * *'  # DB optimization (1st of month)
  
  # User Activity
  inactive_days: 7  # Days to consider user active
  max_inactive_months: 12  # Delete inactive users after N months
  
  # Data Import
  import_enabled: true
  import_backoff_min: 5  # Cooldown between imports (minutes)
  import_max_rate: 24  # Min hours between successful imports
  import_batch_size: 50  # Heartbeats per transaction
  import_hosts_whitelist: []  # Allowed import sources (empty = all)
  
  # Data Retention
  heartbeat_max_age: '4320h'  # Max heartbeat age (180 days)
  data_retention_months: -1  # Keep data N months (-1 = forever)
  
  # Misc
  warm_caches: true  # Warm caches on startup
  
  # Language Mappings
  custom_languages:
    vue: Vue
    jsx: JSX
    tsx: TSX
    svelte: Svelte
  
  canonical_language_names:
    java: Java
    php: PHP
    xml: XML
  
  # Avatar URL template
  avatar_url_template: api/avatar/{username_hash}.svg
  
  # Date Formats
  date_format: Mon, 02 Jan 2006
  datetime_format: Mon, 02 Jan 2006 15:04
```

### Security

```yaml
security:
  password_salt: ""  # REQUIRED: Secret salt for password hashing
  insecure_cookies: true  # Set false for HTTPS
  cookie_max_age: 172800  # Session cookie lifetime (seconds)
  
  # Signup
  allow_signup: true  # Allow new user registration
  signup_captcha: false  # Require CAPTCHA on signup
  invite_codes: true  # Enable invite code system
  
  # Authentication
  disable_local_auth: false  # Disable username/password login
  disable_webauthn: false  # Disable WebAuthn/FIDO2
  oidc_allow_signup: true  # Allow OIDC registration
  oidc_insecure: false  # Skip TLS verification for OIDC
  
  # Access Control
  disable_frontpage: false  # Hide landing page
  expose_metrics: false  # Public Prometheus metrics
  trusted_header_auth: false  # Trust reverse proxy headers (DANGEROUS)
  
  # Proxy
  enable_proxy: false  # Enable WakaTime relay proxy
```

**IMPORTANT**: Set `password_salt` to a random string (32+ characters)

### Mail

```yaml
mail:
  enabled: false  # Enable email sending
  provider: smtp  # smtp only
  smtp:
    host: smtp.example.com
    port: 587
    username: user@example.com
    password: secret
    from: noreply@example.com
    tls: true  # Use STARTTLS
```

### Sentry

```yaml
sentry:
  dsn: ""  # Sentry DSN for error tracking
  enable_tracing: false  # Enable performance tracing
  sample_rate: 0.75  # Error sample rate
  sample_rate_traces: 0.75  # Trace sample rate
```

### OIDC Providers

```yaml
oidc_providers:
  - id: google
    name: Google
    issuer: https://accounts.google.com
    client_id: your-client-id
    client_secret: your-client-secret
    redirect_url: http://localhost:3000/oidc/google/callback
    scopes:
      - openid
      - email
      - profile
```

## Environment Variables

Override config with environment variables using `WAKA3X_` prefix:

```bash
# Database
export WAKA3X_DB_DIALECT=postgres
export WAKA3X_DB_HOST=localhost
export WAKA3X_DB_PORT=5432
export WAKA3X_DB_USER=waka3x
export WAKA3X_DB_PASSWORD=secret
export WAKA3X_DB_NAME=waka3x

# Security
export WAKA3X_PASSWORD_SALT=your-random-salt

# Server
export WAKA3X_SERVER_PORT=3000
export WAKA3X_SERVER_PUBLIC_URL=https://waka3x.example.com

# Mail
export WAKA3X_MAIL_SMTP_HOST=smtp.example.com
export WAKA3X_MAIL_SMTP_PORT=587
export WAKA3X_MAIL_SMTP_USERNAME=user@example.com
export WAKA3X_MAIL_SMTP_PASSWORD=secret
```

**Nested Keys**: Use underscores for nesting (e.g., `WAKA3X_DB_HOST`)

## Docker Configuration

### Docker Secrets

For sensitive values in Docker Compose:

```yaml
# compose.yml
secrets:
  password_salt:
    environment: WAKA3X_PASSWORD_SALT
  db_password:
    environment: WAKA3X_DB_PASSWORD
  smtp_password:
    environment: WAKA3X_MAIL_SMTP_PASS
```

```bash
# .env file
WAKA3X_PASSWORD_SALT=your-random-salt
WAKA3X_DB_PASSWORD=db-password
WAKA3X_MAIL_SMTP_PASS=smtp-password
```

### Docker Environment

```bash
docker run -d \
  -e WAKA3X_PASSWORD_SALT=your-salt \
  -e WAKA3X_DB_DIALECT=postgres \
  -e WAKA3X_DB_HOST=postgres \
  -v waka3x-data:/data \
  waka3x:latest
```

## Configuration Validation

**On Startup**: Config is validated, errors logged

**Required Fields**:
- `security.password_salt` (production)

**Warnings**:
- `insecure_cookies: true` in production
- `allow_signup: true` without invite codes
- Missing mail config when reports enabled

## Common Configurations

### Development (Local)

```yaml
env: development
server:
  port: 3000
  listen_ipv4: 127.0.0.1
  public_url: http://localhost:3000
db:
  dialect: sqlite3
  name: waka3x.db
security:
  insecure_cookies: true
  allow_signup: true
  password_salt: dev-salt-change-in-production
```

### Production (Docker + PostgreSQL)

```yaml
env: production
server:
  port: 3000
  listen_ipv4: 0.0.0.0
  public_url: https://waka3x.example.com
db:
  dialect: postgres
  host: postgres
  port: 5432
  user: waka3x
  password: ${WAKA3X_DB_PASSWORD}
  name: waka3x
  ssl: true
security:
  insecure_cookies: false
  allow_signup: false
  invite_codes: true
  password_salt: ${WAKA3X_PASSWORD_SALT}
mail:
  enabled: true
  smtp:
    host: smtp.example.com
    port: 587
    username: noreply@example.com
    password: ${WAKA3X_MAIL_SMTP_PASS}
```

### Production (Standalone + MySQL)

```yaml
env: production
server:
  port: 3000
  listen_ipv4: 127.0.0.1
  public_url: https://waka3x.example.com
db:
  dialect: mysql
  host: localhost
  port: 3306
  user: waka3x
  password: ${WAKA3X_DB_PASSWORD}
  name: waka3x
  charset: utf8mb4
security:
  insecure_cookies: false
  password_salt: ${WAKA3X_PASSWORD_SALT}
```

## Cron Schedule Format

Extended cron format (6 fields):

```
┌───────────── second (0-59)
│ ┌───────────── minute (0-59)
│ │ ┌───────────── hour (0-23)
│ │ │ ┌───────────── day of month (1-31)
│ │ │ │ ┌───────────── month (1-12)
│ │ │ │ │ ┌───────────── day of week (0-6, 0=Sunday)
│ │ │ │ │ │
* * * * * *
```

**Examples**:
- `0 15 2 * * *` - Daily at 2:15 AM
- `0 0 18 * * 5` - Friday at 6 PM
- `0 0 6 * * 0` - Sunday at 6 AM
- `0 0 8 1 * *` - 1st of month at 8 AM

## Configuration Files Location

- **Default**: `config.default.yml` (in repository)
- **User**: `waka3x.yml` (create from default)
- **Custom**: Specify with `-config` flag

```bash
./waka3x -config /etc/waka3x/config.yml
```

## Security Best Practices

1. **Always set `password_salt`** to random string (32+ chars)
2. **Use `insecure_cookies: false`** in production (requires HTTPS)
3. **Disable `allow_signup`** or enable `invite_codes` for private instances
4. **Use environment variables** for secrets (don't commit to git)
5. **Enable `ssl: true`** for database connections in production
6. **Set `trusted_header_auth: false`** unless you know what you're doing
7. **Use strong database passwords**
8. **Restrict `listen_ipv4`** to localhost if behind reverse proxy

## Troubleshooting

**Config not loading**: Check YAML syntax, file permissions

**Database connection fails**: Verify credentials, host, port

**Emails not sending**: Check SMTP settings, test with telnet

**Sessions not persisting**: Check `cookie_max_age`, `insecure_cookies`

**Migrations failing**: Check database permissions, set `automigrate_fail_silently: true` (not recommended)
