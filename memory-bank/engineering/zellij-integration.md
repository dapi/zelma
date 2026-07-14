---
title: Zellij Integration Research
doc_kind: engineering
doc_function: reference
purpose: Зафиксировать актуальные способы интеграции zelma с zellij, ссылки на API/CLI docs и Go libraries перед проектированием zellij adapter.
derived_from:
  - ../dna/governance.md
  - architecture.md
  - ../domain/context-map.md
  - ../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
canonical_for:
  - zellij_integration_research
  - zellij_adapter_reference_links
  - zellij_resurrection_behavior
  - zellij_pane_command_discovery
  - zellij_session_cache_files
---

# Zellij Integration Research

Snapshot date: `2026-07-07`.

Local probe:

- `zellij --version` returned `zellij 0.44.0`.
- Upstream source spot-check used tag `v0.44.3` in a temporary checkout under
  `/tmp`; local installation was not modified.
- `go version` failed because `go` is not currently in `PATH`.
- `zellij --session zelma action list-panes --json --all` works locally and
  returns pane metadata including terminal/plugin type, title, tab fields,
  geometry, state and observed command/cwd fields for terminal panes.

## Integration Direction

MVP integration should use the Zellij CLI automation surface from Go through
`os/exec`. Do not start with a custom zellij client library or Zellij plugin.

Reasoning:

- Zellij `0.44.0` added/expanded CLI automation for pane listing, sending keys,
  dumping screens, subscribing to pane output and returning created pane IDs.
- `zellij action list-panes --json --all` gives the core facts needed for
  `instances detect`.
- `zellij run` and `zellij action new-pane` return created pane IDs, which is
  enough for `instances create`.
- A Go CLI can keep the integration explicit, testable and fixture-driven by
  wrapping external commands and parsing JSON.

## Primary Zellij Surfaces

| Surface | Link | Use in zelma | Notes |
| --- | --- | --- | --- |
| Zellij CLI control overview | https://zellij.dev/documentation/controlling-zellij-through-cli.html | Entry point for supported automation surfaces | Links to run/edit, action, plugin/pipe and subscribe docs |
| `zellij run` | https://zellij.dev/documentation/zellij-run-and-edit.html | Candidate for `zelma instances create` | Launches command panes, supports `--cwd`, `--name`, layout options and returns `terminal_<id>` |
| `zellij action new-pane` | https://zellij.dev/documentation/cli-actions | Alternative for `instances create` | Returns `terminal_<id>` or `plugin_<id>`; supports `--cwd`, `--name`, `--direction`, `--floating`, `--stacked` |
| `zellij action list-panes --json --all` | https://zellij.dev/documentation/cli-actions | Primary source for `instances detect` and reconciliation | Include `--session <name>` globally when targeting a specific session |
| `zellij list-sessions --short --no-formatting` | https://zellij.dev/documentation/commands.html | Enumerate candidate zellij sessions | Default output is human-formatted; use parse-friendly flags |
| `zellij attach --create-background` | https://zellij.dev/documentation/cli-recipes.html | Future background/session bootstrap | Useful if zelma later creates named background zellij sessions |
| `zellij subscribe --format json` | https://zellij.dev/documentation/zellij-subscribe.html | Future live observation | Emits NDJSON pane update/closed events |
| `zellij action dump-screen` | https://zellij.dev/documentation/cli-actions | Future diagnostics or Codex confirmation | Can dump a pane viewport/scrollback by pane id |
| `zellij action send-keys` | https://zellij.dev/documentation/cli-actions | Future control actions, not MVP create/detect/list | Sends human-readable keys to a pane |
| `zellij action save-session` | https://zellij.dev/documentation/cli-actions | Future manual persistence checkpoint | Forces current session state serialization for resurrection |
| Zellij session resurrection | https://zellij.dev/documentation/session-resurrection.html | Reference for exited session restore and cache layout behavior | Session files are human-readable KDL layouts in the cache folder |
| Zellij options | https://zellij.dev/documentation/options.html | Reference for serialization knobs | Current generated config uses `serialize_pane_viewport`; docs may mention `pane_viewport_serialization` |
| Zellij plugin and pipe | https://zellij.dev/documentation/zellij-plugin-and-pipe.html | Future plugin-based integration | Useful for richer in-session behavior; not MVP |
| Zellij plugin API | https://zellij.dev/documentation/plugin-api.html | Future plugin strategy | Docs are Rust-oriented through `zellij-tile`; Go/WASM path requires separate proof |
| Zellij repository | https://github.com/zellij-org/zellij | Source/reference for behavior and changelog | Use for release notes and issue validation |

