# Use the official Go image as build environment
FROM golang:1.21-alpine AS builder

# Install required system dependencies for SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

# Set working directory
WORKDIR /workspace

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o bin/aura cmd/aura/main.go

# Production stage
FROM alpine:latest

# Install SQLite
RUN apk add --no-cache sqlite

# Create app directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /workspace/bin/aura .

# Create data directory
RUN mkdir -p /data

# Set environment variables
ENV AURA_DB_PATH=/data/aura.db
ENV AURA_ENV=production

# Create non-root user
RUN addgroup -g 1001 aura && \
    adduser -D -s /bin/sh -u 1001 -G aura aura

# Change ownership
RUN chown -R aura:aura /app /data

# Switch to non-root user
USER aura

# Expose port (if needed for future web interface)
EXPOSE 8080

# Command to run
ENTRYPOINT ["./aura"]
