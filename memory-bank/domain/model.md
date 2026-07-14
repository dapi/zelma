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
| `RepositoryWorkspace` | aggregate | Repo root, в котором существует `.zelma/` и локальный `instance registry` | Owns one `InstanceRegistry`; contains many `ZelmaInstance` records | Определяется из текущего пути запуска CLI |
| `InstanceRegistry` | aggregate | `.zelma/instances.json` как repo-local source of truth для known instances | Belongs to `RepositoryWorkspace`; stores `ZelmaInstance` records | Должен иметь versioned schema |
| `ZelmaInstance` | entity | Управляемый экземпляр Codex runtime в `zellij pane` | Has `ZelmaInstanceID`; references `ZellijSessionRef`, optional `ZellijTabRef`, `ZellijPaneRef`, `CodexSessionRef`, `OpenedPath`; has `InstanceOrigin` and lifecycle state | Главная domain entity |
| `ZelmaInstanceID` | value object | Короткий positive integer identifier `zelma instance` внутри repo-local registry | Belongs to exactly one `ZelmaInstance` record in a `InstanceRegistry` | Начинается с `1`; не является глобальным ID |
| `ZellijSessionRef` | value object / external ref | Идентификатор runtime `zellij session` | Parent for `ZellijTabRef` and `ZellijPaneRef` | Не принадлежит `zelma`; приходит из `zellij` |
| `ZellijTabRef` | value object / external ref | Идентификатор tab внутри `zellij session` | Groups `ZellijPaneRef` records when observed from zellij | Может отсутствовать в старых registry records |
| `ZellijPaneRef` | value object / external ref | Идентификатор pane внутри `zellij session` | Belongs to `ZellijSessionRef`; may be located inside `ZellijTabRef`; hosts Codex runtime | Может стать stale после закрытия pane |
| `CodexSessionRef` | value object / external ref | Идентификатор или ссылка на Codex session | Runs inside `ZellijPaneRef`; belongs to `ZelmaInstance` record | Способ получения уточняется в implementation/ADR |
| `OpenedPath` | value object | Нормализованный абсолютный путь, открытый в pane | Used to bind instance to repo/worktree context | Не должен быть относительным в registry |
| `InstanceOrigin` | value object | Способ регистрации: `create` или `detect` | Attribute of `ZelmaInstance` | Не равен lifecycle state |
| `DetectionCandidate` | value object | Набор evidence о pane, который может содержать Codex session | May become a persisted `candidate` `ZelmaInstance` after registry identity refs are known; may become `active` after `CodexSessionRef` resolves | Не является active instance |
| `InstanceCommand` | actor/action | Команда пользователя или skill, меняющая/читающая registry | Invokes registry, zellij adapter, Codex identification | `list` read-only; `create`/`detect` write |

## Relationship Map

Опиши связи на уровне бизнеса, а не таблиц базы данных.

- `RepositoryWorkspace` owns exactly one active `.zelma/instances.json`.
- `InstanceRegistry` contains zero or more `ZelmaInstance` records.
- `InstanceRegistry` assigns each `ZelmaInstance` one positive
  `ZelmaInstanceID`, unique inside that registry.
- `ZelmaInstance` references exactly one `ZellijSessionRef`.
- `ZelmaInstance` may store one `ZellijTabRef` and tab name when observed from
  `zellij list-panes`.
- `ZelmaInstance` references exactly one `ZellijPaneRef` inside that
  `ZellijSessionRef`.
- `ZelmaInstance` references exactly one `CodexSessionRef` when it is `active`.
- `ZelmaInstance` stores exactly one `OpenedPath`.
- `DetectionCandidate` can become a persisted `candidate` `ZelmaInstance` only
  after `zellij session`, `zellij pane` and normalized `opened path` are known.
- Persisted `candidate` `ZelmaInstance` records become `active` only after
  `CodexSessionRef` is resolved.
- `InstanceCommand` may read or mutate `InstanceRegistry` only through documented
  CLI/domain operations.

## Concept Ownership

| Concept | Canonical owner | Allowed writers | Allowed readers | Notes |
| --- | --- | --- | --- | --- |
| `InstanceRegistry` | Instance Registry context | `zelma instances create`, `zelma instances detect`, future migrations | `zelma instances list`, skills, diagnostics | Direct external edits are unsupported except manual recovery |
| `ZelmaInstance` | Instance Registry context | CLI commands and migrations | CLI, skills, docs/tests | Must follow domain invariants |
| `ZelmaInstanceID` | Instance Registry context | Instance Registry normalization and mutating commands | CLI, skills, docs/tests | Repo-local identifier, not owned by zellij or Codex |
| `ZellijSessionRef` | Zellij Integration context | `zellij` runtime; `zelma` records observed refs | CLI, skills | External fact, not owned by registry |
| `ZellijPaneRef` | Zellij Integration context | `zellij` runtime; `zelma` records observed refs | CLI, skills | Must be revalidated before control actions |
| `CodexSessionRef` | Codex Runtime context | Codex runtime; `zelma` records identified refs | CLI, skills | Identification strategy may evolve |
| `OpenedPath` | Instance Registry context | `create` from requested path; `detect` from observed pane cwd/evidence | CLI, skills | Store normalized absolute form |

## Model Boundaries

- `MB-01` Codex conversation contents are not a `zelma` domain concept.
- `MB-02` `zellij` layout, tabs and UI styling are external runtime details unless
  they affect instance identity or pane control.
- `MB-03` Shell history and transient commands typed inside pane are not part of
  `ZelmaInstance`.
- `MB-04` Global user-level instance inventory is outside the MVP domain; first
  registry boundary is repo-local.

## Related Documents

- Бизнес-правила фиксируются в [`rules.md`](rules.md).
- Состояния и transitions фиксируются в [`states.md`](states.md).
- Bounded contexts фиксируются в [`context-map.md`](context-map.md).
