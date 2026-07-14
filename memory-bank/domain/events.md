---
title: Domain Events
doc_kind: domain
doc_function: canonical
purpose: Каноничное место для domain events как бизнес-значимых фактов, их meaning, producers, consumers и минимального payload contract.
derived_from:
  - ../dna/governance.md
  - model.md
  - rules.md
status: active
audience: humans_and_agents
canonical_for:
  - domain_events
  - business_events
---

# Domain Events

Этот документ описывает события, которые являются значимыми фактами предметной области. Technical logs, analytics events и infrastructure messages живут в engineering/ops/product docs, если у них нет domain meaning.

## Events

| Event ID | Event | Meaning | Producer | Consumers | Minimal facts |
| --- | --- | --- | --- | --- | --- |
| `DE-01` | `ZelmaInstanceCreated` | `instances create` produced a live Codex pane and registered it | CLI / Instance Registry | CLI output, skills, logs/tests | Instance id, zellij session, zellij pane, codex session, opened path, origin `create` |
| `DE-02` | `CodexPaneDetected` | Detection found a zellij pane with evidence of Codex runtime through explicit `instances detect` or default `instances list` auto-detect | Detection | Instance Registry | Zellij session, zellij pane, opened path evidence, Codex evidence |
| `DE-03` | `ZelmaInstanceRegistered` | Registry now contains an instance record created from create or detect workflow | Instance Registry | CLI output, skills, tests | Registry path, instance id, state, origin |
| `DE-04` | `ZelmaInstanceBecameStale` | Previously active record no longer validates against runtime state | Reconciliation | CLI output, skills, cleanup workflow | Instance id, previous refs, observed missing evidence |
| `DE-05` | `ZelmaInstanceRevalidated` | Stale or uncertain record was confirmed live again | Reconciliation | CLI output, skills | Instance id, confirmed refs |
| `DE-06` | `ZelmaInstanceClosed` | User intentionally closed or removed a managed instance record | Future close/remove command | CLI output, registry, skills | Instance id, closure reason, timestamp if available |
| `DE-07` | `InstanceRegistryUpdated` | `.zelma/instances.json` was written with a valid schema | Instance Registry | CLI output, tests, diagnostics | Registry path, schema version, affected instance ids |

## Event Rules

- Событие называется в прошедшем времени или как факт, который уже произошел.
- Событие не должно означать command или request.
- Если event меняет allowed state transitions, обнови [`states.md`](states.md).
- Если event переносит responsibility между contexts, обнови [`context-map.md`](context-map.md).

## Delivery Semantics

- Duplicate detection-related events must be harmless: applying the same
  observed pane twice must not create duplicate active records.
- Registry update events are meaningful only after the JSON write has succeeded
  and validation passed.
- Ordering matters for create/register flows: `ZelmaInstanceCreated` precedes
  `ZelmaInstanceRegistered` conceptually, even if implementation logs a single
  combined operation.
- Technical retry, queue, lock и error handling rules фиксируй в
  [`../engineering/architecture.md`](../engineering/architecture.md).
