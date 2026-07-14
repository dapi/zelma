---
title: Codex Runtime Identification Design
doc_kind: engineering
doc_function: reference
purpose: Зафиксировать design правил, evidence sources и verdicts для извлечения CodexSessionRef из live zellij panes.
derived_from:
  - ../dna/governance.md
  - ../domain/model.md
  - ../domain/rules.md
  - ../domain/states.md
  - ../domain/context-map.md
  - zellij-integration.md
status: active
audience: humans_and_agents
canonical_for:
  - codex_runtime_identification
  - codex_session_ref_extraction
  - codex_detection_evidence_rules
  - codex_adapter_design
---

# Codex Runtime Identification Design

Snapshot date: `2026-07-07`.

This document defines how `zelma` should identify that a `zellij pane` is
running Codex and resolve a `CodexSessionRef` for `instances detect` and
`instances create`.

## Verified Local Facts

Local probe on `2026-07-07`:

- `codex --version` returned `codex-cli 0.142.3`.
- `codex resume --help` states that the positional `SESSION_ID` is a UUID or
  session name, with UUID taking precedence when it parses.
- `codex delete --help` accepts the same UUID-or-session-name identity.
- Local Codex session files exist under `~/.codex/sessions/YYYY/MM/DD/` and use
  names shaped like `rollout-<timestamp>-<uuid>.jsonl`.
- The first JSONL record inspected was `type = "session_meta"` and its
  `payload` contained keys including `id`, `session_id`, `cwd`, `cli_version`
  and `timestamp`.
- Local `zellij action list-panes --json --all` exposes pane command and cwd,
  but not pane PID.
- Local process argv can expose Codex command lines, including `resume <uuid>`
  for resumed sessions. Process argv can also contain user prompt text; it must
  not be logged or persisted raw.

No Codex conversation contents were inspected for this design. Future fixtures
must use synthetic `session_meta` records, not real transcripts.

## Design Goals

- Prefer false negatives over false positives. A pane that only maybe contains
  Codex remains a `DetectionCandidate`, not an `active` `ZelmaInstance`.
- Resolve `CodexSessionRef` without reading Codex conversation messages.
- Keep zellij facts, Codex facts and registry writes separated by module
  boundary.
- Make ambiguity visible to CLI/skills so a future command can ask for manual
  resolution instead of guessing.

## Non-Goals

- Do not parse or summarize Codex conversations.
- Do not mutate Codex session logs.
- Do not depend on private Codex internals beyond conservative filesystem
  metadata and first-line `session_meta` parsing.
- Do not use zellij resurrection files as the primary live-detection source.

## CodexSessionRef Shape

This is a design-level value object, not a final JSON schema:

```text
CodexSessionRef {
  session_id: UUID
  source: argv_resume | argv_external_session_uuid | session_file_created_by_zelma | session_file_unique_match | manual_override
  session_file?: path
  confidence: strong | medium
}
```

Rules:

- `session_id` must parse as a UUID.
- `source` must explain how the UUID was resolved.
- `session_file` is optional and should be repo/user local; do not require it to
  be stable across machines.
- Do not store full Codex argv, prompts, or transcript content in the registry.

## Evidence Sources

| Evidence | Owner module | Strength | Use |
| --- | --- | --- | --- |
| `zellij` terminal pane command contains Codex binary | `zellij-adapter` | weak | Establishes candidate only |
| `zellij` pane cwd equals or is inside repo root | `zellij-adapter` | weak | Supports `OpenedPath` resolution |
| Process argv contains `codex resume <uuid>` | `codex-adapter` | strong | Directly resolves `CodexSessionRef` |
| Process argv contains `CODEX_EXTERNAL_SESSION_UUID=<uuid>` or `External session UUID: <uuid>` | `codex-adapter` | strong external ref | Resolves `CodexSessionRef` as wrapper-provided external identity |
| Session file `session_meta.payload.session_id` or `id` is a UUID | `codex-adapter` | medium/strong depending on correlation | Candidate UUID source |
| Session file cwd matches normalized opened path | `codex-adapter` | medium | Correlates Codex file to pane/repo |
| Session file was created after a `zelma instances create` launch timestamp | `codex-adapter` + `detection` | strong if unique | Resolves create result |
| Exactly one live/recent session file matches cwd and observation window | `codex-adapter` + `detection` | medium | May resolve manual detect only if unambiguous |
| User-supplied UUID in a future explicit command/flag | `cli` + `codex-adapter` | strong after validation | Manual ambiguity resolution |

Weak evidence is never enough for `active`.

## CWD Evidence Policy

`cwd` is intentionally a dual-source fact:

- Zellij `pane_cwd` is the primary source for live `OpenedPath`, because it is
  observed from the actual terminal pane being registered.
- Codex `session_meta.payload.cwd` is corroborating evidence for matching a
  Codex session file to the same repo/path.

Rules:

- `active_ready` requires both sources to be compatible when both are present.
- Compatible means both normalized paths are equal, or both are inside the same
  target repo root and the selected `OpenedPath` remains the Zellij `pane_cwd`.
- If Zellij cwd is missing but Codex cwd is available, Detection may keep a
  candidate but must not create `active` unless another strong live-pane binding
  exists.
- If Zellij cwd and Codex cwd conflict across different repo roots, return
  `candidate_ambiguous` or `candidate_unresolved`; do not guess.
- One cwd match is not a `CodexSessionRef`. It only narrows the set of possible
  Codex session files.

