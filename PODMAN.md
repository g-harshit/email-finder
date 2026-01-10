# Running with Podman

This guide shows how to run the Email Finder service using Podman instead of Docker.

## Prerequisites

- ✅ Podman (version 4.0+)
- ✅ podman-compose (for easier management)

## Quick Start

1. **Start all services:**
   ```bash
   podman-compose up -d --build
   ```

2. **Check status:**
   ```bash
   podman ps
   ```

3. **View logs:**
   ```bash
   podman-compose logs -f
   # or individual services
   podman logs email-finder
   podman logs check-if-email-exists
   ```

4. **Stop services:**
   ```bash
   podman-compose down
   ```

## Differences from Docker

### Image Registry
The `docker-compose.yml` has been updated to use `docker.io` instead of `ghcr.io` for better Podman compatibility:
- ✅ `docker.io/reacherhq/check-if-email-exists:latest` (works with Podman)
- ❌ `ghcr.io/reacherhq/check-if-email-exists:latest` (may have auth issues)

### Platform Warnings
You may see platform warnings (amd64 vs arm64) - this is normal and Podman handles it automatically.

## Manual Podman Commands

If you prefer not to use podman-compose:

### Build the email-finder image:
```bash
podman build -t email-finder:latest .
```

### Run check-if-email-exists:
```bash
podman run -d \
  --name check-if-email-exists \
  -p 8081:8081 \
  docker.io/reacherhq/check-if-email-exists:latest
```

### Run email-finder:
```bash
podman run -d \
  --name email-finder \
  -p 8080:8080 \
  --network podman \
  -e EMAIL_VERIFICATION_API_URL=http://check-if-email-exists:8081 \
  email-finder:latest
```

## Troubleshooting

### Image Pull Issues
If you get 403 errors pulling images:
- Try using `docker.io/` prefix instead of `ghcr.io/`
- Check your Podman registry configuration

### Network Issues
If services can't communicate:
```bash
# Create a podman network
podman network create email-finder-network

# Use --network flag when running containers
podman run --network email-finder-network ...
```

### Port Conflicts
If ports 8080 or 8081 are already in use:
- Change ports in `docker-compose.yml`
- Or stop conflicting services

## Verify Service is Running

```bash
# Health check
curl http://localhost:8080/health

# Test API
curl -X POST http://localhost:8080/api/v1/find-email \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "company": "Zepto"
  }'
```

## Service URLs

- **Email Finder API**: http://localhost:8080
- **check-if-email-exists**: http://localhost:8081
