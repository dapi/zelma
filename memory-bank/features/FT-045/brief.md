---
title: "FT-045: Numeric Zelma Session IDs"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для коротких уникальных numeric ID у каждой zelma instance."
derived_from:
  - ../../product/roadmap.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../FT-006/brief.md
  - ../FT-009/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-045: Numeric Zelma Session IDs

## Что

### Проблема

`instances list` показывает длинные external refs (`zellij_session`,
`zellij_pane`, `codex_session`), но у `zelma instance` нет короткого
repo-local identifier. Пользователю и future commands сложно ссылаться на
конкретную запись без копирования pane/session refs.

### Результат

Каждая запись `ZelmaInstance` имеет уникальный положительный integer `id`,
начинающийся с `1` внутри repo-local `.zelma/instances.json`.

### Объем Работ

- `REQ-01` Добавить поле `sessions[].id` в schema v1 registry records.
- `REQ-02` Backfill старых registry records без `id` при чтении и следующей
  записи без ручной migration-команды.
- `REQ-03` Сохранять уже назначенные positive IDs и назначать новым records
  следующий positive ID.
- `REQ-04` Reject duplicate positive IDs и invalid negative IDs.
- `REQ-05` Показывать ID в `instances list` и `instances list --live` как первый
  table column.
- `REQ-06` Включать ID в JSON output для sessions, stale records и stale
  candidate summaries.

### Что Не Входит

- `NS-01` Нет отдельной команды renumber/migrate.
- `NS-02` Нет глобальной уникальности между разными repositories.
- `NS-03` Нет пользовательских string aliases вместо numeric ID.

## Проверка

### Критерии Готовности

- `EC-01` Старый registry без `sessions[].id` читается и выводится с ID `1..n`.
- `EC-02` Mutating commands write positive IDs into `.zelma/instances.json`.
- `EC-03` Existing positive IDs remain stable across repeated detect/list.
- `EC-04` Duplicate or negative IDs produce registry diagnostics.

### Обязательное Покрытие

- Registry tests для backfill, next-ID assignment и validation failures.
- CLI tests для table/JSON output with ID.
- Machine-readable compatibility tests for strict JSON consumers.
