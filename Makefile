.PHONY: clear backend backend-logs database database-logs all up down restart

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

up:
	docker compose up -d

down:
	docker compose down

restart: down up
