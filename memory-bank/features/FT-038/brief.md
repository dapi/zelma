---
title: "FT-038: Stale Cleanup E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия stale cleanup."
derived_from:
  - ../../use-cases/UC-006-stale-cleanup.md
status: active
delivery_status: implemented
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-038: Stale Cleanup E2E

## Что

### Проблема

Cleanup stale registry entries должен быть безопасным: сначала proposal, потом explicit confirm, без удаления live sessions.

### Результат

E2E-тест проверяет stale proposal, no-op без confirm, confirm deletion и идемпотентный повтор.

### Объем Работ

- `REQ-01` Подготовить registry с active и stale entries.
- `REQ-02` Проверить `sessions cleanup --json` как read-only proposal.
- `REQ-03` Проверить `sessions cleanup --confirm --json`.
- `REQ-04` Проверить повторный cleanup как empty proposal.

### Что Не Входит

- `NS-01` Нет удаления live zellij pane.

## Проверка

### Критерии Готовности

- `EC-01` Без confirm registry не меняется.
- `EC-02` Confirm удаляет только stale entries.
- `EC-03` Повторный cleanup безопасен.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста proposal plus confirm cleanup.
