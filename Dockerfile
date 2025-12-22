# Zero - Security Analysis Toolkit
# Multi-stage build for minimal image size

# ============================================
# Stage 1: Build Go binary
# ============================================
FROM golang:1.23-alpine AS builder

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
# Stage 2: Prepare Evidence template
# ============================================
FROM node:20-alpine AS evidence-builder

WORKDIR /evidence

# Copy Evidence template
COPY reports/template/package*.json ./
RUN npm ci --omit=dev

# Copy template files
COPY reports/template/ ./

# ============================================
# Stage 3: Final runtime image
# ============================================
FROM node:20-alpine

LABEL org.opencontainers.image.title="Zero"
LABEL org.opencontainers.image.description="Security analysis toolkit with AI-powered agents"
LABEL org.opencontainers.image.source="https://github.com/crashappsec/zero"
LABEL org.opencontainers.image.vendor="Crash Override"
LABEL org.opencontainers.image.licenses="GPL-3.0"

# Install runtime dependencies
RUN apk add --no-cache \
    git \
    ca-certificates \
    tzdata

# Create non-root user
RUN addgroup -g 1000 zero && \
    adduser -u 1000 -G zero -s /bin/sh -D zero

# Copy Go binary
COPY --from=builder /build/zero /usr/local/bin/zero

# Copy Evidence template with pre-installed dependencies
COPY --from=evidence-builder /evidence /opt/zero/reports/template

# Set up working directory
WORKDIR /home/zero

# Environment variables
ENV ZERO_HOME=/home/zero/.zero
ENV ZERO_TEMPLATE_PATH=/opt/zero/reports/template

# Create .zero directory
RUN mkdir -p /home/zero/.zero && chown -R zero:zero /home/zero

# Switch to non-root user
USER zero

# Default command
ENTRYPOINT ["zero"]
CMD ["--help"]
