---
title: "FT-017: Create Confirmation And Reconciliation"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для подтверждения created zellij pane и записи active registry record."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-004/brief.md
  - ../../epics/EP-002/brief.md
  - ../../epics/EP-003/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-017: Create Confirmation And Reconciliation

## Что

### Проблема

После zellij create нельзя сразу считать session active: pane могла не
создаться, Codex мог не стартовать, а registry write может быть преждевременным.

### Результат

`sessions create` подтверждает наличие pane и enough launch evidence перед
созданием active registry record.

### Объем Работ

- `REQ-01` Проверить созданную pane через zellij read adapter.
- `REQ-02` Сопоставить pane с launch request.
- `REQ-03` Записать active/candidate registry record только после confirmation.
- `REQ-04` Вернуть summary created/registered/skipped.

### Что Не Входит

- `NS-01` Нет окончательного CodexSessionRef, если identity недоступна.
- `NS-02` Нет cleanup pane при registry failure без отдельного design.
- `NS-03` Нет stale lifecycle cleanup.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Confirmation связывает zellij state и registry writes. | `design.md` |

## Проверка

- `SC-01` Created pane подтверждается и registry получает record.
- `SC-02` Если pane не подтверждена, registry не меняется.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-03` | create workflow tests with fake zellij | confirmed record written | `artifacts/ft-017/verify/chk-01/` |
| `CHK-02` | `REQ-02`, `REQ-04` | failed confirmation tests | no write; clear summary | `artifacts/ft-017/verify/chk-02/` |

### Доказательства

- `EVID-01` Confirmed create test output.
- `EVID-02` Failed confirmation test output.
