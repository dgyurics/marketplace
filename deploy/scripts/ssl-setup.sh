#!/bin/bash

# SSL setup script
# - Obtains initial SSL certificate for the given domain and Docker volume
# Usage: ./deploy/scripts/ssl-setup.sh <domain> <ssl-volume-name>
# - NOTE: Stop nginx manually before running (e.g. docker compose stop nginx)
#         certbot --standalone binds to port 80 directly, so nginx must not be running

set -euo pipefail

print_info() { echo "[INFO] $1"; }
print_error() { echo "[ERROR] $1"; }

if [ "$#" -ne 2 ]; then
    print_error "Usage: $0 <domain> <ssl-volume-name>"
    exit 1
fi

DOMAIN="$1"
VOLUME="$2"

cd "$(dirname "$0")/../prod"

if [ -z "$DOMAIN" ]; then
    print_error "Domain cannot be empty."
    exit 1
fi

if [ -z "$VOLUME" ]; then
    print_error "Volume name cannot be empty."
    exit 1
fi

if ! docker volume inspect "$VOLUME" >/dev/null 2>&1; then
    print_info "Volume '$VOLUME' not found. Creating it..."
    docker volume create "$VOLUME" >/dev/null
fi

if docker run --rm -v "$VOLUME:/certs" alpine test -f "/certs/live/$DOMAIN/fullchain.pem"; then
    print_info "Certificate already exists for '$DOMAIN' in volume '$VOLUME'."
    print_info "If needed, use ssl-renew.sh to force renewal."
    exit 0
fi

print_info "Obtaining SSL certificate for '$DOMAIN' into volume '$VOLUME'..."
docker run --rm \
    -v "$VOLUME:/etc/letsencrypt" \
    -p 80:80 \
    certbot/certbot:v2.11.0 certonly --standalone \
    --non-interactive --agree-tos --no-eff-email \
    --register-unsafely-without-email \
    -d "$DOMAIN" -d "www.$DOMAIN" || {
    print_error "SSL setup failed."
    exit 1
}

print_info "Verifying certificate files..."
docker run --rm -v "$VOLUME:/certs" alpine ls -la "/certs/live/$DOMAIN/"

print_info "SSL setup complete."
