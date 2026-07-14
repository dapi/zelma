---
title: "FT-006: Sessions Schema V1"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для versioned schema `.zelma/instances.json`."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-002/brief.md
  - ../../domain/model.md
  - ../../domain/rules.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-006: Sessions Schema V1

## Что

### Проблема

Registry должен хранить zellij session, zellij pane, Codex session и opened path
в устойчивом формате. Без schema v1 downstream команды будут по-разному
интерпретировать один и тот же файл.

### Результат

Есть versioned JSON schema v1 для `.zelma/instances.json` и fixtures, по которым
можно проверять чтение/запись registry.

### Объем Работ

- `REQ-01` Определить top-level version и instances collection.
- `REQ-02` Определить поля session record: zellij session, zellij pane, Codex session ref, opened path и state.
- `REQ-03` Добавить fixtures для empty, minimal и representative registry.

### Что Не Входит

- `NS-01` Нет atomic write implementation.
- `NS-02` Нет live zellij reconciliation.
- `NS-03` Нет migration между версиями старше v1.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Schema является file format contract. | `design.md` |

## Проверка

- `SC-01` Пустой registry v1 парсится как пустой список sessions.
- `SC-02` Representative registry сохраняет все обязательные references.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | schema/golden fixture tests | valid v1 fixtures pass | `artifacts/ft-006/verify/chk-01/` |
| `CHK-02` | `REQ-03` | fixture review | fixtures cover empty/minimal/representative | `artifacts/ft-006/verify/chk-02/` |

### Доказательства

- `EVID-01` Schema/fixture test output.
- `EVID-02` Fixture inventory.
