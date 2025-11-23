# Stripe Environment Variables

## API Keys

### Secret Keys (Server-side)
- `STRIPE_SECRET_KEY`
- `STRIPE_API_KEY`
- `STRIPE_SK`
- `STRIPE_SECRET`
- Pattern: `sk_test_*` (test mode)
- Pattern: `sk_live_*` (live mode)
- Pattern: `rk_test_*` (restricted keys test)
- Pattern: `rk_live_*` (restricted keys live)

### Publishable Keys (Client-side)
- `STRIPE_PUBLISHABLE_KEY`
- `STRIPE_PUBLIC_KEY`
- `STRIPE_PK`
- `NEXT_PUBLIC_STRIPE_PUBLISHABLE_KEY` (Next.js)
- `REACT_APP_STRIPE_PUBLISHABLE_KEY` (Create React App)
- `VITE_STRIPE_PUBLISHABLE_KEY` (Vite)
- Pattern: `pk_test_*` (test mode)
- Pattern: `pk_live_*` (live mode)

### Webhook Secrets
- `STRIPE_WEBHOOK_SECRET`
- `STRIPE_WEBHOOK_SIGNING_SECRET`
- `STRIPE_ENDPOINT_SECRET`
- Pattern: `whsec_*`

## Additional Configuration

### Account/Connect
- `STRIPE_ACCOUNT_ID`
- `STRIPE_CONNECT_CLIENT_ID`
- `STRIPE_PLATFORM_SECRET_KEY`

### Testing
- `STRIPE_TEST_SECRET_KEY`
- `STRIPE_TEST_PUBLISHABLE_KEY`
- `STRIPE_TEST_MODE`

### API Version
- `STRIPE_API_VERSION`

### Timeout/Retry
- `STRIPE_MAX_NETWORK_RETRIES`
- `STRIPE_TIMEOUT`

## Configuration File Patterns

### .env files
```bash
STRIPE_SECRET_KEY=sk_test_***
STRIPE_PUBLISHABLE_KEY=pk_test_***
STRIPE_WEBHOOK_SECRET=whsec_***
```

### config.yml / application.yml
```yaml
stripe:
  api_key: sk_test_***
  publishable_key: pk_test_***
  webhook_secret: whsec_***
```

### config.json
```json
{
  "stripe": {
    "secretKey": "sk_test_***",
    "publishableKey": "pk_test_***",
    "webhookSecret": "whsec_***"
  }
}
```

### Rails secrets.yml / credentials.yml
```yaml
stripe:
  secret_key: <%= ENV["STRIPE_SECRET_KEY"] %>
  publishable_key: <%= ENV["STRIPE_PUBLISHABLE_KEY"] %>
```

### Django settings.py
```python
STRIPE_SECRET_KEY = os.environ.get('STRIPE_SECRET_KEY')
STRIPE_PUBLISHABLE_KEY = os.environ.get('STRIPE_PUBLISHABLE_KEY')
STRIPE_WEBHOOK_SECRET = os.environ.get('STRIPE_WEBHOOK_SECRET')
```

## Key Prefixes and Patterns

### Secret Keys
- `sk_test_` - Test mode secret key (51 chars after prefix)
- `sk_live_` - Live mode secret key (51 chars after prefix)
- `rk_test_` - Test mode restricted key
- `rk_live_` - Live mode restricted key

### Publishable Keys
- `pk_test_` - Test mode publishable key
- `pk_live_` - Live mode publishable key

### Webhook Secrets
- `whsec_` - Webhook signing secret

### Legacy/Other
- `sk_` - Legacy format
- `ca_` - Connect application identifier

## Security Considerations

### HIGH RISK if found in:
- Source code files
- Git history
- Public repositories
- Client-side JavaScript
- HTML files
- Log files
- Error messages

### Expected Locations:
- Environment variable files (.env, .env.local)
- Secret management systems (AWS Secrets Manager, HashiCorp Vault)
- CI/CD environment configurations
- Server configuration files (with restricted permissions)

## Detection Confidence

- **HIGH**: Environment variables with "STRIPE" and "KEY" or "SECRET"
- **HIGH**: Key patterns matching Stripe format (sk_*, pk_*, whsec_*)
- **MEDIUM**: Configuration files with stripe sections
- **CRITICAL**: Secret keys found in source code or public locations (security issue)
