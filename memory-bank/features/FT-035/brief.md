---
title: "FT-035: Managed Agent Launch E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия управляемого запуска agent-сессии."
derived_from:
  - ../../use-cases/UC-003-managed-agent-launch.md
status: active
delivery_status: implemented
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-035: Managed Agent Launch E2E

## Что

### Проблема

Create flow должен быть доказан end-to-end: запуск pane, подтверждение zellij state и запись registry обязаны работать как единый agent-facing контракт.

### Результат

`zelma sessions create --json` покрыт e2e-тестом create-to-list с fake zellij executable и проверкой registry output.

### Объем Работ

- `REQ-01` Подготовить fake zellij для запуска new-pane и последующего list.
- `REQ-02` Проверить успешный create и registry write.
- `REQ-03` Проверить create-to-list reconciliation.
- `REQ-04` Проверить failure path без ложной active записи.

### Что Не Входит

- `NS-01` Нет изменения Codex launch command contract.

## Проверка

### Критерии Готовности

- `EC-01` CLI возвращает created session JSON.
- `EC-02` `sessions list --live --json` видит созданную session.
- `EC-03` Failure path содержит recovery hint.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста, который исполняет `zelma sessions create --json` и затем `zelma sessions list --live --json`.
