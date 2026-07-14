---
title: "FT-049: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "Feature-local журнал решений для FT-049. Фиксирует FPF-обоснования review-improve без владения scope, selected design или execution sequencing."
derived_from:
  - brief.md
  - design.md
  - implementation-plan.md
  - ../../flows/feature-flow.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_049_scope
  - ft_049_selected_design
  - ft_049_acceptance_criteria
  - implementation_sequence
---

# FT-049: Decision Log

Этот журнал фиксирует решения, принятые во время подготовки feature package. Он
не подменяет canonical owners: scope и verify живут в `brief.md`, solution facts
живут в `design.md`, execution sequencing живет в `implementation-plan.md`.

## DL-001: Promote Issue 103 To FT-049 Feature Package

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | 1 |
| Closed question | Можно ли создать feature package для issue 103 без human gate? |
| FPF frame | Bounded contexts + evidence graph |

### Available Facts

- GitHub issue 103 is open and asks for a read-only TUI monitor for live zelma
  sessions.
- The repository has feature packages through `FT-048`; no existing feature
  package references issue 103 or this TUI monitor.
- `../../flows/feature-flow.md` requires feature work to live in
  `memory-bank/features/FT-XXX/`.
- Issue 103 references existing upstream contracts: `UC-001`, `UC-010`,
  `FT-042`, `README.md` list defaults and issue 102 non-scope.

### Decision

Create `memory-bank/features/FT-049/` with `README.md`, `brief.md`,
`design.md`, `ui-reference/README.md`, `implementation-plan.md` and this
`decision-log.md`.

### Rationale

The missing feature package is a process/documentation gap. Upstream facts are
sufficient to define the delivery unit boundaries: a TUI monitor over existing
session status/list/focus contracts, not a new registry schema, web dashboard,
daemon or transcript reader.

### Human Gate

None.

## DL-002: Use `zelma monitor` As The Canonical Command

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | 1 |
| Closed question | Which command name should the TUI feature standardize on? |
| FPF frame | Abductive-deductive-inductive reasoning cycle |

### Available Facts

- Issue 103 proposes `zelma monitor` and allows naming to be decided during
  design among examples like `zelma tui`, `zelma instances monitor` or
  `zelma monitor`.
- The first screen's question is operational and broad: which sessions are live
  and worth attention right now?
- Existing root help already has top-level user workflow commands such as
  `zelma status` and nested session management commands under
  `zelma instances`.

### Alternatives

| Alternative | Fit | Rejection / selection reason |
| --- | --- | --- |
| `zelma monitor` | selected | Short, issue-aligned, and names the operational surface rather than an implementation technology. |
| `zelma tui` | rejected | Names the UI technology, not the live-session monitoring outcome. |
| `zelma instances monitor` | rejected for first slice | Accurate but longer; can be added later only if product wants nested aliases. |

### Decision

Set `zelma monitor` as the canonical command in `brief.md` and `design.md`.

### Rationale

FPF reasoning: propose the command that best fits the issue wording, derive the
consequence that help text must route users from root command map and session
context, then verify through CLI/help tests. This choice is supported by issue
103 and does not require inventing additional product facts.

### Human Gate

None.

## DL-003: Design Required And C3 Is The Minimum C4 Level

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | 1 |
| Closed question | Does FT-049 need `design.md`, and what C4 level is required? |
| FPF frame | Bounded context + assurance boundary classification |

### Available Facts

- `../../flows/feature-flow.md` requires `design.md` when a feature changes a
  CLI contract, UI surface, integration contract or internal component
  boundary.
- Issue 103 adds a new TUI command, keyboard navigation, refresh/polling,
  focus action and a UI data-source boundary.
- The implementation is still inside the existing Go CLI container and does not
  add a new deployable container, queue, storage engine or external service.

### Decision

Set `Design required: yes` and use a C3 component table in `design.md`.

### Rationale

The feature changes internal collaboration among CLI command, provider,
TUI model/view and focus adapter. C3 is enough to make those component
responsibilities and data/action directions explicit; C2 would overstate the
change as a new runtime container, and C4 would prematurely decide class-level
implementation.

### Human Gate

None.

## DL-004: Prefer Status Snapshot Semantics For The TUI Provider

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | 1 |
| Closed question | Should the TUI provider primarily consume status backend semantics or live list semantics? |
| FPF frame | Evidence graph + trust/assurance separation |

### Available Facts

- Issue 103 allows the TUI to consume `zelma status --json`,
  `zelma instances list --live --json`, or an internal service behind those same
  contracts.
- FT-042 status backend returns a versioned snapshot with dashboard status,
  live status and recovery hints.
- FT-027 live list owns live/unreachable reachability semantics but treats
  transient adapter errors as command errors.
- Issue 103 explicitly requires degraded states and recovery hints when zellij
  is unavailable or a pane cannot be revalidated.

### Alternatives

| Alternative | Fit | Rejection / selection reason |
| --- | --- | --- |
| Status snapshot semantics | selected | Already contains active/stale grouping inputs, degraded state and recovery hints required by issue 103. |
| Live list JSON only | rejected as primary | Good for reachability, but does not own the richer dashboard status/recovery hint contract. |
| Direct registry parsing | rejected | Explicitly forbidden by issue 103 and feature boundaries. |

### Decision

Use `internal/status.Snapshot` semantics as the preferred TUI provider contract.
The implementation may call an internal service instead of shelling out to
`zelma status --json`, but it must remain behaviorally equivalent to the public
status/list contracts.

### Rationale

The status snapshot has stronger fit to the monitor's first-screen and degraded
state requirements. The decision keeps design-time claims separate from runtime
evidence: the TUI may sort and render presentation state, but status and
recovery claims remain owned by the status/list contracts.

### Human Gate

None.

## DL-005: Use Bubble Tea With A 5s Default Refresh

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | implementation |
| Closed question | Which TUI runtime and default refresh interval should FT-049 use? |
| FPF frame | Options comparison + assurance boundary classification |

### Available Facts

- The repository had no existing TUI dependency before FT-049.
- `implementation-plan.md` required deterministic tests for provider, render,
  navigation, refresh and focus behavior without requiring a live terminal.
- `brief.md` requires bounded refresh/polling and a manual refresh key, but does
  not require a new config surface.
- Existing session list auto-detect uses a default fresh-enough TTL of `5s`,
  documented in `README.md`.

### Alternatives

| Alternative | Fit | Rejection / selection reason |
| --- | --- | --- |
| Bubble Tea with fakeable monitor seams | selected | Provides a proven Go TUI runtime while keeping behavior testable in `internal/monitor`. |
| Hand-rolled raw terminal handling | rejected | Would increase terminal edge-case risk without improving feature-specific behavior. |
| Add a new configurable refresh option | rejected for FT-049 | Issue 103 requires bounded refresh, not a new config contract. |

### Decision

Use Bubble Tea for the interactive runtime, with monitor behavior isolated
behind fakeable status provider and focus adapter seams. Use `5s` as the
default refresh interval.

### Rationale

This closes the execution ambiguity without expanding scope. Bubble Tea handles
terminal interaction, while FT-049 tests focus on the feature-owned model and
action semantics. A `5s` interval is conservative, bounded and consistent with
the existing fresh-enough inventory timing.

### Human Gate

None.
