---
title: "FT-034: Manual Pane Adoption E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия detect flow вручную созданных Codex pane."
derived_from:
  - ../../use-cases/UC-002-manual-pane-adoption.md
status: active
delivery_status: implemented
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-034: Manual Pane Adoption E2E

## Что

### Проблема

Пользователь может создать zellij pane и запустить Codex вручную. Zelma обязан принимать такую pane под контроль без дублей и ложных active-записей.

### Результат

`zelma instances detect --json` покрыт e2e-тестом с несколькими pane: уверенная Codex pane добавляется, не-Codex pane пропускается, повторный запуск идемпотентен.

### Объем Работ

- `REQ-01` Смоделировать zellij panes с Codex и non-Codex командами.
- `REQ-02` Проверить upsert в `.zelma/instances.json`.
- `REQ-03` Проверить повторный detect без дублей.
- `REQ-04` Проверить JSON summary added/known/skipped.

### Что Не Входит

- `NS-01` Нет изменения classifier policy beyond current rules.

## Проверка

### Критерии Готовности

- `EC-01` Codex pane записана как session.
- `EC-02` Non-Codex pane не записана как active.
- `EC-03` Повторный запуск detect не меняет registry.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста, который исполняет `zelma instances detect --json` через CLI entrypoint.
