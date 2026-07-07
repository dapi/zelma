---
title: "FT-028: Stale Detection"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для правил detection stale records в registry."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-007/brief.md
  - ../../domain/states.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-028: Stale Detection

## Что

### Проблема

Registry records могут ссылаться на закрытые или недоступные panes. Нужны
правила, которые отделяют stale state от transient zellij failures.

### Результат

`zelma` определяет stale candidates и объясняет причину без автоматического
удаления.

### Объем Работ

- `REQ-01` Определить stale criteria.
- `REQ-02` Отличать stale от временной ошибки zellij.
- `REQ-03` Возвращать reason codes для stale candidates.

### Что Не Входит

- `NS-01` Нет удаления stale records.
- `NS-02` Нет user confirmation flow.
- `NS-03` Нет background scans.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Stale criteria являются lifecycle/domain contract. | `design.md` |

## Проверка

- `SC-01` Missing pane получает stale candidate reason.
- `SC-02` Zellij unavailable не превращает все records в stale.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-03` | lifecycle tests | stale reason emitted | `artifacts/ft-028/verify/chk-01/` |
| `CHK-02` | `REQ-02` | zellij failure tests | transient error preserved | `artifacts/ft-028/verify/chk-02/` |

### Доказательства

- `EVID-01` Stale criteria test output.
- `EVID-02` Transient failure test output.
