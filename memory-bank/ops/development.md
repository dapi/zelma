---
title: Development Environment
doc_kind: ops
doc_function: canonical
purpose: Локальная разработка zelma: текущий bootstrap status, зависимости и команды проверки CLI/runtime.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Development Environment

Runtime-код `zelma` создан, поэтому этот документ фиксирует текущие проверки
документации, Go CLI и Docker/zellij e2e.

## Setup

Минимальная подготовка для текущего состояния репозитория:

```bash
mise install
mise exec -- go version
python3 --version
docker --version
zellij --version
codex --version
```

Stack: Go. Docker нужен для local e2e образа, `zellij` запускается внутри этого
образа, а Codex заменяется deterministic fake runtime. Go toolchain фиксируется
через [../../.mise.toml](../../.mise.toml); если shell не активирует mise shims,
запускай Go-команды через `mise exec -- <command>`.

Local probe on `2026-07-07`:

- `zellij --version` returned `zellij 0.44.0`.
- `go version` failed because `go` is not currently in shell `PATH`.
- `mise exec -- go version` uses project Go from `.mise.toml`.
- `docker --version` returned `Docker version 29.4.1`.

## Daily Commands

Canonical проверки:

```bash
python3 scripts/check_memory_bank_index.py
git diff --check
rg -n "zelima|Zelima" .
```

```bash
mise exec -- go test ./...
mise exec -- go vet ./...
mise exec -- go test ./... -race
mise exec -- go build ./cmd/zelma
make test-e2e
```

## Docker Zellij E2E Checks

Local e2e mirrors the intended CI shape and runs real `zellij` inside Docker
while using a deterministic fake Codex runtime. Do not run these checks against
the developer's ambient `zellij` session or real Codex home.

Target:

```bash
make test-e2e
```

Equivalent expanded shape:

```bash
CGO_ENABLED=0 GOOS=linux GOARCH=<docker-arch> go build -o ./bin/zelma-e2e-linux-<docker-arch> ./cmd/zelma
docker build --build-arg TARGETARCH=<docker-arch> -f Dockerfile.e2e -t zelma-e2e .
docker run --rm \
  -v "$(pwd)/bin/zelma-e2e-linux-<docker-arch>:/test/zelma:ro" \
  -v "$(pwd)/scripts/e2e:/test/scripts:ro" \
  -v "$(pwd)/testdata/e2e:/test/testdata:ro" \
  zelma-e2e \
  /test/scripts/docker-runner.sh
```

Runner contract:

- create isolated `HOME`, `CODEX_HOME` and `.zelma` paths inside the container;
- put the mounted `zelma` binary and fake `codex` wrapper on `PATH`;
- start a named `zellij` session through `script`;
- wait for readiness with `zellij list-sessions`;
- run focused assertions for `instances create`, `instances detect`,
  `instances list --live` and `--json` outputs;
- always kill the test `zellij` session before exiting.

## Browser Testing

У проекта нет browser UI. Не запускай dev server и не добавляй browser
verification без отдельной feature или продукта, который вводит UI.

## Database And Services

Внешние runtime dependencies CLI:

- Go toolchain для сборки и тестирования;
- Docker для e2e checks in CI-compatible environment;
- `zellij` для создания и обнаружения panes;
- Codex CLI/runtime для запуска и идентификации Codex-сессий;
- локальная файловая система для `.zelma/instances.json`.

База данных, background service или daemon не входят в текущий scope.

## Adoption Checklist

- [x] указаны текущие проверки документации
- [x] зафиксировано отсутствие browser UI
- [x] перечислены будущие runtime dependencies
- [x] выбран Go stack
- [x] добавлен Docker e2e workflow для `zellij`
- [x] Go toolchain pinned through `.mise.toml`
- [x] после создания scaffold указаны реальные setup/test/lint commands
- [x] после реализации CLI добавлены integration checks для `zellij` и Codex
