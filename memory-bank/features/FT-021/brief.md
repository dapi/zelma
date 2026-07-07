---
title: "FT-021: Candidate Vs Active State Rules"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для правил, когда detected session остается candidate, а когда становится active."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-005/brief.md
  - ../../domain/states.md
  - ../../domain/rules.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-021: Candidate Vs Active State Rules

## Что

### Проблема

Detected pane может быть только кандидатом, пока evidence недостаточно. Нужны
единые правила, чтобы create/detect/list одинаково показывали state.

### Результат

Зафиксированы и проверены state transition rules для candidate, active и
insufficient-evidence cases.

### Объем Работ

- `REQ-01` Определить evidence threshold для active state.
- `REQ-02` Определить candidate state при неполном evidence.
- `REQ-03` Обновить output/status semantics для list/detect.

### Что Не Входит

- `NS-01` Нет parser implementation.
- `NS-02` Нет stale lifecycle cleanup.
- `NS-03` Нет ручного force-promote без отдельного feature.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | State rules являются domain contract и влияют на registry schema. | `design.md` |

## Проверка

- `SC-01` Full evidence переводит record в active.
- `SC-02` Partial evidence оставляет record candidate.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01` | state transition tests | active only with full evidence | `artifacts/ft-021/verify/chk-01/` |
| `CHK-02` | `REQ-02`, `REQ-03` | candidate output tests | candidate visible and explainable | `artifacts/ft-021/verify/chk-02/` |

### Доказательства

- `EVID-01` Active transition test output.
- `EVID-02` Candidate state output test.
