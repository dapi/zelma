---
title: "EP-007: Brief Reconciliation And Lifecycle"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для live/stale lifecycle reconciliation без разрушительных действий по умолчанию."
derived_from:
  - ../../product/roadmap.md
  - ../../domain/states.md
  - ../../domain/rules.md
  - ../../engineering/zellij-integration.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-007: Brief Reconciliation And Lifecycle

## Проблема

Registry может устаревать: zellij panes закрываются, переезжают или становятся
недоступны. `zelma` должна показывать это состояние без автоматического
удаления полезных records.

## Результат

CLI различает registered/live/stale states, показывает live status и предлагает
cleanup как явное действие, а не destructive default.

## Набросок Объема

- `EP-007-REQ-01` Добавить live reconciliation view для list.
- `EP-007-REQ-02` Определить stale detection rules.
- `EP-007-REQ-03` Добавить explicit cleanup/remove proposal path.
- `EP-007-REQ-04` Зафиксировать lifecycle state tests.

## Что Не Входит

- `EP-007-NS-01` Нет background watcher/daemon.
- `EP-007-NS-02` Нет destructive cleanup без подтверждения.
- `EP-007-NS-03` Нет cross-repo global registry.

## Briefs Фич

- [FT-027: Sessions List Live](../../features/FT-027/README.md)
- [FT-028: Stale Detection](../../features/FT-028/README.md)
- [FT-029: Cleanup Remove Proposal](../../features/FT-029/README.md)
- [FT-030: Lifecycle State Tests](../../features/FT-030/README.md)

## Заметки О Готовности

- Требуется согласование lifecycle states с registry schema и zellij read adapter.
