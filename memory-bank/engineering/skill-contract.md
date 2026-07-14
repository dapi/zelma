---
title: Codex Skill Contract
doc_kind: engineering
doc_function: canonical
purpose: "Agent-facing contract for Codex skills that manage zelma instances through the CLI."
derived_from:
  - ../features/FT-023/brief.md
  - ../features/FT-024/brief.md
  - ../features/FT-025/brief.md
  - ../features/FT-026/brief.md
  - ../features/FT-027/brief.md
  - ../features/FT-028/brief.md
  - ../features/FT-029/brief.md
  - ../features/FT-047/brief.md
  - ../features/FT-035/brief.md
  - ../features/FT-101/brief.md
status: active
audience: humans_and_agents
---

# Codex Skill Contract

Codex skills manage `zelma instances` by calling the `zelma` CLI. The skill layer
does not read `.zelma/instances.json` directly, does not parse zellij output, and
does not call `zellij` directly. `zelma` owns registry schema, live zellij
inspection, stale detection and cleanup behavior.

## Purpose

Use the skill contract when an agent needs to inspect, create, discover or clean
up Codex instances for the current repository.

The skill should choose commands from the user's intent:

| Intent | Command | Why |
| --- | --- | --- |
| Show current managed instances | `zelma instances list --json` | Primary inventory command; auto-detects fresh-enough manual panes before returning schema v1 JSON. |
| Show registry only without probing | `zelma instances list --no-detect --json` | Stable registry-only inventory for callers that must avoid zellij/Codex probing. |
| Check whether known instances still have live panes | `zelma instances list --live --json` | Auto-detects first unless cache is fresh, then adds live status. |
| Create a managed Codex pane | `zelma instances create [path] --json` | Controlled workflow that creates and registers a confirmed pane. |
| Preview create inputs | `zelma instances create [path] --dry-run --json` | Resolve Codex command and opened path without side effects. |
| Run an explicit diagnostic detect pass | `zelma instances detect --json` | Detect live zellij panes and upsert candidate or active records outside the normal list workflow. |
| Focus a known instance pane | `zelma instances focus <id> --json` | Switch zellij UI to a registry-backed tab/pane without registry mutation. |
| Send a message to a known instance pane | `zelma instances send <id> [message] --json` or `zelma instances send <id> --stdin --json` | Deliver text only after `zelma` revalidates active Codex readiness; never use direct `zellij` fallback. |
| Observe an instance pane screen | `zelma instances buffer <id> --json` | Read bounded current zellij screen/scrollback for an explicit repo-local instance id without registry mutation. |
| Observe Codex transcript events | `zelma instances transcript <id> --json` | Read bounded Codex transcript events for an explicit repo-local instance id with a resolved `codex_session` without registry mutation. |
| Review stale cleanup | `zelma instances cleanup --json` | Propose stale records without mutation. |
| Remove stale records after explicit user intent | `zelma instances cleanup --confirm --json` | Mutating cleanup for records already marked `stale`. |

If the agent needs current inventory, prefer `instances list --json`. Use
`--no-detect` only when the caller explicitly needs a registry-only read with no
zellij/Codex probing. Use `--live` when live reachability matters; it may contact
zellij even when auto-detect is skipped by cache freshness. Keep standalone
`detect` for diagnostics/manual reconciliation, not normal inventory. Use
`cleanup --confirm` only after explicit user intent to remove stale registry
records.

Use observation commands only when the user or orchestrator explicitly asks to
inspect an instance's current work. `list`, `status`, `detect`, `focus` and
`cleanup` must not read pane buffers or Codex transcript contents implicitly.

`zelma instances` is the only public resource command. The pre-rename resource
command path is intentionally removed, and schema v1 JSON uses
`instances`/`instance` fields.

## Command Contracts

All commands run from inside the target git worktree. Successful agent-facing
commands write data to stdout and exit `0`. Failures write diagnostics to stderr
and exit non-zero. For commands invoked with `--json`, failure diagnostics on
stderr are a stable JSON object with:

- `code`: machine-readable failure code;
- `retryable`: whether the same operation can be retried without required
  manual repair;
- `manual_action_required`: whether a person or environment repair is required
  before continuing safely;
- `recovery_hint`: human-readable recovery guidance;
- `next_command`: safe public `zelma` command to run next, or an empty array
  when no automatic command is safe.

### `zelma instances list --json`

Runs auto-detect by default unless the last successful auto-detect timestamp is
fresh according to `instances_list.auto_detect_ttl` in `.zelma/config.json`
(default `5s`), then reads the repository-local `.zelma/instances.json`. A
missing registry is treated as an empty registry.

