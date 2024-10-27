.PHONY: build run test clean docker-build docker-up docker-down

GO=go
GOTEST=$(GO) test
GOVET=$(GO) vet
GOBUILD=$(GO) build

DOCKER=docker
DOCKER_COMPOSE=docker-compose

APP_NAME=spacetrouble
BINARY_NAME=spacetrouble

BUILD_FLAGS=-v

all: test build

build:
	$(GOBUILD) $(BUILD_FLAGS) -o $(BINARY_NAME) ./cmd/api/main.go

test:
	$(GOTEST) -v ./...

vet:
	$(GOVET) ./...

clean:
	rm -f $(BINARY_NAME)
	rm -rf dist/

docker-build:
	$(DOCKER) build -t $(APP_NAME) .

docker-up:
	$(DOCKER_COMPOSE) up -d

docker-down:
	$(DOCKER_COMPOSE) down

migrate-up:
	docker-compose exec app migrate -path ./migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" up

migrate-down:
	docker-compose exec app migrate -path ./migrations -database "postgres://${POSTGRES_USER}:${POSTGRES_PASSWORD}@${POSTGRES_HOST}:${POSTGRES_PORT}/${POSTGRES_DB}?sslmode=disable" down