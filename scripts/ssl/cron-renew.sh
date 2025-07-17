#!/bin/bash

# Script for automatic SSL certificate renewal via cron
# Add to crontab: 0 */12 * * * /path/to/translation/scripts/ssl/cron-renew.sh

# Change to project directory
cd "$(dirname "$0")/../.."

# Load environment variables
if [ -f .env.prod ]; then
    export $(cat .env.prod | grep -v '^#' | xargs)
fi

# Log file
LOG_FILE="./logs/ssl-renewal.log"

# Create logs directory if it doesn't exist
mkdir -p ./logs

# Function to log messages
log() {
    echo "$(date '+%Y-%m-%d %H:%M:%S') - $1" | tee -a "$LOG_FILE"
}

log "Starting SSL certificate renewal check..."

# Check if certificates need renewal
if docker-compose -f docker-compose.prod.yml run --rm certbot certificates | grep -q "VALID"; then
    log "Certificates are valid, checking if renewal is needed..."
    
    # Try to renew certificates
    if docker-compose -f docker-compose.prod.yml run --rm certbot renew --quiet; then
        log "Certificates renewed successfully"
        
        # Reload nginx to use new certificates
        if docker-compose -f docker-compose.prod.yml exec nginx nginx -s reload; then
            log "Nginx reloaded successfully"
        else
            log "ERROR: Failed to reload nginx"
            exit 1
        fi
    else
        log "No certificates needed renewal"
    fi
else
    log "ERROR: No valid certificates found"
    exit 1
fi

log "SSL certificate renewal check completed" 