Output is schema v1 registry JSON and preserves all registry records for
machine-readable use, including active, candidate, stale, closed and
archived states. The human table output of `zelma instances list` includes
`active` and `candidate` records by default; use `zelma instances list --all`
when a human needs stale, closed or archived records too.

```json
{
  "version": 1,
  "instances": [
    {
      "id": 1,
      "zellij_session": "zelma-main",
      "zellij_tab": "tab_1",
      "zellij_tab_name": "work",
      "zellij_pane": "terminal_7",
      "codex_session": "11111111-1111-4111-8111-111111111111",
      "opened_path": "/workspace/zelma",
      "state": "active"
    }
  ]
}
```

With `--no-detect`, the command skips auto-detect and reads only
`.zelma/instances.json`. With `--live`, each instance also includes `live_status`
with `live` or `unreachable`. The live view does not persist `live_status`.

### `zelma instances create [path] --json`

Creates a zellij pane through `zelma`, confirms launch evidence, and registers a
record only after confirmation. Omit `path` to open the repository root. If
`path` is present, it must be an existing directory equal to or inside the repo
root.

Before launching a new pane, create checks for an existing live `active` record
with the same `opened_path`. In that handoff case it does not create a duplicate
pane and returns `created: 0`, `registered: 0`, `skipped: 1` plus the existing
instance, so the caller can continue polling or focus that instance.

Otherwise, the successful JSON object includes the stable create counters plus
the registered instance row returned by the registry upsert. `instance` is the
matching `active` record when one exists for the zellij pane key; otherwise it
is the matching `candidate` record. Historical `closed` and `stale` records are
not returned as the create result instance.

```json
{
  "created": 1,
  "registered": 1,
  "skipped": 0,
  "instance": {
    "id": 1,
    "zellij_session": "zelma-main",
    "zellij_tab": "tab_1",
    "zellij_tab_name": "work",
    "zellij_pane": "terminal_7",
    "codex_session": "11111111-1111-4111-8111-111111111111",
    "opened_path": "/workspace/zelma",
    "state": "active"
  }
}
```

New records are `candidate` unless Codex session evidence resolves
unambiguously. If pane creation succeeds but confirmation or registry write
fails, `zelma` does not claim to clean up the zellij pane; recovery should run
`zelma instances detect --json` after the environment issue is understood. This
explicit detect command bypasses the `instances list` auto-detect cache.

### `zelma instances create [path] --dry-run --json`

Resolves the launch contract without creating a pane or writing the registry.

The JSON object includes:

- `opened_path`
- `working_directory`
- `binary`
- `args`

Use dry run when the agent needs to validate inputs, explain what would run, or
debug Codex binary/path resolution before a mutating create.

### `zelma instances detect --json`

Reads live zellij sessions and panes through `zelma`, classifies Codex panes,
and upserts registry records. This command is kept for diagnostic/manual detect
passes; normal inventory should use `zelma instances list --json`. It does not
create panes and does not delete stale records.

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

Use `zelma instances detect --json --explain` when the agent needs per-candidate
evidence diagnostics. The output adds optional `candidate_explanations` records
with zellij identity, `opened_path`, `codex_session` when resolved, and
`evidence_verdict` / `evidence_source` / `evidence_reason`. Default
`--json` omits this field for compatibility.

### `zelma instances focus <id> --json`

Reads the registry, finds the record by positive repo-local `id`, and sends
zellij focus actions through `zelma`. This command changes zellij UI focus but
does not mutate `.zelma/instances.json`.

The successful JSON object is the focused instance record:

```json
{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_tab": "tab_6",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": "/workspace/zelma",
  "state": "active"
}
```

### `zelma instances send <id> [message] --json`

Sends a message to an existing active Codex instance after `zelma` revalidates
that the registry record still points to the intended live terminal pane. The
target selector is only the positive repo-local `id` from `zelma instances list`.
The message source is exactly one of:

- one positional `message` argument;
- `--stdin`, which reads the full stdin stream and allows multiline content.

Do not pass both sources. Do not call `zellij` directly and do not type into the
terminal manually as fallback.

The successful JSON object includes target/instance identity and message
metadata, but never the message body:

```json
{
  "id": 2,
  "zellij_session": "zelma-main",
  "zellij_tab": "tab_6",
  "zellij_pane": "terminal_75",
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "opened_path": "/workspace/zelma",
  "state": "active",
  "message": {
    "source": "argument",
    "byte_count": 18,
    "line_count": 1,
    "submitted": true
  }
}
```

Before any write, `zelma` checks that the registry record exists, is `active`,
the zellij session and recorded pane are reachable, the pane is terminal, and
live Codex evidence is compatible with the recorded `codex_session` and
`opened_path`. A live Codex launch command does not need to repeat the
`codex_session` UUID when the active registry record was already resolved from
process/session metadata and the live pane still matches the recorded
`opened_path`; an explicitly different live UUID remains an identity mismatch.

