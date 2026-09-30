<<<<<<< HEAD
.PHONY: test demo match-play match-verify client-snapshot janus-roma-test compose-up compose-down compose-test edge-infer proto proto-check build-services loadpredict aigc-worker aigc-stub-check
=======
.PHONY: test demo match-play match-verify janus-roma-test compose-e2e-test compose-up compose-down compose-test edge-infer proto proto-check build-services loadpredict aigc-worker aigc-stub-check
>>>>>>> a99d35f (feat(phase2): deepen Janus/Roma placeholders after PR #10)

COMPOSE ?= docker compose --profile dev
PROTOC ?= protoc
PROTO_GEN_GO ?= $(shell go env GOPATH)/bin/protoc-gen-go
PROTO_GEN_GO_GRPC ?= $(shell go env GOPATH)/bin/protoc-gen-go-grpc
PROTO_FILES := proto/common/v1/common.proto \
	proto/gateway/v1/gateway.proto \
	proto/lares/v1/lares.proto \
	proto/roma/v1/roma.proto \
	proto/senate/v1/senate.proto \
	proto/chat/v1/chat.proto

test:
	go test ./...

demo:
	go run ./cmd/demo

match-play:
	go run ./cmd/match play -seed 0xcafe

match-verify:
	@tmp=$$(mktemp /tmp/match-XXXX.rgz); \
	go run ./cmd/match play -seed 0xcafe -out $$tmp && \
	go run ./cmd/match verify -in $$tmp; \
	rm -f $$tmp

client-snapshot:
	go run ./cmd/client-snapshot

janus-roma-test:
	go test ./services/janus -run TestJanusToRomaTacticalLockstepPath -v

compose-e2e-test:
	@test -f .env || cp .env.example .env
	COMPOSE_E2E=1 go test ./pkg/integration -run TestComposeJanusRomaTacticalSequence -v -count=1

edge-infer:
	go run ./services/edge-infer

proto:
	@command -v $(PROTOC) >/dev/null || (echo "install protoc to regenerate; checked-in gen/go is CI default" && exit 0)
	@test -x $(PROTO_GEN_GO) || go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.34.2
	@test -x $(PROTO_GEN_GO_GRPC) || go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.5.1
	$(PROTOC) -I proto \
		--go_out=./gen/go --go_opt=module=github.com/iiooiioo888/qianjunpo-reborn/gen/go \
		--go-grpc_out=./gen/go --go-grpc_opt=module=github.com/iiooiioo888/qianjunpo-reborn/gen/go \
		$(PROTO_FILES)

proto-check: proto
	@git diff --exit-code gen/go || (echo "gen/go out of date; commit or run make proto" && exit 1)

build-services:
	CGO_ENABLED=0 go build -o /tmp/janus ./services/janus
	CGO_ENABLED=0 go build -o /tmp/roma ./services/roma
	CGO_ENABLED=0 go build -o /tmp/lares ./services/lares
	CGO_ENABLED=0 go build -o /tmp/senate ./services/senate
	CGO_ENABLED=0 go build -o /tmp/chatserver ./services/chatserver
	CGO_ENABLED=0 go build -o /tmp/loadpredict ./services/loadpredict
	CGO_ENABLED=0 go build -o /tmp/aigc-worker ./services/aigc-worker

loadpredict:
	go run ./services/loadpredict

aigc-worker:
	go run ./services/aigc-worker

aigc-stub-check:
	@curl -sf http://127.0.0.1:8096/health >/dev/null 2>&1 || (echo "start with: make aigc-worker" && exit 0)

compose-up:
	@test -f .env || cp .env.example .env
	$(COMPOSE) up -d redis mysql etcd lares roma janus

compose-up-ops:
	@test -f .env || cp .env.example .env
	docker compose --profile dev --profile ops up -d senate chatserver

compose-down:
	$(COMPOSE) down

compose-test:
	@test -f .env || cp .env.example .env
	$(COMPOSE) build app
	$(COMPOSE) run --rm app make test
