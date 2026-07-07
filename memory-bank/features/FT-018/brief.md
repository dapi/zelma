---
title: "FT-018: Create Failure Recovery Hints"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для agent-friendly recovery hints при частичных сбоях `zelma sessions create`."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-004/brief.md
  - ../../engineering/architecture.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-018: Create Failure Recovery Hints

## Что

### Проблема

Create workflow может частично завершиться: zellij pane создана, Codex не
стартовал, registry write не прошел или confirmation не нашла pane. Агенту
нужны следующие безопасные действия.

### Результат

Ошибки create содержат stable reason codes и recovery hints: retry, run detect,
inspect zellij или fix environment.

### Объем Работ

- `REQ-01` Добавить reason codes для основных create failure modes.
- `REQ-02` Добавить recovery hints в CLI diagnostics.
- `REQ-03` Разделить retryable и non-retryable failures.

### Что Не Входит

- `NS-01` Нет автоматического destructive cleanup.
- `NS-02` Нет background retry.
- `NS-03` Нет изменения zellij adapter behavior.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Recovery hints становятся agent-facing error contract. | `design.md` |

## Проверка

- `SC-01` Missing Codex failure предлагает fix environment.
- `SC-02` Unconfirmed pane failure предлагает inspect/detect path.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | CLI error snapshot tests | reason codes and hints present | `artifacts/ft-018/verify/chk-01/` |
| `CHK-02` | `REQ-03` | failure classification tests | retryable flag correct | `artifacts/ft-018/verify/chk-02/` |

### Доказательства

- `EVID-01` Error snapshot test output.
- `EVID-02` Retryability classification test output.
