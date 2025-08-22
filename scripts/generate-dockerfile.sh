#!/usr/bin/env bash

# Generate Dockerfile script
set -euo pipefail

BUILD_DIR=${1:-$(pwd)/build}

cat > "${BUILD_DIR}/Dockerfile" << 'EOF'
# Build stage
FROM golang:1.25-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /workspace

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY cmd/ cmd/
COPY internal/ internal/
COPY pkg/ pkg/
COPY configs/ configs/

# Build arguments for version injection
ARG SERVICE_NAME="apiserver"
ARG VERSION="dev"
ARG COMMIT="unknown"
ARG BRANCH="unknown"
ARG TREE_STATE="unknown"
ARG BUILD_DATE="unknown"

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-X 'github.com/costa92/go-protoc/v2/pkg/version.serviceName=${SERVICE_NAME}' \
              -X 'github.com/costa92/go-protoc/v2/pkg/version.gitVersion=${VERSION}' \
              -X 'github.com/costa92/go-protoc/v2/pkg/version.gitCommit=${COMMIT}' \
              -X 'github.com/costa92/go-protoc/v2/pkg/version.gitBranch=${BRANCH}' \
              -X 'github.com/costa92/go-protoc/v2/pkg/version.gitTreeState=${TREE_STATE}' \
              -X 'github.com/costa92/go-protoc/v2/pkg/version.buildDate=${BUILD_DATE}' \
              -w -s" \
    -a -installsuffix cgo \
    -o ${SERVICE_NAME} ./cmd/${SERVICE_NAME}

# Runtime stage
FROM alpine:3.19

# Import build arguments
ARG SERVICE_NAME="apiserver"

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata && \
    addgroup -g 65532 ${SERVICE_NAME} && \
    adduser -D -u 65532 -G ${SERVICE_NAME} ${SERVICE_NAME}

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /workspace/${SERVICE_NAME} .
COPY --from=builder /workspace/configs ./configs

# Set ownership
RUN chown -R ${SERVICE_NAME}:${SERVICE_NAME} /app

# Switch to non-root user
USER ${SERVICE_NAME}

# Expose ports
EXPOSE 8080 9090

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD ["/app/${SERVICE_NAME}", "--help"] || exit 1

# Command to run the application
ENTRYPOINT ["/app/${SERVICE_NAME}"]
CMD ["--config=/app/configs/apiserver.yaml"]
EOF

echo "Dockerfile generated successfully at ${BUILD_DIR}/Dockerfile"