## Session Resurrection And Pane Command Discovery

Verified on `2026-07-07` against local `zellij 0.44.0`, official docs and
upstream source tag `v0.44.3`.

### What Zellij Resurrects

Zellij session resurrection serializes a session into a regular Zellij layout:
tabs, panes, cwd and the detected command running in each terminal pane. Exited
resurrectable sessions can be listed with `zellij list-sessions` and restored by
attaching to the exited session.

Important behavior:

- Resurrected commands are command panes, not a continuation of the old process.
- Zellij does not auto-run resurrected commands by default; it shows a
  `Press ENTER to run...` style prompt to avoid dangerous accidental restarts.
- Use `zellij attach --force-run-commands <session>` only when automatic command
  restart is explicitly intended.
- The serialized artifact is a human-readable `session-layout.kdl`; it can be
  inspected, edited and loaded with `zellij --layout <session-layout.kdl>`.
- A pane whose only command is the configured default shell with no arguments is
  treated as a plain shell pane, not serialized as a command pane.

### Configuration Knobs

Root-level `config.kdl` knobs relevant to resurrection:

```kdl
session_serialization true
serialization_interval 60
serialize_pane_viewport true
scrollback_lines_to_serialize 10000
post_command_discovery_hook "printf '%s\n' \"$RESURRECT_COMMAND\" | sed 's/^sudo[[:space:]]\\+//'"
```

Notes:

- `session_serialization` defaults to enabled; set it to `false` to disable
  resurrection state writes.
- `serialization_interval` is in seconds.
- `serialize_pane_viewport` controls viewport serialization in the generated
  `0.44.x` config (`zellij setup --dump-config`). Some public docs mention
  `pane_viewport_serialization`; prefer the generated config name for `0.44.x`.
- `scrollback_lines_to_serialize 0` means all scrollback up to the configured
  scrollback limit, but only when viewport serialization is enabled.
- `post_command_discovery_hook` receives the discovered command in
  `$RESURRECT_COMMAND`; stdout replaces the command that will be serialized.

### Command Discovery Algorithm

Source-level behavior from `v0.44.3`:

- Zellij records the PID of each terminal pane child process.
- For each pane, it prefers a foreground child command whose parent PID matches
  the pane process. On Unix this comes from `ps -ao ppid,args`.
- If no child command is found, it falls back to the pane process command from
  `sysinfo`.
- The detected command vector is mapped to `command` plus `args` in the
  serialized layout.
- If the command equals the configured default shell and has no args, Zellij
  clears it from pane metadata so resurrection starts a normal shell.

Known caveats:

- Unix `ps` output and hook output are split with ASCII whitespace, not a shell
  parser. Quoted args can lose their grouping.
- A wrapper process, background helper or multiple child processes can cause the
  wrong command to be detected.
- The hook can normalize or replace the command, but it still returns a string
  that Zellij later splits by whitespace.
- If `zelma` ever needs precise resurrect command args, prefer generating KDL
  with explicit `command` and `args` fields over depending on auto-discovery.

### Cache Files

Use `zellij setup --check` to find the effective cache dir. Local probe:

```text
[CACHE DIR]: /Users/danil/Library/Caches/org.Zellij-Contributors.Zellij
```

For local `0.44.0`, resurrect files are under:

```text
/Users/danil/Library/Caches/org.Zellij-Contributors.Zellij/
└── contract_version_1/
    └── session_info/
        └── <zellij-session-name>/
            ├── session-layout.kdl
            └── session-metadata.kdl
```

`session-layout.kdl` is the restore layout. `session-metadata.kdl` is used for
session listing/metadata; disabling session metadata can reduce or remove this
support and may affect session-manager/listing features.

Do not hard-code the macOS path in `zelma`; resolve the cache dir through
documented diagnostics or isolate path assumptions behind the adapter. The
layout path shape is useful for fixtures and manual inspection, not as the MVP
live-detection API.

### CLI And Plugin Read Surfaces

Stable first-choice CLI surfaces for `zelma`:

```bash
zellij list-sessions --short --no-formatting
zellij --session <name> action list-panes --json --all
zellij --session <name> action dump-layout
zellij --session <name> action save-session
```

Plugin API notes:

- `PaneUpdate` exposes pane manifest data including pane command, focus, title,
  position and geometry.
- `ListClients` exposes connected clients and their focused pane/running command.
- Plugin API is not the MVP integration path because it adds WASM/plugin
  lifecycle and permission concerns.

### Implications For Zelma

- Use live `list-panes --json --all` for `instances detect`, not resurrect files.
- Treat resurrect files as a recovery/inspection surface for exited sessions and
  as fixture material.
