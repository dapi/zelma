---
title: "FT-002: Дерево Команд Cobra"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для delivery slice: завести Cobra command tree для `zelma setup` и `zelma sessions` без registry/zellij behavior."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/roadmap.md
  - ../../epics/EP-001/brief.md
  - ../../epics/EP-001/charter.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
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

### Объем Работ

- `REQ-01` Добавить root command `zelma`, command `setup` и command group `sessions`.
- `REQ-02` Добавить routed stubs для `sessions list`, `sessions create` и `sessions detect`.
- `REQ-03` Сохранить поведение без side effects: без registry writes и live zellij calls.

### Что Не Входит

- `NS-01` Нет `.zelma/sessions.json` behavior.
- `NS-02` Нет zellij integration.
- `NS-03` Нет finalized help templates за пределами route availability.
- `NS-04` Нет изменения `.gitignore`; это scope [FT-031](../FT-031/README.md).

### Ограничения И Предположения

- `CON-01` Имена команд должны соответствовать product roadmap и domain language.
- `ASM-01` FT-001 scaffold существует и собирается.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | CLI command surface является contract для пользователя и агента. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` `zelma setup --help` и `zelma sessions list/create/detect --help` маршрутизируются в существующие команды.
- `EC-02` Запуск command stubs не трогает registry или zellij.

### Матрица Трассировки

| ID требования | Ссылки на проблему | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- | --- |
| `REQ-01` | `CON-01`, `ASM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `CON-01`, `ASM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `ASM-01` | `EC-02`, `SC-02`, `NEG-01` | `CHK-02` | `EVID-02` |

### Сценарии Приемки

- `SC-01` Агент запускает command help для `setup` и каждого session subcommand и получает routed output.
- `SC-02` Агент запускает stub и получает predictable non-implemented behavior без side effects.

### Негативные / Edge Сценарии

- `NEG-01` Stub-команды не создают, не читают и не изменяют `.zelma/sessions.json`
  и не вызывают live `zellij`.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Go CLI command tests | routes существуют | `artifacts/ft-002/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02`, `NEG-01` | static/search или fake adapters | нет registry/zellij behavior | `artifacts/ft-002/verify/chk-02/` |

### Матрица Тестов

| ID проверки | ID доказательств | Путь доказательств |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-002/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-002/verify/chk-02/` |

### Доказательства

- `EVID-01` Output тестов command routing.
- `EVID-02` Output проверки side-effect boundary.

### Контракт Доказательств

| ID доказательства | Artifact | Producer | Path contract | Используется проверками |
| --- | --- | --- | --- | --- |
| `EVID-01` | Test output для Cobra routing | implementer | `artifacts/ft-002/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Test output или review note для side-effect boundary | implementer / reviewer | `artifacts/ft-002/verify/chk-02/` | `CHK-02` |
