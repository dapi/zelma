---
title: Testing Policy
doc_kind: engineering
doc_function: canonical
purpose: Описывает testing policy репозитория: обязательность test case design, требования к automated regression coverage и допустимые manual-only gaps.
derived_from:
  - ../dna/governance.md
  - ../flows/feature-flow.md
status: active
canonical_for:
  - repository_testing_policy
  - feature_test_case_inventory_rules
  - automated_test_requirements
  - sufficient_test_coverage_definition
  - manual_only_verification_exceptions
  - simplify_review_discipline
  - verification_context_separation
must_not_define:
  - feature_acceptance_criteria
  - feature_scope
audience: humans_and_agents
---

# Testing Policy

## Project Adaptation

Runtime-код `zelma` создан. Stack выбран: Go. Текущая обязательная проверка
документационного слоя:

```bash
python3 scripts/check_memory_bank_index.py
```

Canonical local commands:

- `go test ./...`;
- `go vet ./...`;
- `go test ./... -race` для shared state, registry locking и adapter code;
- `go build ./cmd/zelma`;
- `make test-e2e` для focused Docker e2e against real `zellij`.

Primary test framework: Go `testing` package unless a future ADR chooses
otherwise.

## Required CI Suite Model

После появления Go scaffold CI должен быть разделен минимум на эти suites:

| Suite | Purpose | Required before |
| --- | --- | --- |
| `docs-memory-bank` | Проверить governed docs, links, indexing и typo guard | Any PR touching `memory-bank/` |
| `go-unit-contract` | Unit, fixture, schema, CLI help/output and adapter contract tests | Any runtime PR |
| `go-race` | Race check for shared state, locks and concurrent registry writes | PRs touching registry/concurrency |
| `docker-zellij-e2e` | End-to-end checks against real `zellij` in a container | PRs that claim `sessions create/detect/list` live integration is done |

`docker-zellij-e2e` is required for feature closure once a feature depends on
real `zellij` pane creation, live pane discovery or command execution. The local
target is `make test-e2e`.

## Docker Zellij E2E Target

`zelma` reuses the proven shape from `zellij-tab-status`: build the product
binary outside Docker, then mount it into a small test image that contains
pinned `zellij` and a runner script.

Implemented design:

- `Dockerfile.e2e` installs `zellij`, `util-linux` for `script`, `ca-certificates`
  and minimal test helpers. It must pin the `zellij` version and verify the
  downloaded binary checksum.
- The image must not build `zelma`; CI builds `zelma` once and passes the binary
  artifact to the e2e job.
- The runner starts a named `zellij` session through `script`, waits for readiness
  via `zellij list-sessions`, creates a minimal `config.kdl`, and always attempts
  to kill the test session before exit.
- E2E tests must run with isolated `HOME`, `CODEX_HOME` and registry paths. They
  must not read or mutate the developer's real `.codex`, `.config/zellij` or
  `.zelma/` state.
- CI e2e must use a deterministic fake `codex` executable or wrapper that writes
  synthetic `session_meta` JSONL records. It must not require a real Codex
  account, network access, user prompts or conversation transcripts.
- Synthetic Codex fixtures may contain UUID, cwd, timestamp and CLI version
  metadata only. Do not include real prompts, responses or user session logs.
- The runner exercises the smallest real integration path: `sessions create`,
  `sessions detect`, `sessions list --live` and JSON output contracts.
- The e2e job must have an explicit timeout and must surface `zellij` start logs
  on failure.

Expected CI/local flow:

1. Build `zelma` on the host runner.
2. Upload/download the binary artifact between jobs.
3. Build `Dockerfile.e2e` through `make test-e2e`.
4. Run the container with read-only mounts for the binary, e2e scripts and
   synthetic fixtures.
5. Fail the job on any test failure or leaked background `zellij` session.

This target covers the live terminal boundary. Deterministic parser,
registry, Codex metadata and ambiguity cases still need ordinary fixture tests
outside Docker.

## Core Rules

