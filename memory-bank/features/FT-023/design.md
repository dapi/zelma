---
title: "FT-023: Skill Command Wrappers Design"
doc_kind: feature_design
doc_function: canonical
purpose: "Selected design for thin Codex skill wrappers over the public zelma CLI contract."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
  - ../../engineering/skill-contract.md
  - ../FT-024/brief.md
  - ../FT-025/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
---

# FT-023: Skill Command Wrappers Design

## Selected Design

Implement skill wrappers as a thin Go package at `internal/skills`. The package
executes the `zelma` binary with the documented `instances` commands and parses
only successful JSON stdout. CLI stderr is preserved on failures and mapped to a
small recovery response for agents.

The wrappers do not import registry, zellij, detection, create or CLI internals.
They depend on the command contract documented in
[`../../engineering/skill-contract.md`](../../engineering/skill-contract.md).

## Contracts

| ID | Wrapper | CLI command |
| --- | --- | --- |
| `CTR-01` | `ListSessions` | `zelma instances list --json` |
| `CTR-02` | `ListSessions` with live option | `zelma instances list --live --json` |
| `CTR-03` | `PreviewCreateSession` | `zelma instances create [path] --dry-run --json` |
| `CTR-04` | `CreateSession` | `zelma instances create [path] --json` |
| `CTR-05` | `DetectSessions` | `zelma instances detect --json` |
| `CTR-06` | `FocusSession` | `zelma instances focus <id> --json` |
| `CTR-06` | command failure | preserve exit code, stdout, stderr and agent recovery text |

## Invariants

- Wrappers call `zelma`; they do not call `zellij`.
- Wrappers parse CLI stdout; they do not read `.zelma/instances.json`.
- Wrappers keep diagnostics from stderr attached to failures.
- Wrappers reject malformed or trailing JSON on successful CLI exits.

## Failure Modes

| ID | Failure | Response |
| --- | --- | --- |
| `FM-01` | CLI exits non-zero | Return `CommandError` with exit code and stderr. |
| `FM-02` | Registry schema diagnostic | Stop and ask for valid schema v1 recovery before mutation. |
| `FM-03` | Create partial failure | Suggest uncached `zelma instances detect --json` instead of blind retry. |
| `FM-04` | Successful command emits invalid JSON | Return `DecodeError` with preserved stdout for debugging. |

## Verification

- Unit tests use a fake `zelma` binary and assert exact command invocation.
- Error tests assert diagnostics are preserved and recovery responses are
  agent-readable.
- No live zellij or registry fixtures are needed in the wrapper tests because
  FT-023 verifies the skill boundary, not CLI internals.
