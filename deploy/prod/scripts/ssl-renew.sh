#!/bin/bash
# Auto-renewal script for SSL certificates

set -euo pipefail

# Change to project directory
cd "$(dirname "$0")/../../.."

echo "$(date): Starting SSL renewal check"

# Try to renew certificates (only renews if within 30 days of expiry)
docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" \
    certbot/certbot:v2.11.0 renew --quiet

# Reload nginx if containers are running
if docker-compose ps nginx | grep -q "Up"; then
  docker-compose exec nginx nginx -s reload 2>/dev/null || true
fi

echo "$(date): SSL renewal check completed"