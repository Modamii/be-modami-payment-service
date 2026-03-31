APP_NAME := be-payment-service
MAIN := ./cmd/server/main.go

.PHONY: help
help:
	@echo "Targets:"
	@echo "  make build        Build binary"
	@echo "  make run          Run server"
	@echo "  make swagger      Generate swagger docs (swaggo)"
	@echo "  make tidy         go mod tidy"
	@echo "  make migrate-up   Run migrations up (requires DB_DSN)"
	@echo "  make migrate-down Run migrations down (requires DB_DSN)"

.PHONY: build
build:
	go build -o bin/$(APP_NAME) $(MAIN)

.PHONY: run
run:
	go run $(MAIN)

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: swagger
swagger:
	@command -v swag >/dev/null 2>&1 || (echo "swag not found. Install: go install github.com/swaggo/swag/cmd/swag@latest" && exit 1)
	# Scan only directories that actually contain Go files.
	# Note: when using -d, -g is resolved relative to the first -d directory.
	swag init -g main.go -o ./docs/swagger -d ./cmd/server,./module/payment_gateway,./module/core,./module/payment_gateway_adapter,./pkg,./config

.PHONY: migrate-up
migrate-up:
	@test -n "$(DB_DSN)" || (echo "DB_DSN is required" && exit 1)
	migrate -path ./migrations -database "$(DB_DSN)" up

.PHONY: migrate-down
migrate-down:
	@test -n "$(DB_DSN)" || (echo "DB_DSN is required" && exit 1)
	migrate -path ./migrations -database "$(DB_DSN)" down 1

