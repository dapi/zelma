---
title: "FT-026: Agent Recovery Flows Design"
doc_kind: feature_design
doc_function: canonical
purpose: "Selected design for mapping CLI diagnostics to safe skill recovery actions."
derived_from:
  - brief.md
  - ../../engineering/skill-contract.md
status: active
delivery_status: implemented
audience: humans_and_agents
---

# FT-026: Agent Recovery Flows Design

## Selected Design

Agent recovery lives in `internal/skills` beside the thin CLI wrappers. The
skill layer keeps calling only the public `zelma` CLI, preserves CLI stderr on
failures and adds a structured recovery response:

- `action`: one of `setup`, `detect`, `retry`, `inspect` or `stop`;
- `reason_code`: the stable CLI reason code or skill-level scenario code;
- `message`: agent-readable guidance;
- `next_command`: an optional safe `zelma` command.

## Recovery Map

| CLI reason or scenario | Action | Safe next command |
| --- | --- | --- |
| `unsupported repo`, `repo_not_ready`, `repo_not_prepared` | `setup` | `zelma setup` |
| empty `sessions list --json` result while live panes are likely | `detect` | `zelma sessions detect --json` |
| `create_pane_unconfirmed`, `create_confirmation_failed` | `detect` | `zelma sessions detect --json` |
| `create_registry_write_failed` | `detect` | `zelma sessions detect --json` |
| `create_pane_launch_failed`, `zellij_missing_binary`, `zellij_command_failed` | `stop` | none |
| `create_codex_missing_binary` | `stop` | none |
| registry schema/validation reason codes | `stop` | none |
| registry read/adapter compatibility errors | `inspect` | none |
| stale records reported by detect | `inspect` | `zelma sessions cleanup --json` |
| `registry_locked` | `retry` | none |

`zelma sessions cleanup --json` is a read-only proposal command. Recovery flows
must not suggest `cleanup --confirm`; that remains gated by explicit user
intent.

## Boundaries

- Recovery suggestions never call `zellij` directly.
- Recovery suggestions never read or edit `.zelma/sessions.json` directly.
- Recovery suggestions never perform destructive actions automatically.
- `detect` is the only recovery command suggested after partial create
  confirmation or registry write failures, because it reconciles possible live
  panes through the CLI-owned path.

## Verification

Unit tests in `internal/skills` cover:

- repo-not-ready diagnostics suggesting `zelma setup`;
- zellij unavailable diagnostics stopping for environment repair;
- empty registry plus likely live panes suggesting `sessions detect`;
- stale detect output suggesting only cleanup preview;
- recovery next commands staying inside safe `zelma` CLI surfaces.
