E2E_IMAGE ?= zelma-e2e
E2E_TARGETARCH ?= $(shell docker version --format '{{.Server.Arch}}' 2>/dev/null | sed 's/aarch64/arm64/; s/x86_64/amd64/')
E2E_GOARCH ?= $(E2E_TARGETARCH)
E2E_ZELMA_BIN ?= $(CURDIR)/bin/zelma-e2e-linux-$(E2E_GOARCH)
E2E_FAKE_CODEX_BIN ?= $(CURDIR)/bin/fake-codex-e2e-linux-$(E2E_GOARCH)

.DEFAULT_GOAL := all

.PHONY: all build test test-e2e docker-zellij-e2e e2e-image clean-e2e

all: build test

build:
	go build -o ./zelma ./cmd/zelma

test:
	go test ./...

test-e2e: docker-zellij-e2e

docker-zellij-e2e: $(E2E_ZELMA_BIN) $(E2E_FAKE_CODEX_BIN) e2e-image
	docker run --rm \
	  -v "$(E2E_ZELMA_BIN):/test/zelma:ro" \
	  -v "$(E2E_FAKE_CODEX_BIN):/test/codex:ro" \
	  -v "$(CURDIR)/scripts/e2e:/test/scripts:ro" \
	  -v "$(CURDIR)/testdata/e2e:/test/testdata:ro" \
	  $(E2E_IMAGE) \
	  /test/scripts/docker-runner.sh

e2e-image:
	docker build \
	  --build-arg TARGETARCH="$(E2E_TARGETARCH)" \
	  -f Dockerfile.e2e \
	  -t $(E2E_IMAGE) \
	  .

$(E2E_ZELMA_BIN):
	mkdir -p "$(dir $@)"
	CGO_ENABLED=0 GOOS=linux GOARCH="$(E2E_GOARCH)" go build -o "$@" ./cmd/zelma

$(E2E_FAKE_CODEX_BIN): testdata/e2e/fake-codex.go
	mkdir -p "$(dir $@)"
	CGO_ENABLED=0 GOOS=linux GOARCH="$(E2E_GOARCH)" go build -o "$@" ./testdata/e2e/fake-codex.go

clean-e2e:
	rm -f "$(E2E_ZELMA_BIN)" "$(E2E_FAKE_CODEX_BIN)"
