---
title: "FT-016: Zellij Run New-Pane Adapter Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for creating a zellij command pane through the Go adapter."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
  - ../../engineering/zellij-integration.md
status: draft
audience: humans_and_agents
---

# FT-016: Zellij Run New-Pane Adapter Design

## Selected Design

Pane creation lives in `internal/zellij`. The adapter exposes:

- `Client.RunPane(ctx, request)` for creating one command pane in one explicit
  zellij session;
- `PaneRunner` as the downstream interface for future managed create workflow;
- `RunPaneRequest` with explicit `Session`, optional `CWD` and `Name`, and a
  command vector rather than a shell string;
- `PaneRef` with `Session` and typed `PaneID` such as `terminal_7`.

Production execution uses:

```bash
zellij --session <name> run --cwd <path> --name <name> -- <command...>
```

`--cwd` and `--name` are omitted when the caller leaves them empty. The command
vector is passed after `--`, so adapter callers do not depend on user shell
quoting.

## Reference Contract

`zellij run` returns the created terminal pane id on stdout in typed form. The
adapter parses that stdout into the existing `PaneID` type, rejects non-terminal
pane ids for this method and returns a `PaneRef` that also repeats the target
session name supplied by the caller.

The adapter does not confirm pane metadata through `list-panes`. Confirmation,
Codex identity evidence and registry writes belong to downstream create
features.

## Error Contract

Adapter failures are returned as `DiagnosticError` with the existing stable
codes:

| Code | Meaning |
| --- | --- |
| `zellij_invalid_input` | Caller did not provide a session name or command. |
| `zellij_missing_binary` | The zellij executable was not found. |
| `zellij_command_failed` | zellij returned a create/run failure. |
| `zellij_invalid_output` | zellij stdout did not contain a valid terminal pane id. |

Run-pane diagnostics must state that `zelma` did not write registry state. They
must not reuse read-only hints that imply the zellij call had no external side
effect.

## Boundaries

FT-016 does not implement `zelma sessions create`, write `.zelma/sessions.json`,
confirm that Codex started, attach/focus panes or clean up partially created
panes.

## Verification

- Contract tests cover exact `zellij --session <name> run ... -- <command>`
  invocation.
- Parser tests cover valid `terminal_<id>`/`plugin_<id>` references and invalid
  stdout; run-pane tests reject `plugin_<id>` output.
- Failure tests cover missing input, missing zellij binary, command failure,
  session-not-found output and invalid pane reference output.
