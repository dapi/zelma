---
title: "FT-039: Agent Handoff E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия handoff между агентами."
derived_from:
  - ../../use-cases/UC-007-agent-handoff.md
status: active
delivery_status: implemented
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-039: Agent Handoff E2E

## Что

### Проблема

Новый агент должен восстановить active work из registry/live state и не запускать duplicate issue agents.

### Результат

E2E-тест моделирует новый процесс агента, registry reload и принятие решения continue/poll вместо duplicate create.

### Объем Работ

- `REQ-01` Подготовить registry после предыдущего agent run.
- `REQ-02` Запустить `sessions list --live --json` в fresh process.
- `REQ-03` Проверить active/stale classification.
- `REQ-04` Проверить duplicate launch guard для уже активной issue/session.

### Что Не Входит

- `NS-01` Нет persistence beyond `.zelma/sessions.json`.

## Проверка

### Критерии Готовности

- `EC-01` Новый agent process видит active sessions.
- `EC-02` Stale sessions ведут к cleanup guidance.
- `EC-03` Active issue не запускается повторно.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста handoff from persisted registry.
