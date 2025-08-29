#!/bin/bash

set -euo pipefail

# =============================================================================
# Production Environment Secure Socket Layer (SSL) Setup Script
# =============================================================================
# This script initializes the production environment by:
# 1. Checking for existing SSL certificates
# 2. Obtaining or renewing SSL certificates via Let's Encrypt
# =============================================================================

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Utility functions
log_info() {
  echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
  echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
  echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $1" >&2
}

confirm_action() {
  local prompt="$1"
  local default="${2:-N}"
  
  if [[ "$default" == "Y" ]]; then
    read -p "$prompt (Y/n): " -n 1 -r
  else
    read -p "$prompt (y/N): " -n 1 -r
  fi
  echo
  
  if [[ "$default" == "Y" ]]; then
    [[ "$REPLY" != "n" && "$REPLY" != "N" ]]
  else
    [[ "$REPLY" == "y" || "$REPLY" == "Y" ]]
  fi
}

validate_input() {
  local input="$1"
  local field_name="$2"
  
  if [[ -z "$input" ]]; then
    log_error "$field_name cannot be empty"
    exit 1
  fi
}

# SSL certificate functions
check_existing_certificates() {
  local domain="$1"
  
  log_info "Checking for existing SSL certificates for $domain..."
  
  if docker run --rm -v marketplace_ssl-certs:/certs alpine test -f "/certs/live/$domain/fullchain.pem"; then
    log_warning "SSL certificates already exist for $domain"
    
    if confirm_action "Do you want to renew existing certificates?"; then
      renew_certificates "$domain"
    else
      log_info "Using existing certificates"
    fi
    return 0
  fi
  
  return 1
}

renew_certificates() {
  local domain="$1"
  
  log_info "Renewing SSL certificates for $domain..."
  
  if docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" \
      certbot/certbot:v2.11.0 renew --force-renewal; then
    log_success "SSL certificates renewed successfully"
  else
    log_error "Failed to renew SSL certificates"
    return 1
  fi
}

obtain_certificates() {
  local domain="$1"
  
  log_warning "No SSL certificates found for $domain"
  
  if ! confirm_action "Do you want to obtain SSL certificates?" "Y"; then
    log_info "Skipping SSL certificate setup"
    return 0
  fi
  
  log_info "Obtaining SSL certificates for domain: $domain"
  
  # Shut down all containers to free port 80
  log_info "Stopping containers to free port 80..."
  docker-compose down
  
  # GET SSL CERTIFICATES using Let's Encrypt
  log_info "Requesting certificates from Let's Encrypt..."
  if docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" \
      -p 80:80 \
      certbot/certbot:v2.11.0 certonly --standalone \
      --non-interactive --agree-tos --no-eff-email \
      --register-unsafely-without-email \
      -d "$domain" -d "www.$domain"; then
    
    log_success "SSL certificates obtained and stored in marketplace_ssl-certs volume"
    
    # Verify certificates are in the volume
    log_info "Verifying certificates in volume..."
    docker run --rm -v marketplace_ssl-certs:/certs alpine ls -la "/certs/live/$domain/"
    log_success "SSL certificate setup complete!"
  else
    log_error "Failed to obtain SSL certificates"
    return 1
  fi
}

init_ssl() {
  local domain="$1"

  validate_input "$domain" "Domain"

  # Check for existing certificates first
  if check_existing_certificates "$domain"; then
    return 0
  fi
  
  # No certificates found, try to obtain them
  obtain_certificates "$domain"
}

# Setup auto-renewal cron job
setup_auto_renewal() {
  local domain="$1"

  log_info "Setting up auto-renewal for $domain..."
  
  # Add to crontab (runs daily at 2:30 AM)
  PROJECT_ROOT=$(pwd)
  CRON_JOB="30 2 * * * $PROJECT_ROOT/deploy/prod/scripts/ssl-renew.sh >> $PROJECT_ROOT/logs/ssl-renewal.log 2>&1"
  
  # Check if cron job already exists
  if ! crontab -l 2>/dev/null | grep -q "ssl-renew.sh"; then
    (crontab -l 2>/dev/null; echo "$CRON_JOB") | crontab -
    log_success "Auto-renewal cron job added (daily at 2:30 AM)"
  else
    log_warning "Auto-renewal cron job already exists"
  fi
  
  # Create logs directory
  mkdir -p logs
  log_success "Auto-renewal setup complete!"
}

# Allow script to be run standalone
if [[ "${BASH_SOURCE[0]}" == "${0}" ]]; then
  if [[ $# -eq 0 ]]; then
    log_error "Usage: $0 <domain>"
    exit 1
  fi
  init_ssl "$1"
fi
