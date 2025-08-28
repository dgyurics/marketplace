#!/bin/bash
# ssl.sh - SSL setup functions

set -euo pipefail

# Setup SSL certificates
init_ssl() {
  local domain="$1"

  if [[ -z "$domain" ]]; then
    echo "Error: Domain parameter required"
    return 1
  fi

  # Check if certificates already exist
  if docker run --rm -v marketplace_ssl-certs:/certs alpine test -f "/certs/live/$domain/fullchain.pem"; then
    echo "SSL certificates already exist for $domain"
    read -p "Renew certificates? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Using existing certificates"
      return 0
    fi
  fi

  echo "Setting up SSL for domain: $domain"

  # GET SSL CERTIFICATES using Let's Encrypt
  docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" \
      -p 80:80 \
      certbot/certbot:v2.11.0 certonly --standalone \
      --non-interactive --agree-tos --no-eff-email \
      --register-unsafely-without-email \
      -d "$domain" -d "www.$domain"
  echo "SSL certificates obtained and stored in marketplace_ssl-certs volume"

  # Verify certificates are in the volume
  echo "Verifying certificates in volume..."
  docker run --rm -v marketplace_ssl-certs:/certs alpine ls -la /certs/live/$domain/
  echo "SSL certificate setup complete!"
}

# Setup auto-renewal cron job
setup_auto_renewal() {
  local domain="$1"

  if [[ -z "$domain" ]]; then
    echo "Error: Domain parameter required"
    return 1
  fi

  echo "Setting up auto-renewal for $domain..."
  # TODO auto-renewal logic
}
