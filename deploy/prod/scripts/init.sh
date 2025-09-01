#!/bin/bash

set -euo pipefail

# =============================================================================
# Production Environment Setup Script
# =============================================================================
# This script initializes the production environment by:
# 1. Setting up .env configuration file
# 2. Configuring domain and SSL certificates
# 3. Generating cryptographic keys and secrets
# 4. Setting up third-party service credentials
# =============================================================================

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Paths
readonly PROD_DIR="deploy/prod"
readonly ENV_FILE="$PROD_DIR/.env"
readonly EXAMPLE_ENV="$PROD_DIR/example.env"
readonly NGINX_CONF="$PROD_DIR/nginx.conf"
readonly PRIVATE_KEY="$PROD_DIR/private.pem"
readonly PUBLIC_KEY="$PROD_DIR/public.pem"

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

backup_file() {
  local file="$1"
  local backup="${file}.backup.$(date +%Y%m%d_%H%M%S)"
  cp "$file" "$backup"
  log_success "Backed up $file to $backup"
}

replace_placeholder() {
  local placeholder="$1"
  local value="$2"
  local file="${3:-$ENV_FILE}"
  
  # Cross-platform sed: detect OS and use appropriate syntax
  if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' "s/{{$placeholder}}/$value/g" "$file"
  else
    # Linux and others
    sed -i "s/{{$placeholder}}/$value/g" "$file"
  fi
}

# Validation
validate_environment() {
  log_info "Validating environment..."
  
  if [[ ! -f "go.mod" ]] || [[ ! -d "$PROD_DIR" ]]; then
    log_error "Script must be run from project root"
    exit 1
  fi
  
  if [[ ! -f "$EXAMPLE_ENV" ]]; then
    log_error "$EXAMPLE_ENV not found"
    exit 1
  fi
  
  log_success "Environment validation passed"
}

# Environment file setup
setup_env_file() {
  log_info "Setting up environment file..."
  
  if [[ -f "$ENV_FILE" ]]; then
    log_warning "$ENV_FILE already exists"
    if confirm_action "Overwrite existing .env?"; then
      backup_file "$ENV_FILE"
    else
      log_info "Existing .env preserved. Exiting."
      exit 0
    fi
  fi
  
  cp "$EXAMPLE_ENV" "$ENV_FILE"
  log_success "Created $ENV_FILE from template"
}

# Domain configuration
# Domain configuration
setup_domain() {
  log_info "Configuring domain..."
  
  read -p "Enter your domain name (e.g., marketplace.com): " -r domain
  validate_input "$domain" "Domain name"
  
  # Replace domain in both .env and nginx.conf
  replace_placeholder "DOMAIN" "$domain" "$ENV_FILE"
  replace_placeholder "DOMAIN" "$domain" "$NGINX_CONF"
  
  log_success "Domain '$domain' configured in $ENV_FILE and $NGINX_CONF"
}

# Cryptographic setup
setup_auto_generated_secrets() {
  log_info "Setting up auto-generated secrets..."
  
  # HMAC Secret
  if confirm_action "Auto-generate HMAC_SECRET?" "Y"; then
    hmac_secret=$(openssl rand -hex 32)
    replace_placeholder "HMAC_SECRET" "$hmac_secret"
    log_success "Generated and set HMAC_SECRET"
  else
    read -p "Enter your HMAC_SECRET: " -r hmac_secret
    validate_input "$hmac_secret" "HMAC_SECRET"
    replace_placeholder "HMAC_SECRET" "$hmac_secret"
    log_success "Set custom HMAC_SECRET"
  fi
  
  # ImgProxy Keys
  if confirm_action "Auto-generate IMGPROXY keys?" "Y"; then
    imgproxy_key=$(openssl rand -hex 32)
    imgproxy_salt=$(openssl rand -hex 32)
    replace_placeholder "IMGPROXY_KEY" "$imgproxy_key"
    replace_placeholder "IMGPROXY_SALT" "$imgproxy_salt"
    log_success "Generated and set IMGPROXY_KEY and IMGPROXY_SALT"
  else
    read -p "Enter your IMGPROXY_KEY: " -r imgproxy_key
    validate_input "$imgproxy_key" "IMGPROXY_KEY"
    read -p "Enter your IMGPROXY_SALT: " -r imgproxy_salt
    validate_input "$imgproxy_salt" "IMGPROXY_SALT"
    replace_placeholder "IMGPROXY_KEY" "$imgproxy_key"
    replace_placeholder "IMGPROXY_SALT" "$imgproxy_salt"
    log_success "Set custom IMGPROXY keys"
  fi
}

