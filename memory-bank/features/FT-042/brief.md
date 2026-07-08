---
title: "FT-042: Agent Dashboard Status Backend"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для milestone-2 status backend поверх zelma-сессий."
derived_from:
  - ../../use-cases/UC-010-agent-dashboard-status-backend.md
status: draft
delivery_status: planned
milestone: milestone-2
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-042: Agent Dashboard Status Backend

## Что

### Проблема

Dashboard и внешние agent UI не должны зависеть от внутренних деталей `.zelma/sessions.json` или отдельных CLI команд. Нужен status backend с версионированной моделью, но это не должно блокировать milestone-1.

### Результат

В milestone-2 появляется status backend/command, который агрегирует registry, live zellij state, task metadata и recovery hints в один machine-readable snapshot.

### Объем Работ

- `REQ-01` Спроектировать versioned status model.
- `REQ-02` Агрегировать active/stale/blocked/completed sessions.
- `REQ-03` Предоставить backend command или endpoint для dashboard/agent UI.
- `REQ-04` Покрыть active plus stale snapshot e2e-тестом.

### Что Не Входит

- `NS-01` Не входит в milestone-1.
- `NS-02` Нет обязательного frontend dashboard в рамках этой feature.

## Проверка

### Критерии Готовности

- `EC-01` Snapshot model версионирован.
- `EC-02` Dashboard может получить status без чтения registry internals.
- `EC-03` Degraded zellij state возвращает recovery hints.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста backend snapshot with active and stale sessions.
