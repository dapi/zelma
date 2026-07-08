---
title: "FT-023: Skill Command Wrappers"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для Codex skill wrappers, которые вызывают `zelma` CLI вместо дублирования logic."
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

# FT-023: Skill Command Wrappers

## Что

### Проблема

Skills должны помогать агентам управлять sessions, но если они будут напрямую
читать registry или вызывать zellij, появится второй implementation path.

### Результат

Codex skills вызывают `zelma sessions list/create/detect/focus` и интерпретируют
стабильный CLI contract.

### Объем Работ

- `REQ-01` Создать wrappers для list/create/detect/focus.
- `REQ-02` Обрабатывать exit codes и diagnostics CLI.
- `REQ-03` Не дублировать registry или zellij logic внутри skills.

### Что Не Входит

- `NS-01` Нет прямых zellij calls из skills.
- `NS-02` Нет отдельного parser `.zelma/sessions.json`.
- `NS-03` Нет нового command surface в обход `zelma`.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Skill wrappers являются agent integration contract. | `design.md` |

## Проверка

- `SC-01` Skill wrapper вызывает `zelma sessions list`.
- `SC-02` CLI error превращается в agent-readable recovery response.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-03` | wrapper tests with fake CLI | expected command invocation | `artifacts/ft-023/verify/chk-01/` |
| `CHK-02` | `REQ-02` | fake CLI error tests | diagnostics preserved | `artifacts/ft-023/verify/chk-02/` |

### Доказательства

- `EVID-01` Wrapper invocation test output.
- `EVID-02` Wrapper error handling test output.
