---
title: "FT-027: Sessions List Live Design"
doc_kind: feature
doc_function: design
purpose: "Фиксирует выбранный read-only contract для `sessions list --live`."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
---

# FT-027: Sessions List Live Design

## Contract

`zelma sessions list --live` is a read-only view over the existing sessions
registry enriched with current zellij reachability.

The registry remains the canonical owner of:

- `zellij_session`
- `zellij_pane`
- `codex_session`
- `opened_path`
- `state`

The live view adds only `live_status` in command output. It does not persist
`live_status` into `.zelma/sessions.json`.

## Live Status

| Status | Meaning |
| --- | --- |
| `live` | The registry record's `zellij_session` exists in current zellij inventory and the record's `zellij_pane` exists in that session's pane list. |
| `unreachable` | The zellij session is absent or the pane id is absent from that session's current pane list. |

`unreachable` is not a registry state and must not cause automatic deletion,
cleanup or stale-state mutation.

## Output

Human output adds a `LIVE_STATUS` column between `STATE` and zellij identity
columns.

JSON output keeps the existing schema shape:

- root `version`;
- root `sessions`;
- existing session fields;
- per-session `live_status`.

As of GitHub issue #86, plain `sessions list` and `sessions list --json`
auto-detect by default before rendering inventory. Use
`sessions list --no-detect` for the registry-only contract that does not run the
detect pass. `--live` still owns only live-status enrichment and does not persist
`live_status`.

## Reconciliation Flow

1. Read the repo-local registry.
2. Call `zellij list-sessions --short --no-formatting`.
3. For registry records whose `zellij_session` is currently listed, call
   `zellij --session <name> action list-panes --json --all`.
4. Mark records live by exact typed pane id match, for example `terminal_1`.
5. Return the enriched view without writing the registry.

Transient adapter errors are surfaced as command errors because `--live` cannot
truthfully answer reachability when zellij inventory cannot be read.
