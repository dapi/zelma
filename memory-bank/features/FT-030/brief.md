---
title: "FT-030: Lifecycle State Tests"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для test suite, который покрывает lifecycle states registry/live/stale."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-007/brief.md
  - ../../domain/states.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-030: Lifecycle State Tests

## Что

### Проблема

Lifecycle behavior затрагивает list, detect, stale и cleanup. Без общей test
suite легко получить несовместимые state transitions.

### Результат

Есть lifecycle state tests, которые фиксируют основные transitions и запрещают
destructive default behavior.

### Объем Работ

- `REQ-01` Покрыть registered/live/stale state transitions.
- `REQ-02` Покрыть transient zellij failures.
- `REQ-03` Проверить no destructive cleanup by default.

### Что Не Входит

- `NS-01` Нет реализации новых commands.
- `NS-02` Нет performance/load tests.
- `NS-03` Нет UI tests.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Feature тестирует already selected lifecycle rules. | `none` |

## Проверка

- `SC-01` State transitions соответствуют domain rules.
- `SC-02` Destructive cleanup не происходит без explicit action.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | lifecycle test suite | transitions pass | `artifacts/ft-030/verify/chk-01/` |
| `CHK-02` | `REQ-03` | safety regression tests | no destructive default | `artifacts/ft-030/verify/chk-02/` |

### Доказательства

- `EVID-01` Lifecycle test suite output.
- `EVID-02` Safety regression test output.
