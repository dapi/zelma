---
title: "FT-025: Skill Docs"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для документации Codex skills, которые управляют `zelma` sessions."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-006/brief.md
  - ../../product/context.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-025: Skill Docs

## Что

### Проблема

Агентам нужен короткий и точный contract использования skills: когда вызывать
list/create/detect, какие outputs ожидать и какие recovery paths доступны.

### Результат

Skill docs описывают commands, inputs, outputs, failure modes и boundaries без
дублирования internal implementation.

### Объем Работ

- `REQ-01` Документировать skill purpose и trigger conditions.
- `REQ-02` Документировать CLI commands, которые skill вызывает.
- `REQ-03` Документировать output/recovery expectations для агента.

### Что Не Входит

- `NS-01` Нет user marketing docs.
- `NS-02` Нет переписывания CLI help.
- `NS-03` Нет реализации wrappers.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Feature документирует уже выбранный skill/CLI contract. | `none` |

## Проверка

- `SC-01` Agent может по docs выбрать list/create/detect.
- `SC-02` Docs указывают, что skill не работает с zellij напрямую.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | docs review | commands and triggers clear | `artifacts/ft-025/verify/chk-01/` |
| `CHK-02` | `REQ-03` | recovery docs review | failure paths documented | `artifacts/ft-025/verify/chk-02/` |

### Доказательства

- `EVID-01` Skill docs review note.
- `EVID-02` Recovery docs review note.
