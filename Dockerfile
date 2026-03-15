# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /src

# Install build dependencies
RUN apk add --no-cache git

# Copy go mod and sum files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/main ./app/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk add --no-cache ca-certificates tzdata postgresql-client && update-ca-certificates

# Copy binary from builder
COPY --from=builder /app/main .

# Copy migrations and static files
COPY migrations ./migrations
# Ensure docs are generated before build!
COPY docs ./docs
COPY scripts/wait-for-postgres.sh /usr/local/bin/wait-for-postgres.sh

# Set usage permissions
RUN chmod +x /usr/local/bin/wait-for-postgres.sh

EXPOSE 8080

CMD ["wait-for-postgres.sh", "./main"]
