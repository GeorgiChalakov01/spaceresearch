.PHONY: clear backend backend-logs database database-logs objectstorage objectstorage-logs messagebroker messagebroker-logs documentprocessor documentprocessor-logs all up down restart

clear:
	@clear

backend:
	@echo "=== Backend ==="
	@echo "Compiling source code..."
	@cd backend/app && make compile
	@echo "Building image..."
	@cd backend && docker buildx build -t spaceresearch .
	@echo "Recreating container..."
	@docker compose up -d --no-deps --force-recreate backend

backend-logs:
	@docker logs --follow spaceresearch-backend-1

database:
	@echo "=== DataBase ==="
	@echo "Building image..."
	@cd database && docker buildx build -t spaceresearch-postgres .
	@echo "Recreating container..."
	@docker compose up -d --no-deps --force-recreate database

database-logs:
	@docker logs --follow spaceresearch-database-1

objectstorage:
	@echo "=== ObjectStorage ==="
	@echo "Building image..."
	@cd objectstorage && docker buildx build -t spaceresearch-objectstorage .
	@echo "Recreating container..."
	@docker compose up -d --no-deps --force-recreate objectstorage

objectstorage-logs:
	@docker logs --follow spaceresearch-objectstorage-1

messagebroker:
	@echo "=== MessageBroker ==="
	@echo "Recreating container..."
	@docker compose up -d --no-deps --force-recreate messagebroker

messagebroker-logs:
	@docker logs --follow spaceresearch-messagebroker-1

documentprocessor:
	@echo "=== DocProc ==="
	@echo "Compiling source code..."
	@cd documentprocessor/src && make compile
	@echo "Building image..."
	@cd documentprocessor && docker buildx build -t spaceresearch-documentprocessor .
	@echo "Recreating container..."
	@docker compose up -d --no-deps --force-recreate documentprocessor

documentprocessor-logs:
	@docker logs --follow spaceresearch-documentprocessor-1

all: backend database objectstorage messagebroker documentprocessor

up:
	docker compose up -d

down:
	docker compose down

restart: down up
