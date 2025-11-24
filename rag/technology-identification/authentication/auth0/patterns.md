# Auth0

**Category**: authentication
**Description**: Auth0 identity and access management platform
**Homepage**: https://auth0.com

## Package Detection

### NPM
*Auth0 JavaScript/Node.js SDKs*

- `@auth0/auth0-react`
- `@auth0/nextjs-auth0`
- `@auth0/auth0-spa-js`
- `auth0`
- `express-openid-connect`
- `passport-auth0`

### PYPI
*Auth0 Python SDKs*

- `auth0-python`
- `authlib`
- `python-jose`

### RUBYGEMS
*Auth0 Ruby SDKs*

- `auth0`
- `omniauth-auth0`

### MAVEN
*Auth0 Java SDKs*

- `com.auth0:auth0`
- `com.auth0:java-jwt`
- `com.auth0:jwks-rsa`

### GO
*Auth0 Go SDKs*

- `github.com/auth0/go-auth0`
- `github.com/auth0/go-jwt-middleware`

### Related Packages
- `@auth0/auth0-vue`
- `@auth0/auth0-angular`
- `auth0-lock`

## Import Detection

### Javascript

**Pattern**: `from\s+['"]@auth0/auth0-react['"]`
- Type: esm_import

**Pattern**: `from\s+['"]@auth0/nextjs-auth0['"]`
- Type: esm_import

**Pattern**: `from\s+['"]@auth0/auth0-spa-js['"]`
- Type: esm_import

**Pattern**: `require\(['"]auth0['"]\)`
- Type: commonjs_require

**Pattern**: `from\s+['"]express-openid-connect['"]`
- Type: esm_import

### Python

**Pattern**: `from\s+auth0`
- Type: python_import

**Pattern**: `import\s+auth0`
- Type: python_import

### Go

**Pattern**: `"github\.com/auth0/go-auth0`
- Type: go_import

**Pattern**: `"github\.com/auth0/go-jwt-middleware`
- Type: go_import

## Environment Variables

*Auth0 tenant domain*

*Auth0 application client ID*

*Auth0 application client secret*

*Auth0 API audience*

*Auth0 issuer base URL*

*Auth0 session secret (Next.js)*

*Application base URL*

*Default OAuth scopes*


## Detection Notes

- Check for AUTH0_DOMAIN, AUTH0_CLIENT_ID environment variables
- Look for *.auth0.com domain references
- Now part of Okta

## Detection Confidence

- **Package Detection**: 95% (HIGH)
- **Import Detection**: 90% (HIGH)
- **Environment Variable Detection**: 85% (MEDIUM)
- **API Endpoint Detection**: 80% (MEDIUM)
