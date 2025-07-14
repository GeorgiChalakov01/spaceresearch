.PHONY: clear backend database

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

database:
	@echo "=== DataBase ==="
	@echo "Building image..."
	@cd database && docker buildx build -t spaceresearch-postgres .
	@echo "Recreating container..."
	@docker compose up -d --no-deps --force-recreate database
