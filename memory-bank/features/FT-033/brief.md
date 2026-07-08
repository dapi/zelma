---
title: "FT-033: Agent Session Inventory E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия инвентаризации agent-сессий."
derived_from:
  - ../../use-cases/UC-001-agent-session-inventory.md
status: active
delivery_status: implemented
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-033: Agent Session Inventory E2E

## Что

### Проблема

Supervisor должен надежно получать machine-readable inventory active/stale zelma-сессий. Сейчас этот сценарий должен быть закреплен как e2e-contract, чтобы будущие изменения CLI не ломали agentic orchestration.

### Результат

`zelma sessions list --live --json` покрыт e2e-тестом с registry fixture и zellij adapter fixture, который проверяет active, stale и empty states.

### Объем Работ

- `REQ-01` Добавить e2e fixture для registry с несколькими sessions.
- `REQ-02` Смоделировать live zellij state для active и missing pane.
- `REQ-03` Проверить JSON contract для active, stale и empty list.
- `REQ-04` Проверить, что команда не меняет `.zelma/sessions.json`.

### Что Не Входит

- `NS-01` Нет изменения registry schema.
- `NS-02` Нет dashboard UI.

## Проверка

### Критерии Готовности

- `EC-01` E2E-тест запускает CLI binary, а не внутреннюю функцию.
- `EC-02` JSON output пригоден для agent parsing.
- `EC-03` Stale запись определяется без удаления registry entry.

### Сценарии Приемки

- `SC-01` Given registry with active pane, when `sessions list --live --json`, then output contains live session metadata.
- `SC-02` Given registry with missing pane, when `sessions list --live --json`, then output marks it stale.
- `SC-03` Given empty registry, when command runs, then output is an empty sessions list.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста, который исполняет `zelma sessions list --live --json` через CLI entrypoint.
