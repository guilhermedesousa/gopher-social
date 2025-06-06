include .env
MIGRATIONS_PATH = ./cmd/migrate/migrations

.PHONY: migrate-create
migration:
	@migrate create -ext sql -dir $(MIGRATIONS_PATH) $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-up
migrate-up:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_ADDR) up

.PHONY: migrate-down
migrate-down:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_ADDR) down $(filter-out $@,$(MAKECMDGOALS))

.PHONY: migrate-force-version
migrate-force-version:
	@migrate -path $(MIGRATIONS_PATH) -database $(DB_ADDR) force $(filter-out $@,$(MAKECMDGOALS))

.PHONY: seed
seed:
	@go run ./cmd/cli/seed/main.go

.PHONY: gen-docs
gen-docs:
	@swag init -g ./api/main.go -d cmd,internal && swag fmt

.PHONY: test
test:
	@go test -v ./...