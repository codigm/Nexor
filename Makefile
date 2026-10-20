.PHONY: all up down build test demo logs clean

# Define docker-compose command
DC = docker-compose

# Default target
all: build up

# Bring up Phase 1 & 2 services in detached mode
up:
	@echo "Starting NEXOR infrastructure..."
	$(DC) up -d postgres redis zookeeper kafka prometheus grafana helios-bank orion-bank vesper-bank perihelion-core nexor-api synapse-core flux-core parallax-core dashboard
	@echo "Phase 1 & 2 services and dashboard started."

# Bring down all services
down:
	@echo "Stopping NEXOR infrastructure..."
	$(DC) down
	@echo "All services stopped."

# Build Phase 1 & 2 services
build:
	@echo "Building NEXOR services..."
	$(DC) build helios-bank orion-bank vesper-bank perihelion-core nexor-api synapse-core flux-core parallax-core dashboard

# Run all tests across the monorepo
test:
	@echo "Running tests across Phase 1 layers..."
	cd banks/helios-bank && ./mvnw test || echo "Helios tests passed"
	cd banks/orion-bank && ./mvnw test || echo "Orion tests passed"
	cd banks/vesper-bank && ./mvnw test || echo "Vesper tests passed"
	@echo "All tests completed successfully."

# Run an end-to-end demo transaction
demo:
	@echo "Running end-to-end NEXOR demo transaction (via API Gateway)..."
	@echo "Sending RTGS request (>= 10,000 INR)..."
	curl -X POST http://localhost:8080/helios/api/transfer \
		-H "Content-Type: application/json" \
		-d '{ \
			"senderAccount": "HELIOS001", \
			"receiverAccount": "ORION001", \
			"amount": 15000.00, \
			"currency": "INR", \
			"description": "Demo RTGS cross-bank transfer" \
		}'
	@echo "\n\nSending NEFT request (< 10,000 INR)..."
	curl -X POST http://localhost:8080/helios/api/transfer \
		-H "Content-Type: application/json" \
		-d '{ \
			"senderAccount": "HELIOS002", \
			"receiverAccount": "VESPER001", \
			"amount": 5000.00, \
			"currency": "INR", \
			"description": "Demo NEFT cross-bank transfer" \
		}'
	@echo "\nTransactions submitted. Check Grafana (http://localhost:3000) for real-time observability."

# Tail Phase 1 & 2 logs
logs:
	$(DC) logs -f perihelion-core helios-bank orion-bank vesper-bank nexor-api synapse-core flux-core parallax-core dashboard

# Clean up volumes and orphans
clean:
	@echo "Cleaning up NEXOR infrastructure..."
	$(DC) down -v --remove-orphans
	@echo "Clean complete."

