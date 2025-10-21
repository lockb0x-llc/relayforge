.PHONY: help build clean test run dev stop install docker-build docker-up docker-down

# Default target
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Targets:'
	@egrep '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-20s %s\n", $$1, $$2}'

# Development
install: ## Install dependencies
	@echo "Installing Go dependencies..."
	go mod tidy
	@echo "Installing frontend dependencies..."
	cd web && npm install

dev: ## Start development servers
	@echo "Starting development environment..."
	docker-compose -f docker-compose.yml up postgres redis -d
	@echo "Waiting for database to be ready..."
	sleep 5
	@echo "Starting API server..."
	DB_HOST=localhost go run cmd/api/main.go &
	@echo "Starting runner..."
	go run cmd/runner/main.go &
	@echo "Starting web UI..."
	cd web && npm run dev &
	@echo "Development servers started!"
	@echo "API: http://localhost:8080"
	@echo "Web UI: http://localhost:3000"

run: docker-up ## Start all services using Docker Compose

stop: ## Stop all development processes
	@echo "Stopping development servers..."
	pkill -f "go run"
	pkill -f "npm run dev"
	docker-compose down

# Build
build: ## Build all binaries
	@echo "Building API server..."
	go build -o bin/api ./cmd/api
	@echo "Building CLI..."
	go build -o bin/relayforge ./cmd/cli
	@echo "Building runner..."
	go build -o bin/runner ./cmd/runner
	@echo "Building web UI..."
	cd web && npm run build

build-cli: ## Build CLI binary only
	go build -o bin/relayforge ./cmd/cli

# Testing
test: ## Run tests
	@echo "Running Go tests..."
	go test -v ./...
	@echo "Running frontend tests..."
	cd web && npm test

test-coverage: ## Run tests with coverage
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: coverage.html"

# Docker
docker-build: ## Build Docker images
	@echo "Building Docker images..."
	docker-compose build

docker-up: ## Start services with Docker Compose
	@echo "Starting RelayForge with Docker Compose..."
	docker-compose up -d
	@echo "Services started!"
	@echo "API: http://localhost:8080"
	@echo "Web UI: http://localhost:3000"
	@echo "Database: localhost:5432"

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-logs: ## Show Docker Compose logs
	docker-compose logs -f

# Database
db-migrate: ## Run database migrations
	@echo "Running database migrations..."
	migrate -path migrations -database "postgres://relayforge:password@localhost:5432/relayforge?sslmode=disable" up

db-reset: ## Reset database
	@echo "Resetting database..."
	migrate -path migrations -database "postgres://relayforge:password@localhost:5432/relayforge?sslmode=disable" down -all
	migrate -path migrations -database "postgres://relayforge:password@localhost:5432/relayforge?sslmode=disable" up

# Cleanup
clean: ## Clean build artifacts
	@echo "Cleaning build artifacts..."
	rm -rf bin/
	rm -rf web/.next
	rm -rf web/out
	rm -f coverage.out coverage.html

# Linting
lint: ## Run linters
	@echo "Running Go linter..."
	golangci-lint run
	@echo "Running frontend linter..."
	cd web && npm run lint

fmt: ## Format code
	@echo "Formatting Go code..."
	go fmt ./...
	@echo "Formatting frontend code..."
	cd web && npm run lint:fix

# Example workflows
example-workflow: ## Create example multi-cloud workflow
	@echo "Creating example workflow..."
	mkdir -p examples
	@echo 'name: Multi-Cloud VM Deployment\ndescription: Deploy VMs across AWS and GCP\n\njobs:\n  provision-aws:\n    runs-on: aws-runner\n    steps:\n      - name: Create AWS EC2 instance\n        run: |\n          aws ec2 run-instances \\\n            --image-id ami-0abcdef1234567890 \\\n            --instance-type t3.micro \\\n            --key-name my-key \\\n            --tag-specifications "ResourceType=instance,Tags=[{Key=Name,Value=RelayForge-VM}]"\n      - name: Wait for instance\n        run: |\n          aws ec2 wait instance-running --instance-ids $$INSTANCE_ID\n\n  provision-gcp:\n    runs-on: gcp-runner\n    steps:\n      - name: Create GCP VM instance\n        run: |\n          gcloud compute instances create relayforge-vm \\\n            --zone=us-central1-a \\\n            --machine-type=e2-micro \\\n            --image-family=ubuntu-2004-lts \\\n            --image-project=ubuntu-os-cloud\n      - name: Configure firewall\n        run: |\n          gcloud compute firewall-rules create allow-http \\\n            --allow tcp:80,tcp:443 \\\n            --source-ranges 0.0.0.0/0' > examples/multi-cloud-vm.yml
	@echo "Example workflow created: examples/multi-cloud-vm.yml"

# Release
release: build ## Build release binaries
	@echo "Building release binaries..."
	GOOS=linux GOARCH=amd64 go build -o bin/relayforge-linux-amd64 ./cmd/cli
	GOOS=darwin GOARCH=amd64 go build -o bin/relayforge-darwin-amd64 ./cmd/cli
	GOOS=windows GOARCH=amd64 go build -o bin/relayforge-windows-amd64.exe ./cmd/cli
	@echo "Release binaries built in bin/"