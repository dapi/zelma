---
name: zelma
description: Manage repo-local Codex/zellij instances through the public zelma CLI. Use when asked to list zelma instances, check live status, create a Codex pane with zelma, detect manual Codex panes, focus zelma instance 2 or another numeric id, send a message to a verified zelma instance, observe an instance buffer/transcript, preview cleanup stale zelma instances, or confirm stale cleanup after explicit user intent.
---

# zelma

Use this skill to manage Codex instances for the current repository through the
public `zelma` CLI. Run commands from inside the target Git worktree.

## Boundaries

- Call `zelma` only; do not call `zellij` directly.
- Parse `zelma` JSON output only; do not read or parse `.zelma/instances.json`
  directly.
- Preserve CLI stderr/stdout diagnostics when reporting failures.
- Do not run `zelma instances cleanup --confirm --json` unless the user
  explicitly asks to remove stale records.
- Do not infer current task/activity beyond fields exposed by public `zelma`
  JSON. Current public instance fields include identifiers, `opened_path`,
  `state` and optional `live_status`.
- For send-message intents, do not type into terminals manually and do not call
  zellij directly. Use only `zelma instances send` and stop on not-ready
  diagnostics.
- Observation commands are explicit only. Use `zelma instances buffer <id> --json`
  or `zelma instances transcript <id> --json` only when the user asks to inspect
  an instance's current work; do not read pane buffers or Codex transcripts during
  list/status/detect/focus workflows.
## Intent Routing

| User intent | Command | Notes |
| --- | --- | --- |
| Inspect known instances | `zelma instances list --json` | Registry inventory for the current repo. |
| Check live status | `zelma instances list --live --json` | Read-only live/unreachable check; does not persist live status. |
| Create a managed Codex pane | `zelma instances create [path] --json` | Omit `path` to use the repo root; if a live active instance already owns that path, the command returns skipped with the existing instance instead of launching a duplicate. |
| Preview create inputs | `zelma instances create [path] --dry-run --json` | Use before mutating create when path or Codex binary resolution is uncertain. |
| Detect/adopt manual Codex panes | `zelma instances detect --json` | Upserts detected candidates/active records through `zelma`. |
| Explain detect evidence | `zelma instances detect --json --explain` | Use when evidence diagnostics are needed. |
| Focus a known instance by numeric id | `zelma instances focus <id> --json` | Use ids from `instances list`; `<id>` is repo-local and numeric. |
| Send a message to a known instance by numeric id | `zelma instances send <id> [message] --json` or `zelma instances send <id> --stdin --json` | Use ids from `instances list`; `send` revalidates active Codex readiness and never echoes the message body. |
| Observe pane screen/scrollback | `zelma instances buffer <id> --json` | Read-only bounded zellij pane observation; use `--tail <lines>` to reduce output. |
| Observe Codex transcript events | `zelma instances transcript <id> --json` | Read-only bounded Codex transcript observation for instances with `codex_session`; use `--tail <events>` to reduce output. |
| Preview stale cleanup | `zelma instances cleanup --json` | Read-only proposal. |
| Confirm stale cleanup | `zelma instances cleanup --confirm --json` | Only after explicit user intent to remove stale records. |

## Recovery

Choose safe next steps from CLI diagnostics:

- Not in a Git worktree or repo not prepared: move into the target worktree,
  then run `zelma setup`.
- Invalid registry JSON/schema diagnostics: stop and ask the user to restore
  valid schema v1 JSON before mutating instance commands.
- Codex binary missing during create: stop; fix Codex installation or
  `ZELMA_CODEX_BIN`, then retry.
- zellij unavailable or command failure: stop; fix zellij availability/session
  before retrying.
- Created pane cannot be confirmed, or registry write fails after pane
  creation: run `zelma instances detect --json` after the environment issue is
  understood; do not retry blindly.
- Send reports `conflicting_message_sources`, `missing_message` or
  `empty_message`: retry only after fixing the message source.
- Send reports `instance_not_found`, `pane_not_found`, `pane_not_terminal`,
  `instance_state_not_active`, `runtime_unreachable`,
  `codex_runtime_missing`, `codex_identity_mismatch`, `runtime_ambiguous` or
  `target_not_ready`: stop and present the diagnostic. Use the reported public
  `next_command`, usually `zelma instances list --live --json` or
  `zelma instances detect --json`; do not use direct terminal input as fallback.
- Empty registry while live panes are likely: run `zelma instances detect --json`.
- `list --live` or `detect` reports stale instances: present the stale records
  and use `zelma instances cleanup --json` for a read-only proposal.

Use `memory-bank/engineering/skill-contract.md` as the canonical command and
recovery contract when more detail is needed.
