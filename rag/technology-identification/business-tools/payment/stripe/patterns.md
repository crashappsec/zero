# Stripe

**Category**: business-tools/payment
**Description**: Stripe payment processing platform
**Homepage**: https://stripe.com

## Package Detection

### NPM
*Stripe JavaScript SDKs*

- `stripe`
- `@stripe/stripe-js`
- `@stripe/react-stripe-js`

### PYPI
*Stripe Python SDK*

- `stripe`

### GO
*Stripe Go SDK*

- `github.com/stripe/stripe-go`

### MAVEN
*Stripe Java SDK*

- `com.stripe:stripe-java`

### RUBYGEMS
*Stripe Ruby SDK*

- `stripe`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]stripe['"]`

**Pattern**: `require\(['"]stripe['"]\)`

**Pattern**: `from\s+['"]@stripe/`

### Python

**Pattern**: `import\s+stripe`

**Pattern**: `from\s+stripe`

## Environment Variables


## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
