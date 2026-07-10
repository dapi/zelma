---
title: "FT-049: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space документ для TUI monitor live zelma sessions."
derived_from:
  - brief.md
  - ../FT-027/design.md
  - ../FT-042/brief.md
  - ../FT-047/brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_049_scope
  - ft_049_acceptance_criteria
  - ft_049_evidence_contract
  - implementation_sequence
---

# FT-049: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `ui-reference/README.md` | Support UI reference | `UI-*`, screen states, key controls and low-fidelity mockups |
| `../FT-027/design.md` | Existing live list contract | `sessions list --live` read-only live status semantics |
| `../FT-042/brief.md` | Existing status backend contract | Versioned dashboard snapshot and recovery hints |
| `../FT-047/brief.md` | Existing focus behavior | `zelma sessions focus <id>` contract |

## Context

FT-049 turns existing machine-readable status/list/focus capabilities into a
human-facing terminal monitor. The design problem is to keep the TUI ergonomic
without making it a new registry owner or a second status backend.

The first viewport must make live/active work primary. Non-active records and
recovery hints remain visible, but they are supporting context rather than the
main queue.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C3` | The feature adds a new CLI/TUI component inside the existing Go CLI container and defines its collaboration with status provider and focus adapter. | C3 component table below |

### C4 Artifact

| Element ID | Component | Responsibility | Collaborates with |
| --- | --- | --- | --- |
| `C4-E01` | `cmd/zelma` / Cobra root | Exposes `zelma monitor` and help routing | `internal/cli` monitor command |
| `C4-E02` | Monitor command | Initializes TUI model, provider, focus adapter and bounded refresh config | Status provider, focus adapter |
| `C4-E03` | Status provider | Returns snapshots using `status.Build` or the same model as `zelma status --json` | `internal/status`, `internal/live`, registry loader |
| `C4-E04` | TUI model/view | Sorts sessions, renders live-first screen states, handles keys | Status provider, focus adapter |
| `C4-E05` | Focus adapter | Focuses selected live session through existing focus contract | `sessions focus` equivalent zellij adapter path |

## Selected Solution

- `SOL-01` Add `zelma monitor` as the single canonical TUI command. It closes
  `REQ-01` and follows issue 103's preferred example while avoiding multiple
  aliases in the first slice.
- `SOL-02` Build a monitor-specific model over `internal/status.Snapshot`
  semantics rather than parsing registry internals from the UI. This closes
  `REQ-02`, `REQ-03`, `REQ-04`, `REQ-05` and `REQ-09`.
- `SOL-03` Render two ordered groups: live/active sessions first, then
  non-active records and degraded/recovery context. Hidden filtering is allowed
  only as an optional toggle; live-first remains the default.
- `SOL-04` Use bounded refresh with the 5s default interval from `SD-07` and a
  manual refresh key. The monitor has no daemon mode in FT-049.
- `SOL-05` Gate focus action to the selected live/active session and delegate
  through the existing focus behavior. Non-live selections produce a readable
  in-TUI error/status message rather than attempting cleanup or repair.
- `SOL-06` Keep TUI screen semantics in `ui-reference/README.md`; code-level
  widget layout and library details stay in the implementation plan.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | `zelma tui` as canonical command | Too broad for issue 103; future TUI surfaces could outgrow the live monitor. |
| `ALT-02` | `zelma sessions monitor` as canonical command | More explicit but longer; issue 103 asks for a quick operational monitor and proposes `zelma monitor`. |
| `ALT-03` | TUI reads `.zelma/sessions.json` directly and reimplements status logic | Rejected by issue 103, FT-027 and FT-042 boundaries. |
| `ALT-04` | Hide stale/non-active records completely by default | Rejected because issue 103 says stale/missing/historical records should remain available. |
| `ALT-05` | Add a background daemon for live refresh | Rejected by issue 103 non-scope and FT-049 `NS-04`. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Use `zelma monitor` only | Short, discoverable command aligned with issue 103 | Users looking under `sessions` need help text to route them |
| `TRD-02` | Reuse status snapshot semantics for the TUI model | Keeps degraded state and recovery hints consistent with FT-042 | UI may need a thin view model to sort/group without changing status backend |
| `TRD-03` | Make non-active records secondary, not hidden | Preserves operational context | Screen can become crowded if many historical records exist; filter/toggle can mitigate |

## Accepted Local Decisions

- `SD-01` `zelma monitor` is the canonical command name for FT-049.
- `SD-02` `internal/status.Snapshot` semantics are the preferred data contract
  for the TUI provider because they already include dashboard status, live
  status and recovery hints.
- `SD-03` The monitor may use an internal provider instead of shelling out to
  `zelma status --json`, but the provider must remain behaviorally equivalent
  to the public status/list contracts.
- `SD-04` The TUI layer owns presentation state only: selected row, visible
  section/filter, last refresh time and transient action messages.
- `SD-05` Focus action is available only when the selected row represents a
  live/active session id; all other selections surface a guarded message.
- `SD-06` Use Bubble Tea as the TUI runtime dependency and keep monitor
  behavior behind fakeable provider/focus seams.
- `SD-07` Use `5s` as the default monitor refresh interval to match existing
  fresh-enough session inventory timing without introducing a new config
  surface in FT-049.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | `zelma monitor` command invocation | User / Cobra command | Opens read-only TUI; no extra flags required for live-first default |
| `CTR-02` | Status snapshot | Status provider / TUI model | Versioned snapshot with summary, session statuses, live status and recovery hints |
| `CTR-03` | Visible row model | TUI model / view | Rows include id, state/dashboard status, live status, opened path, zellij identity and Codex session ref |
| `CTR-04` | Refresh action | Key/timer / status provider | Bounded refresh replaces visible snapshot and may only trigger status/list-defined reconciliation |
| `CTR-05` | Focus action | Selected live row / focus adapter | Delegates to existing focus contract by numeric id and returns readable success/failure |

## Invariants

- `INV-01` The UI layer does not read or parse `.zelma/sessions.json`
  directly.
- `INV-02` Default ordering keeps live/active sessions above stale/non-active
  records.
- `INV-03` Refresh does not perform cleanup, removal or monitor-specific
  registry mutation.
- `INV-04` Degraded status and recovery hints from the provider are never
  silently dropped.
- `INV-05` Focus action never targets a missing, stale or non-live row without
  an explicit guarded failure message.

## Failure Modes

- `FM-01` Status provider cannot inspect zellij; the TUI must render degraded
  state and recovery hints from FT-042.
- `FM-02` Active/stale grouping is wrong; stale work may crowd out live work.
- `FM-03` Refresh loop is too aggressive; terminal monitor consumes excessive
  resources.
- `FM-04` Focus adapter fails; the TUI must display a structured,
  user-readable failure and keep the monitor running.
- `FM-05` TUI code reaches into registry internals; feature violates issue 103
  and should fail boundary review.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add provider/model/action tests and command skeleton | `brief.md` and `design.md` active | Remove `monitor` command route and keep existing CLI unchanged |
| `RB-02` | Add interactive TUI rendering | Provider/model tests pass | Disable `monitor` command before merge if rendering is not stable |
| `RB-03` | Verify local suites and docs | TUI and CLI tests pass | Keep feature branch unmerged until command/help/docs are corrected |

## ADR / External Design Dependencies

| Artifact | Current status | Used for | Rule |
| --- | --- | --- | --- |
| `../FT-027/design.md` | `active` | Live list read-only semantics | Monitor must not persist `live_status` or cleanup stale records |
| `../FT-042/brief.md` | `active` | Status backend and degraded recovery hints | Monitor should preserve snapshot status semantics |
| `../FT-047/brief.md` | `active` | Focus command behavior | Monitor focus must preserve non-mutating focus contract |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `TRD-01`, `SD-01`, `C4-01` | `CTR-01` | `RB-01` |
| `REQ-02` | `SOL-02`, `TRD-02`, `SD-02`, `SD-03` | `CTR-02`, `INV-01` | `FM-05`, `RB-01` |
| `REQ-03` | `SOL-02`, `SOL-03` | `CTR-03`, `INV-02` | `FM-02` |
| `REQ-04` | `SOL-03`, `TRD-03` | `CTR-03`, `INV-02` | `FM-02` |
| `REQ-05` | `SOL-02`, `SOL-06` | `CTR-03` | `RB-02` |
| `REQ-06` | `SOL-04`, `SD-07` | `CTR-04`, `INV-03` | `FM-03` |
| `REQ-07` | `SOL-06`, `SD-04` | `CTR-03` | `RB-02` |
| `REQ-08` | `SOL-05`, `SD-05` | `CTR-05`, `INV-05` | `FM-04` |
| `REQ-09` | `SOL-02`, `TRD-02` | `CTR-02`, `INV-04` | `FM-01` |
| `REQ-10` | `SOL-01`-`SOL-06`, `SD-06` | `CTR-01`-`CTR-05`, `INV-01`-`INV-05` | `RB-03` |
