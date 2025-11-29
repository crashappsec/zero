# Authentication Tokens

## JWT (JSON Web Tokens)

### Structure
```
Pattern: eyJ[A-Za-z0-9_-]*\.eyJ[A-Za-z0-9_-]*\.[A-Za-z0-9_-]*
Example: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U
Severity: high
```

JWTs have three base64url-encoded parts separated by dots:
1. Header (starts with `eyJ`)
2. Payload (starts with `eyJ`)
3. Signature

### Security Concerns
- Exposed JWTs can be replayed until expiration
- May contain sensitive claims (email, roles, permissions)
- Weak secrets allow token forgery

### JWT Secrets
```
Pattern: JWT_SECRET|jwt[_-]?secret
Context: Environment variables, config files
Severity: critical
```

The signing secret is more dangerous than individual tokens.

---

## OAuth Tokens

### OAuth 2.0 Access Tokens
```
Pattern: ya29\.[A-Za-z0-9_-]+ (Google)
Pattern: [A-Za-z0-9_-]{20,}\.[A-Za-z0-9_-]{20,} (Generic)
Severity: high
```

### OAuth Refresh Tokens
```
Pattern: 1//[A-Za-z0-9_-]+ (Google)
Severity: critical
```

Refresh tokens are more sensitive as they're long-lived.

### OAuth Client Secrets
```
Pattern: client_secret['":\s]+['"]?[A-Za-z0-9_-]{24,}['"]?
Severity: critical
```

---

## Session Tokens

### Express/Connect Sessions
```
Pattern: s:[A-Za-z0-9+/=]+\.[A-Za-z0-9+/=]+
Context: connect.sid cookie
Severity: high
```

### PHP Sessions
```
Pattern: PHPSESSID=[a-z0-9]{26,32}
Severity: high
```

### Django Sessions
```
Pattern: sessionid=[a-z0-9]{32}
Severity: high
```

### ASP.NET Sessions
```
Pattern: ASP\.NET_SessionId=[a-z0-9]{24}
Severity: high
```

---

## API Bearer Tokens

### Generic Bearer
```
Pattern: [Bb]earer\s+[A-Za-z0-9_\-.~+/]+=*
Context: Authorization headers
Severity: high
```

### Basic Auth (encoded)
```
Pattern: [Bb]asic\s+[A-Za-z0-9+/]+=*
Context: Authorization headers
Severity: high
```

Can be decoded to reveal username:password.

---

## Service-Specific Tokens

### Auth0
```
Pattern: [A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+
Context: AUTH0_CLIENT_SECRET, auth0.com domains
Severity: high
```

### Firebase
```
Pattern: [0-9]+:[A-Za-z0-9_-]+:[A-Za-z0-9_-]+
Context: Firebase config, FCM tokens
Severity: medium
```

### Supabase
```
Pattern: eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+
Context: SUPABASE_ANON_KEY, SUPABASE_SERVICE_ROLE_KEY
Severity: high (service role), medium (anon)
```

The service role key bypasses Row Level Security.

### Okta
```
Pattern: 00[A-Za-z0-9_-]{40}
Context: OKTA_API_TOKEN
Severity: high
```

### Clerk
```
Pattern: sk_live_[A-Za-z0-9]{40,}
Pattern: sk_test_[A-Za-z0-9]{40,}
Severity: high (live), low (test)
```

---

## CSRF Tokens

### General Pattern
```
Pattern: csrf[_-]?token['":\s]+['"]?[A-Za-z0-9_-]{32,}['"]?
Severity: medium
```

CSRF tokens are session-specific but shouldn't be logged.

---

## Password Reset Tokens

### General Pattern
```
Pattern: reset[_-]?token['":\s]+['"]?[A-Za-z0-9_-]{32,}['"]?
Severity: high
```

These allow account takeover if exposed.

---

## Verification Tokens

### Email Verification
```
Pattern: verify[_-]?token['":\s]+['"]?[A-Za-z0-9_-]{32,}['"]?
Severity: medium
```

### Magic Links
```
Pattern: /auth/magic/[A-Za-z0-9_-]{32,}
Pattern: /login/[A-Za-z0-9_-]{64,}
Severity: high
```

---

## Detection Notes

### Token Characteristics
- Usually high entropy (random-looking)
- Often base64 or base64url encoded
- May have recognizable prefixes
- Time-limited but dangerous during validity

### Context Matters
- Tokens in code vs configuration
- Client-side vs server-side exposure
- Public vs authenticated endpoints

### False Positives
- Example tokens in documentation
- Mock tokens in test files
- Expired tokens in logs (still flag)
- Non-secret base64 data

### Severity Adjustments
- **Refresh tokens > Access tokens**: Longer validity
- **Admin tokens > User tokens**: Higher privilege
- **Production > Development**: Real data access
