.PHONY: help backend-install backend-run backend-build backend-test backend-lint backend-swagger backend-migrate backend-migrate-down backend-migrate-status backend-generate-key frontend-install frontend-dev frontend-build frontend-test frontend-test-run frontend-tauri-dev frontend-tauri-build docker-up docker-down docker-logs docker-logs-backend docker-restart docker-clean dev dev-full build build-all test clean clean-all setup

SHELL := /bin/bash

help:
	@echo "Media Parser - доступные команды:"
	@echo ""
	@echo "Backend:"
	@echo "  make backend-install       - Установить зависимости бэкенда"
	@echo "  make backend-run           - Запустить бэкенд локально"
	@echo "  make backend-build         - Собрать бэкенд"
	@echo "  make backend-test          - Запустить тесты бэкенда"
	@echo "  make backend-lint          - Запустить линтер бэкенда"
	@echo "  make backend-swagger       - Сгенерировать Swagger документацию"
	@echo "  make backend-migrate       - Применить миграции"
	@echo "  make backend-migrate-down  - Откатить последнюю миграцию"
	@echo "  make backend-migrate-status - Показать статус миграций"
	@echo "  make backend-generate-key  - Сгенерировать API ключ и сохранить в БД"
	@echo ""
	@echo "Frontend:"
	@echo "  make frontend-install      - Установить зависимости фронтенда"
	@echo "  make frontend-dev          - Запустить фронтенд в режиме разработки"
	@echo "  make frontend-build        - Собрать фронтенд"
	@echo "  make frontend-test         - Запустить тесты фронтенда (watch)"
	@echo "  make frontend-test-run     - Запустить тесты фронтенда (один раз)"
	@echo "  make frontend-tauri-dev    - Запустить Tauri приложение (dev)"
	@echo "  make frontend-tauri-build  - Собрать Tauri приложение"
	@echo ""
	@echo "Docker:"
	@echo "  make docker-up             - Запустить все сервисы через Docker Compose"
	@echo "  make docker-down           - Остановить все сервисы"
	@echo "  make docker-logs           - Показать логи всех сервисов"
	@echo "  make docker-logs-backend   - Показать логи бэкенда"
	@echo "  make docker-restart        - Перезапустить все сервисы"
	@echo "  make docker-clean          - Очистить Docker ресурсы"
	@echo ""
	@echo "Development:"
	@echo "  make dev                   - Запустить бэкенд и фронтенд"
	@echo "  make dev-full              - Запустить Docker + бэкенд + фронтенд"
	@echo ""
	@echo "Build:"
	@echo "  make build                 - Собрать бэкенд и фронтенд"
	@echo "  make build-all             - Собрать всё (включая Tauri)"
	@echo ""
	@echo "Test:"
	@echo "  make test                  - Запустить все тесты"
	@echo ""
	@echo "Clean:"
	@echo "  make clean                 - Очистить временные файлы"
	@echo "  make clean-all             - Полная очистка"
	@echo ""
	@echo "Setup:"
	@echo "  make setup                 - Первичная настройка проекта"

backend-install:
	cd backend && go mod download

backend-run:
	cd backend && go run cmd/main.go

backend-build:
	cd backend && go build -o bin/server cmd/main.go

backend-test:
	cd backend && go test ./...

backend-lint:
	cd backend && golangci-lint run

backend-swagger:
	cd backend && swag init -g cmd/main.go -o docs/swagger

backend-migrate:
	goose -dir backend/migrations postgres "postgres://media_parser_user:media_parser_password@localhost:5432/media_parser?sslmode=disable" up

backend-migrate-down:
	goose -dir backend/migrations postgres "postgres://media_parser_user:media_parser_password@localhost:5432/media_parser?sslmode=disable" down

backend-migrate-status:
	goose -dir backend/migrations postgres "postgres://media_parser_user:media_parser_password@localhost:5432/media_parser?sslmode=disable" status

backend-generate-key:
	@echo "Генерация API ключа с полным доступом (бессрочный)..."
	@cd backend && go run cmd/keygen/main.go -name "default" || echo "Ошибка: убедитесь, что PostgreSQL запущен (make docker-up)"

backend-generate-key-readonly:
	@echo "Генерация API ключа только для чтения (media + requests)..."
	@cd backend && go run cmd/keygen/main.go -name "readonly" -parse=false || echo "Ошибка: убедитесь, что PostgreSQL запущен (make docker-up)"

backend-generate-key-temp:
	@echo "Генерация временного API ключа (24 часа)..."
	@cd backend && go run cmd/keygen/main.go -name "temp" -expires 24h || echo "Ошибка: убедитесь, что PostgreSQL запущен (make docker-up)"

frontend-install:
	cd client && npm install

frontend-dev:
	cd client && npm run dev

frontend-build:
	cd client && npm run build

frontend-test:
	cd client && npm run test

frontend-test-run:
	cd client && npm run test:run

frontend-tauri-dev:
	cd client && npm run tauri dev

frontend-tauri-build:
	cd client && npm run tauri build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f

docker-logs-backend:
	docker-compose logs -f backend

docker-restart:
	docker-compose restart

docker-clean:
	docker-compose down -v
	docker system prune -f

dev:
	@echo "Запуск бэкенда и фронтенда..."
	@cd backend && go run cmd/main.go &
	@cd client && npm run dev

dev-full:
	@echo "Запуск полной среды разработки..."
	docker-compose up -d postgres redis rabbitmq
	timeout /t 5 /nobreak 2>/dev/null || sleep 5
	@cd backend && go run cmd/main.go &
	@cd client && npm run tauri dev

build: backend-build frontend-build

build-all: backend-build frontend-build frontend-tauri-build

test: backend-test frontend-test-run

clean:
	-rm -rf backend/bin
	-rm -rf client/dist
	-rm -rf client/src-tauri/target/release/bundle

clean-all: clean docker-clean

setup:
	@if [ ! -f .env ]; then cp .env.template .env; fi
	@if [ ! -f backend/.env ]; then cp backend/.env.template backend/.env; fi
	make backend-install
	make frontend-install
	make docker-up
	timeout /t 5 /nobreak 2>/dev/null || sleep 5
	make backend-migrate
	@echo ""
	@echo "============================================"
	@echo "Проект настроен!"
	@echo "============================================"
	@echo "Бэкенд: http://localhost:8080"
	@echo "Swagger: http://localhost:8080/swagger/index.html"
	@echo "RabbitMQ UI: http://localhost:15672"
	@echo ""
	@echo "Для генерации API ключа выполните:"
	@echo "  make backend-generate-key"
	@echo "============================================"
