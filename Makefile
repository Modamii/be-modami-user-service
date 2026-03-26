.PHONY: build run test migrate-up migrate-down proto docker-up docker-down lint tidy

# Build only our service packages (exclude pre-existing pkg/kafka from other projects)
PKGS := $(shell go list ./... | grep -v 'pkg/kafka')

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