Stable send failure codes include:

- `conflicting_message_sources`
- `missing_message`
- `empty_message`
- `instance_not_found`
- `pane_not_found`
- `pane_not_terminal`
- `instance_state_not_active`
- `runtime_unreachable`
- `codex_runtime_missing`
- `codex_identity_mismatch`
- `runtime_ambiguous`
- `target_not_ready`

On any not-ready send diagnostic, stop and present the diagnostic. Use only the
public `next_command` returned by `zelma`, typically
`zelma instances list --live --json` or `zelma instances detect --json`. The skill
must not repair, focus or send through direct zellij commands.

### `zelma instances send <id> --stdin --json`

Same target and readiness contract as positional send, but the message body is
read from stdin. Use this form for multiline prompts or text that should not be
placed in the command arguments. Empty stdin is rejected with `empty_message`.

### `zelma instances buffer <id> --json`

Reads the registry, finds the active record by positive repo-local `id`, and
reads the current pane screen through the zellij adapter. This command is
read-only and does not persist pane content. Output is bounded by
`--tail <lines>`; default `120`.

```json
{
  "version": 1,
  "session_id": 2,
  "source": "zellij_buffer",
  "captured_at": "2026-07-10T00:00:00Z",
  "truncated": false,
  "limit": 120,
  "items": [
    {
      "line": 1,
      "text": "synthetic pane line"
    }
  ]
}
```

### `zelma instances transcript <id> --json`

Reads the registry, finds the active record by positive repo-local `id`, and
uses its `codex_session` to read matching Codex JSONL events through the codex
adapter. This command is read-only and does not persist prompts, assistant
answers, tool payloads or transcript content in `.zelma/instances.json`. Output
is bounded by `--tail <events>`; default `50`.

```json
{
  "version": 1,
  "session_id": 2,
  "source": "codex_transcript",
  "captured_at": "2026-07-10T00:00:00Z",
  "truncated": false,
  "limit": 50,
  "codex_session": "11111111-1111-4111-8111-111111111111",
  "items": [
    {
      "index": 1,
      "type": "session_meta"
    }
  ]
}
```

### `zelma instances cleanup --json`

Reads the registry and proposes cleanup for records whose state is already
`stale`. Without `--confirm`, this command does not mutate the registry.

The JSON object includes:

- `summary.proposed`
- `summary.removed`
- `summary.kept`
- `stale_records` when stale records exist

### `zelma instances cleanup --confirm --json`

Removes only records whose registry state is `stale`. Active, candidate, closed
and archived records are never removed by this command.

## Recovery Expectations

The skill preserves CLI diagnostics in its response and attaches a structured
recovery response when it can choose a safe next step. The recovery response
contains:

- `action`: `setup`, `detect`, `retry`, `inspect` or `stop`;
- `reason_code`: the CLI reason code, or a skill-level scenario code for
  successful but incomplete states;
- `message`: agent-readable guidance;
- `next_command`: optional safe `zelma` command.

| Situation | Skill response |
| --- | --- |
| Repository is not ready or not a Git worktree | `setup`; move into the target worktree, then run `zelma setup`. |
| Registry JSON is invalid | `stop`; tell the user to restore valid schema v1 JSON before mutating commands. |
| Codex binary is missing during create | `stop`; fix Codex installation or `ZELMA_CODEX_BIN`, then retry. |
| zellij is unavailable or command execution fails | `stop`; fix zellij availability/session before retrying. |
| Created pane cannot be confirmed | `detect`; run `zelma instances detect --json` and inspect the pane. |
| Registry write fails after pane creation | `detect`; fix filesystem/lock issue, then run `zelma instances detect --json` before retrying create. |
| Registry is empty but live panes are likely | `detect`; run `zelma instances detect --json`. |
| `list --live` or `detect` marks instances stale | `inspect`; present stale records and use `cleanup --json` for proposal. |

Recovery `next_command` values must stay within the public `zelma` CLI. They
must not call `zellij` directly, read `.zelma/instances.json`, or suggest
`cleanup --confirm` without explicit user intent.

## Boundaries

- Skills call `zelma`; they do not call `zellij` directly.
- Skills parse `zelma` machine-readable output; they do not parse
  `.zelma/instances.json` as a separate implementation path.
- Skills do not remove records except through `zelma instances cleanup --confirm`.
- Skills do not assume cleanup of created panes after partial `create` failures.
- Skills keep human-readable stderr diagnostics attached to recovery responses.
- Skills do not read pane buffers or Codex transcript files directly; explicit
  observation must go through `zelma instances buffer <id> --json` or
  `zelma instances transcript <id> --json`.
