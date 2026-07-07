---
title: "FT-014: Detect Upsert Idempotency"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для idempotent registry upsert в `zelma sessions detect`."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-003/brief.md
  - ../../epics/EP-002/brief.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-014: Detect Upsert Idempotency

## Что

### Проблема

`zelma sessions detect` может запускаться многократно. Повторный запуск не
должен создавать duplicates или перетирать более точную информацию.

### Результат

Detect upsert добавляет новые candidate sessions и обновляет существующие
records идемпотентно, с понятным summary для агента.

### Объем Работ

- `REQ-01` Сопоставлять detected pane с existing registry record.
- `REQ-02` Добавлять новый candidate record только при отсутствии match.
- `REQ-03` Повторный detect не создает duplicates.
- `REQ-04` Возвращать summary added/unchanged/skipped.

### Что Не Входит

- `NS-01` Нет destructive cleanup stale records.
- `NS-02` Нет create workflow.
- `NS-03` Нет окончательной Codex session identity, если evidence неполное.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Upsert rules меняют registry state и conflict behavior. | `design.md` |

## Проверка

- `SC-01` Первый detect добавляет candidate record.
- `SC-02` Повторный detect оставляет один record и summary `unchanged`.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | detect tests with temp registry | new candidate added | `artifacts/ft-014/verify/chk-01/` |
| `CHK-02` | `REQ-03`, `REQ-04` | repeated detect tests | no duplicate; stable summary | `artifacts/ft-014/verify/chk-02/` |

### Доказательства

- `EVID-01` First detect upsert test output.
- `EVID-02` Repeated detect idempotency test output.
