.PHONY: test-integration test-integration-docker test-integration-containers help

# Поднимаем тестовую БД через docker-compose
up-test-db:
	docker-compose up -d postgres-test
	@echo "Waiting for database to be ready..."
	@sleep 5

# Останавливаем тестовую БД
down-test-db:
	docker-compose down -v

# Запускаем интеграционные тесты с docker-compose
test-integration-docker: up-test-db
	@echo "Running integration tests with docker-compose..."
	TEST_DATABASE_URL="postgres://test:test@localhost:54321/notes_test?sslmode=disable" \
	go test -v -tags=integration ./integration/... -count=1
	$(MAKE) down-test-db

# Запускаем интеграционные тесты с testcontainers
test-integration-containers:
	@echo "Running integration tests with testcontainers..."
	go test -v -tags=integration ./integration/... -count=1

# Запускаем все интеграционные тесты
test-integration: test-integration-docker test-integration-containers

# Показываем справку
help:
	@echo "Available commands:"
	@echo "  make up-test-db              - Start test database"
	@echo "  make down-test-db            - Stop test database"
	@echo "  make test-integration-docker - Run integration tests with docker-compose"
	@echo "  make test-integration-containers - Run integration tests with testcontainers"
	@echo "  make test-integration        - Run all integration tests"

