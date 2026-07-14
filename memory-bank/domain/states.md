---
title: Domain States
doc_kind: domain
doc_function: canonical
purpose: Каноничное место для lifecycle states, allowed transitions, terminal states и state-related invariants.
derived_from:
  - ../dna/governance.md
  - model.md
  - rules.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_states
  - state_transitions
---

# Domain States

Этот документ описывает состояния domain concepts и допустимые transitions. Он не должен превращаться в UI state или implementation state machine, если эти состояния не имеют бизнес-смысла.

## State Machines

| State Machine | Concept | Owner | Notes |
| --- | --- | --- | --- |
| `SM-01` | `ZelmaInstance` | Instance Registry | Covers create, detect, stale handling and close/removal semantics |
| `SM-02` | `DetectionCandidate` | Detection | Covers candidate evidence before active registration |

## States

| State | Meaning | Entry condition | Exit condition | Terminal |
| --- | --- | --- | --- | --- |
| `candidate` | Pane may contain a Codex session, but required identity is incomplete | `instances detect` sees partial evidence but cannot satisfy active invariants | Evidence resolves to active, or candidate is discarded | no |
| `active` | Registry record represents a live, manageable Codex instance in `zellij` | `instances create` succeeds or `instances detect` resolves all required refs | Runtime validation fails, user closes/removes instance, or migration changes state | no |
| `stale` | Registry record no longer matches live runtime state | Reconciliation/list/detect observes missing pane, missing zellij session or missing Codex runtime | Instance reappears/revalidates, user removes, or record is archived | no |
| `closed` | User intentionally ended or removed the managed instance record | Explicit future close/remove command or accepted cleanup | None for same record | yes |
| `archived` | Historical record kept for audit/migration, not shown as active inventory by default | Future retention policy | None for same record | yes |

## Transitions

| Transition ID | From | To | Trigger | Preconditions | Forbidden when |
| --- | --- | --- | --- | --- | --- |
| `TR-01` | none | `active` | `zelma instances create` | Pane created, Codex launched, all required refs resolved, registry write succeeds | Required refs missing |
| `TR-02` | none | `candidate` | `zelma instances detect` | Pane has partial Codex evidence but not enough for active record | Pane clearly not running Codex |
| `TR-03` | `candidate` | `active` | Detection evidence resolved | `zellij session`, `zellij pane`, `codex session`, `opened path` all known | Duplicate active record would be created |
| `TR-04` | `active` | `stale` | Runtime reconciliation | Existing record cannot be verified against live `zellij`/Codex state | Evidence is inconclusive and command is read-only without stale policy |
| `TR-05` | `stale` | `active` | Runtime revalidation | Same pane/session/codex identity is confirmed again | Identity changed and would violate duplicate rules |
| `TR-06` | `active` | `closed` | Future explicit close/remove | User or skill requests close/remove through supported command | Command lacks explicit user intent |
| `TR-07` | `stale` | `closed` | Future cleanup/remove | User accepts cleanup of stale record | Record may still be live or evidence is ambiguous |
| `TR-08` | `closed` | `archived` | Future retention policy | Historical retention is enabled | Project chooses hard delete instead |

## State Invariants

- `SI-01` `active` records must satisfy [`rules.md`](rules.md) `DR-01`.
- `SI-02` `candidate` records must not appear as active instances in default
  `instances list` output.
- `SI-03` `stale` records must not be used for destructive pane control without
  revalidation.
- `SI-04` Terminal states must not be silently reactivated; create a new record
  or run explicit restore/migration if that behavior is added later.

## Implementation Notes

Если runtime implementation использует дополнительные technical states, документируй их в code/API docs или [`../engineering/architecture.md`](../engineering/architecture.md), а здесь оставляй только business-visible states.
