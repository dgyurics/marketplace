#!/bin/bash

set -euo pipefail # Exit on errors, undefined bariables, and pipe failures

# Verify script being ran from project root
if [[ ! -f "go.mod" ]] || [[ ! -d "deploy/prod" ]]; then
  echo "Error: Script must be ran from project root"
  exit 1
fi

# Copy example.env --> .env
copy_example_env() {
  echo "Copying deploy/prod/example.env to deploy/prod/.env..."

  if [[ ! -f "deploy/prod/example.env" ]]; then
    echo "Error: deploy/prod/example.env not found"
    exit 1
  fi

  if [[ -f "deploy/prod/.env" ]]; then
    echo "Warning: deploy/prod/.env already exists"
    read -p "Overwrite existing .env? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Aborted. Existing .env preserved."
      exit 1
    fi
    cp "deploy/prod/.env" "deploy/prod/.env.backup.$(date +%Y%m%d_%H%M%S)"
    echo "Backed up existing .env"
  fi

  cp "deploy/prod/example.env" "deploy/prod/.env"
  echo "Successfully created deploy/prod/.env"
}

copy_example_env

# Replace {{DOMAIN}}
replace_domain() {
  echo "Setting up domain configuration..." >&2

  read -p "Enter your domain name (e.g., marketplace.com): " -r user_domain

  if [[ -z "$user_domain" ]]; then
    echo "Error: Domain name cannot be empty"
    exit 1
  fi

  # Replace {{DOMAIN}} in .env file
  sed -i '' "s/{{DOMAIN}}/$user_domain/g" "deploy/prod/.env"
  echo "Replaced {{DOMAIN}} with $user_domain in deploy/prod/.env" >&2

  # Return the domain
  echo "$user_domain"
}

domain=$(replace_domain)

# Replace {{HMAC_SECRET}}
setup_hmac_secret() {
  echo "Setting up HMAC_SECRET..."

  read -p "Auto-generate HMAC_SECRET? (Y/n): " -n 1 -r
  echo

  if [[ "$REPLY" == "n" || "$REPLY" == "N" ]]; then
    read -p "Enter your HMAC_SECRET: " -r hmac_secret
    if [[ -z "$hmac_secret" ]]; then
      echo "Error: HMAC_SECRET cannot be empty"
      exit 1
    fi
  else
    hmac_secret=$(openssl rand -hex 32)
    echo "Generated HMAC_SECRET: $hmac_secret"
  fi

  # Replace placeholder in .env
  sed -i '' "s/{{HMAC_SECRET}}/$hmac_secret/g" "deploy/prod/.env"
  echo "HMAC_SECRET set in deploy/prod/.env"
}

setup_hmac_secret

# Generate RSA Keys (used by JWT)
setup_rsa_keys() {
  echo "Setting up RSA keys for JWT..."

  if [[ -f "deploy/prod/private.pem" || -f "deploy/prod/public.pem" ]]; then
    echo "Warning: RSA keys already exists"
    read -p "Regenerate RSA keys? (y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
      echo "Using existing RSA keys"
      return 0
    fi
    echo "Backing up existing keys..."
    cp "deploy/prod/private.pem" "deploy/prod/private.pem.backup.$(date +%Y%m%d_%H%M%S)"
    cp "deploy/prod/public.pem" "deploy/prod/public.pem.backup.$(date +%Y%m%d_%H%M%S)"
  fi

  echo "Generating RSA key pair..."

  # Generate private key (2048 bits)
  openssl genpkey -algorithm RSA -out "deploy/prod/private.pem" -pkeyopt rsa_keygen_bits:2048

  # Generate public key from private key
  openssl rsa -pubout -in "deploy/prod/private.pem" -out "deploy/prod/public.pem" 

  echo "Generated RSA keys:"
  echo "  Private key: deploy/prod/private.pem"
  echo "  Public key:  deploy/prod/public.pem"  
}

setup_rsa_keys

# Replace {{IMGPROXY_KEY}} {{IMGPROXY_SALT}}
setup_imgproxy_keys() {
  echo "Setting up IMGPROXY_KEY and IMGPROXY_SALT..."

  read -p "Auto-generate IMGPROXY_KEY and IMGPROXY_SALT? (Y/n): " -n 1 -r
  echo

  if [[ "$REPLY" == "n" || "$REPLY" == "N" ]]; then
    read -p "Enter your IMGPROXY_KEY: " -r imgproxy_key
    if [[ -z "$imgproxy_key" ]]; then
      echo "Error: IMGPROXY_KEY cannot be empty"
      exit 1
    fi
    read -p "Enter your IMGPROXY_SALT: " -r imgproxy_salt
    if [[ -z "$imgproxy_salt" ]]; then
      echo "Error: IMGPROXY_SALT cannot be empty"
      exit 1
    fi    
  else
    imgproxy_key=$(openssl rand -hex 32)
    echo "Generated IMGPROXY_KEY: $imgproxy_key"
    imgproxy_salt=$(openssl rand -hex 32)
    echo "Generated IMGPROXY_SALT: $imgproxy_salt"
  fi

  # Replace placeholder in .env
  sed -i '' "s/{{IMGPROXY_KEY}}/$imgproxy_key/g" "deploy/prod/.env"
  echo "IMGPROXY_KEY set in deploy/prod/.env"

  # Replace placeholder in .env
  sed -i '' "s/{{IMGPROXY_SALT}}/$imgproxy_salt/g" "deploy/prod/.env"
  echo "IMGPROXY_SALT set in deploy/prod/.env"  
}

