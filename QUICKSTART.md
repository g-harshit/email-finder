# Quick Start Guide

Get the Email Finder service up and running in minutes!

## Prerequisites

### For Docker Compose (Recommended):
- ✅ Docker (20.10+)
- ✅ Docker Compose (2.0+)

### For Local Development:
- ✅ Go 1.21+
- ✅ check-if-email-exists service (via Docker or CLI)

**That's it!** No database, message queues, or API keys needed.

See [PREREQUISITES.md](PREREQUISITES.md) for detailed setup instructions.

## Option 1: Docker Compose (Fastest - Recommended)

1. **Start the services:**
   ```bash
   docker-compose up -d
   ```

2. **Verify it's running:**
   ```bash
   curl http://localhost:8080/health
   ```

3. **Test the API:**
   ```bash
   curl -X POST http://localhost:8080/api/v1/find-email \
     -H "Content-Type: application/json" \
     -d '{
       "first_name": "John",
       "last_name": "Doe",
       "company": "example.com"
     }'
   ```

4. **View logs:**
   ```bash
   docker-compose logs -f
   ```

5. **Stop the services:**
   ```bash
   docker-compose down
   ```

## Option 2: Local Development

### Step 1: Start check-if-email-exists Service

**Using Docker:**
```bash
docker run -d -p 8081:8081 --name check-if-email-exists \
  ghcr.io/reacherhq/check-if-email-exists:latest
```

**Or download the CLI binary:**
1. Download from [releases](https://github.com/reacherhq/check-if-email-exists/releases)
2. Extract and note the path
3. Set `EMAIL_VERIFICATION_CLI_PATH` in `.env` file

### Step 2: Configure Environment

```bash
cp .env.example .env
# Edit .env if needed (defaults should work)
```

### Step 3: Install Dependencies

```bash
go mod download
```

### Step 4: Run the Service

```bash
go run ./cmd/server
```

Or using Make:
```bash
make run
```

### Step 5: Test the API

```bash
curl -X POST http://localhost:8080/api/v1/find-email \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "company": "example.com"
  }'
```

## API Examples

### Using curl

```bash
# Health check
curl http://localhost:8080/health

# Find emails
curl -X POST http://localhost:8080/api/v1/find-email \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith",
    "company": "acme.com"
  }'
```

### Using the example script

```bash
chmod +x examples/api_example.sh
./examples/api_example.sh
```

### Using Python

```python
import requests

url = "http://localhost:8080/api/v1/find-email"
payload = {
    "first_name": "John",
    "last_name": "Doe",
    "company": "example.com"
}

response = requests.post(url, json=payload)
print(response.json())
```

### Using JavaScript/Node.js

```javascript
const fetch = require('node-fetch');

const url = 'http://localhost:8080/api/v1/find-email';
const payload = {
  first_name: 'John',
  last_name: 'Doe',
  company: 'example.com'
};

fetch(url, {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify(payload)
})
  .then(res => res.json())
  .then(data => console.log(data));
```

## Troubleshooting

### Service won't start

1. **Check if port 8080 is available:**
   ```bash
   lsof -i :8080
   ```

2. **Check logs:**
   ```bash
   docker-compose logs email-finder
   ```

### Email verification not working

1. **Verify check-if-email-exists is running:**
   ```bash
   curl http://localhost:8081/health
   # or
   docker ps | grep check-if-email-exists
   ```

2. **Check environment variables:**
   ```bash
   cat .env
   # Ensure EMAIL_VERIFICATION_API_URL is correct
   ```

3. **Test check-if-email-exists directly:**
   ```bash
   curl -X POST http://localhost:8081/v0/check_email \
     -H "Content-Type: application/json" \
     -d '{"to_email": "test@example.com"}'
   ```

### No emails found

- This is normal! The service only returns emails that are verified as deliverable
- Try with different names/companies
- Check the `total_checked` field to see how many patterns were tested

## Next Steps

- Read the full [README.md](README.md) for detailed documentation
- Check the [API documentation](#api-usage) in README.md
- Customize configuration in `.env`
- Explore the codebase structure

## Production Deployment

For production deployment:

1. Set proper environment variables
2. Use a reverse proxy (nginx, traefik, etc.)
3. Set up monitoring and logging
4. Configure rate limiting
5. Use HTTPS/TLS
6. Set up health checks

See the main README.md for production deployment details.
