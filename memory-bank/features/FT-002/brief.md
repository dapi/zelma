---
title: "FT-002: Дерево Команд Cobra"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для delivery slice: завести Cobra command tree для `zelma setup` и `zelma sessions` без registry/zellij behavior."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-001/brief.md
  - ../../epics/EP-001/charter.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-002: Дерево Команд Cobra

## Что

### Проблема

После Go scaffold CLI должен иметь стабильную структуру команд, чтобы
downstream features могли добавлять behavior без переименования entrypoints.
В эту структуру входит `zelma setup`, но его filesystem behavior реализуется
отдельной feature.

### Результат

| ID метрики | Метрика | База | Цель | Способ измерения |
| --- | --- | --- | --- | --- |
| `MET-01` | Маршрутизируемые команды | routed stubs отсутствуют | `setup` и `sessions list/create/detect` маршрутизируются через Cobra | CLI tests |

### Scope

- `REQ-01` Добавить root command `zelma`, command `setup` и command group `sessions`.
- `REQ-02` Добавить routed stubs для `sessions list`, `sessions create` и `sessions detect`.
- `REQ-03` Сохранить поведение без side effects: без registry writes и live zellij calls.

### Что Не Входит

- `NS-01` Нет `.zelma/sessions.json` behavior.
- `NS-02` Нет zellij integration.
- `NS-03` Нет finalized help templates за пределами route availability.
- `NS-04` Нет изменения `.gitignore`; это scope `FT-031`.

### Ограничения И Предположения

- `CON-01` Command names must match product roadmap and domain language.
- `ASM-01` FT-001 scaffold exists and builds.

## Решение О Необходимости Design

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | CLI command surface является contract для пользователя и агента. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` `zelma setup --help` и `zelma sessions list/create/detect --help` route to existing commands.
- `EC-02` Running command stubs does not touch registry or zellij.

### Матрица Трассировки

| ID требования | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |

### Сценарии Приемки

- `SC-01` Agent runs command help for `setup` and each session subcommand and receives routed output.
- `SC-02` Agent runs a stub and receives predictable non-implemented behavior with no side effects.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01` | Go CLI command tests | routes exist | `artifacts/ft-002/verify/chk-01/` |
| `CHK-02` | `EC-02` | static/search or fake adapters | no registry/zellij behavior | `artifacts/ft-002/verify/chk-02/` |

### Доказательства

- `EVID-01` Command routing test output.
- `EVID-02` Side-effect boundary check output.
