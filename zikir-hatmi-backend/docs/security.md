# Security Documentation

This document describes the security features and configuration options for the Zikir Hatmi application.

## Security Features

### 1. WebSocket Origin Validation

The WebSocket endpoint now validates the `Origin` header to prevent cross-site WebSocket hijacking attacks.

**Configuration:**
- Default allowed origins: `localhost`, `127.0.0.1`
- Additional origins can be configured via the `ALLOWED_ORIGINS` environment variable (comma-separated)

Example:
```bash
export ALLOWED_ORIGINS="zikirhatmi.example.com,app.example.com"
```

### 2. Rate Limiting

The `/hatims/{shareCode}/join` endpoint is protected against brute force attacks with rate limiting:
- **Limit:** 10 attempts per IP address per hatim per minute
- Returns HTTP 429 (Too Many Requests) when limit is exceeded

### 3. Token Expiration

Authentication tokens now have a 30-day expiration period:
- Tokens are automatically invalidated after 30 days
- Users must re-authenticate (join the hatim again) after token expiration
- The database schema includes `expires_at` column in `hatim_tokens` table

### 4. Authorization on Admin Endpoints

Write operations on hatims require valid authentication:

#### PATCH /hatims/{shareCode}
- Requires `Authorization: Bearer <token>` header
- Token must be valid for the specific hatim

#### DELETE /hatims/{shareCode}
- Requires `Authorization: Bearer <token>` header
- Token must be valid for the specific hatim

#### GET /hatims (List all hatims)
- Requires admin authentication
- Set `ADMIN_KEY` environment variable to enable this endpoint
- Request must include `Authorization: Bearer <admin_key>` header
- Returns HTTP 403 if `ADMIN_KEY` is not configured

Example:
```bash
export ADMIN_KEY="your-secure-admin-key"
```

### 5. Secure Random Generation

Client session IDs are generated using cryptographically secure random numbers (`crypto/rand`) instead of pseudo-random numbers (`math/rand`).

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ALLOWED_ORIGINS` | Comma-separated list of allowed WebSocket origins | localhost,127.0.0.1 |
| `ADMIN_KEY` | Admin key for listing all hatims | (disabled) |
| `TRUST_PROXY` | Trust X-Forwarded-For headers (set to "true" when behind reverse proxy) | false |
| `DATABASE_URL` | PostgreSQL connection string | (required) |
| `PORT` | Server port | 8080 |

## Security Best Practices

### Production Deployment

1. **Set ALLOWED_ORIGINS**: Configure only your production domain(s)
   ```bash
   export ALLOWED_ORIGINS="yourdomain.com"
   ```

2. **Set ADMIN_KEY**: Use a strong, random admin key
   ```bash
   export ADMIN_KEY="$(openssl rand -base64 32)"
   ```

3. **Use HTTPS**: Deploy behind a reverse proxy (nginx, Caddy) with TLS

4. **Database Security**: Use strong database credentials and enable SSL

### Password Requirements

Passwords for protected hatims are hashed using Argon2id with:
- Memory: 64 MB
- Iterations: 1
- Parallelism: 4
- Key length: 32 bytes

## Security Vulnerabilities Fixed

The following security issues were identified and fixed:

1. **WebSocket Origin Bypass (High)**: `CheckOrigin` always returned `true`
   - Fixed: Proper origin validation with configurable allowed origins

2. **Missing Rate Limiting (Medium)**: No protection against brute force attacks
   - Fixed: Rate limiting on join endpoint (10 attempts/min/IP/hatim)

3. **Tokens Never Expire (Medium)**: Authentication tokens had no expiration
   - Fixed: 30-day token expiration

4. **No Authorization on Admin Endpoints (Critical)**: PATCH/DELETE had no auth
   - Fixed: Token-based authorization required

5. **Public Exposure of All Hatims (High)**: GET /hatims exposed all data
   - Fixed: Admin key required for listing

6. **Weak Random Generation (Low)**: Using time-seeded math/rand
   - Fixed: Using crypto/rand for session IDs
