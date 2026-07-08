---
title: Execution Order
doc_kind: product
doc_function: delivery_sequence
purpose: Практический порядок запуска GitHub issues и параллельных волн реализации.
derived_from:
  - roadmap.md
  - ../epics/README.md
  - ../features/README.md
status: active
audience: humans_and_agents
---

# Execution Order

Этот документ фиксирует порядок запуска feature issues в производство. Он
описывает delivery sequencing и параллелизм, но не заменяет brief конкретной
feature.

## Правила Запуска

- Запускай параллельно только задачи с независимыми write scopes и без прямой
  зависимости по runtime contract.
- После merge каждой волны подтягивай `main` перед стартом следующей волны.
- Если задача выявляет новый контракт, downstream задачи должны использовать
  merged contract, а не локальные предположения.
- Review, CI и merge gates обязательны для каждой feature issue.

## Волна 0: CLI And Registry Foundation

Перед zellij/read/detect работами должны быть готовы базовые EP-001/EP-002
контракты:

- `FT-001`: Go Module Scaffold
- `FT-002`: Cobra Command Tree
- `FT-003`: Agent-First Help Templates
- `FT-004`: Output And Error Contract Tests
- `FT-005`: Repo Root Resolver
- `FT-006`: Sessions Schema V1
- `FT-007`: Atomic Registry Writes And Lock
- `FT-008`: Registry Validation And Recovery
- `FT-009`: Sessions List Output
- `FT-031`: Zelma Setup Gitignore

Результат волны: запускаемый CLI, repo root resolution, registry schema/read
surface и setup behavior существуют как merged baseline для downstream issues.

## Волна 1: Zellij Read Path

Можно запускать параллельно:

- `#22` / `FT-010`: Zellij Adapter ListSessions
- `#23` / `FT-011`: Zellij Adapter ListPanes
- `#24` / `FT-012`: Zellij JSON Fixture Tests

Результат волны: устойчивый read-only контракт zellij adapter и fixture база
для downstream detect/create/live flows.

## Волна 2: Detect MVP

После merge `FT-010`...`FT-012` можно запускать параллельно:

- `#25` / `FT-013`: Codex Pane Candidate Classifier
- `#26` / `FT-014`: Detect Upsert Idempotency

Результат волны: `zelma sessions detect` может консервативно находить Codex
pane candidates и идемпотентно обновлять registry.

## Волна 3: Create MVP

После базового zellij adapter contract можно запускать параллельно:

- `#27` / `FT-015`: Codex Launch Contract
- `#28` / `FT-016`: Zellij Run New-Pane Adapter

Затем последовательно:

- `#29` / `FT-017`: Create Confirmation And Reconciliation
- `#30` / `FT-018`: Create Failure Recovery Hints

Результат волны: `zelma sessions create` создает managed Codex pane,
подтверждает ее и дает понятные recovery hints при частичных сбоях.

## Волна 4: Codex Session Identity

Сначала:

- `#31` / `FT-019`: Codex Metadata Source Discovery

После discovery можно запускать параллельно:

- `#32` / `FT-020`: Session Evidence Parser
- `#34` / `FT-022`: Privacy-Safe Evidence Fixtures

Затем:

- `#33` / `FT-021`: Candidate Vs Active State Rules

Результат волны: records получают надежный `CodexSessionRef` или явно остаются
в candidate state при недостатке evidence.

## Волна 5: Lifecycle

После read adapter, registry и detect basics можно запускать параллельно:

- `#39` / `FT-027`: Sessions List Live
- `#40` / `FT-028`: Stale Detection
- `#42` / `FT-030`: Lifecycle State Tests

Затем:

- `#41` / `FT-029`: Cleanup Remove Proposal

Результат волны: CLI различает registered/live/stale state и предлагает cleanup
как явное действие без destructive default.

## Волна 6: Skill Pack

После стабилизации CLI JSON/output contract и команд `list/create/detect`
можно запускать параллельно:

- `#36` / `FT-024`: Machine-Readable Output Compatibility Tests
- `#37` / `FT-025`: Skill Docs

Затем можно запускать параллельно:

- `#35` / `FT-023`: Skill Command Wrappers
- `#38` / `FT-026`: Agent Recovery Flows

Результат волны: Codex skills становятся thin wrappers над `zelma` CLI и
используют тот же stable output contract.

## Краткий Практический Порядок

0. Завершить foundation baseline: `FT-001`...`FT-009`, `FT-031`.
1. Параллельно: `#22`, `#23`, `#24`.
2. Параллельно: `#25`, `#26`.
3. Параллельно: `#27`, `#28`.
4. Последовательно: `#29`, затем `#30`.
5. Сначала: `#31`.
6. Параллельно: `#32`, `#34`.
7. Затем: `#33`.
8. Параллельно: `#39`, `#40`, `#42`.
9. Затем: `#41`.
10. Параллельно: `#36`, `#37`.
11. Затем параллельно: `#35`, `#38`.
