GO ?= go
DOCKER_COMPOSE ?= docker compose

.PHONY: fmt test vet build build-market-api build-market-rpc build-exchange-api build-exchange-rpc build-ucenter-api build-ucenter-rpc build-jobcenter tidy docker-up docker-down docker-logs

fmt:
	$(GO) fmt ./...

test:
	$(GO) test ./...

vet:
	$(GO) vet ./...

build: build-market-api build-market-rpc build-exchange-api build-exchange-rpc build-ucenter-api build-ucenter-rpc build-jobcenter

build-market-api:
	$(GO) build ./app/market/api

build-market-rpc:
	$(GO) build ./app/market/rpc

build-exchange-api:
	$(GO) build ./app/exchange/api

build-exchange-rpc:
	$(GO) build ./app/exchange/rpc

build-ucenter-api:
	$(GO) build ./app/ucenter/api

build-ucenter-rpc:
	$(GO) build ./app/ucenter/rpc

build-jobcenter:
	$(GO) build ./app/jobcenter

tidy:
	$(GO) mod tidy

docker-up:
	$(DOCKER_COMPOSE) up -d --build

docker-down:
	$(DOCKER_COMPOSE) down -v

docker-logs:
	$(DOCKER_COMPOSE) logs -f --tail=200
