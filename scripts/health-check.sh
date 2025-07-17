#!/bin/bash

# Health check script for production environment

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables
if [ -f .env.prod ]; then
    export $(cat .env.prod | grep -v '^#' | xargs)
fi

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ $2${NC}"
    else
        echo -e "${RED}✗ $2${NC}"
    fi
}

echo "=== Production Health Check ==="
echo "Domain: ${DOMAIN:-'Not set'}"
echo "Time: $(date)"
echo

# Check if docker-compose.prod.yml exists
if [ ! -f "docker-compose.prod.yml" ]; then
    echo -e "${RED}✗ docker-compose.prod.yml not found${NC}"
    exit 1
fi

# Check if .env.prod exists
if [ ! -f ".env.prod" ]; then
    echo -e "${YELLOW}⚠ .env.prod not found${NC}"
else
    echo -e "${GREEN}✓ .env.prod found${NC}"
fi

echo
echo "=== Service Status ==="

# Check if services are running
if docker-compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo -e "${GREEN}✓ Services are running${NC}"
    
    # Check individual services
    services=("nginx" "app" "redis" "rabbitmq")
    
    for service in "${services[@]}"; do
        if docker-compose -f docker-compose.prod.yml ps | grep -q "$service.*Up"; then
            print_status 0 "$service is running"
        else
            print_status 1 "$service is not running"
        fi
    done
else
    echo -e "${RED}✗ No services are running${NC}"
    echo "Run 'make prod-up' to start services"
fi

echo
echo "=== SSL Certificate Status ==="

# Check SSL certificates
if [ -d "certbot/conf/live/${DOMAIN}" ]; then
    echo -e "${GREEN}✓ SSL certificates exist${NC}"
    
    # Check certificate expiration
    if command -v openssl >/dev/null 2>&1; then
        cert_file="certbot/conf/live/${DOMAIN}/fullchain.pem"
        if [ -f "$cert_file" ]; then
            expiry=$(openssl x509 -enddate -noout -in "$cert_file" | cut -d= -f2)
            echo "Certificate expires: $expiry"
            
            # Check if certificate expires in less than 30 days
            expiry_date=$(date -d "$expiry" +%s 2>/dev/null || date -j -f "%b %d %H:%M:%S %Y %Z" "$expiry" +%s 2>/dev/null)
            current_date=$(date +%s)
            days_left=$(( (expiry_date - current_date) / 86400 ))
            
            if [ $days_left -lt 30 ]; then
                echo -e "${YELLOW}⚠ Certificate expires in $days_left days${NC}"
            else
                echo -e "${GREEN}✓ Certificate is valid for $days_left days${NC}"
            fi
        fi
    fi
else
    echo -e "${RED}✗ SSL certificates not found${NC}"
    echo "Run 'make ssl-init' to get SSL certificates"
fi

echo
echo "=== Network Connectivity ==="

# Check if domain resolves
if [ -n "$DOMAIN" ]; then
    if nslookup "$DOMAIN" >/dev/null 2>&1; then
        print_status 0 "Domain $DOMAIN resolves"
    else
        print_status 1 "Domain $DOMAIN does not resolve"
    fi
fi

# Check if ports are accessible
if command -v nc >/dev/null 2>&1; then
    if nc -z localhost 80 2>/dev/null; then
        print_status 0 "Port 80 (HTTP) is accessible"
    else
        print_status 1 "Port 80 (HTTP) is not accessible"
    fi
    
    if nc -z localhost 443 2>/dev/null; then
        print_status 0 "Port 443 (HTTPS) is accessible"
    else
        print_status 1 "Port 443 (HTTPS) is not accessible"
    fi
fi

echo
echo "=== Application Health ==="

# Check application health endpoint
if [ -n "$DOMAIN" ]; then
    if command -v curl >/dev/null 2>&1; then
        # Try HTTPS first
        if curl -f -s -k "https://$DOMAIN/api/health" >/dev/null 2>&1; then
            print_status 0 "Application health check (HTTPS) passed"
        elif curl -f -s "http://$DOMAIN/api/health" >/dev/null 2>&1; then
            print_status 0 "Application health check (HTTP) passed"
        else
            print_status 1 "Application health check failed"
        fi
    fi
fi

echo
echo "=== Resource Usage ==="

# Check resource usage
if command -v docker >/dev/null 2>&1; then
    echo "Container resource usage:"
    docker stats --no-stream --format "table {{.Container}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
fi

echo
echo "=== Recent Logs ==="

# Show recent logs
echo "Recent nginx logs:"
docker-compose -f docker-compose.prod.yml logs --tail=5 nginx 2>/dev/null || echo "No nginx logs available"

echo
echo "Recent application logs:"
docker-compose -f docker-compose.prod.yml logs --tail=5 app 2>/dev/null || echo "No application logs available"

echo
echo "=== Recommendations ==="

# Provide recommendations
if [ ! -f ".env.prod" ]; then
    echo -e "${YELLOW}• Create .env.prod file from env.prod.example${NC}"
fi

if [ ! -d "certbot/conf/live/${DOMAIN}" ]; then
    echo -e "${YELLOW}• Run 'make ssl-init' to get SSL certificates${NC}"
fi

if ! docker-compose -f docker-compose.prod.yml ps | grep -q "Up"; then
    echo -e "${YELLOW}• Run 'make prod-up' to start services${NC}"
fi

echo
echo "Health check completed at $(date)" 