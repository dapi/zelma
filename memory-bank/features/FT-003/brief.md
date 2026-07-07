---
title: "FT-003: Agent-First Шаблоны Help"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для agent-first help templates в `zelma` и `zelma sessions`."
derived_from:
  - ../../product/context.md
  - ../../product/vision.md
  - ../../epics/EP-001/brief.md
  - ../../engineering/architecture.md
status: draft
delivery_status: planned
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

### Scope

- `REQ-01` Настроить top-level help для `zelma`.
- `REQ-02` Настроить `zelma sessions help`.
- `REQ-03` Включить agent-first command map, stable output conventions и recovery hint pattern.

### Что Не Входит

- `NS-01` No final behavior for session commands.
- `NS-02` No skill implementation.
- `NS-03` No localization system.

### Ограничения И Предположения

- `CON-01` Help must remain readable by humans after agent-first sections.
- `CON-02` Help must not promise behavior outside implemented or stubbed commands.

## Решение О Необходимости Design

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | Help output является стабильным CLI contract для агентов. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` `zelma help` and bare `zelma` output are agent-first.
- `EC-02` `zelma sessions help` shows subcommand map and expected output modes.

### Матрица Трассировки

| ID требования | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-01`, `EC-02` | `CHK-02` | `EVID-02` |

### Сценарии Приемки

- `SC-01` Agent reads top-level help and can identify next commands without prose parsing.
- `SC-02` Agent reads `sessions` help and can choose list/create/detect.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02` | CLI snapshot tests | expected help layout | `artifacts/ft-003/verify/chk-01/` |
| `CHK-02` | `REQ-03` | review output headings/order | command map precedes prose | `artifacts/ft-003/verify/chk-02/` |

### Доказательства

- `EVID-01` Snapshot test output.
- `EVID-02` Help contract review note.
