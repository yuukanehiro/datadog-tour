.PHONY: help up down logs build restart clean test-api

DOCKER_COMPOSE := docker-compose -f docker/docker-compose.yml --env-file .env

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

up: ## Start all services
	$(DOCKER_COMPOSE) up -d

down: ## Stop all services
	$(DOCKER_COMPOSE) down

logs: ## Show logs from all services
	$(DOCKER_COMPOSE) logs -f

logs-api: ## Show logs from API service
	$(DOCKER_COMPOSE) logs -f api

logs-frontend: ## Show logs from Frontend service
	$(DOCKER_COMPOSE) logs -f frontend

logs-datadog: ## Show logs from Datadog agent
	$(DOCKER_COMPOSE) logs -f datadog

build: ## Build and start all services
	$(DOCKER_COMPOSE) up -d --build

restart: ## Restart all services
	$(DOCKER_COMPOSE) restart

clean: ## Remove all containers, volumes, and images
	$(DOCKER_COMPOSE) down -v
	$(DOCKER_COMPOSE) rm -f

test-api: ## Run API tests
	@echo "Testing health endpoint..."
	@curl -s http://localhost:8080/health | jq
	@echo "\nCreating a user..."
	@curl -s -X POST http://localhost:8080/api/users \
		-H "Content-Type: application/json" \
		-d '{"name":"Test User","email":"test@example.com"}' | jq
	@echo "\nGetting all users..."
	@curl -s http://localhost:8080/api/users | jq
	@echo "\nSetting cache..."
	@curl -s -X POST http://localhost:8080/api/cache/set \
		-H "Content-Type: application/json" \
		-d '{"key":"test-key","value":"test-value"}' | jq
	@echo "\nGetting cache..."
	@curl -s http://localhost:8080/api/cache/get/test-key | jq
	@echo "\nTesting slow endpoint..."
	@curl -s http://localhost:8080/api/slow | jq
	@echo "\nTesting error endpoint..."
	@curl -s http://localhost:8080/api/error | jq

mysql-cli: ## Connect to MySQL CLI
	$(DOCKER_COMPOSE) exec mysql mysql -u demouser -pdemopassword datadog_demo

redis-cli: ## Connect to Redis CLI
	$(DOCKER_COMPOSE) exec redis redis-cli

status: ## Show status of all services
	$(DOCKER_COMPOSE) ps
