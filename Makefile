# Variables
APP_NAME=pr-review-service
DOCKER_COMPOSE=docker-compose
GO=go
GO_MOD=$(GO) mod
GO_BUILD=$(GO) build
GO_TEST=$(GO) test
GO_RUN=$(GO) run

build: 
	$(GO_BUILD) -o bin/$(APP_NAME) ./cmd/server

run: 
	$(GO_RUN) ./cmd/server

lint: 
	@which golangci-lint > /dev/null || (echo "Installing golangci-lint..." && go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest)
	golangci-lint run


docker-up: 
	$(DOCKER_COMPOSE) up -d

docker-down: 
	$(DOCKER_COMPOSE) down

docker-logs: 
	$(DOCKER_COMPOSE) logs -f

docker-logs-app:
	$(DOCKER_COMPOSE) logs -f app

docker-logs-db:
	$(DOCKER_COMPOSE) logs -f postgres

docker-restart:
	$(DOCKER_COMPOSE) restart

docker-rebuild:
	$(DOCKER_COMPOSE) down
	$(DOCKER_COMPOSE) up -d --build

deps:
	$(GO_MOD) download

tidy:
	$(GO_MOD) tidy

verify:
	$(GO_MOD) verify

status:
	$(DOCKER_COMPOSE) ps
