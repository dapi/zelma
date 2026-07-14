---
title: "FT-047: Focus Zelma Session By ID"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для фокусировки zellij pane по numeric zelma instance id."
derived_from:
  - ../../product/roadmap.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../FT-009/brief.md
  - ../FT-027/brief.md
  - ../FT-045/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-047: Focus Zelma Session By ID

## Что

### Проблема

После появления numeric `ZelmaInstanceID` пользователь видит короткий ID в
`instances list`, но все еще не может использовать его для перехода к нужной
работе в `zellij`.

### Результат

`zelma instances focus <id>` фокусирует zellij tab/pane, сохраненные в registry
для указанной `zelma instance`.

### Объем Работ

- `REQ-01` Добавить CLI command `zelma instances focus <id>`.
- `REQ-02` Lookup выполняется по repo-local positive numeric `id` из
  `.zelma/instances.json`.
- `REQ-03` Если registry record содержит `zellij_tab`, команда сначала
  переключает zellij на этот tab через stable tab id.
- `REQ-04` Команда затем фокусирует `zellij_pane`.
- `REQ-05` Команда не создает, не detect-ит, не cleanup-ит и не мутирует
  registry records.
- `REQ-06` Поддержать `--json` с focused session JSON object.

### Что Не Входит

- `NS-01` Нет fuzzy matching по path, pane или Codex session.
- `NS-02` Нет автоматического stale cleanup или registry repair.
- `NS-03` Нет глобальной фокусировки между разными repo registries.

## Проверка

### Критерии Готовности

- `EC-01` `instances focus <id>` вызывает zellij tab focus перед pane focus, если
  `zellij_tab` известен.
- `EC-02` Registry file не меняется после focus command.
- `EC-03` Invalid или missing ID возвращает agent-readable diagnostic.
- `EC-04` Zellij command failures surface through existing zellij diagnostics.

### Обязательное Покрытие

- Unit tests для zellij focus adapter command args and validation.
- CLI tests для success, JSON output and invalid/not-found IDs.
