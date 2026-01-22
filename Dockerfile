# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /build

# Copy go mod files
COPY go.mod go.sum* ./

# Download dependencies
RUN go mod download

# Copy source code
COPY *.go ./

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o porkbun-ssl .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests and su-exec for user switching
RUN apk --no-cache add ca-certificates tzdata su-exec

WORKDIR /app

# Copy the binary from builder
COPY --from=builder /build/porkbun-ssl .

# Copy entrypoint script
COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

# Create directory for certificates
RUN mkdir -p /certs

# Create default user (will be modified at runtime if PUID/PGID differ)
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser && \
    chown -R appuser:appuser /app /certs

ENTRYPOINT ["/entrypoint.sh"]
