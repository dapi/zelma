---
title: "FT-010: Zellij Adapter ListSessions Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for the read-only zellij ListSessions adapter method and diagnostic mapping."
derived_from:
  - brief.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
  - ../../engineering/architecture.md
  - ../../engineering/zellij-integration.md
status: active
audience: humans_and_agents
---

# FT-010: Zellij Adapter ListSessions Design

## Selected Design

`internal/zellij` owns zellij CLI access. `Client.ListSessions(ctx)` runs:

```bash
zellij list-sessions --short --no-formatting
```

The method returns project-owned `zellij.Session` records with a normalized
`Name` field. It does not read panes, inspect Codex, mutate registry state or
call command-layer code.

## Contracts

| Contract ID | Input / Output | Semantics |
| --- | --- | --- |
| `CTR-01` | zellij list output -> `[]Session` | One UTF-8 session name per line, preserved verbatim; empty output means no sessions. Zellij `0.44.x` empty-inventory stderr `No active zellij sessions found.` with exit 1 also maps to an empty slice. |
| `CTR-02` | missing zellij binary -> diagnostic error | Code is `zellij_missing_binary` with a recovery hint to install or configure zellij. |
| `CTR-03` | non-zero command result -> diagnostic error | Code is `zellij_command_failed`; stderr and exit code are preserved when available, except the known empty-inventory result from `CTR-01`. |
| `CTR-04` | malformed output -> diagnostic error | Code is `zellij_invalid_output`; callers do not receive partial sessions. |

## Invariants

- `INV-01` The adapter always uses `--short --no-formatting` and does not parse
  formatted human output.
- `INV-02` External invocations are wrapped in `context.Context` timeout.
- `INV-03` Registry persistence remains outside `internal/zellij`.

## Verification

- Unit tests parse a multi-session fixture and assert the exact zellij command.
- Unit tests cover missing binary, command failure and invalid output mappings.
