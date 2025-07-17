# Translation Service
# Author: Kyrylo Kovalenko (git@kovalenko.tech)
# Website: https://kovalenko.tech
# Build stage
FROM golang:1.23.2-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy dependency files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main ./cmd/server

# Final stage
FROM alpine:latest

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1001 -S appgroup && \
    adduser -u 1001 -S appuser -G appgroup

# Set working directory
WORKDIR /app

# Copy binary from builder stage
COPY --from=builder /app/main .

# Change ownership of files
RUN chown -R appuser:appgroup /app

# Switch to non-privileged user
USER appuser

# Expose port
EXPOSE 8080

# Start command
CMD ["./main"] 