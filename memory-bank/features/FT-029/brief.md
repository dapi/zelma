---
title: "FT-029: Cleanup Remove Proposal"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для явного cleanup/remove proposal flow без destructive default."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-007/brief.md
  - ../../features/FT-028/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-029: Cleanup Remove Proposal

## Что

### Проблема

После stale detection пользователю нужен способ убрать устаревшие records, но
автоматическое удаление опасно.

### Результат

CLI предлагает explicit cleanup/remove path с summary того, какие records будут
изменены.

### Объем Работ

- `REQ-01` Показать cleanup proposal для stale records.
- `REQ-02` Требовать явное действие для удаления/cleanup.
- `REQ-03` Сохранять audit-friendly summary.

### Что Не Входит

- `NS-01` Нет automatic cleanup по умолчанию.
- `NS-02` Нет удаления live records.
- `NS-03` Нет глобального registry cleanup.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Cleanup меняет registry state и требует explicit safety contract. | `design.md` |

## Проверка

- `SC-01` Proposal показывает stale records перед cleanup.
- `SC-02` Без явного действия registry не меняется.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-03` | proposal output tests | expected summary | `artifacts/ft-029/verify/chk-01/` |
| `CHK-02` | `REQ-02` | no-confirm tests | no registry change | `artifacts/ft-029/verify/chk-02/` |

### Доказательства

- `EVID-01` Cleanup proposal output.
- `EVID-02` No-confirm safety test.
