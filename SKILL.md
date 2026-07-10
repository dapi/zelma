---
name: zelma
description: Manage repo-local Codex/zellij sessions through the public zelma CLI. Use when asked to list zelma sessions, check live status, create a Codex pane with zelma, detect manual Codex panes, focus zelma session 2 or another numeric id, send a message to a verified zelma session, observe a session buffer/transcript, preview cleanup stale zelma sessions, or confirm stale cleanup after explicit user intent.
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
- For send-message intents, do not type into terminals manually and do not call
  zellij directly. Use only `zelma sessions send` and stop on not-ready
  diagnostics.
- Observation commands are explicit only. Use `zelma sessions buffer <id> --json`
  or `zelma sessions transcript <id> --json` only when the user asks to inspect
  a session's current work; do not read pane buffers or Codex transcripts during
  list/status/detect/focus workflows.

## Intent Routing

| User intent | Command | Notes |
| --- | --- | --- |
| Inspect known sessions | `zelma sessions list --json` | Registry inventory for the current repo. |
| Check live status | `zelma sessions list --live --json` | Read-only live/unreachable check; does not persist live status. |
| Create a managed Codex pane | `zelma sessions create [path] --json` | Omit `path` to use the repo root; if a live active session already owns that path, the command returns skipped with the existing session instead of launching a duplicate. |
| Preview create inputs | `zelma sessions create [path] --dry-run --json` | Use before mutating create when path or Codex binary resolution is uncertain. |
| Detect/adopt manual Codex panes | `zelma sessions detect --json` | Upserts detected candidates/active records through `zelma`. |
| Explain detect evidence | `zelma sessions detect --json --explain` | Use when evidence diagnostics are needed. |
| Focus a known session by numeric id | `zelma sessions focus <id> --json` | Use ids from `sessions list`; `<id>` is repo-local and numeric. |
| Send a message to a known session by numeric id | `zelma sessions send <id> [message] --json` or `zelma sessions send <id> --stdin --json` | Use ids from `sessions list`; `send` revalidates active Codex readiness and never echoes the message body. |
| Observe pane screen/scrollback | `zelma sessions buffer <id> --json` | Read-only bounded zellij pane observation; use `--tail <lines>` to reduce output. |
| Observe Codex transcript events | `zelma sessions transcript <id> --json` | Read-only bounded Codex transcript observation for sessions with `codex_session`; use `--tail <events>` to reduce output. |
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
- Send reports `conflicting_message_sources`, `missing_message` or
  `empty_message`: retry only after fixing the message source.
- Send reports `session_not_found`, `pane_not_found`, `pane_not_terminal`,
  `session_state_not_active`, `runtime_unreachable`,
  `codex_runtime_missing`, `codex_identity_mismatch`, `runtime_ambiguous` or
  `target_not_ready`: stop and present the diagnostic. Use the reported public
  `next_command`, usually `zelma sessions list --live --json` or
  `zelma sessions detect --json`; do not use direct terminal input as fallback.
- Empty registry while live panes are likely: run `zelma sessions detect --json`.
- `list --live` or `detect` reports stale sessions: present the stale records
  and use `zelma sessions cleanup --json` for a read-only proposal.

Use `memory-bank/engineering/skill-contract.md` as the canonical command and
recovery contract when more detail is needed.
