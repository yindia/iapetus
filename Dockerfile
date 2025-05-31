# syntax=docker/dockerfile:1
# Dockerfile for building a general-purpose iapetus runner image

# --- Build stage ---
FROM golang:1.23-alpine AS builder
WORKDIR /src
# Copy go.mod and go.sum first for better build caching
COPY go.mod go.sum ./
RUN go mod download
# Copy the rest of the source
COPY . .
# Build the iapetus binary (adjust path if needed)
RUN go build -o /iapetus ./cmd/iapetus

# --- Runtime stage ---
FROM alpine:3.18
WORKDIR /app
# Install bash for bash backend, and tini for better signal handling
RUN apk add --no-cache bash tini
# Copy the built binary
COPY --from=builder /iapetus /usr/local/bin/iapetus

# Create a non-root user for security
RUN adduser -D -u 1001 iapetususer
USER iapetususer

# Use tini as the entrypoint for proper signal handling
ENTRYPOINT ["/sbin/tini", "--", "/usr/local/bin/iapetus"]

# ---
# Usage:
#   docker build -t iapetus-runner .
#   docker run --rm -v $(pwd)/workflow.yaml:/app/workflow.yaml iapetus-runner run --workflow /app/workflow.yaml
#   # Or pass any iapetus CLI command/args
#   docker run --rm iapetus-runner --help

