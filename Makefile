# Translation Service Makefile
# Author: Kyrylo Kovalenko (git@kovalenko.tech)
# Website: https://kovalenko.tech

.PHONY: help build run test clean deps docker-up docker-down docker-build docker-run docker-logs docker-clean dev setup deps-up deps-down deps-logs deps-clean swagger generate-api-key

# Variables
BINARY_NAME=translation-server
BUILD_DIR=build
DOCKER_IMAGE=translation-app
DOCKER_TAG=latest

help: ## Show help
	@echo "Available commands:"
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

generate-api-key: ## Generate a secure API key
	@echo "Generating secure API key..."
	@go run scripts/generate_api_key.go

deps: ## Install dependencies
	go mod download
	go mod tidy

swagger: ## Generate Swagger documentation
	swag init -g cmd/server/main.go -o docs

build: ## Build application
	mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/server

run: ## Run application locally
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v ./...

clean: ## Clean build artifacts
	rm -rf $(BUILD_DIR)

# Docker commands
docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE):$(DOCKER_TAG) .

docker-run: ## Run application in Docker
	docker run -p 8080:8080 --env-file .env $(DOCKER_IMAGE):$(DOCKER_TAG)

docker-up: ## Start all services with Docker Compose
	docker-compose up -d

docker-down: ## Stop all Docker Compose services
	docker-compose down

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f

docker-clean: ## Clean Docker images and containers
	docker-compose down -v
	docker system prune -f
	docker image prune -f

# Dependencies only commands
deps-up: ## Start only dependencies (Redis, RabbitMQ)
	docker-compose -f docker-compose.deps.yml up -d

deps-down: ## Stop only dependencies
	docker-compose -f docker-compose.deps.yml down

deps-logs: ## Show dependencies logs
	docker-compose -f docker-compose.deps.yml logs -f

deps-clean: ## Clean dependencies containers and volumes
	docker-compose -f docker-compose.deps.yml down -v

# Development commands
dev: deps-up ## Start development environment (dependencies + application)
	@echo "Starting dependencies..."
	@sleep 5
	@echo "Starting application..."
	@make run

setup: deps deps-up ## Setup project (dependencies + Docker)
	@echo "Project setup complete! Don't forget to create .env file with your OpenAI API key"
	@echo "Copy env.example to .env and add your OPENAI_API_KEY"

# Production commands
prod-build: ## Build for production
	docker-compose -f docker-compose.prod.yml build

prod-up: ## Start production environment
	docker-compose -f docker-compose.prod.yml up -d

prod-down: ## Stop production environment
	docker-compose -f docker-compose.prod.yml down

prod-logs: ## Show production logs
	docker-compose -f docker-compose.prod.yml logs -f

prod-status: ## Show production service status
	docker-compose -f docker-compose.prod.yml ps

# SSL Certificate management
ssl-init: ## Initialize SSL certificates
	./scripts/ssl/init-letsencrypt.sh

ssl-renew: ## Renew SSL certificates
	./scripts/ssl/renew-certs.sh

# Production setup
prod-setup: ## Setup production environment
	@echo "Setting up production environment..."
	@echo "1. Copy env.prod.example to .env.prod and configure your settings"
	@echo "2. Update nginx/conf.d/default.conf with your domain"
	@echo "3. Run 'make ssl-init' to get SSL certificates"
	@echo "4. Run 'make prod-up' to start the production stack"

health-check: ## Check production environment health
	./scripts/health-check.sh

deploy: ## Deploy to production
	./scripts/deploy/deploy.sh

update: ## Update production application
	./scripts/deploy/update.sh

# Utility commands
logs: ## Show application logs
	docker-compose logs -f app

redis-cli: ## Connect to Redis CLI
	docker exec -it translation-redis redis-cli

rabbitmq-management: ## Open RabbitMQ management interface
	@echo "RabbitMQ Management: http://localhost:15672"
	@echo "Username: guest, Password: guest"

redis-commander: ## Open Redis Commander interface
	@echo "Redis Commander: http://localhost:8081"

status: ## Show service status
	docker-compose ps

deps-status: ## Show dependencies status
	docker-compose -f docker-compose.deps.yml ps 