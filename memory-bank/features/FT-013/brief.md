---
title: "FT-013: Codex Pane Candidate Classifier"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для classifier, который находит кандидатов Codex panes без unsafe takeover."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-003/brief.md
  - ../../domain/rules.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-013: Codex Pane Candidate Classifier

## Что

### Проблема

Manual detect должен отличать вероятные Codex panes от обычных shell panes.
Неверная классификация опасна: `zelma` может взять под контроль чужой pane.

### Результат

Classifier возвращает candidate/unknown verdict с reason codes и не создает
active session без достаточного evidence.

### Объем Работ

- `REQ-01` Классифицировать pane metadata как candidate или unknown.
- `REQ-02` Возвращать reason codes для agent review.
- `REQ-03` Быть консервативным при неполном metadata.

### Что Не Входит

- `NS-01` Нет окончательного CodexSessionRef.
- `NS-02` Нет registry write.
- `NS-03` Нет takeover panes без явного upsert workflow.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Classification policy влияет на безопасность detect. | `design.md` |

## Проверка

- `SC-01` Pane с явным Codex command получает candidate verdict.
- `SC-02` Pane с неопределенным command получает unknown, а не active.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | classifier fixture tests | expected verdicts and reasons | `artifacts/ft-013/verify/chk-01/` |
| `CHK-02` | `REQ-03` | partial metadata tests | conservative unknown verdict | `artifacts/ft-013/verify/chk-02/` |

### Доказательства

- `EVID-01` Classifier fixture test output.
- `EVID-02` Conservative unknown test output.
