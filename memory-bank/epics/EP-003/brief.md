---
title: "EP-003: Brief Zellij Read Integration And Detect"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для read-only zellij integration и обнаружения вручную созданных Codex panes."
derived_from:
  - ../../product/roadmap.md
  - ../../engineering/zellij-integration.md
  - ../../domain/model.md
  - ../../domain/rules.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-003: Brief Zellij Read Integration And Detect

## Проблема

Пользователь может вручную создать zellij pane и запустить в ней Codex. `zelma`
должна находить такие panes и заносить их в registry без unsafe takeover и
ложного присвоения чужих процессов.

## Результат

`zelma sessions detect` читает zellij sessions/panes, классифицирует кандидатов
Codex panes и идемпотентно обновляет registry.

## Набросок Объема

- `EP-003-REQ-01` Читать список zellij sessions через adapter.
- `EP-003-REQ-02` Читать panes и связанные metadata через adapter.
- `EP-003-REQ-03` Покрыть zellij JSON fixtures тестами.
- `EP-003-REQ-04` Консервативно классифицировать Codex pane candidates.
- `EP-003-REQ-05` Upsert detected sessions без duplicate records.

## Что Не Входит

- `EP-003-NS-01` Нет создания новых panes.
- `EP-003-NS-02` Нет destructive cleanup stale records.
- `EP-003-NS-03` Нет brittle parsing без fixtures/evidence.

## Briefs Фич

- [FT-010: Zellij Adapter ListSessions](../../features/FT-010/README.md)
- [FT-011: Zellij Adapter ListPanes](../../features/FT-011/README.md)
- [FT-012: Zellij JSON Fixture Tests](../../features/FT-012/README.md)
- [FT-013: Codex Pane Candidate Classifier](../../features/FT-013/README.md)
- [FT-014: Detect Upsert Idempotency](../../features/FT-014/README.md)

## Заметки О Готовности

- Нужен design для zellij CLI contract, fixture corpus и uncertainty states.
