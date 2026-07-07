---
title: "EP-002: Brief Registry And Repo State"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для repo-local registry: `.zelma/sessions.json`, repo root state и безопасные операции чтения/записи."
derived_from:
  - ../../product/roadmap.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-002: Brief Registry And Repo State

## Проблема

Команды `list/create/detect` должны работать с одним repo-local источником
правды, но schema, repo root resolution, atomic writes и recovery rules пока не
зафиксированы в delivery artifacts.

## Результат

`.zelma/sessions.json` становится versioned registry known sessions текущего
repo, с predictable validation, duplicate prevention и безопасными writes.
`zelma setup` подготавливает repo-local state и добавляет `.zelma` в
`.gitignore`, чтобы registry не попадал в git.

## Набросок Объема

- `EP-002-REQ-01` Определить repo root resolution для всех команд.
- `EP-002-REQ-02` Зафиксировать registry schema v1.
- `EP-002-REQ-03` Реализовать atomic writes и lock для concurrent access.
- `EP-002-REQ-04` Добавить validation/recovery для поврежденного или устаревшего registry.
- `EP-002-REQ-05` Реализовать первый read surface через `sessions list`.
- `EP-002-REQ-06` Реализовать `zelma setup`, который идемпотентно добавляет `.zelma` в `.gitignore`.

## Что Не Входит

- `EP-002-NS-01` Нет live zellij discovery.
- `EP-002-NS-02` Нет создания panes.
- `EP-002-NS-03` Нет окончательной Codex session identification.

## Briefs Фич

- [FT-005: Repo Root Resolver](../../features/FT-005/README.md)
- [FT-006: Sessions Schema V1](../../features/FT-006/README.md)
- [FT-007: Atomic Registry Writes And Lock](../../features/FT-007/README.md)
- [FT-008: Registry Validation And Recovery](../../features/FT-008/README.md)
- [FT-009: Sessions List Output](../../features/FT-009/README.md)
- [FT-031: Zelma Setup Gitignore](../../features/FT-031/README.md)

## Заметки О Готовности

- Перед реализацией нужен явный schema/design layer для file format и
  migration/recovery поведения.
