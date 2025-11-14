LOCAL_BIN:=$(CURDIR)/bin
COMPOSE_FILE := docker-compose.yml
# Подключаем .env (не падаем, если файла нет)
-include .env

# Значение по умолчанию, если в .env не задан DB_DSN
DB_DSN ?= postgresql://user:password@127.0.0.1:5432/poker?sslmode=disable

.PHONY: .install-sqlc
.install-sqlc:
	$(info Installing sqlc...)
	GOBIN=$(LOCAL_BIN) go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

.PHONY: .sqlc-generate
.sqlc-generate:
	$(info sqlc-generate...)
	bin/sqlc generate


.PHONY: .install-goose
.install-goose:
	$(info Installing binary dependencies...)
	GOBIN=$(LOCAL_BIN) go install github.com/pressly/goose/v3/cmd/goose@v3.22.1
	
.PHONY: .apply-migrations
.apply-migrations:
	bin/goose -dir migrations postgres "$(DATABASE_URL)" up

.PHONY: .reset-migrations
.reset-migrations:
	bin/goose -dir migrations postgres "$(DATABASE_URL)" down-to 0
	bin/goose -dir migrations postgres "$(DATABASE_URL)" up                                                                               

MIGRATIONS_DIR=./migrations
.PHONY: create-migration
create-migration:
	@read -p "Введите имя миграции: " name; \
	bin/goose -dir $(MIGRATIONS_DIR) create $$name sql


run-all:
	@echo "Building and starting containers..."
	docker-compose -f $(COMPOSE_FILE) up --build -d