---
title: "FT-016: Zellij Run New-Pane Adapter"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для zellij adapter, который создает новую pane и запускает command."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-004/brief.md
  - ../../engineering/zellij-integration.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-016: Zellij Run New-Pane Adapter

## Что

### Проблема

Managed create требует надежно открыть zellij pane и запустить в ней Codex
command, сохранив enough metadata для последующего confirmation.

### Результат

Adapter создает pane через zellij CLI, возвращает pane reference или
recoverable error без записи registry.

### Объем Работ

- `REQ-01` Добавить adapter method для создания pane с command.
- `REQ-02` Вернуть zellij session/pane reference, если zellij это позволяет.
- `REQ-03` Нормализовать ошибки zellij create/run.

### Что Не Входит

- `NS-01` Нет registry write.
- `NS-02` Нет confirmation logic после create.
- `NS-03` Нет attach/focus helpers.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Создание pane меняет внешнее состояние zellij и требует failure-mode design. | `design.md` |

## Проверка

- `SC-01` Adapter формирует ожидаемый zellij command для new pane.
- `SC-02` Ошибка zellij возвращается как normalized diagnostic.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | fake zellij command tests | expected invocation/reference | `artifacts/ft-016/verify/chk-01/` |
| `CHK-02` | `REQ-03` | fake failure tests | normalized errors | `artifacts/ft-016/verify/chk-02/` |

### Доказательства

- `EVID-01` Adapter invocation test output.
- `EVID-02` Failure mapping test output.
