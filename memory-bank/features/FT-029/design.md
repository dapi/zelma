---
title: "FT-029: Cleanup Remove Proposal Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for explicit stale-record cleanup without destructive default."
derived_from:
  - brief.md
  - ../FT-027/design.md
  - ../FT-028/design.md
status: active
audience: humans_and_agents
---

# FT-029: Cleanup Remove Proposal Design

## Selected Design

`zelma instances cleanup` is an explicit cleanup path for stale registry records.
The command reads `.zelma/instances.json` and builds a proposal from records whose
registry `state` is exactly `stale`.

By default, the command is read-only:

- it prints `proposed`, `removed` and `kept` counts;
- it prints each proposed stale record with zellij, Codex and opened-path
  identity;
- it does not contact zellij;
- it does not write `.zelma/instances.json`.

`zelma instances cleanup --confirm` applies the same stale-only selection while
holding the registry write lock. It removes only records still marked `stale` at
write time. Records in `active`, `candidate`, `closed` or `archived` state are
never removed by this command.

## Contracts

| Contract ID | Contract | Owner |
| --- | --- | --- |
| `CTR-01` | Cleanup proposal is the default behavior and is read-only. | CLI |
| `CTR-02` | Cleanup apply requires explicit `--confirm`. | CLI |
| `CTR-03` | Cleanup apply removes only `stale` records. | Session Registry |
| `CTR-04` | Cleanup output includes an audit-friendly summary and stale record identities. | CLI |

## Non-Goals

- No automatic cleanup from `instances detect` or `instances list --live`.
- No deletion of live or unresolved candidate records.
- No global registry cleanup across repositories.
- No registry schema change for cleanup reasons or audit history.

## Verification

- `CHK-01`: CLI tests cover default proposal output with stale record identity.
- `CHK-02`: CLI tests cover no-confirm registry immutability.
- `CHK-03`: CLI and registry tests cover `--confirm` removing only stale
  records.
