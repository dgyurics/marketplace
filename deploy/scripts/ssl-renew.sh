#!/bin/bash

# SSL renewal script
# - Renews the shared SSL certificate through the running nginx webroot
# Usage: ./deploy/scripts/ssl-renew.sh <ssl-volume-name> [--dry-run]

set -euo pipefail

print_info() { echo "[INFO] $1"; }
print_error() { echo "[ERROR] $1"; }

if [ "$#" -lt 1 ]; then
    print_error "Usage: $0 <ssl-volume-name> [--dry-run]"
    exit 1
fi

VOLUME="$1"
shift

cd "$(dirname "$0")/../prod"

if ! docker volume inspect "$VOLUME" >/dev/null 2>&1; then
    print_error "Volume '$VOLUME' not found."
    exit 1
fi

print_info "Renewing SSL certificate for volume '$VOLUME'..."
docker run --rm \
    -v "$VOLUME:/etc/letsencrypt" \
    certbot/certbot:v5.8.0 renew --cert-name shared \
    --webroot -w /etc/letsencrypt/webroot \
    --non-interactive --no-random-sleep-on-renew \
    --logs-dir /etc/letsencrypt/logs "$@" || {
    print_error "SSL renewal failed."
    exit 1
}

docker exec nginx nginx -t
docker exec nginx nginx -s reload

print_info "SSL renewal complete."
