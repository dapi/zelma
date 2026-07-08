---
title: "FT-040: Multi-Agent Parallel Delivery E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия параллельной доставки несколькими агентами."
derived_from:
  - ../../use-cases/UC-008-multi-agent-parallel-delivery.md
status: draft
delivery_status: planned
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-040: Multi-Agent Parallel Delivery E2E

## Что

### Проблема

Parallel delivery требует доказать, что supervisor не запускает дубли, выдерживает startup lock/retry и корректно отслеживает несколько pane.

### Результат

E2E-тест запускает несколько simulated issue agents, проверяет staggered startup, registry uniqueness, polling и completion handling.

### Объем Работ

- `REQ-01` Смоделировать parallel group из independent issues.
- `REQ-02` Проверить задержку или retry при startup lock.
- `REQ-03` Проверить уникальность worktree/session per issue.
- `REQ-04` Проверить completion/merge simulation и запуск следующей задачи.

### Что Не Входит

- `NS-01` Нет реального concurrent merge в GitHub.

## Проверка

### Критерии Готовности

- `EC-01` Несколько agents видны в registry без дублей.
- `EC-02` Lock/retry сценарий проходит без падения supervisor.
- `EC-03` Next task стартует только после dependency completion.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста multi-pane orchestration.
