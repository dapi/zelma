---
title: "FT-010: Zellij Adapter ListSessions"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для adapter method, который читает zellij sessions без изменения внешнего состояния."
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

# FT-010: Zellij Adapter ListSessions

## Что

### Проблема

`instances detect` должен сначала узнать доступные zellij sessions. Этот доступ
нужно изолировать в adapter, чтобы domain logic не зависела напрямую от
`os/exec` и формата zellij CLI.

### Результат

Go adapter предоставляет read-only метод получения zellij sessions и
возвращает нормализованные ошибки для CLI.

### Объем Работ

- `REQ-01` Добавить adapter method для чтения списка zellij sessions.
- `REQ-02` Нормализовать успешный результат в project-owned model.
- `REQ-03` Нормализовать ошибки missing zellij, non-zero exit и invalid output.

### Что Не Входит

- `NS-01` Нет чтения panes.
- `NS-02` Нет записи registry.
- `NS-03` Нет создания или изменения zellij sessions.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Adapter boundary и error mapping являются integration contract. | `design.md` |

## Проверка

- `SC-01` Adapter парсит fixture с несколькими zellij sessions.
- `SC-02` Missing zellij binary возвращает agent-friendly error.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | adapter unit tests with fake command output | sessions parsed | `artifacts/ft-010/verify/chk-01/` |
| `CHK-02` | `REQ-03` | fake command failure tests | normalized errors | `artifacts/ft-010/verify/chk-02/` |

### Доказательства

- `EVID-01` Adapter parsing test output.
- `EVID-02` Adapter error mapping test output.
