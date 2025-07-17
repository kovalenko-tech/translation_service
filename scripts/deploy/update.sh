#!/bin/bash

# Quick update script for production

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Translation API Production Update ===${NC}"
echo

# Check if running as root
if [ "$EUID" -eq 0 ]; then
    echo -e "${RED}Error: Do not run this script as root${NC}"
    exit 1
fi

# Check if .env.prod exists
if [ ! -f ".env.prod" ]; then
    echo -e "${RED}Error: .env.prod not found${NC}"
    exit 1
fi

# Load environment variables
export $(cat .env.prod | grep -v '^#' | xargs)

echo -e "${GREEN}Domain: $DOMAIN${NC}"
echo

# Confirm update
read -p "Do you want to proceed with update? (y/N): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "Update cancelled"
    exit 0
fi

echo
echo -e "${BLUE}Step 1: Stopping production services...${NC}"
make prod-down

echo
echo -e "${BLUE}Step 2: Pulling latest changes...${NC}"
git pull

echo
echo -e "${BLUE}Step 3: Building new Docker images...${NC}"
make prod-build

echo
echo -e "${BLUE}Step 4: Starting production services...${NC}"
make prod-up

echo
echo -e "${BLUE}Step 5: Waiting for services to start...${NC}"
sleep 10

echo
echo -e "${BLUE}Step 6: Running health check...${NC}"
make health-check

echo
echo -e "${GREEN}=== Update completed successfully! ===${NC}"
echo
echo -e "${BLUE}Your application is now available at:${NC}"
echo -e "${GREEN}https://$DOMAIN${NC}"
echo
echo -e "${BLUE}Useful commands:${NC}"
echo -e "  make prod-logs    - View logs"
echo -e "  make prod-status  - Check service status"
echo -e "  make health-check - Run health check" 