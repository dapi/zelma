---
title: "FT-036: Issue Supervisor Orchestration E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия запуска и контроля issue shipping agent."
derived_from:
  - ../../use-cases/UC-004-issue-supervisor-orchestration.md
  - ../../prompts/PROMPT-005-start-issue-shipping-supervisor.md
status: draft
delivery_status: planned
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-036: Issue Supervisor Orchestration E2E

## Что

### Проблема

Supervisor flow должен быть проверяемым end-to-end: запуск issue agent, polling, review/fix loop и cleanup не должны жить только как prompt convention.

### Результат

Добавлен e2e-harness, который моделирует issue agent completion, review with findings, fix pass и final clean review.

### Объем Работ

- `REQ-01` Покрыть launch of `start-issue <issue>` в zellij pane.
- `REQ-02` Проверить polling active pane не реже одного раза в минуту через controllable clock или fixture.
- `REQ-03` Проверить review/fix/re-review цикл до clean review.
- `REQ-04` Проверить cleanup pane/registry после успешного merge simulation.

### Что Не Входит

- `NS-01` Нет настоящего GitHub merge в e2e.
- `NS-02` Нет изменения prompt model selection кроме уже зафиксированных требований.

## Проверка

### Критерии Готовности

- `EC-01` E2E доказывает повторный review после fix.
- `EC-02` E2E доказывает cleanup после completion.
- `EC-03` E2E доказывает structured supervisor state.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста supervisor happy path and review-fix-repeat path.
