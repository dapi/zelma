---
title: "FT-027: Sessions List Live"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для `sessions list --live`, который сверяет registry с текущим zellij state."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-007/brief.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-027: Sessions List Live

## Что

### Проблема

Обычный `sessions list` показывает known sessions из registry, но не отвечает,
существуют ли соответствующие zellij panes прямо сейчас.

### Результат

`sessions list --live` дополняет registry records live status из zellij без
разрушительных изменений registry.

### Объем Работ

- `REQ-01` Сверить registry records с zellij sessions/panes.
- `REQ-02` Показать live/unreachable status в human и JSON output.
- `REQ-03` Не удалять stale records автоматически.

### Что Не Входит

- `NS-01` Нет cleanup/remove.
- `NS-02` Нет background watcher.
- `NS-03` Нет create/detect behavior.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Live view объединяет registry и zellij integration contracts. | `design.md` |

## Проверка

- `SC-01` Existing pane получает live status.
- `SC-02` Missing pane показывается как unreachable без удаления record.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | list tests with fake zellij | live status shown | `artifacts/ft-027/verify/chk-01/` |
| `CHK-02` | `REQ-03` | stale fixture tests | no registry deletion | `artifacts/ft-027/verify/chk-02/` |

### Доказательства

- `EVID-01` Live list test output.
- `EVID-02` No-delete stale test output.
