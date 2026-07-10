---
title: Use Cases Index
doc_kind: use_case
doc_function: index
purpose: Навигация по instantiated use cases проекта. Читать, чтобы найти канонический сценарий продукта или зарегистрировать новый.
derived_from:
  - ../dna/governance.md
  - ../flows/templates/use-case/UC-XXX.md
status: active
audience: humans_and_agents
---

# Use Cases Index

Каталог `memory-bank/use-cases/` хранит канонические пользовательские и операционные сценарии проекта.

Use case нужен для сценария, который живет на уровне продукта, повторяется во времени и может быть upstream для нескольких feature packages. Это не замена `SC-*` внутри `brief.md`: `SC-*` описывают acceptance сценарии delivery-единицы, а `UC-*` описывают устойчивое поведение системы на уровне проекта.

Обычно use case наследует общий product context из [`../product/context.md`](../product/context.md). Если сценарий зависит от предметных правил, states или events, он также должен ссылаться на соответствующие документы из [`../domain/README.md`](../domain/README.md).

## Когда Заводить Use Case

- появляется новый стабильный пользовательский или операционный сценарий;
- несколько features реализуют или меняют один и тот же flow;
- нужен канонический owner для trigger, preconditions, main flow и postconditions.

## Когда Use Case Не Нужен

- сценарий одноразовый и живет только внутри одной feature;
- это implementation detail, а не продуктовый или операционный flow;
- его достаточно описать через `SC-*` в `brief.md`.

## Реестр

| UC ID | Title | Status | Primary actor | Upstream PRD | Implemented by | Last updated |
| --- | --- | --- | --- | --- | --- | --- |
| [`UC-001`](UC-001-agent-session-inventory.md) | Инвентаризация agent-сессий | `draft` | supervising agent | `none` | `FT-033` | 2026-07-08 |
| [`UC-002`](UC-002-manual-pane-adoption.md) | Взятие вручную созданной pane под контроль | `draft` | supervising agent | `none` | `FT-034` | 2026-07-08 |
| [`UC-003`](UC-003-managed-agent-launch.md) | Управляемый запуск новой agent-сессии | `draft` | supervising agent | `none` | `FT-035` | 2026-07-08 |
| [`UC-004`](UC-004-issue-supervisor-orchestration.md) | Supervisor orchestration для GitHub issue | `draft` | shipping supervisor | `none` | `FT-036` | 2026-07-08 |
| [`UC-005`](UC-005-agent-recovery.md) | Восстановление после ошибок agent-сессии | `draft` | supervising agent | `none` | `FT-037` | 2026-07-08 |
| [`UC-006`](UC-006-stale-cleanup.md) | Очистка stale-сессий после завершения задачи | `draft` | supervising agent | `none` | `FT-038` | 2026-07-08 |
| [`UC-007`](UC-007-agent-handoff.md) | Handoff между агентами | `draft` | incoming agent | `none` | `FT-039` | 2026-07-08 |
| [`UC-008`](UC-008-multi-agent-parallel-delivery.md) | Параллельная доставка несколькими агентами | `draft` | shipping supervisor | `none` | `FT-040` | 2026-07-08 |
| [`UC-009`](UC-009-environment-smoke-diagnostics.md) | Smoke-диагностика окружения | `draft` | setup agent | `none` | `FT-041` | 2026-07-08 |
| [`UC-010`](UC-010-agent-dashboard-status-backend.md) | Dashboard/status backend для agent-сессий | `draft` | dashboard agent | `none` | `FT-042` | 2026-07-08 |
| [`UC-011`](UC-011-send-message-to-codex-session.md) | Отправка сообщения в существующую Codex session | `draft` | human / supervising agent | `none` | `FT-101` | 2026-07-10 |

## Naming

- Формат файла: `UC-XXX-short-name.md`
- Вместо `XXX` используй стабильный проектный идентификатор
- Один use case может быть upstream для нескольких feature packages

## Template

- Используй шаблон [`../flows/templates/use-case/UC-XXX.md`](../flows/templates/use-case/UC-XXX.md)
