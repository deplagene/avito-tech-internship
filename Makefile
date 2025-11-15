.PHONY: build run test docker-up docker-down migrate-up migrate-down lint clean

APP_NAME=server
BUILD_DIR=./bin
CMD_DIR=./cmd

# Default values
HTTP_PORT ?= 8080
DATABASE_URL ?= postgres://user:password@localhost:5432/avito_trainee?sslmode=disable
TEST_DATABASE_URL ?= postgres://user:password@localhost:5432/avito_trainee_test?sslmode=disable
ENV ?= development

MODULE_PATH=deplagene/avito-tech-internship

# Собрать приложение
build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd

# Запустить локально
run: build
	@echo "Running $(APP_NAME)..."
	@HTTP_PORT=$(HTTP_PORT) DATABASE_URL=$(DATABASE_URL) ENV=$(ENV) $(BUILD_DIR)/$(APP_NAME)

# Запустить юнит-тесты
test:
	@echo "Running unit tests..."
	@go test -v $(shell go list ./... | grep -v /test)

# Запуск Докер-композа
docker-up:
	@echo "Starting Docker Compose environment..."
	@docker-compose up --build -d

# Остановка Докер-композа
docker-down:
	@echo "Stopping Docker Compose environment..."
	@docker-compose down -v

# Применить миграции
migrate-up:
	@echo "Applying database migrations..."
	@docker-compose run --rm app migrate -path /migrations -database "$(DATABASE_URL)" up

# Откатить миграции
migrate-down:
	@echo "Rolling back database migrations..."
	@docker-compose run --rm app migrate -path /migrations -database "$(DATABASE_URL)" down

# Запустить линтер
lint:
	@echo "Running golangci-lint..."
	@/home/deplagene/go/bin/golangci-lint run ./...

# Очистка
clean:
	@echo "Cleaning up..."
	@rm -rf $(BUILD_DIR)
	@go clean