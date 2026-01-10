# Prerequisites

This document outlines what you need to run the Email Finder service.

## Option 1: Docker Compose (Recommended - Easiest)

**Minimum Requirements:**
- ✅ **Docker** (version 20.10 or higher)
- ✅ **Docker Compose** (version 2.0 or higher)

**That's it!** Docker Compose will automatically:
- Pull the required images (including check-if-email-exists)
- Set up networking between services
- Configure everything automatically

**Check if you have Docker:**
```bash
docker --version
docker-compose --version
# or for newer versions
docker compose version
```

**Install Docker (if needed):**
- **macOS**: Download [Docker Desktop](https://www.docker.com/products/docker-desktop)
- **Linux**: Follow [Docker installation guide](https://docs.docker.com/engine/install/)
- **Windows**: Download [Docker Desktop](https://www.docker.com/products/docker-desktop)

## Option 2: Local Development

**Requirements:**
- ✅ **Go 1.21 or higher**
- ✅ **check-if-email-exists service** (can run via Docker or CLI binary)

**Check if you have Go:**
```bash
go version
```

**Install Go (if needed):**
- Download from [golang.org](https://golang.org/dl/)
- Or use package manager: `brew install go` (macOS), `apt install golang-go` (Ubuntu)

**For check-if-email-exists, you have two options:**

### Option A: Using Docker (Easier)
- Just need Docker installed (same as Option 1)
- Run: `docker run -d -p 8081:8081 ghcr.io/reacherhq/check-if-email-exists:latest`

### Option B: Using CLI Binary
- Download from [releases page](https://github.com/reacherhq/check-if-email-exists/releases)
- Extract and set `EMAIL_VERIFICATION_CLI_PATH` in `.env` file

## Summary

| Setup Method | Prerequisites | Complexity |
|-------------|---------------|------------|
| **Docker Compose** | Docker + Docker Compose | ⭐ Easiest |
| **Local (Docker)** | Go + Docker | ⭐⭐ Easy |
| **Local (CLI)** | Go + Binary Download | ⭐⭐⭐ Moderate |

## Quick Verification

After installation, verify everything works:

**Docker Compose:**
```bash
docker-compose up -d
curl http://localhost:8080/health
```

**Local Development:**
```bash
go version  # Should show 1.21+
docker ps   # Should show check-if-email-exists running (if using Docker)
```

## No Additional Services Required

✅ **No database needed** - Everything is in-memory  
✅ **No message queue needed** - Direct HTTP calls  
✅ **No external APIs required** - Uses open-source check-if-email-exists  
✅ **No API keys needed** - Fully self-contained  

## Network Requirements

The service needs:
- **Outbound internet access** for:
  - DNS lookups (domain resolution)
  - Email verification (SMTP connections)
  - Pulling Docker images (first time only)

No inbound firewall rules needed unless exposing to external networks.

## System Resources

**Minimum:**
- 512 MB RAM
- 1 CPU core
- 1 GB disk space

**Recommended:**
- 1 GB RAM
- 2 CPU cores
- 2 GB disk space

## Troubleshooting

**Docker not starting?**
- Ensure Docker Desktop is running (macOS/Windows)
- Check Docker daemon: `docker info`

**Port conflicts?**
- Default ports: 8080 (Email Finder), 8081 (check-if-email-exists)
- Change in `docker-compose.yml` or `.env` if needed

**Go version issues?**
- Ensure Go 1.21+: `go version`
- Update if needed: `brew upgrade go` or download from golang.org