- Любое изменение поведения, которое можно проверить детерминированно, обязано получить automated regression coverage.
- Любой новый или измененный contract обязан получить contract-level automated verification.
- Любой bugfix обязан добавить regression test на воспроизводимый сценарий.
- Required automated tests считаются закрывающими риск только если они проходят локально и в CI.
- Manual-only verify допустим только как явное исключение и не заменяет automated coverage там, где automation реалистична.

## Ownership Split

- Canonical test cases delivery-единицы задаются в `brief.md` через `SC-*`, feature-specific `NEG-*`, `CHK-*` и `EVID-*`.
- `design.md`, если нужен, владеет selected design, C4 applicability/model, `CTR-*`, `INV-*`, `FM-*` и локальными `RB-*`, но не подменяет canonical verify contract.
- `implementation-plan.md` владеет только стратегией исполнения: какие test surfaces будут добавлены или обновлены, какие gaps временно остаются manual-only и почему.

## Feature Flow Expectations

Canonical lifecycle gates живут в [../flows/feature-flow.md](../flows/feature-flow.md):

- к `Problem Ready` `brief.md` уже фиксирует test case inventory;
- к `Solution Ready` required `design.md` фиксирует selected design, C4 applicability/model, contracts и solution-level failure modes;
- к `Plan Ready` `implementation-plan.md` содержит `Test Strategy` с planned automated coverage и manual-only gaps;
- к `Done` required tests добавлены, локальные команды зелёные и CI не противоречит локальному verify.

## Что Считается Sufficient Coverage

- Покрыт основной changed behavior и ближайший regression path.
- Покрыты новые или измененные contracts, события, schema или integration boundaries.
- Покрыты критичные failure modes из `FM-*` в required `design.md`, bug history или acceptance risks.
- Покрыты feature-specific negative/edge scenarios, если они меняют verdict.
- Процент line coverage сам по себе недостаточен: нужен scenario- и contract-level coverage.

## Когда Manual-Only Допустим

- Сценарий зависит от live infra, внешних систем, hardware, недетерминированной среды или human оценки UI.
- Для каждого manual-only gap: причина, ручная процедура, owner follow-up.
- Если manual-only gap оставляет без regression protection критичный путь, feature не считается завершённой.

## Simplify Review

Отдельный проход верификации после функционального тестирования. Цель: убедиться, что реализация минимально сложна.

- Выполняется после прохождения tests, но до closure gate.
- Паттерны: premature abstractions, глубокая вложенность, дублирование логики, dead code, overengineering.
- Три похожие строки лучше premature abstraction. Абстракция оправдана только когда она реально уменьшает риск или повтор.

## Verification Context Separation

Разные этапы верификации — отдельные проходы:

1. **Функциональная верификация** — tests проходят, acceptance scenarios покрыты
2. **Simplify review** — код минимально сложен
3. **Acceptance test** — end-to-end по `SC-*`

Для small features допустимо в одной сессии, но simplify review не пропускается.

## Project-Specific Conventions

- Registry schema changes require contract tests against sample
  `.zelma/sessions.json` files.
- `sessions detect` requires fixtures for `zellij` inspection output and Codex
  identification evidence before it can be considered regression-covered.
- `sessions create` requires integration coverage or a documented manual-only
  gap until reliable local automation around `zellij` pane creation exists.
- Reliable local automation around `zellij` means the Docker e2e target above,
  not an ad hoc manual run inside a developer's current terminal session.
- CLI output intended for skills requires machine-readable contract tests.
- `zelma`, `zelma help` and command-specific help require snapshot/contract tests
  that assert agent-first ordering: command map and copy-ready examples appear
  before human explanatory prose.
- Do not use Codex conversation contents as test fixtures unless a future
  feature explicitly requires it and privacy constraints are documented.

## Checklist For Project Adoption

- [x] documented current memory-bank check
- [x] выбран Go stack
- [x] указаны ожидаемые Go CLI local test commands
- [x] перечислена обязательная CI suite model
- [x] задокументирован deterministic fixture pattern для `zellij`/Codex/registry
- [x] описан Docker e2e target вместо manual-only live terminal integration
- [ ] policy не противоречит [../flows/feature-flow.md](../flows/feature-flow.md)
