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

Runtime-код `zelma` пока не создан. Stack выбран: Go. Текущая обязательная
проверка документационного слоя:

```bash
python3 scripts/check_memory_bank_index.py
```

После выбора CLI стека этот раздел должен зафиксировать:

- основной test framework: Go `testing` package unless an ADR chooses otherwise;
- стратегию fixtures для `zellij` outputs, Codex session evidence и registry files;
- canonical local commands: `go test ./...`, `go vet ./...`, and targeted
  integration tests once scaffold exists;
- обязательные CI jobs;
- допустимые manual-only исключения для live `zellij`/Codex integration.

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
- CLI output intended for skills requires machine-readable contract tests.
- `zelma`, `zelma help` and command-specific help require snapshot/contract tests
  that assert agent-first ordering: command map and copy-ready examples appear
  before human explanatory prose.
- Do not use Codex conversation contents as test fixtures unless a future
  feature explicitly requires it and privacy constraints are documented.

## Checklist For Project Adoption

- [x] documented current memory-bank check
- [x] выбран Go stack
- [ ] указаны реальные CLI local test commands
- [ ] перечислены обязательные CI suites
- [ ] задокументирован deterministic fixture pattern для `zellij`/Codex/registry
- [ ] описаны manual-only exceptions for live terminal integration
- [ ] policy не противоречит [../flows/feature-flow.md](../flows/feature-flow.md)
