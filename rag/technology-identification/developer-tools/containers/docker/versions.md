# Docker and Docker Compose Versions

## Docker Engine

### Current Versions (2025)

#### Docker Engine (Community Edition)
- **Latest Stable**: 27.x
- **Previous Stable**: 26.x, 25.x, 24.x
- **LTS**: 24.0 LTS (Long-term support)
- **Status**: Active development
- **Repository**: https://github.com/moby/moby

#### Version History
- **27.x** (2025) - Current
- **26.x** (2024) - Stable
- **25.x** (2024) - Stable
- **24.x** (2023) - LTS
- **23.x** (2023) - Stable
- **20.10.x** (2020-2023) - Previous LTS
- **19.03.x** (2019-2021) - EOL
- **18.09.x** (2018-2020) - EOL
- **17.x** (2017-2018) - EOL

### Version Detection

#### Check Docker Version
```bash
docker --version
docker version
```

#### Output Format
```
Docker version 27.0.0, build abc1234
```

#### API Version
```bash
docker version --format '{{.Server.APIVersion}}'
```

### Docker Engine API Versions
- **1.45** - Docker 27.x
- **1.44** - Docker 26.x
- **1.43** - Docker 24.x / 25.x
- **1.42** - Docker 23.x
- **1.41** - Docker 20.10.x
- **1.40** - Docker 19.03.x

## Docker Compose

### Docker Compose v2 (Current)

#### Standalone Binary
- **Latest**: 2.x (2.30.x as of 2025)
- **Status**: Active development
- **Install**: Standalone binary or Docker Desktop
- **Repository**: https://github.com/docker/compose

#### Docker Compose Plugin (Preferred)
- **Command**: `docker compose` (note: no hyphen)
- **Included**: Docker Desktop, Docker Engine 20.10.13+
- **Status**: Recommended approach

#### Version Check
```bash
docker compose version
# Output: Docker Compose version v2.30.0
```

### Docker Compose v1 (Legacy)

#### Standalone Tool
- **Latest**: 1.29.x (final release)
- **Status**: End of Life (EOL) as of June 2023
- **Command**: `docker-compose` (note: hyphen)
- **Python**: Required Python runtime

#### Version Check
```bash
docker-compose --version
# Output: docker-compose version 1.29.2, build unknown
```

### Compose File Versions

#### Compose Specification (Modern)
```yaml
# No version field needed
services:
  app:
    image: myapp
```
- **Status**: Current standard
- **Features**: All modern features
- **Compatibility**: Docker Compose v2

#### Version 3 (Swarm-focused)
```yaml
version: '3.8'
services:
  app:
    image: myapp
```
- **Versions**: 3.0 - 3.9
- **3.9**: Latest in v3 series
- **3.8**: Widely used
- **Features**: Deploy, secrets, configs
- **Compatibility**: Docker Compose v1.27+ and v2

#### Version 2 (Legacy)
```yaml
version: '2.4'
services:
  app:
    image: myapp
```
- **Versions**: 2.0 - 2.4
- **Status**: Legacy
- **Features**: resource limits, healthchecks
- **Compatibility**: Docker Compose v1

#### Version Compatibility Matrix

| Compose File | Docker Compose v1 | Docker Compose v2 | Features |
|--------------|-------------------|-------------------|----------|
| Spec (no version) | ❌ | ✅ | All modern features |
| 3.9 | ✅ (1.27+) | ✅ | Deploy, secrets, configs |
| 3.8 | ✅ (1.25+) | ✅ | Most v3 features |
| 3.7 | ✅ (1.22+) | ✅ | Init, rollback |
| 3.0-3.6 | ✅ | ✅ | Basic v3 features |
| 2.4 | ✅ (1.12+) | ✅ | Platform, runtime |
| 2.0-2.3 | ✅ | ✅ | Basic v2 features |

## Docker Desktop

### Current Versions
- **Latest**: 4.x (4.35.x as of 2025)
- **Status**: Active development
- **Platforms**: macOS, Windows, Linux
- **Includes**: Docker Engine, Docker Compose, Kubernetes

