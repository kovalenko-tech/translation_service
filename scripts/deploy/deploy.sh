#!/bin/bash

# Quick deployment script for production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Translation API Production Deployment ===${NC}"
echo

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo -e "${RED}Error: Do not run this script as root${NC}"
    exit 1
fi

# Check if Docker is installed
if ! command -v docker &> /dev/null; then
    echo -e "${RED}Error: Docker is not installed${NC}"
    exit 1
fi

# Check if Docker Compose is installed
if ! command -v docker-compose &> /dev/null; then
    echo -e "${RED}Error: Docker Compose is not installed${NC}"
    exit 1
fi

# Check if .env.prod exists
if [ ! -f ".env.prod" ]; then
    echo -e "${YELLOW}Warning: .env.prod not found${NC}"
    echo "Creating .env.prod from template..."
    cp env.prod.example .env.prod
    echo -e "${YELLOW}Please edit .env.prod with your configuration and run this script again${NC}"
    exit 1
fi

# Load environment variables
export $(cat .env.prod | grep -v '^#' | xargs)

# Check if domain is set
if [ -z "$DOMAIN" ] || [ "$DOMAIN" = "your-domain.com" ]; then
    echo -e "${RED}Error: Please set your domain in .env.prod${NC}"
    exit 1
fi

echo -e "${GREEN}Domain: $DOMAIN${NC}"
echo -e "${GREEN}Email: $CERTBOT_EMAIL${NC}"
echo

# Confirm deployment
read -p "Do you want to proceed with deployment? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Deployment cancelled"
    exit 0
fi

echo
echo -e "${BLUE}Step 1: Building Docker images...${NC}"
make prod-build

echo
echo -e "${BLUE}Step 2: Checking SSL certificates...${NC}"
if [ ! -d "certbot/conf/live/$DOMAIN" ]; then
    echo "SSL certificates not found. Initializing..."
    make ssl-init
else
    echo -e "${GREEN}SSL certificates found${NC}"
fi

echo
echo -e "${BLUE}Step 3: Starting production services...${NC}"
make prod-up

echo
echo -e "${BLUE}Step 4: Waiting for services to start...${NC}"
sleep 10

echo
echo -e "${BLUE}Step 5: Running health check...${NC}"
make health-check

echo
echo -e "${GREEN}=== Deployment completed successfully! ===${NC}"
echo
echo -e "${BLUE}Your application is now available at:${NC}"
echo -e "${GREEN}https://$DOMAIN${NC}"
echo
echo -e "${BLUE}Useful commands:${NC}"
echo -e "  make prod-logs    - View logs"
echo -e "  make prod-status  - Check service status"
echo -e "  make health-check - Run health check"
echo -e "  make prod-down    - Stop services"
echo
echo -e "${YELLOW}Don't forget to:${NC}"
echo -e "  • Set up automatic SSL renewal: 0 */12 * * * $(pwd)/scripts/ssl/cron-renew.sh"
echo -e "  • Configure monitoring and alerts"
echo -e "  • Set up regular backups" 