setup_imgproxy_keys

# Replace {{STRIPE_WEBHOOK_SIGNING_SECRET}} {{STRIPE_SECRET_KEY}} {{STRIPE_PUBLISHABLE_KEY}}
setup_stripe_secrets() {
  echo "Setting up STRIPE_WEBHOOK_SIGNING_SECRET, STRIPE_SECRET_KEY, STRIPE_PUBLISHABLE_KEY..."

  read -p "Enter your STRIPE_WEBHOOK_SIGNING_SECRET: " -r stripe_webhook_signing_secret
  if [[ -z "$stripe_webhook_signing_secret" ]]; then
    echo "Error: STRIPE_WEBHOOK_SIGNING_SECRET cannot be empty"
    exit 1
  fi

  read -p "Enter your STRIPE_SECRET_KEY: " -r stripe_secret_key
  if [[ -z "$stripe_secret_key" ]]; then
    echo "Error: STRIPE_SECRET_KEY cannot be empty"
    exit 1
  fi

  read -p "Enter your STRIPE_PUBLISHABLE_KEY: " -r stripe_publishable_key
  if [[ -z "$stripe_publishable_key" ]]; then
    echo "Error: STRIPE_PUBLISHABLE_KEY cannot be empty"
    exit 1
  fi

  # Replace placeholder in .env
  sed -i '' "s/{{STRIPE_WEBHOOK_SIGNING_SECRET}}/$stripe_webhook_signing_secret/g" "deploy/prod/.env"
  echo "STRIPE_WEBHOOK_SIGNING_SECRET set in deploy/prod/.env"

  # Replace placeholder in .env
  sed -i '' "s/{{STRIPE_SECRET_KEY}}/$stripe_secret_key/g" "deploy/prod/.env"
  echo "STRIPE_SECRET_KEY set in deploy/prod/.env"

  # Replace placeholder in .env
  sed -i '' "s/{{STRIPE_PUBLISHABLE_KEY}}/$stripe_publishable_key/g" "deploy/prod/.env"
  echo "STRIPE_PUBLISHABLE_KEY set in deploy/prod/.env"
}

setup_stripe_secrets

# Replace {{MAIL_API_KEY}} {{MAIL_API_SECRET}} {{MAIL_FROM_EMAIL}} {{MAIL_FROM_NAME}}
setup_mailjet() {
  echo "Setting up MAIL_API_KEY, MAIL_API_SECRET, MAIL_FROM_EMAIL, MAIL_FROM_NAME..."

  read -p "Enter your MAIL_API_KEY: " -r mail_api_key
  if [[ -z "$mail_api_key" ]]; then
    echo "Error: MAIL_API_KEY cannot be empty"
    exit 1
  fi

  read -p "Enter your MAIL_API_SECRET: " -r mail_api_secret
  if [[ -z "$mail_api_secret" ]]; then
    echo "Error: MAIL_API_SECRET cannot be empty"
    exit 1
  fi

  read -p "Enter your MAIL_FROM_EMAIL: " -r mail_from_email
  if [[ -z "$mail_from_email" ]]; then
    echo "Error: MAIL_FROM_EMAIL cannot be empty"
    exit 1
  fi
 
   read -p "Enter your MAIL_FROM_NAME: " -r mail_from_name
  if [[ -z "$mail_from_name" ]]; then
    echo "Error: MAIL_FROM_NAME cannot be empty"
    exit 1
  fi

  # Replace placeholder in .env
  sed -i '' "s/{{MAIL_API_KEY}}/$mail_api_key/g" "deploy/prod/.env"
  echo "MAIL_API_KEY set in deploy/prod/.env"

  # Replace placeholder in .env
  sed -i '' "s/{{MAIL_API_SECRET}}/$mail_api_secret/g" "deploy/prod/.env"
  echo "MAIL_API_SECRET set in deploy/prod/.env"

  # Replace placeholder in .env
  sed -i '' "s/{{MAIL_FROM_EMAIL}}/$mail_from_email/g" "deploy/prod/.env"
  echo "MAIL_FROM_EMAIL set in deploy/prod/.env"

  # Replace placeholder in .env
  sed -i '' "s/{{MAIL_FROM_NAME}}/$mail_from_name/g" "deploy/prod/.env"
  echo "MAIL_FROM_NAME set in deploy/prod/.env"
}

setup_mailjet

# Generate SSL certificates and store them in docker volume
# references ./ssl.sh script
setup_ssl() {
  read -p "Setup SSL certificates? (Y/n): " -n 1 -r
  echo

  if [[ "$REPLY" == "n" || "$REPLY" == "N" ]]; then
    return 0
  else
    # Source the SSL functions
    source deploy/prod/scripts/ssl.sh
    echo "Setting up SSL certificates..."
    init_ssl "$domain"  
  fi   
}

setup_ssl
