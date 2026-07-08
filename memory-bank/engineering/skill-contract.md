---
title: Codex Skill Contract
doc_kind: engineering
doc_function: canonical
purpose: "Agent-facing contract for Codex skills that manage zelma sessions through the CLI."
derived_from:
  - ../features/FT-023/brief.md
  - ../features/FT-024/brief.md
  - ../features/FT-025/brief.md
  - ../features/FT-026/brief.md
  - ../features/FT-027/brief.md
  - ../features/FT-028/brief.md
  - ../features/FT-029/brief.md
status: active
audience: humans_and_agents
---

# Codex Skill Contract

Codex skills manage `zelma sessions` by calling the `zelma` CLI. The skill layer
does not read `.zelma/sessions.json` directly, does not parse zellij output, and
does not call `zellij` directly. `zelma` owns registry schema, live zellij
inspection, stale detection and cleanup behavior.

## Purpose

Use the skill contract when an agent needs to inspect, create, discover or clean
up Codex sessions for the current repository.

The skill should choose commands from the user's intent:

| Intent | Command | Why |
| --- | --- | --- |
| Show known managed sessions | `zelma sessions list --json` | Stable registry inventory for agents. |
| Check whether known sessions still have live panes | `zelma sessions list --live --json` | Read-only live status without registry mutation. |
| Create a managed Codex pane | `zelma sessions create [path] --json` | Controlled workflow that creates and registers a confirmed pane. |
| Preview create inputs | `zelma sessions create [path] --dry-run --json` | Resolve Codex command and opened path without side effects. |
| Register manually created Codex panes | `zelma sessions detect --json` | Detect live zellij panes and upsert candidate or active records. |
| Review stale cleanup | `zelma sessions cleanup --json` | Propose stale records without mutation. |
| Remove stale records after explicit user intent | `zelma sessions cleanup --confirm --json` | Mutating cleanup for records already marked `stale`. |

If the agent only needs current registry data, prefer `sessions list --json`.
Use `--live` only when live reachability matters, because it contacts zellij.
Use `detect` when the user says a Codex pane already exists outside `zelma` or
when recovery guidance suggests reconciling live panes. Use `cleanup --confirm`
only after explicit user intent to remove stale registry records.

## Command Contracts

All commands run from inside the target git worktree. Successful agent-facing
commands write data to stdout and exit `0`. Failures write diagnostics to stderr
and exit non-zero.

### `zelma sessions list --json`

Reads the repository-local `.zelma/sessions.json`. A missing registry is treated
as an empty registry. Without `--live`, this command does not contact zellij and
does not mutate the registry.

Output is schema v1 registry JSON:

```json
{
  "version": 1,
  "sessions": [
    {
      "zellij_session": "zelma-main",
      "zellij_pane": "terminal_7",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
```

With `--live`, each session also includes `live_status` with `live` or
`unreachable`. The live view is read-only; it does not persist `live_status`.

### `zelma sessions create [path] --json`

Creates a zellij pane through `zelma`, confirms launch evidence, and registers a
record only after confirmation. Omit `path` to open the repository root. If
`path` is present, it must be an existing directory equal to or inside the repo
root.

The successful JSON summary is:

```json
{
  "created": 1,
  "registered": 1,
  "skipped": 0
}
```

New records are `candidate` unless Codex session evidence resolves
unambiguously. If pane creation succeeds but confirmation or registry write
fails, `zelma` does not claim to clean up the zellij pane; recovery should run
`zelma sessions detect --json` after the environment issue is understood.

### `zelma sessions create [path] --dry-run --json`

Resolves the launch contract without creating a pane or writing the registry.

The JSON object includes:

- `opened_path`
- `working_directory`
- `binary`
- `args`

Use dry run when the agent needs to validate inputs, explain what would run, or
debug Codex binary/path resolution before a mutating create.

### `zelma sessions detect --json`

Reads live zellij sessions and panes through `zelma`, classifies Codex panes,
and upserts registry records. It does not create panes and does not delete stale
records.

The successful JSON summary is:

```json
{
  "added": 1,
  "unchanged": 0,
  "skipped": 0,
  "active": 0,
  "candidate": 1,
  "stale": 0
}
```

When existing active records are proven missing by a successful live inventory,
the output may include `stale_candidates` with reason codes. Those records are
marked stale; removal is a separate cleanup command.

### `zelma sessions cleanup --json`

Reads the registry and proposes cleanup for records whose state is already
`stale`. Without `--confirm`, this command does not mutate the registry.

The JSON object includes:

- `summary.proposed`
- `summary.removed`
- `summary.kept`
- `stale_records` when stale records exist

### `zelma sessions cleanup --confirm --json`

Removes only records whose registry state is `stale`. Active, candidate, closed
and archived records are never removed by this command.

## Recovery Expectations

The skill should preserve CLI diagnostics in its response and choose the next
safe command from the diagnostic:

| Situation | Skill response |
| --- | --- |
| Not inside a git repository | Ask the user to run from a repository worktree. |
| Registry JSON is invalid | Stop; tell the user to restore valid schema v1 JSON before mutating commands. |
| Codex binary is missing during create | Stop; fix Codex installation or `ZELMA_CODEX_BIN`, then retry. |
| zellij command fails before pane confirmation | Stop; inspect zellij availability/session, then retry only after the environment is fixed. |
| Created pane cannot be confirmed | Do not retry blindly; run `zelma sessions detect --json` and inspect the zellij pane. |
| Registry write fails after pane creation | Fix filesystem/lock issue, then run `zelma sessions detect --json` before retrying create. |
| `list --live` or `detect` marks sessions stale | Present stale records; use `cleanup --json` for proposal and `cleanup --confirm --json` only on explicit user intent. |

## Boundaries

- Skills call `zelma`; they do not call `zellij` directly.
- Skills parse `zelma` machine-readable output; they do not parse
  `.zelma/sessions.json` as a separate implementation path.
- Skills do not remove records except through `zelma sessions cleanup --confirm`.
- Skills do not assume cleanup of created panes after partial `create` failures.
- Skills keep human-readable stderr diagnostics attached to recovery responses.
