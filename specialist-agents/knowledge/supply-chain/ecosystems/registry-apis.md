# Package Registry API Reference

## Overview

This document provides a reference for interacting with package registry APIs across different ecosystems. Use these APIs to fetch package metadata, version information, and security advisories.

## npm Registry

### Base URL
```
https://registry.npmjs.org
```

### Get Package Metadata
```bash
curl https://registry.npmjs.org/{package}
```

Response includes:
- All versions with metadata
- Maintainers list
- Time of each release
- Dist tags (latest, next, etc.)

### Get Specific Version
```bash
curl https://registry.npmjs.org/{package}/{version}
```

### Get Download Stats
```bash
# Last day
curl https://api.npmjs.org/downloads/point/last-day/{package}

# Last week
curl https://api.npmjs.org/downloads/point/last-week/{package}

# Last month
curl https://api.npmjs.org/downloads/point/last-month/{package}

# Date range
curl https://api.npmjs.org/downloads/range/{start}:{end}/{package}
```

### Search Packages
```bash
curl "https://registry.npmjs.org/-/v1/search?text={query}&size=20"
```

### Bulk Security Advisory
```bash
curl -X POST https://registry.npmjs.org/-/npm/v1/security/advisories/bulk \
  -H "Content-Type: application/json" \
  -d '{"package-name": ["1.0.0", "2.0.0"]}'
```

## PyPI

### Base URL
```
https://pypi.org
```

### Get Package Metadata (JSON API)
```bash
curl https://pypi.org/pypi/{package}/json
```

Response includes:
- Package info (author, license, description)
- All releases with download URLs
- Vulnerabilities (if any)

### Get Specific Version
```bash
curl https://pypi.org/pypi/{package}/{version}/json
```

### Simple Index (PEP 503)
```bash
curl https://pypi.org/simple/{package}/
```

Returns HTML with links to all available distributions.

### Download Stats (via pypistats.org)
```bash
# Recent downloads
curl https://pypistats.org/api/packages/{package}/recent

# By Python version
curl https://pypistats.org/api/packages/{package}/python_minor

# By system
curl https://pypistats.org/api/packages/{package}/system
```

## RubyGems

### Base URL
```
https://rubygems.org
```

### Get Gem Info
```bash
curl https://rubygems.org/api/v1/gems/{gem}.json
```

### Get All Versions
```bash
curl https://rubygems.org/api/v1/versions/{gem}.json
```

### Get Dependencies
```bash
curl https://rubygems.org/api/v1/dependencies?gems={gem1},{gem2}
```

### Search Gems
```bash
curl "https://rubygems.org/api/v1/search.json?query={query}"
```

### Download Stats
```bash
curl https://rubygems.org/api/v1/downloads/{gem}-{version}.json
```

## Maven Central

### Base URL
```
https://search.maven.org
```

### Search Artifacts
```bash
# By group and artifact
curl "https://search.maven.org/solrsearch/select?q=g:{groupId}+AND+a:{artifactId}&rows=20&wt=json"

# By class name
curl "https://search.maven.org/solrsearch/select?q=fc:{className}&rows=20&wt=json"
```

### Get Artifact Versions
```bash
curl "https://search.maven.org/solrsearch/select?q=g:{groupId}+AND+a:{artifactId}&core=gav&rows=100&wt=json"
```

### Direct Download
```bash
curl https://repo1.maven.org/maven2/{groupId}/{artifactId}/{version}/{artifactId}-{version}.pom
```

## crates.io (Rust)

### Base URL
```
https://crates.io/api/v1
```

### Get Crate Info
```bash
curl https://crates.io/api/v1/crates/{crate}
```

### Get Specific Version
```bash
curl https://crates.io/api/v1/crates/{crate}/{version}
```

### Get Dependencies
```bash
curl https://crates.io/api/v1/crates/{crate}/{version}/dependencies
```

### Search Crates
```bash
curl "https://crates.io/api/v1/crates?q={query}&per_page=20"
```

## Go Modules (proxy.golang.org)

### Base URL
```
https://proxy.golang.org
```

### List Versions
```bash
curl https://proxy.golang.org/{module}/@v/list
```

### Get Module Info
```bash
curl https://proxy.golang.org/{module}/@v/{version}.info
```

### Get go.mod
```bash
curl https://proxy.golang.org/{module}/@v/{version}.mod
```

### Download Source
```bash
curl https://proxy.golang.org/{module}/@v/{version}.zip
```

## GitHub Packages

### Base URL (npm)
```
https://npm.pkg.github.com
```

### Authentication Required
```bash
curl -H "Authorization: Bearer ${GITHUB_TOKEN}" \
  https://npm.pkg.github.com/{org}/{package}
```

## Rate Limiting

| Registry | Rate Limit | Notes |
|----------|------------|-------|
| npm | Generous, no strict limit | May throttle abusive patterns |
| PyPI | ~100 req/min | Use mirrors for high volume |
| RubyGems | ~10 req/sec | Rate limit headers in response |
| Maven Central | No strict limit | Cache results |
| crates.io | ~1 req/sec | User-Agent required |
| Go proxy | Generous | Caches upstream |

## Best Practices

### Caching
- Cache package metadata locally
- Respect `Cache-Control` headers
- Use ETags for conditional requests

### User-Agent
Always set a descriptive User-Agent:
```bash
curl -H "User-Agent: MyApp/1.0 (contact@example.com)" ...
```

### Error Handling
- 404: Package/version doesn't exist
- 429: Rate limited, back off
- 5xx: Server error, retry with exponential backoff

### Security
- Verify checksums (sha512 for npm, sha256 for PyPI)
- Check signatures when available
- Use HTTPS exclusively
- Validate package names to prevent injection

## OSV (Open Source Vulnerabilities)

### Query Vulnerabilities
```bash
curl -X POST https://api.osv.dev/v1/query \
  -H "Content-Type: application/json" \
  -d '{
    "package": {
      "name": "package-name",
      "ecosystem": "npm"
    },
    "version": "1.0.0"
  }'
```

### Batch Query
```bash
curl -X POST https://api.osv.dev/v1/querybatch \
  -H "Content-Type: application/json" \
  -d '{
    "queries": [
      {"package": {"name": "pkg1", "ecosystem": "npm"}, "version": "1.0.0"},
      {"package": {"name": "pkg2", "ecosystem": "PyPI"}, "version": "2.0.0"}
    ]
  }'
```

### Get Vulnerability by ID
```bash
curl https://api.osv.dev/v1/vulns/{vuln_id}
```

## deps.dev API

### Base URL
```
https://api.deps.dev/v3alpha
```

### Get Package Info
```bash
curl "https://api.deps.dev/v3alpha/systems/{ecosystem}/packages/{package}"
```

### Get Version Info
```bash
curl "https://api.deps.dev/v3alpha/systems/{ecosystem}/packages/{package}/versions/{version}"
```

### Get Dependencies
```bash
curl "https://api.deps.dev/v3alpha/systems/{ecosystem}/packages/{package}/versions/{version}:dependencies"
```

Ecosystems: `npm`, `pypi`, `maven`, `go`, `cargo`, `nuget`
