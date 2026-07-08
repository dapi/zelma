---
title: "FT-011: Zellij Adapter ListPanes"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для adapter method, который читает panes и metadata zellij session."
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

# FT-011: Zellij Adapter ListPanes

## Что

### Проблема

Для detect недостаточно знать zellij sessions; нужно получить panes, их ids,
команды и рабочие каталоги, если zellij CLI предоставляет эти данные.

### Результат

Adapter читает panes выбранной zellij session и возвращает нормализованные pane
records для classifier и detect workflow.

### Объем Работ

- `REQ-01` Добавить adapter method для чтения panes zellij session.
- `REQ-02` Нормализовать pane id, command/process metadata и path, если доступны.
- `REQ-03` Обработать отсутствующие/частичные metadata без panic.

### Что Не Входит

- `NS-01` Нет classifier логики Codex panes.
- `NS-02` Нет registry upsert.
- `NS-03` Нет focus/attach управления zellij panes.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Pane metadata contract влияет на detect и create confirmation. | `design.md` |

## Проверка

- `SC-01` Adapter парсит fixture с несколькими panes.
- `SC-02` Partial pane metadata сохраняет record в uncertain state, а не падает.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | adapter fixture tests | panes parsed | `artifacts/ft-011/verify/chk-01/` |
| `CHK-02` | `REQ-03` | partial fixture tests | no panic; explicit missing fields | `artifacts/ft-011/verify/chk-02/` |

### Доказательства

- `EVID-01` Pane parsing test output.
- `EVID-02` Partial metadata test output.
