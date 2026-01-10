# Zero - Engineering Intelligence Platform
# Multi-stage build for minimal image size

# ============================================
# Stage 1: Build Go binary
# ============================================
FROM golang:1.24-alpine AS builder

WORKDIR /build

# Install build dependencies
RUN apk add --no-cache git

# Copy go modules first (better caching)
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o zero ./cmd/zero

# ============================================
# Stage 2: Final runtime image
# ============================================
FROM node:20-alpine

LABEL org.opencontainers.image.title="Zero"
LABEL org.opencontainers.image.description="Engineering intelligence platform with AI-powered agents"
LABEL org.opencontainers.image.source="https://github.com/crashappsec/zero"
LABEL org.opencontainers.image.vendor="Crash Override"
LABEL org.opencontainers.image.licenses="GPL-3.0"

# Install runtime dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Create non-root user (use different GID/UID to avoid conflicts with node user)
RUN addgroup -g 10000 zero && \
    adduser -u 10000 -G zero -s /bin/sh -D zero

# Copy Go binary
COPY --from=builder /build/zero /usr/local/bin/zero

# Set up working directory
WORKDIR /home/zero

# Environment variables
ENV ZERO_HOME=/home/zero/.zero

# Create .zero directory
RUN mkdir -p /home/zero/.zero && chown -R zero:zero /home/zero

# Switch to non-root user
USER zero

# Default command
ENTRYPOINT ["zero"]
CMD ["--help"]