setup_rsa_keys() {
  log_info "Setting up RSA key pair for JWT..."
  
  if [[ -f "$PRIVATE_KEY" || -f "$PUBLIC_KEY" ]]; then
    log_warning "RSA keys already exist"
    if confirm_action "Regenerate RSA keys?"; then
      backup_file "$PRIVATE_KEY"
      backup_file "$PUBLIC_KEY"
    else
      log_success "Using existing RSA keys"
      return 0
    fi
  fi
  
  # Generate private key (2048 bits)
  openssl genpkey -algorithm RSA -out "$PRIVATE_KEY" -pkeyopt rsa_keygen_bits:2048
  
  # Generate public key from private key
  openssl rsa -pubout -in "$PRIVATE_KEY" -out "$PUBLIC_KEY"
  
  log_success "Generated RSA key pair:"
  echo "  Private key: $PRIVATE_KEY"
  echo "  Public key:  $PUBLIC_KEY"
}

# Third-party service credentials
setup_stripe_credentials() {
  log_info "Setting up Stripe credentials..."
  
  read -p "Enter your STRIPE_WEBHOOK_SIGNING_SECRET: " -r webhook_secret
  validate_input "$webhook_secret" "STRIPE_WEBHOOK_SIGNING_SECRET"
  
  read -p "Enter your STRIPE_SECRET_KEY: " -r secret_key
  validate_input "$secret_key" "STRIPE_SECRET_KEY"
  
  read -p "Enter your STRIPE_PUBLISHABLE_KEY: " -r publishable_key
  validate_input "$publishable_key" "STRIPE_PUBLISHABLE_KEY"
  
  replace_placeholder "STRIPE_WEBHOOK_SIGNING_SECRET" "$webhook_secret"
  replace_placeholder "STRIPE_SECRET_KEY" "$secret_key"
  replace_placeholder "STRIPE_PUBLISHABLE_KEY" "$publishable_key"

  if confirm_action "Enable Test Mode?" "Y"; then
    replace_placeholder "TEST_MODE" "true"
    log_success "Test Mode enabled"
  else
    replace_placeholder "TEST_MODE" "false"
    log_success "Test Mode disabled"
  fi
  
  log_success "Stripe credentials configured"
}

setup_mail_credentials() {
  log_info "Setting up Mailjet credentials..."
  
  read -p "Enter your MAIL_API_KEY: " -r api_key
  validate_input "$api_key" "MAIL_API_KEY"
  
  read -p "Enter your MAIL_API_SECRET: " -r api_secret
  validate_input "$api_secret" "MAIL_API_SECRET"
  
  read -p "Enter your MAIL_FROM_EMAIL: " -r from_email
  validate_input "$from_email" "MAIL_FROM_EMAIL"
  
  read -p "Enter your MAIL_FROM_NAME: " -r from_name
  validate_input "$from_name" "MAIL_FROM_NAME"
  
  replace_placeholder "MAIL_API_KEY" "$api_key"
  replace_placeholder "MAIL_API_SECRET" "$api_secret"
  replace_placeholder "MAIL_FROM_EMAIL" "$from_email"
  replace_placeholder "MAIL_FROM_NAME" "$from_name"
  
  log_success "Mailjet credentials configured"
}

# SSL setup
setup_ssl_certificates() {
  local domain="$1"
  
  if [[ ! -f "$PROD_DIR/scripts/ssl.sh" ]]; then
      log_error "SSL script not found: $PROD_DIR/scripts/ssl.sh"
      exit 1
  fi
  
  log_info "Setting up SSL certificates..."
  # shellcheck source=./ssl.sh
  source "$PROD_DIR/scripts/ssl.sh"
  init_ssl "$domain"
}

# Main execution
# Main execution
main() {
  echo "========================================"
  echo "Production Environment Setup"
  echo "========================================"
  echo
  
  # Validation
  validate_environment
  
  # Environment setup
  setup_env_file
  
  # Domain configuration
  setup_domain
  
  # Extract domain from .env file for SSL setup (be specific to avoid STRIPE_BASE_URL)
  domain=$(grep "^BASE_URL=https://" "$ENV_FILE" | cut -d'/' -f3)
  
  # Cryptographic setup
  setup_auto_generated_secrets
  setup_rsa_keys
  
  # Third-party credentials
  setup_stripe_credentials
  setup_mail_credentials
  
  # SSL certificates
  setup_ssl_certificates "$domain"
  
  echo
  echo "========================================"
  log_success "Production environment setup complete!"
  echo "========================================"
  echo
  echo "Next steps:"
  echo "1. Review your configuration in $ENV_FILE"
  echo "2. Run 'docker compose -f deploy/prod/docker-compose.yaml up -d' to start services"
  echo "3. Verify your application is accessible at https://$domain"
}

# Run main function
main "$@"
