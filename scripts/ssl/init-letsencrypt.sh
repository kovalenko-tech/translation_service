#!/bin/bash

# Load environment variables
if [ -f .env.prod ]; then
    export $(cat .env.prod | grep -v '^#' | xargs)
fi

# Check if domain is set
if [ -z "$DOMAIN" ]; then
    echo "Error: DOMAIN is not set in .env.prod file"
    exit 1
fi

if [ -z "$CERTBOT_EMAIL" ]; then
    echo "Error: CERTBOT_EMAIL is not set in .env.prod file"
    exit 1
fi

# Create directories for certbot
mkdir -p certbot/conf
mkdir -p certbot/www

# Check if certificates already exist
if [ -d "certbot/conf/live/$DOMAIN" ]; then
    echo "Certificates already exist for $DOMAIN"
    echo "To renew certificates, run: docker-compose -f docker-compose.prod.yml run --rm certbot renew"
    exit 0
fi

echo "Initializing SSL certificates for domain: $DOMAIN"

# Start nginx temporarily for certificate validation
docker-compose -f docker-compose.prod.yml up -d nginx

# Wait for nginx to start
echo "Waiting for nginx to start..."
sleep 10

# Get SSL certificate
docker-compose -f docker-compose.prod.yml run --rm certbot certonly \
    --webroot \
    --webroot-path=/var/www/certbot \
    --email $CERTBOT_EMAIL \
    --agree-tos \
    --no-eff-email \
    -d $DOMAIN \
    -d www.$DOMAIN

# Reload nginx to use new certificates
docker-compose -f docker-compose.prod.yml exec nginx nginx -s reload

echo "SSL certificates initialized successfully!"
echo "You can now start the full production stack with: docker-compose -f docker-compose.prod.yml up -d" 