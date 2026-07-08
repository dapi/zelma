---
title: Domain Rules
doc_kind: domain
doc_function: canonical
purpose: Каноничное место для бизнес-правил, инвариантов, policies и rule ownership.
derived_from:
  - ../dna/governance.md
  - model.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_rules
  - domain_invariants
---

# Domain Rules

Этот документ фиксирует правила предметной области, которые обязана соблюдать любая реализация. Он не описывает UI behavior, test plan или technical exception handling, если они не являются частью business rule.

## Invariants

| Rule ID | Rule | Applies to | Why it exists | Source |
| --- | --- | --- | --- | --- |
| `DR-01` | Active `ZelmaSession` MUST include `zellij session`, `zellij pane`, `codex session` and normalized `opened path` | `ZelmaSession` | Без полного набора свойств нельзя надежно list/control сессию | Product prompt `2026-07-07` |
| `DR-02` | `.zelma/sessions.json` is the repo-local canonical registry for `zelma sessions` | `SessionRegistry` | Нужен один source of truth для CLI и skills | Product prompt `2026-07-07` |
| `DR-03` | There MUST NOT be two active records for the same `(repo root, zellij session, zellij pane)` | `SessionRegistry` | Detect должен быть идемпотентным и не плодить дубликаты | Domain decision |
| `DR-04` | `sessions detect` MUST NOT register panes without evidence that Codex is running there | Detection | Защищает non-Codex terminal work from accidental takeover | Product constraint |
| `DR-05` | `sessions list` MUST NOT create, detect, close or mutate sessions as its primary behavior | CLI commands | Inventory command должен быть predictable и safe | Domain decision |
| `DR-06` | `OpenedPath` stored in registry MUST be normalized and absolute | `OpenedPath` | Relative paths become ambiguous across shells and skills | Domain decision |
| `DR-07` | Any registry schema change MUST be versioned or migrated | `SessionRegistry` | Skills and older CLI versions need stable contracts | Product metric guardrail |
| `DR-08` | A `DetectionCandidate` without required Codex identity MUST remain non-active until resolved | Detection lifecycle | Сохраняет смысл active `zelma session` | Domain decision |
| `DR-09` | Every persisted `ZelmaSession` MUST have a positive repo-local numeric `id`, unique within `.zelma/sessions.json` | `SessionRegistry` | Пользователь и future commands должны ссылаться на короткий stable identifier | User request `2026-07-08` |

## Policies

| Policy ID | Policy | Input | Output / Verdict | Owner |
| --- | --- | --- | --- | --- |
| `POL-01` | Register create result | Requested path, created `zellij pane`, launched Codex evidence, Codex session ref | Create/update one `active` `ZelmaSession` with origin `create` | Session Registry |
| `POL-02` | Register detected pane | Zellij pane facts, Codex evidence, opened path, existing registry | Create/update one `active` `ZelmaSession` or keep `candidate`/skip | Session Registry + Detection |
| `POL-03` | Resolve duplicate candidate | Existing active record and newly observed candidate with same pane key | Update existing record, do not append duplicate | Session Registry |
| `POL-04` | Mark stale | Registry record and runtime evidence showing pane/session/codex missing | Transition active record to `stale` without deleting by default | Reconciliation |

## Cross-Context Rules

- `XDR-01` Session Registry context must treat `zellij` facts as observed
  external state, not as registry-owned state.
- `XDR-02` Skill Integration must use stable CLI output/schema rather than
  inventing a separate `.zelma/sessions.json` contract.
- `XDR-03` Codex Runtime Identification must provide enough evidence for
  `CodexSessionRef`; if it cannot, Detection must not create an active record.

## Rule Change Policy

- Если feature меняет domain invariant, обнови этот документ до или вместе с feature `brief.md` / required `design.md`.
- Если правило локально только для одной delivery-единицы, держи его в feature package, пока оно не станет shared domain rule.
- Если правило является архитектурным решением, фиксируй его в ADR и ссылайся отсюда.
