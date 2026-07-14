---
title: "FT-028: Stale Detection Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for stale detection rules and reason-code output."
derived_from:
  - brief.md
  - ../../domain/states.md
  - ../../domain/rules.md
status: active
audience: humans_and_agents
---

# FT-028: Stale Detection Design

## Selected Design

`zelma instances detect` uses the same successful zellij inventory pass for
candidate detection and stale reconciliation. It records observed zellij
sessions and pane keys while reading `list-sessions` and `list-panes`. After
detected pane upsert, it compares existing `active` registry records with that
snapshot:

- missing zellij session returns reason `missing_zellij_session`;
- present zellij session with missing pane key returns reason `missing_pane`;
- live pane keys are not stale, even when they are not detected as Codex panes.

Stale detection is gated on a complete successful inventory pass. Any zellij
adapter error returns the original command diagnostic and skips registry writes,
so transient zellij failures do not convert records to `stale`.

## Contracts

| Contract ID | Contract | Owner |
| --- | --- | --- |
| `CTR-01` | Only `active` records are transitioned to `stale` by FT-028. | Session Registry |
| `CTR-02` | Stale candidates include a machine-readable reason code. | CLI |
| `CTR-03` | Zellij inventory errors preserve the existing registry without stale transitions. | Detection |

## Invariants

- `INV-01` FT-028 does not delete stale records.
- `INV-02` Unresolved `candidate` records are not transitioned to `stale`
  because the current registry schema requires identity fields for `stale`.
- `INV-03` Reason codes are returned in detect output and are not stored as a
  registry schema change.

## Verification

- `CHK-01`: Go registry and CLI tests cover missing pane stale reasons.
- `CHK-02`: Go CLI tests cover zellij command failure without registry
  mutation.
