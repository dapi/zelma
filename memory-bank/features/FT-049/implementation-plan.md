---
title: "FT-049: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-049 без переопределения canonical problem и solution facts."
derived_from:
  - brief.md
  - design.md
  - ui-reference/README.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_049_scope
  - ft_049_selected_design
  - ft_049_acceptance_criteria
  - ft_049_blocker_state
---

# План имплементации

## Цель текущего плана

Добавить `zelma monitor` как read-only terminal monitor поверх существующих
status/list/focus contracts, с live-first rendering, bounded refresh,
keyboard navigation and focus action.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` first |
| `ui-reference/README.md` | support UI reference | `UI-*`, states, controls and mockups | Promote requirement/design changes to canonical owners first |
| `../FT-027/design.md` | live list baseline | read-only live status semantics | Update owning feature only if the live contract changes |
| `../FT-042/brief.md` | status backend baseline | dashboard snapshot and recovery hints | Update owning feature only if status semantics change |
| `../FT-047/brief.md` | focus baseline | focus by numeric id | Update owning feature only if focus contract changes |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `../../../cmd/zelma/main.go` | CLI binary entrypoint | New command must be available through existing root command | Keep Cobra construction in `internal/cli` |
| `../../../internal/cli/cli.go` | Existing Cobra commands, help text, status/list/focus flows | `zelma monitor` routing, help, provider wiring and action behavior belong near current CLI code unless split is justified | Mirror existing command/test style; avoid duplicating registry parsing in TUI |
| `../../../internal/status/snapshot.go` | Dashboard snapshot model | Preferred provider semantics for the monitor | Reuse snapshot fields for live-first view model |
| `../../../internal/live/reconcile.go` | Live reachability reconciliation | Status provider already depends on live reconciliation | Do not invent TUI-specific reconciliation rules |
| `../../../internal/zellij/zellij.go` | Focus adapter implementation | Monitor focus action should preserve existing focus behavior | Reuse existing focus request path or factor shared helper conservatively |
| `../../../internal/cli/cli_test.go` | CLI/help/status/focus tests with fakes | Expected location/pattern for command tests | Add monitor command/action tests in same style or adjacent package |
| `../../../internal/status/snapshot_test.go` | Status snapshot tests | Existing degraded/active/stale semantics | Add monitor model tests against fake snapshots, not live zellij |
| `../../../go.mod` | Dependency manifest | TUI library may be needed | Prefer established Go TUI library only if implementation cannot stay simple without it |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CLI/help command routing | `REQ-01`, `SC-01`, `SOL-01` | Existing help tests cover current command map | Add `monitor` root help and command help tests | `go test ./...` | Same Go suite | N/A | `none` |
| TUI model/render ordering | `REQ-03`, `REQ-04`, `REQ-05`, `REQ-09`, `SC-01`-`SC-04`, `SC-07`, `SOL-02`, `SOL-03` | Status snapshot tests exist, but no TUI ordering tests | Add deterministic tests for active+stale ordering, empty live and degraded hints | `go test ./...` | Same Go suite | Full terminal pixel rendering may remain manual if model/render strings are deterministic | `none` |
| Refresh/navigation/focus actions | `REQ-06`, `REQ-07`, `REQ-08`, `SC-05`, `SC-06`, `NEG-02`, `SOL-04`, `SOL-05` | Focus command tests exist | Add fake provider/focus adapter tests for refresh, selection and guarded focus | `go test ./...` | Same Go suite | Live terminal interaction can be manually spot-checked after automated action tests | `none` |
| Boundary review | `REQ-02`, `NEG-01`, `INV-01` | Existing code separates status/live/registry | Static search/code review for direct UI-layer registry parsing | `rg -n "sessions.json|registry\\.Load|registry\\.Read|os\\.ReadFile" internal/cli internal/monitor` plus code review | N/A unless docs lint exists | N/A | `none` |
| Repo checks | `CHK-05` | Existing required checks | Run memory-bank and whitespace checks | `go test ./...`; `python3 scripts/check_memory_bank_index.py`; `git diff --check` | Same suites in CI when available | N/A | `none` |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Which Go TUI library should be used? | Resolved during implementation. See `decision-log.md` `DL-005` and `design.md` `SD-06`. | `none` | Use Bubble Tea behind fakeable monitor seams. |
| `OQ-02` | What exact refresh interval should be default? | Resolved during implementation. See `decision-log.md` `DL-005` and `design.md` `SD-07`. | `none` | Use `5s` default interval without adding a new config surface in FT-049. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Commands run from repository worktree root | `STEP-01`-`STEP-08` | Relative paths or memory-bank checks resolve wrong files |
| test | Required local checks are `go test ./...`, `python3 scripts/check_memory_bank_index.py` and `git diff --check` | `STEP-07`, `STEP-08` | Verification cannot support `CHK-05` |
| terminal | TUI behavior must be testable through deterministic model/action tests without requiring a live zellij terminal | `STEP-03`-`STEP-06` | Tests become flaky or require manual zellij setup |
| access / network / secrets | No secrets or external services are required; dependency download may use normal Go module network access | `STEP-03`, `STEP-08` | If dependency acquisition is blocked, stop and select a no-new-dependency fallback or human-approved path |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `SD-01`, `CTR-01` | `zelma monitor` remains the canonical command name | `STEP-01`, `STEP-02` | yes |
| `PRE-02` | `SD-02`, `SD-03`, `CTR-02` | Status snapshot semantics remain available for monitor provider | `STEP-02`, `STEP-03`, `STEP-04` | yes |
| `PRE-03` | `SD-05`, `CTR-05` | Existing focus contract remains the delegated action | `STEP-05` | yes |
| `PRE-04` | `OQ-01`, `OQ-02` | Library and refresh defaults are settled within design constraints | `STEP-03`, `STEP-04` | no |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `SOL-01`, `CTR-01` | `zelma monitor` command and help route exist | agent | `PRE-01` |
| `WS-2` | `REQ-02`-`REQ-07`, `REQ-09`, `SOL-02`-`SOL-04`, `SOL-06` | Monitor provider, model, render and refresh/navigation behavior exist | agent | `PRE-02`, `PRE-04` |
| `WS-3` | `REQ-08`, `SOL-05`, `CTR-05` | Focus action delegates by selected live id and handles failure | agent | `PRE-03`, `WS-2` |
| `WS-4` | `REQ-10`, `CHK-01`-`CHK-05` | Automated coverage and required checks exist | agent | `WS-1`-`WS-3` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Implementation needs a background daemon, registry schema change or transcript/pane-buffer reader | `STEP-03`-`STEP-06` | These are explicitly non-scope in `brief.md` | Human approval plus updated feature docs |
| `AG-02` | Chosen TUI dependency materially changes packaging, licensing or runtime constraints | `STEP-03` | Dependency risk may exceed feature scope | Human approval or accepted design update |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `SOL-01` | Add command/help skeleton for `zelma monitor` | `internal/cli/cli.go`, `cmd/zelma/main.go` if needed | Command route and help text | `CHK-01` | `EVID-01` | CLI/help tests | `PRE-01` | `none` | Command naming pressure conflicts with `SD-01` |
| `STEP-02` | agent | `REQ-02`, `SOL-02`, `CTR-02` | Introduce monitor provider interface over status snapshot semantics | `internal/cli`, optional `internal/monitor`, `internal/status` | Fakeable provider | `CHK-02`, `CHK-04` | `EVID-02`, `EVID-04` | Unit tests with fake provider and boundary review | `PRE-02` | `none` | Provider needs direct UI-layer registry parsing |
| `STEP-03` | agent | `REQ-03`-`REQ-05`, `REQ-07`, `REQ-09`, `SOL-03`, `SOL-06` | Implement view model, ordering and render states | `internal/monitor` or equivalent | Deterministic model/render behavior | `CHK-02` | `EVID-02` | Model/render tests | `STEP-02`, `OQ-01` | `AG-02` if dependency risk appears | TUI library requires architecture outside design |
| `STEP-04` | agent | `REQ-06`, `SOL-04`, `CTR-04` | Implement bounded refresh and manual refresh action | `internal/monitor`, `internal/cli` | Refresh behavior | `CHK-03` | `EVID-03` | Fake timer/provider tests | `STEP-03`, `OQ-02` | `none` | Refresh implies background daemon |
| `STEP-05` | agent | `REQ-08`, `SOL-05`, `CTR-05` | Implement guarded focus action | `internal/monitor`, existing focus adapter path | Focus action behavior | `CHK-03` | `EVID-03` | Fake focus adapter tests | `PRE-03`, `STEP-03` | `none` | Focus requires mutating registry or repairing stale records |
| `STEP-06` | agent | `REQ-10`, `INV-01`-`INV-05` | Tighten boundary and failure handling | TUI/provider/action code | Boundary-safe implementation | `CHK-02`, `CHK-03`, `CHK-04` | `EVID-02`, `EVID-03`, `EVID-04` | Static search and unit tests | `STEP-02`-`STEP-05` | `none` | Direct registry parsing or hidden degraded state remains |
| `STEP-07` | agent | `REQ-10` | Update docs/help if command behavior changed during implementation | `README.md`, `memory-bank/features/FT-049/*` as needed | Consistent docs | `CHK-01`, `CHK-05` | `EVID-01`, `EVID-05` | Docs review and memory-bank audit | `STEP-01`-`STEP-06` | `none` | Docs need new requirements or design facts |
| `STEP-08` | agent | `CHK-05` | Run required checks | Full repo | Test/check output | `CHK-05` | `EVID-05` | `go test ./...`; `python3 scripts/check_memory_bank_index.py`; `git diff --check` | `STEP-01`-`STEP-07` | `none` | Failing checks reveal scope/design contradiction |

## Parallelizable Work

- `PAR-01` CLI/help tests and monitor model tests can be developed in parallel
  after `STEP-02` defines the provider seam.
- `PAR-02` Focus action tests should wait until row selection semantics from
  `STEP-03` are stable.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `CHK-01` | Command is discoverable and documented in help | `EVID-01` |
| `CP-02` | `STEP-02`, `STEP-03`, `CHK-02`, `CHK-04` | Provider/model/render preserve status boundary and live-first ordering | `EVID-02`, `EVID-04` |
| `CP-03` | `STEP-04`, `STEP-05`, `CHK-03` | Refresh/navigation/focus actions are deterministic and guarded | `EVID-03` |
| `CP-04` | `STEP-08`, `CHK-05` | Required local checks pass | `EVID-05` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | TUI library adds broad dependency or testing friction | Scope expands beyond first slice | Keep TUI behind fakeable model/provider/actions and use `AG-02` if risk is material | Dependency requires runtime assumptions not in `go.mod` today |
| `ER-02` | UI code duplicates status/list logic | Drift from FT-027/FT-042 contracts | Provider seam over status snapshot; boundary static search | Tests need registry fixtures in UI layer |
| `ER-03` | Live terminal behavior is hard to automate | Acceptance becomes manual-heavy | Test model/render/action deterministically with fakes | Tests require real zellij terminal |
| `ER-04` | Many non-active records crowd screen | Live-first objective weakens | Default grouping plus secondary toggle/filter | Render tests show non-active before live rows |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `NS-03`, `INV-01`, `NEG-01` | Implementation requires new registry schema or UI-layer direct registry parsing | Stop and update design; do not ship TUI | Existing CLI/status/focus behavior unchanged |
| `STOP-02` | `NS-04`, `AG-01` | Implementation requires a daemon/background service | Stop and request human approval | Feature remains documentation/planning only |
| `STOP-03` | `NS-01`, `AG-01` | Implementation requires transcript or pane-buffer reader | Stop and route to issue 102 scope | Monitor excludes observation commands |
| `STOP-04` | `INV-05`, `NEG-02` | Focus cannot be guarded to live/active selected rows | Stop and redesign focus affordance | Monitor can render read-only rows without focus action |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | TUI dependency decision note | implementer / reviewer | `artifacts/ft-049/plan/dependency-decision/` | `CP-02` |
| `EVID-10` | Simplify-review verdict for monitor model/actions | implementer / reviewer | `artifacts/ft-049/plan/simplify-review/` | `CP-03` |

## Готово для приемки

- Все workstreams завершены или явно остановлены через `STOP-*`.
- Все checkpoints имеют evidence.
- Required local suites зелёные, а CI не противоречит local verify.
- Manual-only gaps закрыты через approved `AG-*` или остаются blockers для
  `delivery_status: done`.
- Support docs не расходятся с canonical `brief.md`, `design.md` и этим планом.
- Финальная приемка идет по `brief.md` `Verify`, а не по этому checklist.
