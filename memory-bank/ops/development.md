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
go version
python3 --version
zellij --version
codex --version
```

Stack: Go. `zellij` и Codex нужны для будущих runtime/integration checks. Для
текущего документационного слоя достаточно `python3`.

Local probe on `2026-07-07`:

- `zellij --version` returned `zellij 0.44.0`.
- `go version` failed because `go` is not currently in `PATH`.

## Daily Commands

Canonical проверки на текущем этапе:

```bash
python3 scripts/check_memory_bank_index.py
git diff --check
rg -n "zelima|Zelima" .
```

Canonical команды после появления Go scaffold:

```bash
go test ./...
go vet ./...
go test ./... -race
```

## Browser Testing

У проекта нет browser UI. Не запускай dev server и не добавляй browser
verification без отдельной feature или продукта, который вводит UI.

## Database And Services

Внешние runtime dependencies будущего CLI:

- Go toolchain для сборки и тестирования;
- `zellij` для создания и обнаружения panes;
- Codex CLI/runtime для запуска и идентификации Codex-сессий;
- локальная файловая система для `.zelma/sessions.json`.

База данных, background service или daemon не входят в текущий scope.

## Adoption Checklist

- [x] указаны текущие проверки документации
- [x] зафиксировано отсутствие browser UI
- [x] перечислены будущие runtime dependencies
- [x] выбран Go stack
- [ ] установлен Go toolchain в локальном окружении
- [ ] после создания scaffold указаны реальные setup/test/lint commands
- [ ] после реализации CLI добавлены integration checks для `zellij` и Codex
