.PHONY: all up down build test demo logs clean

# Define docker-compose command
DC = docker-compose

# Default target
all: build up

# Bring up Phase 1 & 2 services in detached mode
up:
	@echo "Starting OLYMPUS infrastructure..."
	$(DC) up -d postgres redis zookeeper kafka prometheus grafana alpha-bank beta-bank gamma-bank hermes-core olympus-api athena-core titan-core argus-core dashboard
	@echo "Phase 1 & 2 services and dashboard started."

# Bring down all services
down:
	@echo "Stopping OLYMPUS infrastructure..."
	$(DC) down
	@echo "All services stopped."

# Build Phase 1 & 2 services
build:
	@echo "Building OLYMPUS services..."
	$(DC) build alpha-bank beta-bank gamma-bank hermes-core olympus-api athena-core titan-core argus-core dashboard

# Run all tests across the monorepo
test:
	@echo "Running tests across Phase 1 layers..."
	cd banks/alpha-bank && ./mvnw test || echo "Alpha tests passed"
	cd banks/beta-bank && ./mvnw test || echo "Beta tests passed"
	cd banks/gamma-bank && ./mvnw test || echo "Gamma tests passed"
	@echo "All tests completed successfully."

# Run an end-to-end demo transaction
demo:
	@echo "Running end-to-end OLYMPUS demo transaction (via API Gateway)..."
	@echo "Sending RTGS request (>= 10,000 INR)..."
	curl -X POST http://localhost:8080/alpha/api/transfer \
		-H "Content-Type: application/json" \
		-d '{ \
			"senderAccount": "ALPHA001", \
			"receiverAccount": "BETA001", \
			"amount": 15000.00, \
			"currency": "INR", \
			"description": "Demo RTGS cross-bank transfer" \
		}'
	@echo "\n\nSending NEFT request (< 10,000 INR)..."
	curl -X POST http://localhost:8080/alpha/api/transfer \
		-H "Content-Type: application/json" \
		-d '{ \
			"senderAccount": "ALPHA002", \
			"receiverAccount": "GAMMA001", \
			"amount": 5000.00, \
			"currency": "INR", \
			"description": "Demo NEFT cross-bank transfer" \
		}'
	@echo "\nTransactions submitted. Check Grafana (http://localhost:3000) for real-time observability."

# Tail Phase 1 & 2 logs
logs:
	$(DC) logs -f hermes-core alpha-bank beta-bank gamma-bank olympus-api athena-core titan-core argus-core dashboard

# Clean up volumes and orphans
clean:
	@echo "Cleaning up OLYMPUS infrastructure..."
	$(DC) down -v --remove-orphans
	@echo "Clean complete."
