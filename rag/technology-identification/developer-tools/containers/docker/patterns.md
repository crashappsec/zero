# Docker

**Category**: developer-tools/containers
**Description**: Container platform for building, shipping, and running applications
**Homepage**: https://www.docker.com

## Configuration Files

- `Dockerfile`
- `Dockerfile.*` (e.g., Dockerfile.dev, Dockerfile.prod)
- `*.Dockerfile`
- `.dockerignore`
- `docker-compose.yml`
- `docker-compose.yaml`
- `docker-compose.*.yml`

## Package Detection

Docker is detected primarily through configuration files rather than package dependencies.

## Environment Variables

- `DOCKER_HOST`
- `DOCKER_TLS_VERIFY`
- `DOCKER_CERT_PATH`
- `DOCKER_BUILDKIT`
- `COMPOSE_PROJECT_NAME`
- `COMPOSE_FILE`

## Detection Notes

- Look for Dockerfile in repository root or subdirectories
- Check for docker-compose.yml files
- Look for .dockerignore files
- Multi-stage builds indicate more mature Docker usage

## Detection Confidence

- **Dockerfile Detection**: 95% (HIGH)
- **docker-compose.yml Detection**: 95% (HIGH)
- **.dockerignore Detection**: 90% (HIGH)
- **Docker commands in scripts**: 80% (MEDIUM)
