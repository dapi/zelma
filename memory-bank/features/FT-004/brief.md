---
title: "FT-004: Тесты Output И Error Contract"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для contract tests, которые защищают базовый CLI output и diagnostics."
derived_from:
  - ../../product/context.md
  - ../../epics/EP-001/brief.md
  - ../../engineering/testing-policy.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-004: Тесты Output И Error Contract

## Что

### Проблема

Agent-facing CLI быстро ломается, если output/error messages меняются без
контрактных тестов. Нужно зафиксировать baseline до registry и zellij behavior.

### Результат

| ID метрики | Метрика | База | Цель | Способ измерения |
| --- | --- | --- | --- | --- |
| `MET-01` | Покрытие contract | CLI contract tests отсутствуют | root/help/stub outputs покрыты | Go tests |

### Объем Работ

- `REQ-01` Добавить tests для root и help output.
- `REQ-02` Добавить tests для command stub status и diagnostics.
- `REQ-03` Разделить stdout machine/human output и stderr diagnostics там, где применимо.

### Что Не Входит

- `NS-01` Нет tests для registry/zellij behavior.
- `NS-02` Нет final JSON schema tests для sessions.
- `NS-03` Нет настройки CI pipeline.

### Ограничения И Предположения

- `ASM-01` FT-002 и FT-003 задают CLI surfaces под тестами.
- `CON-01` Tests должны падать при случайном fallback к generic Cobra help.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Feature тестирует уже выбранные CLI surfaces и не задает новый solution contract. | `none` |

## Проверка

### Критерии Готовности

- `EC-01` Contract tests падают при изменении порядка help.
- `EC-02` Command stubs возвращают predictable status и diagnostics.

### Матрица Трассировки

| ID требования | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-02`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |

### Сценарии Приемки

- `SC-01` Агент меняет code, и test failure указывает на изменение help contract.
- `SC-02` Агент вызывает unimplemented session command и получает стабильную diagnostic shape.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02` | `go test ./...` | contract tests проходят | `artifacts/ft-004/verify/chk-01/` |
| `CHK-02` | `REQ-03` | stdout/stderr assertions | streams разделены | `artifacts/ft-004/verify/chk-02/` |

### Доказательства

- `EVID-01` Output Go tests.
- `EVID-02` Output stream assertions.
