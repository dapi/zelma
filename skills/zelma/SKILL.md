---
name: zelma
description: Manage repo-local Codex/zellij sessions through the public zelma CLI. Use when asked to list zelma sessions, check live status, create a Codex pane with zelma, detect manual Codex panes, focus zelma session 2 or another numeric id, preview cleanup stale zelma sessions, or confirm stale cleanup after explicit user intent.
---

# zelma

Use this skill to manage Codex sessions for the current repository through the
public `zelma` CLI. Run commands from inside the target Git worktree.

## Boundaries

- Call `zelma` only; do not call `zellij` directly.
- Parse `zelma` JSON output only; do not read or parse `.zelma/sessions.json`
  directly.
- Preserve CLI stderr/stdout diagnostics when reporting failures.
- Do not run `zelma sessions cleanup --confirm --json` unless the user
  explicitly asks to remove stale records.
- Do not infer current task/activity beyond fields exposed by public `zelma`
  JSON. Current public session fields include identifiers, `opened_path`,
  `state` and optional `live_status`.

## Intent Routing

| User intent | Command | Notes |
| --- | --- | --- |
| Inspect known sessions | `zelma sessions list --json` | Registry inventory for the current repo. |
| Check live status | `zelma sessions list --live --json` | Read-only live/unreachable check; does not persist live status. |
| Create a managed Codex pane | `zelma sessions create [path] --json` | Omit `path` to use the repo root. |
| Preview create inputs | `zelma sessions create [path] --dry-run --json` | Use before mutating create when path or Codex binary resolution is uncertain. |
| Detect/adopt manual Codex panes | `zelma sessions detect --json` | Upserts detected candidates/active records through `zelma`. |
| Explain detect evidence | `zelma sessions detect --json --explain` | Use when evidence diagnostics are needed. |
| Focus a known session by numeric id | `zelma sessions focus <id> --json` | Use ids from `sessions list`; `<id>` is repo-local and numeric. |
| Preview stale cleanup | `zelma sessions cleanup --json` | Read-only proposal. |
| Confirm stale cleanup | `zelma sessions cleanup --confirm --json` | Only after explicit user intent to remove stale records. |

## Recovery

Choose safe next steps from CLI diagnostics:

- Not in a Git worktree or repo not prepared: move into the target worktree,
  then run `zelma setup`.
- Invalid registry JSON/schema diagnostics: stop and ask the user to restore
  valid schema v1 JSON before mutating session commands.
- Codex binary missing during create: stop; fix Codex installation or
  `ZELMA_CODEX_BIN`, then retry.
- zellij unavailable or command failure: stop; fix zellij availability/session
  before retrying.
- Created pane cannot be confirmed, or registry write fails after pane
  creation: run `zelma sessions detect --json` after the environment issue is
  understood; do not retry blindly.
- Empty registry while live panes are likely: run `zelma sessions detect --json`.
- `list --live` or `detect` reports stale sessions: present the stale records
  and use `zelma sessions cleanup --json` for a read-only proposal.

Use `memory-bank/engineering/skill-contract.md` as the canonical command and
recovery contract when more detail is needed.
