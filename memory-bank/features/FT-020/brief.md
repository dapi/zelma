---
title: "FT-020: Session Evidence Parser"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для parser, который извлекает privacy-safe evidence для CodexSessionRef."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-005/brief.md
  - ../../features/FT-019/brief.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-020: Session Evidence Parser

## Что

### Проблема

Даже если metadata source найден, parser должен извлекать только минимальные
identity поля и устойчиво обрабатывать отсутствующие/неполные данные.

### Результат

Parser возвращает `CodexSessionRef` или explicit insufficient-evidence verdict,
не сохраняя приватный content.

### Объем Работ

- `REQ-01` Извлекать identity metadata из выбранных sources.
- `REQ-02` Возвращать confidence/insufficient evidence.
- `REQ-03` Исключить хранение conversation content.

### Что Не Входит

- `NS-01` Нет выбора новых metadata sources.
- `NS-02` Нет записи registry.
- `NS-03` Нет чтения приватного content сверх минимального metadata.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Parser contract влияет на privacy, identity и state transitions. | `design.md` |

## Проверка

- `SC-01` Valid evidence возвращает CodexSessionRef.
- `SC-02` Partial evidence возвращает insufficient evidence.
- `SC-03` Content fields не появляются в output.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | parser fixture tests | expected ref/verdict | `artifacts/ft-020/verify/chk-01/` |
| `CHK-02` | `REQ-03` | output/content scan tests | no private content stored | `artifacts/ft-020/verify/chk-02/` |

### Доказательства

- `EVID-01` Parser fixture test output.
- `EVID-02` Privacy scan test output.
