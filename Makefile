.PHONY: build run test migrate-up migrate-down migrate-down-one migrate-force migrate-version migrate-create swagger proto docker-up docker-down lint tidy

# Build only our service packages (exclude pre-existing pkg/kafka from other projects)
PKGS := $(shell go list ./... | grep -v 'pkg/kafka')

# Migration config (override via env vars or DATABASE_URL)
DB_HOST     ?= localhost
DB_PORT     ?= 5432
DB_USER     ?= postgres
DB_PASSWORD ?= postgres_password
DB_NAME     ?= be-modami-user-service
DB_SCHEMA   ?= public
DB_SSLMODE  ?= disable
DATABASE_URL ?= postgres://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)&search_path=$(DB_SCHEMA)

build:
	go build -o bin/server ./cmd/server

run:
	go run ./cmd/server

test:
	go test $(PKGS)

migrate-up:
	migrate -path migrations -database "$(DATABASE_URL)" up

migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

migrate-down-one:
	migrate -path migrations -database "$(DATABASE_URL)" down 1

migrate-force:
	migrate -path migrations -database "$(DATABASE_URL)" force $(VERSION)

migrate-version:
	migrate -path migrations -database "$(DATABASE_URL)" version

migrate-create:
	migrate create -ext sql -dir migrations -seq $(NAME)

swagger:
	swag init -g cmd/server/main.go -o docs --parseDependency --parseInternal

proto:
	protoc --go_out=. --go-grpc_out=. proto/user/user.proto

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

lint:
	golangci-lint run

tidy:
	go mod tidy
