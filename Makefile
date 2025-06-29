.PHONY: help start stop restart build logs setup-config clean test

# Default target
help:
	@echo "Available commands:"
	@echo "  setup-config  - Create config.yaml from template"
	@echo "  start         - Start the API Gateway"
	@echo "  stop          - Stop the API Gateway"
	@echo "  restart       - Restart the API Gateway"
	@echo "  build         - Build the Docker image"
	@echo "  logs          - Show container logs"
	@echo "  test          - Test the /ping endpoint"
	@echo "  clean         - Remove containers and images"

# Setup configuration from template
setup-config:
	@if [ ! -f config.yaml ]; then \
		cp config.yaml.template config.yaml; \
		echo "✅ Created config.yaml from template"; \
	else \
		echo "⚠️  config.yaml already exists"; \
	fi

# Start the API Gateway
start: setup-config
	@echo "🚀 Starting API Gateway..."
	docker-compose up --build -d
	@echo "✅ API Gateway started at http://localhost:8080"

# Stop the API Gateway
stop:
	@echo "🛑 Stopping API Gateway..."
	docker-compose down
	@echo "✅ API Gateway stopped"

# Restart the API Gateway
restart: stop start

# Build the Docker image
build:
	@echo "🔨 Building Docker image..."
	docker-compose build
	@echo "✅ Docker image built"

# Show container logs
logs:
	docker-compose logs -f

# Test the /ping endpoint
test:
	@echo "🧪 Testing /ping endpoint..."
	@curl -s http://localhost:8080/ping | jq . || echo "❌ Test failed - is the service running?"

# Clean up containers and images
clean:
	@echo "🧹 Cleaning up..."
	docker-compose down --rmi all --volumes --remove-orphans
	@echo "✅ Cleanup complete"

# Development helpers
dev-start: start logs

# Check if service is healthy
health:
	@echo "🏥 Checking service health..."
	@docker-compose ps