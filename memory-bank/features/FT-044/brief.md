---
title: "FT-044: Detect Evidence Explain And Indexed Lookup"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для ускорения `instances detect` и объяснения evidence decisions."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-005/brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../FT-021/brief.md
  - ../FT-043/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-044: Detect Evidence Explain And Indexed Lookup

## Что

### Проблема

`instances detect` может быть медленным и непрозрачным: каждый unresolved
candidate заново сканирует `$CODEX_HOME/sessions`, а пользователь не видит,
почему pane стала `active` или осталась `candidate`.

### Результат

`instances detect` строит индекс Codex session evidence один раз за запуск и
умеет показывать per-candidate explanation через `--explain`.

### Объем Работ

- `REQ-01` Построить `session_meta` evidence index один раз за detect/create
  enrichment run.
- `REQ-02` Не строить index, если все candidates уже имеют `codex_session` из
  сильного evidence, например argv.
- `REQ-03` Добавить `instances detect --explain` с evidence verdict/source/reason
  по каждому detected candidate.
- `REQ-04` Поддержать `instances detect --json --explain` через optional
  `candidate_explanations`.
- `REQ-05` Не менять default text output и default JSON output без `--explain`.

### Что Не Входит

- `NS-01` Нет кеша между запусками CLI.
- `NS-02` Нет фильтрации detect по одной zellij session.
- `NS-03` Нет ручной disambiguation-команды.

## Проверка

### Критерии Готовности

- `EC-01` Existing detect JSON remains backward compatible without
  `--explain`.
- `EC-02` `--explain` показывает resolved/insufficient evidence по candidates.
- `EC-03` Session metadata lookup preserves ambiguity rules.

### Обязательное Покрытие

- Unit tests для `SessionEvidenceIndex`.
- CLI tests для text and JSON `--explain`.
- Existing detect/create tests remain green.
