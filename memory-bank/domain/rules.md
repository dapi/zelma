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
| `DR-01` | Active `ZelmaInstance` MUST include `zellij session`, `zellij pane`, `codex session` and normalized `opened path` | `ZelmaInstance` | Без полного набора свойств нельзя надежно list/control instance | Product prompt `2026-07-07` |
| `DR-02` | `.zelma/instances.json` is the repo-local canonical registry for `zelma instances` | `InstanceRegistry` | Нужен один source of truth для CLI и skills | Product prompt `2026-07-07` |
| `DR-03` | There MUST NOT be two active records for the same `(repo root, zellij session, zellij pane)` | `InstanceRegistry` | Detect должен быть идемпотентным и не плодить дубликаты | Domain decision |
| `DR-04` | `instances detect` MUST NOT register panes without evidence that Codex is running there | Detection | Защищает non-Codex terminal work from accidental takeover | Product constraint |
| `DR-05` | `instances list` MAY run bounded auto-detect and mutate registry records before rendering inventory; callers that require a registry-only read MUST use `instances list --no-detect` | CLI commands | Inventory command должен быть ergonomic by default while preserving an explicit predictable read-only path | GitHub issue #86 |
| `DR-06` | `OpenedPath` stored in registry MUST be normalized and absolute | `OpenedPath` | Relative paths become ambiguous across shells and skills | Domain decision |
| `DR-07` | Any registry schema change MUST be versioned or migrated | `InstanceRegistry` | Skills and older CLI versions need stable contracts | Product metric guardrail |
| `DR-08` | A `DetectionCandidate` without required Codex identity MUST remain non-active until resolved | Detection lifecycle | Сохраняет смысл active `zelma instance` | Domain decision |
| `DR-09` | Every persisted `ZelmaInstance` MUST have a positive repo-local numeric `id`, unique within `.zelma/instances.json` | `InstanceRegistry` | Пользователь и commands должны ссылаться на короткий stable identifier | User request `2026-07-08` |

## Policies

| Policy ID | Policy | Input | Output / Verdict | Owner |
| --- | --- | --- | --- | --- |
| `POL-01` | Register create result | Requested path, created `zellij pane`, launched Codex evidence, Codex session ref | Create/update one `active` `ZelmaInstance` with origin `create` | Instance Registry |
| `POL-02` | Register detected pane | Zellij pane facts, Codex evidence, opened path, existing registry | Create/update one `active` `ZelmaInstance` or keep `candidate`/skip | Instance Registry + Detection |
| `POL-03` | Resolve duplicate candidate | Existing active record and newly observed candidate with same pane key | Update existing record, do not append duplicate | Instance Registry |
| `POL-04` | Mark stale | Registry record and runtime evidence showing pane/session/codex missing | Transition active record to `stale` without deleting by default | Reconciliation |

## Cross-Context Rules

- `XDR-01` Instance Registry context must treat `zellij` facts as observed
  external state, not as registry-owned state.
- `XDR-02` Skill Integration must use stable CLI output/schema rather than
  inventing a separate `.zelma/instances.json` contract.
- `XDR-03` Codex Runtime Identification must provide enough evidence for
  `CodexSessionRef`; if it cannot, Detection must not create an active record.

## Rule Change Policy

- Если feature меняет domain invariant, обнови этот документ до или вместе с feature `brief.md` / required `design.md`.
- Если правило локально только для одной delivery-единицы, держи его в feature package, пока оно не станет shared domain rule.
- Если правило является архитектурным решением, фиксируй его в ADR и ссылайся отсюда.
