# Variables
COMPOSE_FILE := docker_compose.yml
PROJECT_NAME := payment_gateway

# Default target
.PHONY: help
help:
	@echo "Usage:"
	@echo "  make build       Build the application and services"
	@echo "  make start       Start the services"
	@echo "  make stop        Stop the services"
	@echo "  make deploy      Build and start the services (combined build and start)"
	@echo "  make clean       Remove all containers, volumes, and networks"
	@echo "  make logs        Tail the logs of all services"
	@echo "  make app-logs    Tail the logs of the app service"
	@echo "  make restart     Restart all services"

# Build the application and services
.PHONY: build
build:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) build

# Start the services
.PHONY: start
start:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) up

# Stop the services
.PHONY: stop
stop:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) stop

# Clean up containers, volumes, and networks
.PHONY: clean
clean:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) down -v --remove-orphans

# Combined build and start
.PHONY: deploy
deploy: build start

# Tail logs for all services
.PHONY: logs
logs:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f

# Tail logs for the app service
.PHONY: app-logs
app-logs:
	docker-compose -f $(COMPOSE_FILE) -p $(PROJECT_NAME) logs -f app

# Restart the services
.PHONY: restart
restart:
	$(MAKE) stop
	$(MAKE) start
