---
title: "FT-007: Atomic Registry Writes And Lock"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для atomic writes и lock механизма вокруг `.zelma/sessions.json`."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-002/brief.md
  - ../../engineering/architecture.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-007: Atomic Registry Writes And Lock

## Что

### Проблема

Несколько `zelma` команд или агентов могут одновременно обновлять registry.
Прямые writes могут повредить JSON или потерять изменения.

### Результат

Registry writes выполняются атомарно и защищены lock-механизмом с понятными
ошибками при конфликте доступа.

### Объем Работ

- `REQ-01` Реализовать atomic write contract для registry file.
- `REQ-02` Добавить lock/guard для concurrent writers.
- `REQ-03` Вернуть recoverable diagnostics при lock/write failures.

### Что Не Входит

- `NS-01` Нет изменения schema v1.
- `NS-02` Нет distributed lock между машинами.
- `NS-03` Нет background writer/daemon.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Atomic write и lock меняют filesystem failure modes. | `design.md` |

## Проверка

- `SC-01` Успешная запись не оставляет частичный JSON.
- `SC-02` Concurrent write conflict не повреждает существующий registry.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01` | filesystem tests | atomic replacement завершается успешно | `artifacts/ft-007/verify/chk-01/` |
| `CHK-02` | `REQ-02`, `REQ-03` | concurrency/failure tests | no corruption; clear error | `artifacts/ft-007/verify/chk-02/` |

### Доказательства

- `EVID-01` Atomic write test output.
- `EVID-02` Concurrent write/failure test output.
