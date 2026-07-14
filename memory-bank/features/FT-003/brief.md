---
title: "FT-003: Agent-First Шаблоны Help"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для agent-first help templates в `zelma` и `zelma instances`."
derived_from:
  - ../../product/context.md
  - ../../product/vision.md
  - ../../epics/EP-001/brief.md
  - ../../engineering/architecture.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-003: Agent-First Шаблоны Help

## Что

### Проблема

Default CLI help обычно оптимизирован для человека. `zelma` должна сначала
показывать агенту command map, stable flags, output modes и recovery hints, а
описательный текст оставлять вторичным.

### Результат

| ID метрики | Метрика | База | Цель | Способ измерения |
| --- | --- | --- | --- | --- |
| `MET-01` | Agent-first help | generic/default help | command map и machine-use hints выводятся раньше prose | snapshot tests |

### Объем Работ

- `REQ-01` Настроить top-level help для `zelma`.
- `REQ-02` Настроить `zelma instances help`.
- `REQ-03` Включить agent-first command map, stable output conventions и recovery hint pattern.

### Что Не Входит

- `NS-01` Нет final behavior для session commands.
- `NS-02` Нет реализации skills.
- `NS-03` Нет localization system.

### Ограничения И Предположения

- `CON-01` Help должен оставаться читаемым человеком после agent-first sections.
- `CON-02` Help не должен обещать behavior вне реализованных или stubbed commands.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Help output является стабильным CLI contract для агентов. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` `zelma help` и bare `zelma` output являются agent-first.
- `EC-02` `zelma instances help` показывает subcommand map и ожидаемые output modes.

### Матрица Трассировки

| ID требования | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-01`, `EC-02` | `CHK-02` | `EVID-02` |

### Сценарии Приемки

- `SC-01` Агент читает top-level help и может выбрать next commands без prose parsing.
- `SC-02` Агент читает `sessions` help и может выбрать list/create/detect.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02` | CLI snapshot tests | ожидаемый help layout | `artifacts/ft-003/verify/chk-01/` |
| `CHK-02` | `REQ-03` | review output headings/order | command map идет перед prose | `artifacts/ft-003/verify/chk-02/` |

### Доказательства

- `EVID-01` Output snapshot tests.
- `EVID-02` Review note по help contract.
