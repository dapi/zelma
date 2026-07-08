---
title: "EP-005: Brief Codex Session Identity"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для Codex session identity: reliable `CodexSessionRef` и evidence boundaries."
derived_from:
  - ../../product/roadmap.md
  - ../../engineering/codex-runtime-identification.md
  - ../../domain/model.md
status: active
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-005: Brief Codex Session Identity

## Проблема

`zelma session` должна знать не только zellij pane, но и Codex session. Этот
identity слой рискованный: источники metadata могут меняться, а session logs
могут содержать приватные данные.

## Результат

Registry records получают надежный `CodexSessionRef` или явно остаются в
candidate state, если evidence недостаточно.

## Набросок Объема

- `EP-005-REQ-01` Найти доступные Codex metadata sources.
- `EP-005-REQ-02` Добавить parser/process evidence extraction.
- `EP-005-REQ-03` Зафиксировать rules для candidate vs active sessions.
- `EP-005-REQ-04` Создать privacy-safe fixture corpus.

## Что Не Входит

- `EP-005-NS-01` Нет чтения/хранения полного содержимого Codex разговоров.
- `EP-005-NS-02` Нет небезопасной эвристики, которая делает active record без evidence.
- `EP-005-NS-03` Нет зависимости от private internal API без fallback.

## Briefs Фич

- [FT-019: Codex Metadata Source Discovery](../../features/FT-019/README.md)
- [FT-020: Session Evidence Parser](../../features/FT-020/README.md)
- [FT-021: Candidate Vs Active State Rules](../../features/FT-021/README.md)
- [FT-022: Privacy-Safe Evidence Fixtures](../../features/FT-022/README.md)

## Заметки О Готовности

- Требуется отдельный design/evidence document перед реализацией parser или
  privacy-sensitive fixtures.
