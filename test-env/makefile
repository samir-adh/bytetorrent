.PHONY: start stop clean status logs restart

# Start the environment
start:
	@./start.sh

# Stop containers (keep data)
stop:
	@docker-compose stop

# Clean everything
clean:
	@./clean.sh

# Show status
status:
	@./status.sh

# Show logs
logs:
	@docker-compose logs -f

# Restart containers
restart: stop start

# Quick test - create and check
test: start
	@sleep 5
	@echo "Testing environment..."
	@./status.sh