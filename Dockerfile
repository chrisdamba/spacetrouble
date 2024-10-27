# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git make

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o spacetrouble ./cmd/api/main.go

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates for HTTPS calls
RUN apk --no-cache add ca-certificates

# Copy the binary from builder
COPY --from=builder /app/spacetrouble .
COPY --from=builder /app/migrations ./migrations

# Install migrate tool for database migrations
RUN wget -O /usr/local/bin/migrate https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz && \
    tar -xf /usr/local/bin/migrate && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

EXPOSE 5000

CMD ["./spacetrouble"]