### Version Components
```
Docker Desktop 4.35.0 includes:
- Docker Engine 27.0.0
- Docker Compose 2.30.0
- Kubernetes 1.29.2
```

## Base Image Versions

### Programming Languages

#### Node.js
```dockerfile
FROM node:20        # Latest 20.x
FROM node:18        # Latest 18.x
FROM node:18.19.0   # Specific version
FROM node:18-alpine # Alpine variant
```

#### Python
```dockerfile
FROM python:3.12       # Latest 3.12.x
FROM python:3.11       # Latest 3.11.x
FROM python:3.11.8     # Specific version
FROM python:3.11-slim  # Slim variant
```

#### Go
```dockerfile
FROM golang:1.22       # Latest 1.22.x
FROM golang:1.21       # Latest 1.21.x
FROM golang:1.21.6     # Specific version
FROM golang:1.21-alpine # Alpine variant
```

#### Java
```dockerfile
FROM openjdk:21        # Latest 21
FROM openjdk:17        # Latest 17
FROM openjdk:17-slim   # Slim variant
FROM eclipse-temurin:17 # Adoptium
```

### Operating Systems

#### Alpine Linux
```dockerfile
FROM alpine:3.19       # Latest stable
FROM alpine:3.18       # Previous stable
FROM alpine:latest     # Latest (not recommended)
```

#### Ubuntu
```dockerfile
FROM ubuntu:24.04      # Noble Numbat (LTS)
FROM ubuntu:22.04      # Jammy Jellyfish (LTS)
FROM ubuntu:20.04      # Focal Fossa (LTS)
```

#### Debian
```dockerfile
FROM debian:bookworm   # Debian 12 (current)
FROM debian:bullseye   # Debian 11
FROM debian:buster     # Debian 10
```

### Databases

#### PostgreSQL
```dockerfile
FROM postgres:16       # Latest 16.x
FROM postgres:15       # Latest 15.x
FROM postgres:15.5     # Specific version
FROM postgres:15-alpine # Alpine variant
```

#### MySQL
```dockerfile
FROM mysql:8.3         # Latest 8.3
FROM mysql:8           # Latest 8.x
FROM mysql:8.0.36      # Specific version
```

#### Redis
```dockerfile
FROM redis:7.2         # Latest 7.2.x
FROM redis:7           # Latest 7.x
FROM redis:7-alpine    # Alpine variant
```

## BuildKit

### BuildKit Versions
- **Latest**: 0.12.x (as of 2025)
- **Status**: Active development
- **Default**: Docker Engine 23.0+

### Enable BuildKit
```bash
export DOCKER_BUILDKIT=1
```

### BuildKit Features
- Multi-platform builds
- Build secrets
- SSH forwarding
- Cache mounts
- Improved layer caching

## Version Detection Patterns

### Docker Version in CI/CD

#### GitHub Actions
```yaml
- name: Set up Docker Buildx
  uses: docker/setup-buildx-action@v3
  with:
    version: latest
```

#### GitLab CI
```yaml
image: docker:27-dind
services:
  - docker:27-dind
```

### Dockerfile Version Hints

#### ARG VERSION
```dockerfile
ARG DOCKER_VERSION=27.0.0
FROM docker:${DOCKER_VERSION}
```

### Compose File Version Declaration
```yaml
version: '3.8'
```

## Deprecations and Migrations

### Docker Compose v1 → v2
- **Timeline**: v1 EOL June 2023
- **Command**: `docker-compose` → `docker compose`
- **Installation**: Python package → Docker plugin
- **Breaking Changes**: Minor syntax differences

### Compose File v2 → v3
- **Removed**: `volume_driver`, `volumes_from`
- **Changed**: Resource limits syntax
- **Added**: Deploy, secrets, configs

### Legacy Docker Versions
- **17.x and earlier**: End of support
- **18.09**: Extended support ended
- **19.03**: No longer maintained

## Detection Confidence

- **HIGH**: `docker --version` output
- **HIGH**: `docker-compose --version` or `docker compose version` output
- **HIGH**: Specific version in Dockerfile FROM statement
- **MEDIUM**: Compose file version declaration
- **MEDIUM**: Docker API version in code
- **LOW**: Inferred from features used
