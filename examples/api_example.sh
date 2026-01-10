#!/bin/bash

# Example API usage script for Email Finder Service

BASE_URL="http://localhost:8080"

echo "=== Email Finder API Examples ==="
echo ""

# Health Check
echo "1. Health Check:"
curl -s "${BASE_URL}/health" | jq .
echo ""
echo "---"
echo ""

# Find Emails Example 1 - Using company name (domain will be resolved automatically)
echo "2. Finding emails for John Doe at Google (domain auto-resolved):"
curl -s -X POST "${BASE_URL}/api/v1/find-email" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "last_name": "Doe",
    "company": "Google"
  }' | jq .
echo ""
echo "---"
echo ""

# Find Emails Example 2 - Using company name with suffix
echo "3. Finding emails for Jane Smith at Microsoft Inc (domain auto-resolved):"
curl -s -X POST "${BASE_URL}/api/v1/find-email" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Jane",
    "last_name": "Smith",
    "company": "Microsoft Inc"
  }' | jq .
echo ""
echo "---"
echo ""

# Find Emails Example 3 - Using direct domain
echo "4. Finding emails using direct domain:"
curl -s -X POST "${BASE_URL}/api/v1/find-email" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "Bob",
    "last_name": "Johnson",
    "company": "example.com"
  }' | jq .
echo ""
echo "---"
echo ""

# Invalid Request Example
echo "4. Invalid Request (missing fields):"
curl -s -X POST "${BASE_URL}/api/v1/find-email" \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John"
  }' | jq .
echo ""
