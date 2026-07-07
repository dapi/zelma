---
title: "EP-004: Brief Managed Create Workflow"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для managed `sessions create`: zellij pane, Codex launch и registry record."
derived_from:
  - ../../product/roadmap.md
  - ../../engineering/zellij-integration.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-004: Brief Managed Create Workflow

## Проблема

Пользователю нужна команда, которая создает новую zelma session без ручной
настройки zellij pane и без ручной записи `.zelma/sessions.json`.

## Результат

`zelma sessions create` создает zellij pane, запускает Codex в нужном path,
получает достаточно evidence для подтверждения и записывает active record в
registry.

## Набросок Объема

- `EP-004-REQ-01` Зафиксировать Codex launch contract.
- `EP-004-REQ-02` Реализовать zellij run/new-pane adapter.
- `EP-004-REQ-03` Подтверждать созданную pane перед registry write.
- `EP-004-REQ-04` Давать agent-friendly recovery hints при частичных сбоях.

## Что Не Входит

- `EP-004-NS-01` Нет глобального daemon/watch mode.
- `EP-004-NS-02` Нет automatic cleanup без явного lifecycle feature.
- `EP-004-NS-03` Нет UI поверх zellij.

## Briefs Фич

- [FT-015: Codex Launch Contract](../../features/FT-015/README.md)
- [FT-016: Zellij Run New-Pane Adapter](../../features/FT-016/README.md)
- [FT-017: Create Confirmation And Reconciliation](../../features/FT-017/README.md)
- [FT-018: Create Failure Recovery Hints](../../features/FT-018/README.md)

## Заметки О Готовности

- Create path требует design по failure modes: pane created but Codex failed,
  Codex running but registry write failed, existing matching pane found.
