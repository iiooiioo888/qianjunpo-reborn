.PHONY: test demo match-play match-verify client-snapshot janus-roma-test janus-discovery-test compose-e2e-test timedilation-drill compose-up compose-down compose-test edge-infer test-edge-infer test-ai-rag test-agones-room agones-room-demo validate-agones-manifests check-agones proto proto-check build-services loadpredict aigc-worker aigc-stub-check observability-check promtool-check-rules check-observability

OBS_ALERT_RULES := deploy/observability/prometheus/alerts/qjp-production.rules.yml
PROMTOOL_IMAGE ?= prom/prometheus:v2.54.1

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

janus-discovery-test:
	go test ./internal/janus -run 'TestNormalizeZoneID|TestDiscovery|TestParseZoneEndpointMap' -count=1

timedilation-drill:
	go test ./pkg/timedilation -run 'TestOverloadDrill|TestDegradeRecovers' -count=1
	go run ./cmd/timedilation-drill

compose-e2e-test:
	@test -f .env || cp .env.example .env
	COMPOSE_E2E=1 go test ./pkg/integration -run TestComposeJanusRomaTacticalSequence -v -count=1

edge-infer:
	go run ./services/edge-infer

test-edge-infer:
	go test ./services/edge-infer/... -v -count=1

test-agones-room:
	go test ./pkg/agones/... -count=1

validate-agones-manifests:
	go test ./pkg/agones/ -run TestAgonesDeploy -count=1

check-agones: validate-agones-manifests test-agones-room

agones-room-demo:
	@ROMA_AGONES_BACKEND=mock ROMA_AGONES_ALLOCATOR=mock \
		go run ./cmd/agones-room allocate -zone demo -shard 0
	@ROMA_AGONES_BACKEND=mock ROMA_AGONES_ALLOCATOR=mock \
		go run ./cmd/agones-room ready
	@ROMA_AGONES_BACKEND=mock ROMA_AGONES_ALLOCATOR=mock \
		go run ./cmd/agones-room shutdown

test-ai-rag:
	go test ./pkg/ai ./pkg/rag -v -count=1

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

observability-check:
	go test ./pkg/observability/... -count=1

promtool-check-rules:
	@test -f $(OBS_ALERT_RULES)
	@if command -v promtool >/dev/null 2>&1; then \
		promtool check rules $(OBS_ALERT_RULES); \
	elif command -v docker >/dev/null 2>&1 && docker info >/dev/null 2>&1; then \
		docker run --rm --entrypoint promtool -v "$(CURDIR):/work:ro" $(PROMTOOL_IMAGE) \
			check rules /work/$(OBS_ALERT_RULES); \
	else \
		echo "promtool/docker not found; skipped (Go checks in observability-check still run)"; \
	fi

check-observability: observability-check promtool-check-rules
