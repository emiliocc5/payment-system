.PHONY: build up down logs clean restart

# Build all services
build:
	docker compose build

# Start all services
up:
	docker compose up -d

# Stop all services
down:
	docker compose down

# View logs
logs:
	docker compose logs -f

# Clean up everything
clean:
	docker compose down -v
	docker system prune -f

# Restart services
restart:
	docker compose down
	docker compose up

# Start only infrastructure
infra:
	docker compose up -d postgres-main rabbitmq kafka prometheus grafana

# Start services
services:
	docker compose up -d payment-wallet-service