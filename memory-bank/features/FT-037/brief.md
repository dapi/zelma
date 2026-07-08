---
title: "FT-037: Agent Recovery E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия recovery diagnostics."
derived_from:
  - ../../use-cases/UC-005-agent-recovery.md
status: draft
delivery_status: planned
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-037: Agent Recovery E2E

## Что

### Проблема

При ошибках registry, zellij или create flow агент должен получать actionable recovery hints, а не неструктурированный текст.

### Результат

E2E suite покрывает основные failure modes и проверяет JSON diagnostics with retryability/manual-action fields.

### Объем Работ

- `REQ-01` Покрыть corrupted registry diagnostic.
- `REQ-02` Покрыть unavailable zellij diagnostic.
- `REQ-03` Покрыть partial create failure diagnostic.
- `REQ-04` Проверить suggested next command в JSON.

### Что Не Входит

- `NS-01` Нет destructive auto-repair без confirm.

## Проверка

### Критерии Готовности

- `EC-01` Каждая ошибка имеет machine-readable code.
- `EC-02` Retryable/manual action отличимы.
- `EC-03` Recovery command не теряет registry data.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-тестов основных классов recovery.
