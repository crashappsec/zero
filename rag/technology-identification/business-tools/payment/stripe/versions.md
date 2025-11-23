# Stripe SDK Versions and API Versions

## Current Stable Versions (as of 2025)

### Official SDKs

#### Node.js
- **Package**: `stripe`
- **Latest**: v17.x
- **Minimum Supported**: v8.x
- **Node.js Requirements**: Node 12+
- **Repository**: https://github.com/stripe/stripe-node

#### Python
- **Package**: `stripe`
- **Latest**: v11.x
- **Minimum Supported**: v5.x
- **Python Requirements**: Python 3.6+
- **Repository**: https://github.com/stripe/stripe-python

#### Ruby
- **Package**: `stripe`
- **Latest**: v13.x
- **Minimum Supported**: v5.x
- **Ruby Requirements**: Ruby 2.3+
- **Repository**: https://github.com/stripe/stripe-ruby

#### PHP
- **Package**: `stripe/stripe-php`
- **Latest**: v15.x
- **Minimum Supported**: v7.x
- **PHP Requirements**: PHP 7.3+
- **Repository**: https://github.com/stripe/stripe-php

#### Go
- **Package**: `github.com/stripe/stripe-go`
- **Latest**: v78.x
- **Minimum Supported**: v72.x
- **Go Requirements**: Go 1.18+
- **Repository**: https://github.com/stripe/stripe-go

#### Java
- **Package**: `com.stripe:stripe-java`
- **Latest**: v27.x
- **Minimum Supported**: v20.x
- **Java Requirements**: Java 8+
- **Repository**: https://github.com/stripe/stripe-java

#### .NET
- **Package**: `Stripe.net`
- **Latest**: v47.x
- **Minimum Supported**: v40.x
- **.NET Requirements**: .NET Standard 2.0+
- **Repository**: https://github.com/stripe/stripe-dotnet

### Browser SDKs

#### Stripe.js
- **Package**: `@stripe/stripe-js`
- **Latest**: v4.x
- **CDN**: https://js.stripe.com/v3/
- **Repository**: https://github.com/stripe/stripe-js

#### React Stripe.js
- **Package**: `@stripe/react-stripe-js`
- **Latest**: v2.x
- **React Requirements**: React 16.8+
- **Repository**: https://github.com/stripe/react-stripe-js

## API Versions

### Current API Version
- **Latest**: 2024-12-18
- **Recommended**: Use latest or pin to specific version

### Recent API Versions
- 2024-12-18
- 2024-11-20
- 2024-10-28
- 2024-09-30
- 2024-06-20
- 2023-10-16
- 2023-08-16
- 2022-11-15

### API Version Compatibility
- SDKs are forward-compatible
- New API versions may introduce breaking changes
- Always test before upgrading API version
- Can specify version per request or globally

### Setting API Version

#### Node.js
```javascript
const stripe = require('stripe')('sk_test_***', {
  apiVersion: '2024-12-18'
});
```

#### Python
```python
stripe.api_version = '2024-12-18'
```

#### Ruby
```ruby
Stripe.api_version = '2024-12-18'
```

#### PHP
```php
\Stripe\Stripe::setApiVersion('2024-12-18');
```

#### Go
```go
stripe.APIVersion = "2024-12-18"
```

## Version Detection Patterns

### Package Manager Files

#### package.json
```json
"stripe": "^17.0.0"
"@stripe/stripe-js": "^4.0.0"
```

#### requirements.txt
```
stripe==11.0.0
stripe>=11.0.0
```

#### Gemfile
```ruby
gem 'stripe', '~> 13.0'
```

#### composer.json
```json
"stripe/stripe-php": "^15.0"
```

#### go.mod
```
github.com/stripe/stripe-go/v78 v78.0.0
```

### Source Code Patterns

#### Version Checks
```javascript
Stripe.VERSION
stripe.__version__
Stripe::VERSION
```

#### API Version Headers
- HTTP Header: `Stripe-Version: 2024-12-18`
- Request parameter: `?version=2024-12-18`

## Deprecated/Legacy Versions

### EOL SDK Versions
- **Node.js**: v7.x and below (deprecated)
- **Python**: v4.x and below (deprecated)
- **Ruby**: v4.x and below (deprecated)
- **PHP**: v6.x and below (deprecated)
- **Go**: v71.x and below (unsupported)

### Migration Notes
- Checkout v2 (legacy) → Checkout Sessions
- Charges API → Payment Intents API
- Tokens API → Payment Methods API
- Sources API → Payment Methods API

## Breaking Changes to Watch For

### Major SDK Version Bumps
- Often include breaking changes
- Review migration guides
- Test thoroughly before upgrading

### API Version Changes
- New required parameters
- Deprecated endpoints
- Changed response formats
- New validation rules

## Version Compatibility Matrix

### Node.js SDK
| SDK Version | Node.js Version | API Version Default |
|-------------|----------------|---------------------|
| v17.x       | 12+            | Latest              |
| v16.x       | 12+            | 2024-06-20          |
| v15.x       | 12+            | 2023-10-16          |
| v14.x       | 12+            | 2023-08-16          |

### Python SDK
| SDK Version | Python Version | API Version Default |
|-------------|---------------|---------------------|
| v11.x       | 3.6+          | Latest              |
| v10.x       | 3.6+          | 2024-06-20          |
| v9.x        | 3.6+          | 2023-10-16          |

### Go SDK
| SDK Version | Go Version | API Version Default |
|-------------|-----------|---------------------|
| v78.x       | 1.18+     | Latest              |
| v77.x       | 1.18+     | 2024-06-20          |
| v76.x       | 1.18+     | 2023-10-16          |

## Detection Confidence

- **HIGH**: Exact version in package manager file
- **HIGH**: Version constant in source code
- **MEDIUM**: API version in configuration
- **LOW**: Version inferred from API usage patterns
