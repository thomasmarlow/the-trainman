.PHONY: help start stop restart build logs setup-config clean test test-proxy test-users test-orders test-all

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
	@echo "  test-proxy    - Test proxy functionality"
	@echo "  test-users    - Test users service proxy"
	@echo "  test-orders   - Test orders service proxy"
	@echo "  test-all      - Run all tests"
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

# Test proxy functionality
test-proxy:
	@echo "🧪 Testing proxy functionality..."
	@echo "Testing users service..."
	@curl -s http://localhost:8080/api/users/users | jq . || echo "❌ Users test failed"
	@echo "Testing orders service..."
	@curl -s http://localhost:8080/api/orders/orders | jq . || echo "❌ Orders test failed"

# Test users service endpoints
test-users:
	@echo "🧪 Testing users service endpoints..."
	@echo "GET /api/users/users:"
	@curl -s http://localhost:8080/api/users/users | jq . || echo "❌ Failed"
	@echo "GET /api/users/profile:"
	@curl -s http://localhost:8080/api/users/profile | jq . || echo "❌ Failed"

# Test orders service endpoints
test-orders:
	@echo "🧪 Testing orders service endpoints..."
	@echo "GET /api/orders/orders:"
	@curl -s http://localhost:8080/api/orders/orders | jq . || echo "❌ Failed"
	@echo "GET /api/orders/status:"
	@curl -s http://localhost:8080/api/orders/status | jq . || echo "❌ Failed"

# Run all tests
test-all: test test-users test-orders
	@echo "✅ All tests completed"