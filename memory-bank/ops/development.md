---
title: Development Environment
doc_kind: ops
doc_function: canonical
purpose: Локальная разработка zelma: текущий bootstrap status, ожидаемые зависимости и команды проверки документации до появления runtime-кода.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Development Environment

Runtime-код `zelma` пока не создан, поэтому этот документ фиксирует текущие
проверки документации и ожидаемые внешние зависимости для будущего CLI.

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

Stack: Go. Docker, `zellij` и Codex нужны для будущих runtime/integration
checks. Для текущего документационного слоя достаточно `python3`. Go toolchain
фиксируется через [../../.mise.toml](../../.mise.toml); если shell не активирует
mise shims, запускай Go-команды через `mise exec -- <command>`.

Local probe on `2026-07-07`:

- `zellij --version` returned `zellij 0.44.0`.
- `go version` failed because `go` is not currently in shell `PATH`.
- `mise exec -- go version` uses project Go from `.mise.toml`.
- `docker --version` returned `Docker version 29.4.1`.

## Daily Commands

Canonical проверки на текущем этапе:

```bash
python3 scripts/check_memory_bank_index.py
git diff --check
rg -n "zelima|Zelima" .
```

Canonical команды после появления Go scaffold:

```bash
mise exec -- go test ./...
mise exec -- go vet ./...
mise exec -- go test ./... -race
mise exec -- go build ./cmd/zelma
```

## Docker Zellij E2E Checks

Future local e2e should mirror CI and run real `zellij` inside Docker while
using a deterministic fake Codex runtime. Do not run these checks against the
developer's ambient `zellij` session or real Codex home.

Expected target shape after CLI integration exists:

```bash
go build -o ./tmp/zelma ./cmd/zelma
docker build -f Dockerfile.e2e -t zelma-e2e .
docker run --rm \
  -v "$(pwd)/tmp/zelma:/test/zelma:ro" \
  -v "$(pwd)/scripts/e2e:/test/scripts:ro" \
  -v "$(pwd)/testdata/e2e:/test/testdata:ro" \
  zelma-e2e \
  /test/scripts/docker-runner.sh
```

Expected future convenience command:

```bash
make test-e2e
```

Runner contract:

- create isolated `HOME`, `CODEX_HOME` and `.zelma` paths inside the container;
- put the mounted `zelma` binary and fake `codex` wrapper on `PATH`;
- start a named `zellij` session through `script`;
- wait for readiness with `zellij list-sessions`;
- run focused assertions for `sessions create`, `sessions detect`,
  `sessions list` and `--json` outputs when those commands exist;
- always kill the test `zellij` session before exiting.

## Browser Testing

У проекта нет browser UI. Не запускай dev server и не добавляй browser
verification без отдельной feature или продукта, который вводит UI.

## Database And Services

Внешние runtime dependencies будущего CLI:

- Go toolchain для сборки и тестирования;
- Docker для future e2e checks in CI-compatible environment;
- `zellij` для создания и обнаружения panes;
- Codex CLI/runtime для запуска и идентификации Codex-сессий;
- локальная файловая система для `.zelma/sessions.json`.

База данных, background service или daemon не входят в текущий scope.

## Adoption Checklist

- [x] указаны текущие проверки документации
- [x] зафиксировано отсутствие browser UI
- [x] перечислены будущие runtime dependencies
- [x] выбран Go stack
- [x] описан будущий Docker e2e workflow для `zellij`
- [x] Go toolchain pinned through `.mise.toml`
- [ ] после создания scaffold указаны реальные setup/test/lint commands
- [ ] после реализации CLI добавлены integration checks для `zellij` и Codex
