# Email Finder Service

A production-ready email finder service that generates possible email addresses based on first name, last name, and company, then verifies which emails actually exist. This service is a backend replica of [Dropcontact](https://www.dropcontact.com/).

## Features

- 🎯 **Smart Email Pattern Generation**: Generates 20+ common email patterns (firstname.lastname, f.lastname, etc.)
- ✅ **Email Verification**: Integrates with [check-if-email-exists](https://github.com/reacherhq/check-if-email-exists) to verify email deliverability
- 🚀 **Production Ready**: Built with Go, includes proper error handling, logging, and configuration
- 🐳 **Docker Support**: Easy deployment with Docker and Docker Compose
- 📊 **Confidence Scoring**: Returns emails with confidence levels (high, medium, low)
- 🔍 **RESTful API**: Clean REST API for easy integration

## Architecture

```
┌─────────────┐
│   Client    │
└──────┬──────┘
       │
       │ HTTP Request
       │ (first_name, last_name, company)
       ▼
┌─────────────────────┐
│   Email Finder API  │
│   (Gin HTTP Server) │
└──────────┬──────────┘
           │
           │ 1. Generate Patterns
           ▼
┌─────────────────────┐
│  Email Generator    │
│  (Pattern Engine)   │
└──────────┬──────────┘
           │
           │ 2. Verify Emails
           ▼
┌─────────────────────┐
│  Email Verifier     │
│  (HTTP/CLI Client)  │
└──────────┬──────────┘
           │
           │ 3. Check Email Existence
           ▼
┌─────────────────────┐
│ check-if-email-     │
│ exists Service      │
│ (Rust Backend)      │
└─────────────────────┘
```

## Prerequisites

### Option 1: Docker Compose (Recommended - Easiest)
- ✅ **Docker** (version 20.10+)
- ✅ **Docker Compose** (version 2.0+)

That's it! Docker Compose handles everything automatically.

### Option 2: Local Development
- ✅ **Go 1.21 or higher**
- ✅ **check-if-email-exists service** (runs via Docker or CLI binary)

**See [PREREQUISITES.md](PREREQUISITES.md) for detailed installation instructions.**

## Quick Start

### Option 1: Using Docker Compose (Recommended)

1. Clone the repository:
```bash
git clone <repository-url>
cd Email-Finder
```

2. Start all services:
```bash
make docker-up
# or
docker-compose up -d
```

This will start:
- Email Finder API on `http://localhost:8080`
- check-if-email-exists service on `http://localhost:8081`

### Option 2: Local Development

1. Install dependencies:
```bash
make install-deps
# or
go mod download
```

2. Set up environment variables:
```bash
cp .env.example .env
# Edit .env with your configuration
```

3. Start the check-if-email-exists service:
```bash
# Option A: Using Docker
docker run -d -p 8081:8081 ghcr.io/reacherhq/check-if-email-exists:latest

# Option B: Using the CLI binary
# Download from https://github.com/reacherhq/check-if-email-exists/releases
# Set EMAIL_VERIFICATION_CLI_PATH in .env
```

4. Run the service:
```bash
make run
# or
go run ./cmd/server
```

## API Usage

### Find Emails

**Endpoint:** `POST /api/v1/find-email`

**Request:**
```json
{
  "first_name": "John",
  "last_name": "Doe",
  "company": "Google"
}
```

**Note:** The `company` field can be either:
- A company name (e.g., "Google", "Microsoft Inc", "Acme Corporation") - the service will automatically resolve it to a domain
- A domain name (e.g., "example.com", "google.com") - will be used directly

**Response:**
```json
{
  "found_emails": [
    {
      "email": "john.doe@google.com",
      "pattern": "firstname.lastname",
      "is_reachable": "safe",
      "is_valid": true,
      "is_deliverable": true,
      "confidence": "high"
    },
    {
      "email": "j.doe@google.com",
      "pattern": "f.lastname",
      "is_reachable": "risky",
      "is_valid": true,
      "is_deliverable": true,
      "confidence": "medium"
    }
  ],
  "total_checked": 20,
  "total_found": 2,
  "domain": "google.com",
  "domain_resolved": true,
  "request": {
    "first_name": "John",
    "last_name": "Doe",
    "company": "Google"
  }
}
```

### Health Check

**Endpoint:** `GET /health`

**Response:**
```json
{
  "status": "healthy",
  "service": "email-finder"
}
```

## Configuration

Configuration is done via environment variables. See `.env.example` for all available options:

| Variable | Description | Default |
|----------|-------------|---------|
| `SERVER_PORT` | Server port | `8080` |
| `SERVER_HOST` | Server host | `0.0.0.0` |
| `EMAIL_VERIFICATION_API_URL` | URL of check-if-email-exists HTTP API | `http://localhost:8081` |
| `EMAIL_VERIFICATION_CLI_PATH` | Path to CLI binary (if using CLI mode) | `` |
| `LOG_LEVEL` | Logging level (debug, info, warn, error) | `info` |
| `LOG_FORMAT` | Log format (json, text) | `json` |
| `RATE_LIMIT` | Rate limit per IP (requests per minute) | `60` |
| `VERIFICATION_TIMEOUT` | Timeout for email verification (seconds) | `30` |
| `MAX_EMAIL_PATTERNS` | Maximum patterns to generate | `20` |

## Domain Resolution

The service automatically resolves company names to domains using multiple strategies (in order of priority):

1. **In-Memory Company Map**: For well-known companies, uses a pre-built map for instant resolution (e.g., "Zepto" → "zeptonow.com", "Google" → "google.com")
2. **Direct Domain Detection**: If the input already looks like a domain (contains a dot), it's used directly
3. **DNS Verification**: Attempts to verify domains via DNS lookups (MX, A, CNAME records) for companies not in the map
4. **Pattern Matching**: Generates common domain patterns (company.com, company.io, company.co, etc.) as fallback

**Examples:**
- "Zepto" → "zeptonow.com" (from company map - instant, no DNS lookup)
- "Google" → "google.com" (from company map)
- "Microsoft Inc" → "microsoft.com" (from company map, handles suffixes)
- "Acme Corporation" → "acme.com" (via DNS verification or pattern matching)
- "example.com" → "example.com" (used directly)

**Well-Known Companies Included:**
The service includes an in-memory map of 100+ well-known companies across:
- Tech companies (Google, Microsoft, Apple, Amazon, Meta, etc.)
- Indian companies (Zepto, Swiggy, Zomato, Flipkart, Razorpay, etc.)
- Financial services (Goldman Sachs, JPMorgan, Visa, etc.)
- Consulting firms (McKinsey, BCG, Deloitte, PwC, etc.)
- And many more...

For companies not in the map, the service falls back to DNS verification and pattern matching. The resolved domain is included in the API response, so you know which domain was used for email generation.

## Email Patterns Generated

The service generates **~200 email patterns** including:

### Base Patterns (20)
1. `firstname.lastname@company.com`
2. `firstnamelastname@company.com`
3. `f.lastname@company.com`
4. `flastname@company.com`
5. `firstname.l@company.com`
6. `firstnamel@company.com`
7. `firstname@company.com`
8. `lastname@company.com`
9. `lastname.firstname@company.com`
10. `lastnamefirstname@company.com`
11. `l.firstname@company.com`
12. `lfirstname@company.com`
13. `f_lastname@company.com`
14. `firstname_lastname@company.com`
15. `lastname_firstname@company.com`
16. `fl@company.com`
17. `firstname.f.lastname@company.com`
18. `lastname.f@company.com`
19. `f.firstname.lastname@company.com`
20. `firstname-lastname@company.com`

### Numbered Variations
- **Single digits (0-9)**: `firstname.lastname0` through `firstname.lastname9` (30 patterns)
- **Double digits (1-50)**: `firstname.lastname1` through `firstname.lastname50` (150 patterns)
  - Also includes: `firstnamelastname1-50` and `f.lastname1-50`

**Total: ~200 unique email patterns per request**

All patterns are verified in parallel for optimal performance.

## Project Structure

```
Email-Finder/
├── cmd/
│   └── server/
│       └── main.go          # Application entry point
├── config/
│   └── config.go            # Configuration management
├── internal/
│   ├── generator/
│   │   └── email_generator.go  # Email pattern generation
│   ├── resolver/
│   │   └── domain_resolver.go  # Domain resolution from company name
│   ├── verifier/
│   │   └── email_verifier.go   # Email verification logic
│   ├── service/
│   │   └── email_finder_service.go  # Business logic
│   └── handler/
│       └── email_handler.go      # HTTP handlers
├── .env.example            # Example environment variables
├── Dockerfile              # Docker build file
├── docker-compose.yml      # Docker Compose configuration
├── Makefile                # Build automation
├── go.mod                  # Go module definition
└── README.md              # This file
```

## Development

### Build
```bash
make build
```

### Run Tests
```bash
make test
```

### Format Code
```bash
make fmt
```

### Run Linter
```bash
make lint
```

### Clean Build Artifacts
```bash
make clean
```

## Docker Commands

### Build Docker Image
```bash
make docker-build
```

### Start Services
```bash
make docker-up
```

### Stop Services
```bash
make docker-down
```

### View Logs
```bash
make docker-logs
```

## Integration with check-if-email-exists

This service integrates with [check-if-email-exists](https://github.com/reacherhq/check-if-email-exists) in two ways:

1. **HTTP API Mode** (Recommended): The service makes HTTP requests to the check-if-email-exists HTTP backend
2. **CLI Mode**: The service calls the check-if-email-exists CLI binary directly

**Note:** The HTTP API endpoint format may vary depending on your check-if-email-exists setup. The default endpoint is `/v0/check_email`. If your setup uses a different endpoint, you may need to modify the `VerifyEmail` function in `internal/verifier/email_verifier.go` or adjust the API URL configuration.

### Setting up check-if-email-exists

**Option 1: Using Docker (Recommended)**
```bash
docker run -d -p 8081:8081 ghcr.io/reacherhq/check-if-email-exists:latest
```

**Option 2: Building from Source**
See the [check-if-email-exists documentation](https://github.com/reacherhq/check-if-email-exists) for building instructions.

## Production Deployment

### Environment Variables
Ensure all required environment variables are set in your production environment.

### Health Checks
The service exposes a `/health` endpoint that can be used for health checks in production.

### Rate Limiting
Configure rate limiting based on your needs using the `RATE_LIMIT` environment variable.

### Monitoring
The service uses structured logging (JSON format) which can be easily integrated with log aggregation tools.

## License

This project is licensed under the MIT License.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## Acknowledgments

- [check-if-email-exists](https://github.com/reacherhq/check-if-email-exists) - Email verification library
- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [Dropcontact](https://www.dropcontact.com/) - Inspiration for this service
