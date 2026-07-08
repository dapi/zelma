---
title: "FT-012: Zellij JSON Fixture Tests"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для fixture corpus и tests, которые защищают zellij JSON parsing."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-003/brief.md
  - ../../engineering/zellij-integration.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-012: Zellij JSON Fixture Tests

## Что

### Проблема

Zellij JSON output может меняться между версиями. Без fixtures parser будет
ломаться незаметно или начнет принимать неверные формы данных.

### Результат

В repo есть fixture corpus для поддерживаемых zellij outputs и tests, которые
фиксируют допустимые и недопустимые shapes.

### Объем Работ

- `REQ-01` Добавить fixtures для sessions output.
- `REQ-02` Добавить fixtures для panes output.
- `REQ-03` Добавить negative fixtures для invalid/partial output.

### Что Не Входит

- `NS-01` Нет live zellij integration tests.
- `NS-02` Нет поддержки всех прошлых версий zellij без evidence.
- `NS-03` Нет Codex classification.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Feature фиксирует test corpus для already selected adapter contract. | `none` |

## Проверка

- `SC-01` Valid fixtures проходят parser tests.
- `SC-02` Invalid fixtures дают контролируемую ошибку.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | `go test ./...` | valid fixtures pass | `artifacts/ft-012/verify/chk-01/` |
| `CHK-02` | `REQ-03` | negative fixture tests | invalid output rejected | `artifacts/ft-012/verify/chk-02/` |

### Доказательства

- `EVID-01` Valid fixture test output.
- `EVID-02` Negative fixture test output.
