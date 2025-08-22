#!/bin/bash

# chmod +x deploy/prod/init.sh
DOMAIN="marketplace.com"

# Shut down all containers
echo "Stopping all running containers..."
docker-compose down

# GET SSL CERTIFICATES using Let's Encrypt
docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" \
    -p 80:80 \
    certbot/certbot:v2.11.0 certonly --standalone \
    --non-interactive --agree-tos --no-eff-email \
    --register-unsafely-without-email \
    -d "$DOMAIN" -d "www.$DOMAIN"
echo "SSL certificates obtained and stored in marketplace_ssl-certs volume"

# Verify certificates are in the volume
echo "Verifying certificates in volume..."
docker run --rm -v marketplace_ssl-certs:/certs alpine ls -la /certs/live/$DOMAIN/

echo "SSL certificate setup complete!"

# TODO automate certificate renewal
