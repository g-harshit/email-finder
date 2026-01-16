#!/bin/bash
set -e

echo "🔨 Building Email Finder for Render..."

# Download check-if-email-exists CLI binary for Linux
echo "📥 Downloading check-if-email-exists binary..."
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
    BINARY_URL="https://github.com/reacherhq/check-if-email-exists/releases/latest/download/check_if_email_exists-x86_64-unknown-linux-gnu.tar.gz"
elif [ "$ARCH" = "aarch64" ]; then
    BINARY_URL="https://github.com/reacherhq/check-if-email-exists/releases/latest/download/check_if_email_exists-aarch64-unknown-linux-gnu.tar.gz"
else
    echo "⚠️  Unknown architecture: $ARCH, trying x86_64..."
    BINARY_URL="https://github.com/reacherhq/check-if-email-exists/releases/latest/download/check_if_email_exists-x86_64-unknown-linux-gnu.tar.gz"
fi

# Download and extract
curl -L -o check_if_email_exists.tar.gz "$BINARY_URL" || {
    echo "⚠️  Failed to download binary, will use HTTP API mode instead"
    exit 0
}

tar -xzf check_if_email_exists.tar.gz
chmod +x check_if_email_exists
mv check_if_email_exists check_if_email_exists_linux || true
rm check_if_email_exists.tar.gz

echo "✅ Binary downloaded successfully"

# Build Go application
echo "🔨 Building Go application..."
go mod download
go build -buildvcs=false -o email-finder ./cmd/server

echo "✅ Build complete!"
