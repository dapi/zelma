---
title: "FT-049: TUI Monitor For Live Sessions"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для delivery-единицы, добавляющей human-friendly read-only TUI monitor для live zelma sessions."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../../use-cases/UC-001-agent-session-inventory.md
  - ../../use-cases/UC-010-agent-dashboard-status-backend.md
  - ../FT-027/brief.md
  - ../FT-042/brief.md
  - ../FT-045/brief.md
  - ../FT-047/brief.md
status: active
delivery_status: in_progress
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-049: TUI Monitor For Live Sessions

## What

### Problem

Issue 103 notes that `zelma sessions list --live --json` and
`zelma status --json` already expose machine-readable session state, but a
human operator has no live terminal monitor that quickly answers: which sessions
are running and worth attention now?

Plain list/status output is useful for agents and scripts, but it does not give
an interactive first screen with live work visually and navigationally primary.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Live-first TUI availability | No TUI monitor command | `zelma monitor` opens a read-only TUI with live/active sessions first by default | CLI/help review and TUI rendering tests |
| `MET-02` | Active/stale prioritization | Mixed records require interpreting JSON/table output | Mixed active and stale snapshot renders live/active records above non-active records by default | Deterministic render/order tests |
| `MET-03` | Reuse of existing contracts | UI could drift into registry internals | TUI consumes status/list-equivalent service contracts and does not parse `.zelma/sessions.json` directly in the UI layer | Code review/static search and tests with fake provider |

### Scope

- `REQ-01` Add one canonical TUI command, `zelma monitor`, and document it in
  CLI help.
- `REQ-02` Render a terminal UI over `zelma status --json`,
  `zelma sessions list --live --json`, or an internal service behind those same
  contracts; the UI layer must not parse `.zelma/sessions.json` directly.
- `REQ-03` Make live/active sessions visually and navigationally primary in the
  default view.
- `REQ-04` Keep stale, blocked, completed or otherwise non-active records
  visible through a secondary section, filter or toggle without placing them
  above live work by default.
- `REQ-05` Show enough session identity to choose the right work item:
  repo-local numeric id, state/dashboard status, live status, opened path,
  zellij session/tab/pane and Codex session ref when available.
- `REQ-06` Support bounded refresh/polling plus a manual refresh key.
- `REQ-07` Support keyboard navigation over visible sessions.
- `REQ-08` Focus the selected live session through existing
  `zelma sessions focus <id>` behavior or an equivalent internal adapter that
  preserves that command contract.
- `REQ-09` Surface degraded states and recovery hints from the status backend
  when zellij is unavailable or a pane cannot be revalidated.
- `REQ-10` Add automated tests for rendering/order logic and command/action
  behavior with fake status provider and fake focus adapter.

### Non-Scope

- `NS-01` No transcript or pane-buffer reader; issue 102 owns future read-only
  observation commands.
- `NS-02` No web dashboard.
- `NS-03` No new registry schema just for the TUI.
- `NS-04` No background daemon requirement for the first slice.
- `NS-05` No direct registry mutation from refresh except behavior already
  defined by status/list reconciliation contracts.
- `NS-06` No automatic cleanup, removal or stale-state repair from the monitor.

### Constraints / Assumptions

- `ASM-01` GitHub issue 103 is the tracker source for this delivery unit.
- `ASM-02` FT-042 status backend is implemented and exposes a versioned
  snapshot with dashboard status, live status and recovery hints.
- `ASM-03` FT-047 focus command is implemented and provides the existing focus
  behavior by repo-local numeric id.
- `CON-01` The TUI must remain read-only for session data: refresh may invoke
  status/list reconciliation, but the monitor itself must not write registry
  records or perform cleanup.
- `CON-02` The first viewport must answer the operational question "which
  sessions are live and worth attention right now?"
- `CON-03` Terminal refresh must be bounded; unbounded tight loops are outside
  the accepted operating model.

