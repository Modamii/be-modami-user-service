# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o bin/server ./cmd/server

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for TLS
RUN apk add --no-cache ca-certificates tzdata

# Copy binary from builder
COPY --from=builder /app/bin/server .

# Copy migrations
COPY --from=builder /app/migrations ./migrations

# Expose ports
EXPOSE 8080 9090

# Run
CMD ["./server"]
