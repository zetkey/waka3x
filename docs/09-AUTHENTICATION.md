# Waka3x Authentication

**AI Agent Reference Document**  
Last Updated: 2026-05-17

## Overview

Waka3x supports multiple authentication methods: local (username/password), OIDC (OpenID Connect), and WebAuthn (FIDO2). Each method can be enabled/disabled independently.

## Authentication Methods

### 1. Local Authentication (Username/Password)

**Default**: Enabled

**Configuration**:
```yaml
security:
  disable_local_auth: false
  password_salt: your-secret-salt  # REQUIRED
```

**Password Hashing**: Argon2id (secure, memory-hard)

**Flow**:
1. User submits username/password to `/api/login`
2. Server verifies credentials
3. Server creates session cookie
4. Client stores cookie for subsequent requests

**Endpoints**:
- `POST /api/login` - Login
- `POST /api/signup` - Register new account
- `GET /api/logout` - Logout
- `POST /api/reset-password` - Request password reset
- `POST /api/set-password` - Set new password with token

**Handler**: `routes/api/auth.go`

### 2. API Key Authentication

**Purpose**: For editor plugins and programmatic access

**Format**: 
```
Authorization: Basic <base64(api_key)>
```

**API Key Storage**: 
- Primary key: `users.api_key` (unique, auto-generated)
- Additional keys: `api_keys` table (optional, named keys)

**Flow**:
1. Client sends API key in Authorization header
2. Middleware extracts and validates key
3. User loaded from database
4. Request proceeds with authenticated user

**Middleware**: `middlewares/authenticate.go`

**Key Management**:
- View API key: Settings page
- Generate additional keys: `POST /api/settings/api-keys`
- Revoke keys: `DELETE /api/settings/api-keys/:id`

### 3. OIDC (OpenID Connect)

**Purpose**: Single Sign-On with external providers (Google, GitHub, etc.)

**Configuration**:
```yaml
security:
  oidc_allow_signup: true  # Allow new user registration via OIDC
  oidc_insecure: false  # Skip TLS verification (dev only)

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

**Flow**:
1. User clicks "Login with Google" button
2. Redirected to `/oidc/google/login`
3. Redirected to Google OAuth consent screen
4. User authorizes application
5. Google redirects to `/oidc/google/callback` with code
6. Server exchanges code for tokens
7. Server extracts user info from ID token
8. Server creates/links user account
9. Server creates session cookie
10. User redirected to dashboard

**User Linking**:
- OIDC users identified by `sub` (subject) claim
- Stored in `users.sub` with `users.auth_type = 'oidc:<provider>'`
- Unique constraint on (auth_type, sub)

**Endpoints**:
- `GET /oidc/{provider}/login` - Initiate OIDC flow
- `GET /oidc/{provider}/callback` - Handle OIDC callback

**Handler**: `routes/api/auth.go`, `routes/utils/oidc.go`

**Supported Providers**: Any OIDC-compliant provider (Google, GitHub, Keycloak, etc.)

### 4. WebAuthn (FIDO2)

**Purpose**: Passwordless authentication with security keys, biometrics

**Configuration**:
```yaml
security:
  disable_webauthn: false
```

**Flow (Registration)**:
1. User navigates to Settings → Security
2. Clicks "Add Security Key"
3. Server generates challenge
4. Browser prompts for security key/biometric
5. User authenticates with device
6. Browser sends credential to server
7. Server verifies and stores credential

**Flow (Login)**:
1. User enters username on login page
2. Clicks "Use Security Key"
3. Server generates challenge
4. Browser prompts for security key/biometric
5. User authenticates with device
6. Browser sends assertion to server
7. Server verifies signature
8. Server creates session cookie

**Credential Storage**: `webauthn_credentials` table

**Endpoints**:
- `POST /api/webauthn/register/begin` - Start registration
- `POST /api/webauthn/register/finish` - Complete registration
- `POST /api/webauthn/login/begin` - Start login
- `POST /api/webauthn/login/finish` - Complete login

**Handler**: `routes/api/auth.go`

**Library**: `github.com/go-webauthn/webauthn`

## Session Management

### Session Storage

**Method**: Encrypted cookies (gorilla/sessions)

**Configuration**:
```yaml
security:
  cookie_max_age: 172800  # 48 hours in seconds
  insecure_cookies: true  # Set false for HTTPS