No unresolved blocking problem-space decisions remain after
`decision-log.md` entries `DL-001` through `DL-004`.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature adds a new CLI/TUI surface, keyboard actions, refresh behavior and a UI/backend contract over existing status/focus boundaries. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` `zelma monitor` appears in root help and opens the TUI without extra
  flags.
- `EC-02` Live/active sessions render before stale/non-active records in the
  default view.
- `EC-03` With no live sessions, the TUI shows an empty-live state plus known
  stale/non-active records or recovery hints.
- `EC-04` The UI layer does not parse `.zelma/sessions.json` directly.
- `EC-05` Selecting a live session and invoking focus delegates to the existing
  focus contract and reports a user-readable failure on focus errors.
- `EC-06` Refresh is bounded and does not mutate registry state beyond
  already-defined status/list reconciliation behavior.
- `EC-07` Rendering/order and command/action behavior are covered by automated
  tests with fakes.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01` | `EC-01`, `SC-01` | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` |
| `REQ-02` | `ASM-02`, `CON-01` | `EC-04`, `SC-04`, `NEG-01` | `CHK-02`, `CHK-04` | `EVID-02`, `EVID-04` |
| `REQ-03` | `CON-02` | `EC-02`, `SC-01`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-04` | `CON-02` | `EC-02`, `EC-03`, `SC-02`, `SC-03` | `CHK-02` | `EVID-02` |
| `REQ-05` | `ASM-02` | `SC-01`, `SC-02`, `SC-03` | `CHK-02` | `EVID-02` |
| `REQ-06` | `CON-03` | `EC-06`, `SC-05` | `CHK-03` | `EVID-03` |
| `REQ-07` | `CON-02` | `SC-01`, `SC-05` | `CHK-03` | `EVID-03` |
| `REQ-08` | `ASM-03` | `EC-05`, `SC-06`, `NEG-02` | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` |
| `REQ-09` | `ASM-02` | `EC-03`, `SC-03`, `SC-07` | `CHK-02` | `EVID-02` |
| `REQ-10` | `ASM-01` | `EC-07` | `CHK-01`, `CHK-02`, `CHK-03`, `CHK-04`, `CHK-05` | `EVID-01`-`EVID-05` |

### Acceptance Scenarios

- `SC-01` Running `zelma monitor` with live sessions shows live/active sessions
  first without requiring extra flags.
- `SC-02` With mixed active and stale records, active/live sessions remain
  visually and navigationally primary; stale records do not appear above live
  work by default.
- `SC-03` With no live sessions, the TUI shows an empty-live state and still
  exposes stale/non-active records or recovery hints.
- `SC-04` The TUI obtains session data through status/list-equivalent contracts
  and not through direct UI-layer registry parsing.
- `SC-05` Manual refresh and bounded polling update the visible snapshot without
  a tight loop or unbounded background daemon.
- `SC-06` Selecting a live session and invoking focus switches to the stored
  zellij tab/pane or reports a structured, user-readable failure.
- `SC-07` When zellij is unavailable or pane revalidation degrades, the TUI
  displays status backend recovery hints instead of hiding the problem.

### Negative / Edge Scenarios

- `NEG-01` A TUI implementation reads `.zelma/sessions.json` directly in the UI
  layer; the feature must be rejected.
- `NEG-02` Focus is offered for a non-live or missing selected session without a
  guarded user-readable failure; the feature must be rejected.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `REQ-01` | CLI/help tests for root command map and `zelma monitor --help` if present | Canonical command is documented and discoverable | `artifacts/ft-049/verify/chk-01/` |
| `CHK-02` | `EC-02`, `EC-03`, `SC-01`-`SC-04`, `SC-07` | Deterministic TUI model/render tests with fake active/stale/degraded snapshots | Ordering, empty-live state and recovery hint rendering match contract | `artifacts/ft-049/verify/chk-02/` |
| `CHK-03` | `EC-05`, `EC-06`, `SC-05`, `SC-06`, `NEG-02` | TUI command/action tests with fake provider and fake focus adapter | Refresh is bounded; focus delegates by selected live id and errors are readable | `artifacts/ft-049/verify/chk-03/` |
| `CHK-04` | `EC-04`, `NEG-01` | Code review/static search for direct registry parsing from TUI layer | UI layer uses status/list-equivalent provider only | `artifacts/ft-049/verify/chk-04/` |
| `CHK-05` | `EC-07` | Run `go test ./...`, `python3 scripts/check_memory_bank_index.py` and `git diff --check` | All required local checks pass | `artifacts/ft-049/verify/chk-05/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-049/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-049/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-049/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-049/verify/chk-04/` |
| `CHK-05` | `EVID-05` | `artifacts/ft-049/verify/chk-05/` |

### Evidence

- `EVID-01` CLI/help test output for `zelma monitor`.
- `EVID-02` TUI model/render test output for live-first, mixed stale and
  degraded snapshots.
- `EVID-03` TUI action test output for refresh, navigation and focus.
- `EVID-04` Boundary review/static search showing no UI-layer direct registry
  parsing.
- `EVID-05` Required local check output for `go test ./...`,
  `python3 scripts/check_memory_bank_index.py` and `git diff --check`.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | CLI/help test output | implementer / CI | `artifacts/ft-049/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Render/order test output | implementer / CI | `artifacts/ft-049/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Action behavior test output | implementer / CI | `artifacts/ft-049/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Static search and code review note | implementer / reviewer | `artifacts/ft-049/verify/chk-04/` | `CHK-04` |
| `EVID-05` | Local command output summary | implementer / CI | `artifacts/ft-049/verify/chk-05/` | `CHK-05` |