## Resolution Algorithm

Input facts:

- repo root and current command context;
- observed `zellij session`;
- observed `zellij pane` facts from `list-panes --json --all`;
- optional process evidence if implementation can safely correlate it;
- Codex session metadata discovered under Codex home.

Steps:

1. Reject non-terminal panes.
2. Build `OpenedPath` from `pane_cwd` first. Normalize to an absolute path and
   require it to be equal to or inside the target repo root.
3. Check the observed pane command. If it does not identify Codex, skip unless a
   future stronger process-level proof identifies Codex for the pane.
4. Try explicit UUID extraction from process argv:
   - accept `codex resume <uuid>` when the UUID parses;
   - accept wrapper-provided external UUID evidence when the command also
     identifies Codex;
   - ignore raw prompts and unrelated args;
   - do not persist full argv.
5. If explicit UUID is unavailable, scan Codex session metadata:
   - discover candidate JSONL files under `$CODEX_HOME/sessions` if
     `CODEX_HOME` is set, otherwise `~/.codex/sessions`;
   - read only the first `session_meta` JSONL record;
   - extract `payload.session_id` first, then `payload.id` if needed;
   - normalize `payload.cwd`;
   - keep only metadata whose cwd is equal to or inside the target repo root.
6. For `instances create`, correlate by launch timestamp and uniqueness:
   - record the timestamp before starting Codex;
   - after pane creation, wait briefly for a new session file;
   - accept the UUID only if exactly one matching session file appears for the
     opened path in the launch window.
7. For `instances detect`, correlate conservatively:
   - accept explicit `resume <uuid>` as strong;
   - accept a session-file match only if exactly one candidate matches cwd and
     recency policy;
   - otherwise return an unresolved candidate with ambiguity details.
8. Register `active` only when `zellij session`, `zellij pane`,
   `CodexSessionRef` and `OpenedPath` are all resolved.

## Verdicts

| Verdict | Meaning | Registry effect |
| --- | --- | --- |
| `not_codex` | Pane does not provide Codex evidence | No registry write |
| `candidate_unresolved` | Pane likely runs Codex but lacks UUID evidence | No active record; future CLI may show candidate |
| `candidate_ambiguous` | More than one Codex metadata match is plausible | No active record; require manual resolution |
| `active_ready` | All required refs are resolved and duplicate rules pass | Create/update one active record |
| `stale_candidate` | Existing active record no longer validates against live Codex evidence | Reconciliation may mark stale per policy |

`instances list` should not silently promote unresolved candidates. Promotion
belongs to `instances detect`, `instances create`, or a future explicit resolution
command.

## Ambiguity Policy

Ambiguity is expected when several Codex panes run in the same repo.

Rules:

- Multiple matching session files for the same cwd means no automatic
  `CodexSessionRef` unless one match has strong evidence.
- A Codex command without explicit `resume <uuid>` plus several recent session
  files for the same repo is `candidate_ambiguous`.
- A wrapper-provided external UUID may resolve a `CodexSessionRef`, but it must
  keep source semantics separate from Codex `session_meta` identity.
- A stale session file with matching cwd is not enough to claim a live pane.
- Future manual resolution should accept a user-provided Codex UUID and validate
  it against session metadata before registry write.

## Privacy And Safety Rules

- Read only Codex `session_meta` for automatic detection.
- Never store Codex prompts, responses or full argv in `.zelma/instances.json`.
- Process argv may contain prompt text; adapters must extract only safe tokens
  such as binary name, `resume` UUID and `--cd` path.
- Diagnostics may include counts and redacted IDs, but not transcript content.
- Fixtures must be synthetic and contain only metadata required by tests.

## Module Responsibilities

| Module | Responsibilities | Must not do |
| --- | --- | --- |
| `zellij-adapter` | List sessions/panes, expose typed pane facts, command/cwd and pane identity | Parse Codex session files |
| `codex-adapter` | Locate Codex home, parse safe session metadata, extract UUIDs, inspect safe process evidence if available | Write registry or parse zellij layout |
| `detection` | Combine zellij facts and Codex facts into verdicts | Shell out directly or mutate JSON |
| `registry` | Persist validated `active`/`stale` records | Guess missing refs |
| `cli` | Present unresolved/ambiguous candidates and explicit next steps | Hide ambiguity behind success |

## Test Fixtures

Required deterministic fixtures before implementation is considered
regression-covered:

- `zellij list-panes --json --all` output with non-Codex panes.
- Codex pane with command/cwd but no resolvable UUID -> `candidate_unresolved`.
- Codex `resume <uuid>` command -> `active_ready`.
- Codex command with wrapper-provided external UUID -> `active_ready` with
  external source semantics.
- Synthetic `session_meta` JSONL with matching cwd and UUID.
- Multiple matching session files -> `candidate_ambiguous`.
- Create flow launch-window match -> `active_ready`.
- Session metadata with missing/invalid UUID -> unresolved.
- Session metadata whose cwd is outside repo -> ignored.

## Open Questions

- Should a future `zelma instances resolve` command be introduced, or should
  `instances detect --codex-session-id <uuid> --pane-id <pane>` be enough?
- What exact recency window is acceptable for manual detect without explicit
  UUID? This should be tuned with real fixtures after the first adapter exists.
- Should `CodexSessionRef` store both `id` and `session_id` if Codex metadata
  keeps both fields with distinct meaning? Until proven necessary, use one
  canonical UUID and retain source evidence separately.