```

**Cookie Name**: `waka3x_session`

**Cookie Attributes**:
- `HttpOnly: true` - Prevent JavaScript access
- `Secure: !insecure_cookies` - HTTPS only (if configured)
- `SameSite: Lax` - CSRF protection
- `MaxAge: cookie_max_age` - Expiration time

### Session Data

Stored in cookie (encrypted):
- User ID
- Username
- Login timestamp
- CSRF token

### Session Validation

**Middleware**: `middlewares/authenticate.go`

**Process**:
1. Extract session cookie
2. Decrypt and validate
3. Check expiration
4. Load user from database
5. Store user in request context

**Context Key**: `"principal"` - Access via `middlewares.GetPrincipal(r)`

## Authentication Middleware

**File**: `middlewares/authenticate.go`

**Usage**:
```go
router.Group(func(r chi.Router) {
    r.Use(middlewares.NewAuthenticateMiddleware(userService))
    r.Get("/api/summary", summaryHandler.Get)
})
```

**Behavior**:
1. Check for API key in Authorization header
2. If no API key, check for session cookie
3. If neither, return 401 Unauthorized
4. If valid, load user and store in context
5. Call next handler

**Optional Authentication**:
```go
r.Use(middlewares.NewOptionalAuthenticateMiddleware(userService))
```
Allows both authenticated and anonymous access.

## User Registration

### Signup Flow

**Endpoint**: `POST /api/signup`

**Request**:
```json
{
  "username": "newuser",
  "email": "user@example.com",
  "password": "password123",
  "password_repeat": "password123",
  "location": "America/New_York",
  "invite_code": "optional-code"
}
```

**Validation**:
- Username: 3-32 chars, alphanumeric + underscore
- Email: Valid email format
- Password: 8+ chars
- Password repeat: Must match password
- Invite code: Required if `invite_codes: true` and `allow_signup: false`

**Process**:
1. Validate input
2. Check if username/email already exists
3. Verify invite code (if required)
4. Hash password with Argon2id
5. Generate API key (UUID)
6. Create user record
7. Send welcome email (if mail enabled)
8. Create session
9. Return success

**Configuration**:
```yaml
security:
  allow_signup: true  # Allow public registration
  signup_captcha: false  # Require CAPTCHA
  invite_codes: true  # Require invite code when signup disabled
```

### Invite Codes

**Purpose**: Control registration when public signup disabled

**Generation**: Admin creates codes via settings

**Storage**: `key_value_store` table with key prefix `invite_code:`

**Usage**: User provides code during signup

## Password Reset

### Reset Flow

**Endpoints**:
- `POST /api/reset-password` - Request reset
- `POST /api/set-password` - Set new password

**Process**:
1. User submits email to `/api/reset-password`
2. Server generates reset token (UUID)
3. Server stores token in `users.reset_token`
4. Server sends email with reset link
5. User clicks link (contains token)
6. User submits new password with token
7. Server validates token
8. Server updates password
9. Server clears reset token

**Token Expiration**: 1 hour (hardcoded)

**Email Template**: HTML email with reset link

## Security Considerations

### Password Security

- **Hashing**: Argon2id (memory-hard, resistant to GPU attacks)
- **Salt**: Configured in `security.password_salt`
- **Parameters**: Time=1, Memory=64MB, Threads=4, KeyLen=32
- **Library**: `github.com/alexedwards/argon2id`

### Session Security

- **Encryption**: AES-256-GCM (gorilla/securecookie)
- **CSRF Protection**: SameSite=Lax cookie attribute
- **HttpOnly**: Prevents XSS attacks
- **Secure**: HTTPS-only in production

### API Key Security

- **Format**: UUID v4 (128-bit random)
- **Storage**: Plain text in database (not hashed)
- **Transmission**: Base64-encoded in Authorization header
- **Rotation**: Users can regenerate keys

### Rate Limiting

**Not implemented** - Use reverse proxy (nginx) for rate limiting

**Recommended**:
- Login: 5 attempts per 15 minutes per IP
- Signup: 3 attempts per hour per IP
- Password reset: 3 attempts per hour per email

## Multi-Factor Authentication

**Status**: Not implemented

**Future**: Could add TOTP (Time-based One-Time Password) support

## Trusted Header Authentication

**Configuration**:
```yaml
security:
  trusted_header_auth: false  # DANGEROUS - only for reverse proxy setups
```

**Purpose**: Trust authentication headers from reverse proxy (e.g., `X-Forwarded-User`)

**WARNING**: Only enable if reverse proxy properly authenticates users and sets headers. Misconfiguration allows authentication bypass.

## Admin Users

**Flag**: `users.is_admin = true`

**Capabilities**:
- View all users
- Delete users
- Access admin endpoints

**Creation**: Set manually in database or via first user

## Testing Authentication

### Local Auth
```bash
curl -X POST http://localhost:3000/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test"}' \
  -c cookies.txt
```

### API Key
```bash
curl http://localhost:3000/api/summary \
  -H "Authorization: Basic $(echo -n 'your-api-key' | base64)"
```

### Session
```bash
curl http://localhost:3000/api/summary -b cookies.txt
```

## Troubleshooting

**Login fails**: Check password, username, database connection

**Session not persisting**: Check `cookie_max_age`, `insecure_cookies` (must be true for HTTP)

**API key not working**: Check format, Base64 encoding, key exists in database

**OIDC fails**: Check provider config, redirect URL, client ID/secret

**WebAuthn fails**: Requires HTTPS (or localhost), check browser support
