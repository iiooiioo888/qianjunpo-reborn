.PHONY: test demo compose-up compose-down compose-test

COMPOSE ?= docker compose --profile dev

test:
	go test ./...

demo:
	go run ./cmd/demo

compose-up:
	@test -f .env || cp .env.example .env
	$(COMPOSE) up -d redis mysql etcd

compose-down:
	$(COMPOSE) down

compose-test:
	@test -f .env || cp .env.example .env
	$(COMPOSE) build app
	$(COMPOSE) run --rm app make test
