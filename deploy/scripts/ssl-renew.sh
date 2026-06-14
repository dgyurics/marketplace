#!/bin/bash

# SSL renewal script
# - Renews SSL certificate for the given Docker volume
# - NOTE: Stop nginx manually before running (e.g. docker compose stop nginx)
#         certbot --standalone binds to ports 80 and 443 directly, so nginx must not be running

set -euo pipefail

print_info() { echo "[INFO] $1"; }
print_error() { echo "[ERROR] $1"; }

if [ "$#" -ne 1 ]; then
    print_error "Usage: $0 <ssl-volume-name>"
    exit 1
fi

VOLUME="$1"

cd "$(dirname "$0")/../prod"

if ! docker volume inspect "$VOLUME" >/dev/null 2>&1; then
    print_error "Volume '$VOLUME' not found."
    exit 1
fi

print_info "Renewing SSL certificate for volume '$VOLUME'..."
docker run --rm \
    -v "$VOLUME:/etc/letsencrypt" \
    -p 80:80 \
    -p 443:443 \
    certbot/certbot:v2.11.0 renew --force-renewal --standalone --logs-dir /etc/letsencrypt/logs -v || {
    print_error "SSL renewal failed."
    exit 1
}

print_info "SSL renewal complete."