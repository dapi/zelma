---
title: "FT-011: Zellij Adapter ListPanes Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for the read-only zellij ListPanes adapter contract."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
  - ../../engineering/zellij-integration.md
status: draft
audience: humans_and_agents
---

# FT-011: Zellij Adapter ListPanes Design

## Selected Design

Pane inspection lives in `internal/zellij`. The adapter exposes:

- `Client.ListPanes(ctx, session)` for read-only pane discovery in one explicit
  zellij session;
- `PaneLister` as the downstream interface for detect/classifier code;
- the existing FT-010 injected runner seam for fixture tests without a live
  terminal.

Production execution uses:

```bash
zellij --session <name> action list-panes --json --all
```

The binary defaults to `zellij` and can be overridden with `ZELMA_ZELLIJ_BIN`.

## Pane Record Contract

The adapter returns project-owned `Pane` records instead of exposing raw zellij
JSON. It normalizes:

- pane identity to typed IDs such as `terminal_1` and `plugin_3`;
- pane kind through typed `PaneID`;
- command metadata from `pane_command`;
- working directory metadata from `pane_cwd`;
- tab, focus, floating, exited and plugin URL fields when present.

Missing command or cwd metadata does not fail parsing. The record is returned
with nil optional metadata fields, so later classifier logic can make an
uncertainty decision without panicking.

## Error Contract

Adapter failures are returned as `DiagnosticError` with stable codes:

| Code | Meaning |
| --- | --- |
| `zellij_invalid_input` | Caller did not provide a session name. |
| `zellij_missing_binary` | The zellij executable was not found. |
| `zellij_command_failed` | zellij returned a command/runtime error. |
| `zellij_invalid_output` | JSON could not be parsed into pane records. |

## Verification

- Fixture test parses multiple terminal/plugin panes and normalized metadata.
- Partial fixture test keeps a pane record with explicit missing command/cwd
  fields.
- Contract tests cover explicit `--session` invocation and diagnostic mapping.
