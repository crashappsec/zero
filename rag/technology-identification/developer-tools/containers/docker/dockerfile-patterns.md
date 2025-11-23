# Dockerfile Patterns

## File Names

### Standard Dockerfile Names
- `Dockerfile`
- `Dockerfile.*` (e.g., Dockerfile.dev, Dockerfile.prod)
- `*.Dockerfile`
- `.dockerignore`

### Multi-stage Build Files
- `Dockerfile` (with multiple FROM statements)
- `Dockerfile.multistage`

## Dockerfile Instructions

### Base Image Instructions
```dockerfile
FROM ubuntu:22.04
FROM node:18-alpine
FROM python:3.11-slim
FROM golang:1.21
FROM nginx:alpine
FROM postgres:15
FROM redis:7-alpine
FROM alpine:latest
FROM scratch
```

### Build Instructions
```dockerfile
RUN apt-get update && apt-get install -y
RUN npm install
RUN pip install -r requirements.txt
RUN go build -o app
RUN cargo build --release
```

### Copy Instructions
```dockerfile
COPY . /app
COPY --from=builder /app/dist /usr/share/nginx/html
ADD https://example.com/file.tar.gz /tmp/
WORKDIR /app
```

### Environment and Configuration
```dockerfile
ENV NODE_ENV=production
ENV PATH="/app/bin:${PATH}"
ARG BUILD_DATE
ARG VERSION
LABEL maintainer="email@example.com"
LABEL version="1.0"
```

### Exposure and Networking
```dockerfile
EXPOSE 8080
EXPOSE 443/tcp
EXPOSE 8080/udp
```

### Execution Instructions
```dockerfile
CMD ["node", "server.js"]
CMD npm start
ENTRYPOINT ["python", "app.py"]
ENTRYPOINT ["/bin/sh", "-c"]
```

### User and Permissions
```dockerfile
USER node
USER 1000
RUN useradd -m appuser
```

### Health Checks
```dockerfile
HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost/ || exit 1
```

### Volume Mounts
```dockerfile
VOLUME ["/data"]
VOLUME /var/log
```

### Shell Configuration
```dockerfile
SHELL ["/bin/bash", "-c"]
```

### Build Stage Names
```dockerfile
FROM node:18 AS builder
FROM node:18 AS development
FROM nginx:alpine AS production
```

## Common Base Images

### Programming Languages

#### Node.js
```dockerfile
FROM node:18
FROM node:18-alpine
FROM node:18-slim
FROM node:20-bookworm
```

#### Python
```dockerfile
FROM python:3.11
FROM python:3.11-slim
FROM python:3.11-alpine
FROM python:3.12-bookworm
```

#### Go
```dockerfile
FROM golang:1.21
FROM golang:1.21-alpine
FROM golang:1.22-bookworm
```

#### Java
```dockerfile
FROM openjdk:17
FROM openjdk:17-slim
FROM eclipse-temurin:17
FROM amazoncorretto:17
```

#### Ruby
```dockerfile
FROM ruby:3.2
FROM ruby:3.2-alpine
FROM ruby:3.3-slim
```

#### PHP
```dockerfile
FROM php:8.2
FROM php:8.2-fpm
FROM php:8.2-apache
FROM php:8.3-cli
```

#### Rust
```dockerfile
FROM rust:1.75
FROM rust:1.75-alpine
FROM rust:1.75-slim
```

### Web Servers
```dockerfile
FROM nginx:alpine
FROM nginx:latest
FROM httpd:alpine
FROM caddy:alpine
```

### Databases
```dockerfile
FROM postgres:15
FROM mysql:8
FROM mariadb:10
FROM mongodb:7
FROM redis:7-alpine
```

### Operating Systems
```dockerfile
FROM ubuntu:22.04
FROM ubuntu:24.04
FROM debian:bookworm
FROM debian:bullseye
FROM alpine:3.19
FROM alpine:latest
FROM centos:7
FROM rockylinux:9
FROM amazonlinux:2023
```

