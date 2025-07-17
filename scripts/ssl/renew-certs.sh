#!/bin/bash

# Load environment variables
if [ -f .env.prod ]; then
    export $(cat .env.prod | grep -v '^#' | xargs)
fi

echo "Renewing SSL certificates..."

# Renew certificates
docker-compose -f docker-compose.prod.yml run --rm certbot renew

# Reload nginx to use renewed certificates
docker-compose -f docker-compose.prod.yml exec nginx nginx -s reload

echo "SSL certificates renewed successfully!" 