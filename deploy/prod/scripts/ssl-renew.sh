#!/bin/bash

# SSL renewal script
cd "$(dirname "$0")/../../.."

# Stop nginx
docker compose -f deploy/prod/docker-compose.yaml stop nginx

# Renew SSL certificate
docker run --rm -v "marketplace_ssl-certs:/etc/letsencrypt" -p 80:80 -p 443:443 \
    certbot/certbot:v2.11.0 renew --force-renewal --standalone

# Start nginx
docker compose -f deploy/prod/docker-compose.yaml start nginx