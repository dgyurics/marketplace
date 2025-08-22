#!/bin/bash

set -euo pipefail # Exit on errors, undefined bariables, and pipe failures

# Verify script being ran from project root
if [[ ! -f "go.mod" ]] || [[ ! -d "deploy/prod" ]]; then
  echo "Error: Script must be ran from project root"
  exit 1
fi

# Step 1: Copy example.env to .env
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

# Step 2: Replace {{DOMAIN}} with user input
replace_domain() {
  echo "Setting up domain configuration..."

  read -p "Enter your domain name (e.g., marketplace.com): " -r domain

  if [[ -z "$domain" ]]; then
    echo "Error: Domain name cannot be empty"
    exit 1
  fi

  # Replace {{DOMAIN}} in .env file
  sed -i "s/{{DOMAIN}}/$domain/g" "deploy/prod/.env"
  echo "Replaced {{DOMAIN}} with $domain in deploy/prod/.env"
}

replace_domain

# Generate HMAC_SECRET and set value in deploy/prod/.env where HMAC_SECRET=
# Generate IMGPROXY_KEY and IMGPROXY_SALT and set value in deploy/prod/.env

# GENERATE RSA KEYS
# openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
# openssl rsa -pubout -in private.pem -out public.pem