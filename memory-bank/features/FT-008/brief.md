---
title: "FT-008: Registry Validation And Recovery"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для validation, diagnostics и recovery hints при некорректном registry state."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-002/brief.md
  - ../../domain/rules.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-008: Registry Validation And Recovery

## Что

### Проблема

Registry может быть поврежден вручную, устареть или содержать duplicate
records. Команды должны безопасно диагностировать это состояние, а не молча
перезаписывать файл.

### Результат

`zelma` валидирует registry перед использованием и выдает agent-friendly
recovery hints без разрушительных действий по умолчанию.

### Объем Работ

- `REQ-01` Валидировать JSON, version и обязательные поля.
- `REQ-02` Находить duplicate/conflicting session records.
- `REQ-03` Возвращать recovery diagnostics с machine-readable error codes.

### Что Не Входит

- `NS-01` Нет автоматической destructive repair.
- `NS-02` Нет live zellij reconciliation.
- `NS-03` Нет migration framework за пределами v1.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Recovery behavior и error codes становятся CLI/storage contract. | `design.md` |

## Проверка

- `SC-01` Invalid JSON приводит к понятной ошибке и не меняет файл.
- `SC-02` Duplicate record определяется как validation problem.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01` | invalid fixture tests | clear validation error | `artifacts/ft-008/verify/chk-01/` |
| `CHK-02` | `REQ-02`, `REQ-03` | duplicate/conflict fixture tests | error code + recovery hint | `artifacts/ft-008/verify/chk-02/` |

### Доказательства

- `EVID-01` Invalid fixture test output.
- `EVID-02` Duplicate/conflict diagnostic output.
