---
title: Domain Model
doc_kind: domain
doc_function: canonical
purpose: Каноничное описание ключевых domain concepts, relationships, ownership и model boundaries.
derived_from:
  - ../dna/governance.md
  - glossary.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_model
  - domain_concepts
---

# Domain Model

Этот документ описывает conceptual model предметной области. Он не должен подменять database schema, API contract или code module layout.

## Concepts

| Concept | Kind | Owns / Represents | Key relationships | Notes |
| --- | --- | --- | --- | --- |
| `RepositoryWorkspace` | aggregate | Repo root, в котором существует `.zelma/` и локальный `session registry` | Owns one `SessionRegistry`; contains many `ZelmaSession` records | Определяется из текущего пути запуска CLI |
| `SessionRegistry` | aggregate | `.zelma/sessions.json` как repo-local source of truth для known sessions | Belongs to `RepositoryWorkspace`; stores `ZelmaSession` records | Должен иметь versioned schema |
| `ZelmaSession` | entity | Управляемая запись о Codex-сессии в `zellij pane` | Has `ZelmaSessionID`; references `ZellijSessionRef`, optional `ZellijTabRef`, `ZellijPaneRef`, `CodexSessionRef`, `OpenedPath`; has `SessionOrigin` and lifecycle state | Главная domain entity |
| `ZelmaSessionID` | value object | Короткий positive integer identifier `zelma session` внутри repo-local registry | Belongs to exactly one `ZelmaSession` record in a `SessionRegistry` | Начинается с `1`; не является глобальным ID |
| `ZellijSessionRef` | value object / external ref | Идентификатор runtime `zellij session` | Parent for `ZellijTabRef` and `ZellijPaneRef` | Не принадлежит `zelma`; приходит из `zellij` |
| `ZellijTabRef` | value object / external ref | Идентификатор tab внутри `zellij session` | Groups `ZellijPaneRef` records when observed from zellij | Может отсутствовать в старых registry records |
| `ZellijPaneRef` | value object / external ref | Идентификатор pane внутри `zellij session` | Belongs to `ZellijSessionRef`; may be located inside `ZellijTabRef`; hosts Codex runtime | Может стать stale после закрытия pane |
| `CodexSessionRef` | value object / external ref | Идентификатор или ссылка на Codex session | Runs inside `ZellijPaneRef`; belongs to `ZelmaSession` record | Способ получения уточняется в implementation/ADR |
| `OpenedPath` | value object | Нормализованный абсолютный путь, открытый в pane | Used to bind session to repo/worktree context | Не должен быть относительным в registry |
| `SessionOrigin` | value object | Способ регистрации: `create` или `detect` | Attribute of `ZelmaSession` | Не равен lifecycle state |
| `DetectionCandidate` | value object | Набор evidence о pane, который может содержать Codex session | May become `ZelmaSession` if required refs resolved | Не является active session |
| `SessionCommand` | actor/action | Команда пользователя или skill, меняющая/читающая registry | Invokes registry, zellij adapter, Codex identification | `list` read-only; `create`/`detect` write |

## Relationship Map

Опиши связи на уровне бизнеса, а не таблиц базы данных.

- `RepositoryWorkspace` owns exactly one active `.zelma/sessions.json`.
- `SessionRegistry` contains zero or more `ZelmaSession` records.
- `SessionRegistry` assigns each `ZelmaSession` one positive
  `ZelmaSessionID`, unique inside that registry.
- `ZelmaSession` references exactly one `ZellijSessionRef`.
- `ZelmaSession` may store one `ZellijTabRef` and tab name when observed from
  `zellij list-panes`.
- `ZelmaSession` references exactly one `ZellijPaneRef` inside that
  `ZellijSessionRef`.
- `ZelmaSession` references exactly one `CodexSessionRef` when it is `active`.
- `ZelmaSession` stores exactly one `OpenedPath`.
- `DetectionCandidate` can become `ZelmaSession` only after required identity
  evidence is resolved.
- `SessionCommand` may read or mutate `SessionRegistry` only through documented
  CLI/domain operations.

## Concept Ownership

| Concept | Canonical owner | Allowed writers | Allowed readers | Notes |
| --- | --- | --- | --- | --- |
| `SessionRegistry` | Session Registry context | `zelma sessions create`, `zelma sessions detect`, future migrations | `zelma sessions list`, skills, diagnostics | Direct external edits are unsupported except manual recovery |
| `ZelmaSession` | Session Registry context | CLI commands and migrations | CLI, skills, docs/tests | Must follow domain invariants |
| `ZelmaSessionID` | Session Registry context | Session Registry normalization and mutating commands | CLI, skills, docs/tests | Repo-local identifier, not owned by zellij or Codex |
| `ZellijSessionRef` | Zellij Integration context | `zellij` runtime; `zelma` records observed refs | CLI, skills | External fact, not owned by registry |
| `ZellijPaneRef` | Zellij Integration context | `zellij` runtime; `zelma` records observed refs | CLI, skills | Must be revalidated before control actions |
| `CodexSessionRef` | Codex Runtime context | Codex runtime; `zelma` records identified refs | CLI, skills | Identification strategy may evolve |
| `OpenedPath` | Session Registry context | `create` from requested path; `detect` from observed pane cwd/evidence | CLI, skills | Store normalized absolute form |

## Model Boundaries

- `MB-01` Codex conversation contents are not a `zelma` domain concept.
- `MB-02` `zellij` layout, tabs and UI styling are external runtime details unless
  they affect session identity or pane control.
- `MB-03` Shell history and transient commands typed inside pane are not part of
  `ZelmaSession`.
- `MB-04` Global user-level session inventory is outside the MVP domain; first
  registry boundary is repo-local.

## Related Documents

- Бизнес-правила фиксируются в [`rules.md`](rules.md).
- Состояния и transitions фиксируются в [`states.md`](states.md).
- Bounded contexts фиксируются в [`context-map.md`](context-map.md).
