---
title: "FT-024: Machine-Readable Output Compatibility Tests"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для tests, которые защищают JSON output contract, используемый skills."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-006/brief.md
  - ../../engineering/testing-policy.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-024: Machine-Readable Output Compatibility Tests

## Что

### Проблема

Skills зависят от machine-readable output. Если JSON shape меняется без
compatibility tests, agents могут начать ошибочно интерпретировать sessions.

### Результат

Compatibility tests фиксируют JSON schemas/examples для CLI outputs,
используемых skill wrappers.

### Объем Работ

- `REQ-01` Зафиксировать JSON examples для list/detect/create summaries.
- `REQ-02` Добавить tests, которые валидируют compatibility с skill wrappers.
- `REQ-03` Проверять error JSON/diagnostics, если такой mode выбран.

### Что Не Входит

- `NS-01` Нет изменения runtime behavior commands.
- `NS-02` Нет поддержки устаревших schemas без migration decision.
- `NS-03` Нет skill UX docs.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Feature тестирует уже выбранный CLI output contract. | `none` |

## Проверка

- `SC-01` Skill wrapper tests проходят на current JSON examples.
- `SC-02` Несовместимое изменение output ломает compatibility test.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | compatibility tests | CLI examples parse in wrappers | `artifacts/ft-024/verify/chk-01/` |
| `CHK-02` | `REQ-03` | diagnostic compatibility tests | errors parsed predictably | `artifacts/ft-024/verify/chk-02/` |

### Доказательства

- `EVID-01` Compatibility test output.
- `EVID-02` Diagnostic compatibility test output.
