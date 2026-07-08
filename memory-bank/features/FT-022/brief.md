---
title: "FT-022: Privacy-Safe Evidence Fixtures"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для fixture corpus, который покрывает Codex evidence без приватного conversation content."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-005/brief.md
  - ../../features/FT-019/brief.md
  - ../../features/FT-020/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-022: Privacy-Safe Evidence Fixtures

## Что

### Проблема

Parser и state rules требуют fixtures, но реальные Codex logs могут содержать
приватные запросы, код и контекст пользователя.

### Результат

Fixture corpus содержит synthetic/redacted evidence, достаточный для tests, и
не содержит приватных conversation payloads.

### Объем Работ

- `REQ-01` Создать synthetic fixtures для valid evidence.
- `REQ-02` Создать partial/invalid fixtures.
- `REQ-03` Добавить privacy scan/review для fixture corpus.

### Что Не Входит

- `NS-01` Нет включения реальных conversations.
- `NS-02` Нет хранения secrets, tokens или repo-private paths.
- `NS-03` Нет расширения parser behavior.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Feature создает test corpus по уже выбранным privacy boundaries. | `none` |

## Проверка

- `SC-01` Valid synthetic fixtures проходят parser tests.
- `SC-02` Privacy scan не находит conversation content/secrets.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | parser fixture tests | fixtures exercise expected cases | `artifacts/ft-022/verify/chk-01/` |
| `CHK-02` | `REQ-03` | privacy scan/review | no private content | `artifacts/ft-022/verify/chk-02/` |

### Доказательства

- `EVID-01` Fixture parser test output.
- `EVID-02` Privacy scan/review output.
