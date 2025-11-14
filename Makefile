.PHONY: build run up down test clean

APP_NAME := server
DOCKER_IMAGE_NAME := avito-backend-trainee

# Build the Go application binary
build:
	@echo "Building Go application..."
	CGO_ENABLED=0 GOOS=linux go build -o $(APP_NAME) ./main.go

# Run the Go application locally (without Docker)
run: build
	@echo "Running Go application locally..."
	./$(APP_NAME)

# Build and run Docker containers using docker-compose
up:
	@echo "Starting Docker containers..."
	docker-compose up --build -d

# Stop and remove Docker containers
down:
	@echo "Stopping and removing Docker containers..."
	docker-compose down

# Run tests (placeholder for now)
test:
	@echo "Running tests..."
	go test ./...

# Clean up build artifacts
clean:
	@echo "Cleaning up build artifacts..."
	rm -f $(APP_NAME)
	docker rmi $(DOCKER_IMAGE_NAME)_app || true