- Never assume resurrected command panes already started; account for suspended
  command panes unless `--force-run-commands` was used intentionally.
- Do not rely on Zellij command discovery for stable Codex identity extraction.
  It is a useful signal, but `CodexSessionRef` needs separate Codex evidence.
- If `zelma` creates its own layouts in the future, emit explicit KDL command
  and `args` nodes rather than round-tripping a shell command string.

## Local JSON Findings

Observed with:

```bash
zellij --session zelma action list-panes --json --all
```

Relevant fields seen in local `0.44.0` output:

- `id`
- `is_plugin`
- `is_focused`
- `is_floating`
- `is_suppressed`
- `title`
- `exited`
- `exit_status`
- `tab_id`
- `tab_position`
- `tab_name`
- `plugin_url`
- `pane_command`
- `pane_cwd`

Adapter rule: treat these as zellij adapter facts, not domain field names. Store
normalized domain values in `.zelma/instances.json` after mapping and validation.

## MVP Adapter Rules

- Always target sessions explicitly when operating outside the current pane:
  `zellij --session <name> ...`.
- Use `context.Context` timeouts around every external `zellij` invocation.
- Use `zellij list-sessions --short --no-formatting` for session discovery.
- Use `zellij --session <name> action list-panes --json --all` for pane
  discovery.
- Filter out plugin panes for Codex detection unless a future feature explicitly
  supports plugin-hosted workflows.
- Normalize pane identity as typed pane id, for example `terminal_0`, not bare
  `0`; terminal and plugin pane ids can overlap.
- Prefer `zellij run --cwd <path> --name <name> -- codex` for the first create
  prototype. Re-evaluate `action new-pane` if it gives better control in tests.
- For verified send delivery, keep text submission behind a typed adapter method
  such as `SendTextToPane`. FT-101 uses one explicit-pane
  `zellij --session <name> action write-chars --pane-id <pane> <message + "\n">`
  call when `submit=true`. Adapter diagnostics for this path must redact the
  message body.
- Never parse ANSI-colored human output when a parse-friendly flag or JSON
  output exists.

## Detection Strategy Notes

For `instances detect`, first candidate signal is the pane command/cwd metadata
from `list-panes --json --all`:

- candidate pane: `is_plugin == false`;
- candidate command: command line contains Codex executable/entrypoint;
- candidate path: `pane_cwd` exists and is inside or equal to the current repo
  root.

`CodexSessionRef` extraction rules live in
[`codex-runtime-identification.md`](codex-runtime-identification.md). Do not
mark a record `active` until those rules produce a resolved UUID.

## Go Libraries

| Library | Link | Candidate use | Adoption status |
| --- | --- | --- | --- |
| `os/exec` | https://pkg.go.dev/os/exec | Run `zellij` with context, stdout/stderr capture and stdin pipes | Primary, standard library |
| `encoding/json` | https://pkg.go.dev/encoding/json | Parse zellij JSON output and `.zelma/instances.json` | Primary, standard library |
| `github.com/spf13/cobra` | https://pkg.go.dev/github.com/spf13/cobra | Command tree for `zelma instances create/detect/list` | Selected CLI framework |
| `os.CreateTemp` + `os.Rename` | https://pkg.go.dev/os | Atomic replacement of `.zelma/instances.json` on Unix and best-effort replace on Windows | Primary persistence helper |
| `github.com/gofrs/flock` | https://pkg.go.dev/github.com/gofrs/flock | Cross-process lock around registry writes | Candidate locking helper |
| `github.com/calico32/kdl-go` | https://pkg.go.dev/github.com/calico32/kdl-go | Parse/emit KDL layouts if zelma starts generating zellij layouts | Optional, not MVP |
| `github.com/njreid/gokdl2` | https://pkg.go.dev/github.com/njreid/gokdl2 | Alternative KDL parser with v1/v2 support and compliance focus | Optional, compare before adopting KDL |

No official Go zellij client library was found. The conservative first
implementation should define an internal `zellij-adapter` over `os/exec` and
cover it with fixtures from supported zellij versions.

## Risks

- Zellij CLI JSON shape is a compatibility contract for zelma even if it is not
  a Go API. Pin supported zellij versions and keep fixtures.
- Commands without `--session` may target ambient state and are unsuitable for
  reliable automation.
- `list-sessions` default output includes formatting; use `--short
  --no-formatting`.
- A pane id alone is not enough; keep pane type and zellij session name.
- Plugin API is powerful but adds WASM/plugin lifecycle and permission concerns.
  It should be a later ADR, not the MVP path.
