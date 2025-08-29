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
    read -p "Do you want to renew existing certificates? (y/N): " -n 1 -r
    echo
    if [[ "$REPLY" != "y" && "$REPLY" != "Y" ]]; then
      echo "Using existing certificates"
      return 0
    fi
    
    echo "Renewing SSL certificates for $domain..."
    docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" \
        certbot/certbot:v2.11.0 renew --force-renewal
    echo "SSL certificates renewed successfully"
    return 0
  fi

  # No certificates found - ask to obtain them
  echo "No SSL certificates found for $domain"
  read -p "Do you want to obtain SSL certificates? (Y/n): " -n 1 -r
  echo
  if [[ "$REPLY" == "n" || "$REPLY" == "N" ]]; then
    echo "Skipping SSL certificate setup"
    return 0
  fi
  
  echo "Obtaining SSL certificates for domain: $domain"
  
  # Shut down all containers to free port 80
  echo "Stopping containers to free port 80..."
  docker-compose down
  
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

# FIXME not yet used/tested
# Setup auto-renewal cron job
setup_auto_renewal() {
  local domain="$1"

  echo "Setting up auto-renewal for $domain..."
  
  # Add to crontab (runs daily at 2:30 AM)
  PROJECT_ROOT=$(pwd)
  CRON_JOB="30 2 * * * $PROJECT_ROOT/deploy/prod/scripts/ssl-renew.sh >> $PROJECT_ROOT/logs/ssl-renewal.log 2>&1"
  
  # Check if cron job already exists
  if ! crontab -l 2>/dev/null | grep -q "ssl-renew.sh"; then
    (crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -
    echo "Auto-renewal cron job added (daily at 2:30 AM)"
  else
    echo "Auto-renewal cron job already exists"
  fi
  
  # Create logs directory
  mkdir -p logs
  
  echo "Auto-renewal setup complete!"
}