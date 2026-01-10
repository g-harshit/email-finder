# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

# Install git (needed for some Go dependencies)
RUN apk add --no-cache git

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o email-finder ./cmd/server

# Final stage - use Debian for better binary compatibility
FROM debian:bookworm-slim

# Install required libraries for check_if_email_exists binary
RUN apt-get update && \
    apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    libc6 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/email-finder .

# Copy check_if_email_exists Linux binary (for CLI mode)
# Prefer check_if_email_exists_linux (ARM64), fallback to any Linux binary
COPY check_if_email_exists_linux* check_if_email_exists* ./
RUN if [ -f check_if_email_exists_linux ]; then \
        mv check_if_email_exists_linux check_if_email_exists && \
        chmod +x check_if_email_exists; \
    elif ls check_if_email_exists* 2>/dev/null | grep -v "check_if_email_exists$" | head -1 | xargs file | grep -q "ELF.*Linux"; then \
        mv $(ls check_if_email_exists* | grep -v "check_if_email_exists$" | head -1) check_if_email_exists && \
        chmod +x check_if_email_exists; \
    fi

EXPOSE 8080

CMD ["./email-finder"]