### Specialized
```dockerfile
FROM scratch (minimal, no OS)
FROM distroless/static (Google distroless)
FROM distroless/base
FROM chainguard/static (Chainguard)
```

## Multi-Stage Build Patterns

### Node.js Multi-Stage
```dockerfile
FROM node:18 AS builder
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

FROM node:18-alpine
WORKDIR /app
COPY --from=builder /app/dist ./dist
COPY --from=builder /app/node_modules ./node_modules
CMD ["node", "dist/server.js"]
```

### Go Multi-Stage
```dockerfile
FROM golang:1.21 AS builder
WORKDIR /app
COPY go.* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o app

FROM alpine:latest
COPY --from=builder /app/app /app
CMD ["/app"]
```

### Python Multi-Stage
```dockerfile
FROM python:3.11 AS builder
WORKDIR /app
COPY requirements.txt .
RUN pip install --user -r requirements.txt
COPY . .

FROM python:3.11-slim
WORKDIR /app
COPY --from=builder /root/.local /root/.local
COPY --from=builder /app .
CMD ["python", "app.py"]
```

## Security Best Practices Patterns

### Non-Root User
```dockerfile
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
USER appuser
```

### Minimal Base Images
```dockerfile
FROM alpine:latest
FROM distroless/static
FROM scratch
```

### Layer Optimization
```dockerfile
RUN apt-get update && apt-get install -y \
    package1 \
    package2 \
    && rm -rf /var/lib/apt/lists/*
```

### Secrets Management
```dockerfile
RUN --mount=type=secret,id=mysecret \
    secret=$(cat /run/secrets/mysecret) && \
    # use secret
```

### Build Arguments
```dockerfile
ARG VERSION=latest
ARG BUILD_DATE
LABEL build-date=$BUILD_DATE
```

## .dockerignore Patterns

### Common Exclusions
```
node_modules
npm-debug.log
.git
.gitignore
.env
.env.local
*.md
Dockerfile*
docker-compose*.yml
.dockerignore
__pycache__
*.pyc
.pytest_cache
.vscode
.idea
.DS_Store
dist
build
target
*.log
```

## Anti-Patterns to Detect

### Security Issues
```dockerfile
# Bad: Running as root
USER root

# Bad: Hardcoded secrets
ENV API_KEY=secret123

# Bad: Using latest tag
FROM ubuntu:latest

# Bad: Exposing unnecessary ports
EXPOSE 22
```

### Build Issues
```dockerfile
# Bad: Not combining RUN commands
RUN apt-get update
RUN apt-get install -y package1
RUN apt-get install -y package2

# Bad: Not cleaning up
RUN apt-get install -y something
# Missing: && rm -rf /var/lib/apt/lists/*
```

## Dockerfile Commands Reference

### Core Instructions
- `FROM` - Base image
- `RUN` - Execute command during build
- `CMD` - Default command to run
- `ENTRYPOINT` - Configured container executable
- `COPY` - Copy files from host
- `ADD` - Copy files (with URL/tar support)
- `WORKDIR` - Set working directory
- `ENV` - Set environment variable
- `ARG` - Build-time variable
- `EXPOSE` - Document port
- `VOLUME` - Create mount point
- `USER` - Set user/UID
- `LABEL` - Add metadata
- `HEALTHCHECK` - Container health check
- `SHELL` - Override default shell
- `ONBUILD` - Trigger for downstream builds
- `STOPSIGNAL` - Signal to stop container

## Detection Confidence

- **HIGH**: File named "Dockerfile" or "*.Dockerfile"
- **HIGH**: Presence of FROM, RUN, CMD, or ENTRYPOINT instructions
- **HIGH**: .dockerignore file
- **MEDIUM**: Docker-related comments in build scripts
- **LOW**: Generic build commands without Docker specifics
