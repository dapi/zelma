---
title: "FT-026: Agent Recovery Flows"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для skill-level recovery flows поверх CLI diagnostics."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-006/brief.md
  - ../../features/FT-023/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-026: Agent Recovery Flows

## Что

### Проблема

Когда CLI возвращает ошибку, agent skill должен не просто пересказать stderr, а
предложить безопасный следующий шаг: setup, detect, retry, inspect или stop.

### Результат

Skills имеют recovery flows для common failures и сохраняют CLI reason codes.

### Объем Работ

- `REQ-01` Map CLI reason codes to agent actions.
- `REQ-02` Предлагать `zelma setup`, если repo не подготовлен.
- `REQ-03` Предлагать `instances detect`, если registry пустой, но zellij panes вероятны.

### Что Не Входит

- `NS-01` Нет автоматических destructive actions.
- `NS-02` Нет обхода CLI через direct filesystem/zellij access.
- `NS-03` Нет полной incident automation.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Recovery map является agent behavior contract. | `design.md` |

## Проверка

- `SC-01` Repo-not-ready error приводит к suggestion `zelma setup`.
- `SC-02` Zellij unavailable error приводит к stop/fix environment guidance.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | recovery map tests | expected setup suggestion | `artifacts/ft-026/verify/chk-01/` |
| `CHK-02` | `REQ-03` | recovery scenario tests | expected detect suggestion | `artifacts/ft-026/verify/chk-02/` |

### Доказательства

- `EVID-01` Recovery map test output.
- `EVID-02` Recovery scenario test output.
