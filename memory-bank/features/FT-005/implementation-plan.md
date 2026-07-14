---
title: "FT-005: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-005. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical problem и solution фактов."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_005_scope
  - ft_005_selected_design
  - ft_005_acceptance_criteria
  - ft_005_blocker_state
---

# FT-005: Implementation Plan

## Цель текущего плана

Реализовать централизованный repo root resolver для Git worktree root и
подготовить CLI-visible unsupported-repo diagnostic, не добавляя registry IO,
`.gitignore` mutation или zellij behavior.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem / verify owner | `REQ-*`, `NS-*`, `SC-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` first |
| `decision-log.md` | decision evidence | `DL-001` rationale for Git worktree boundary | Update decision log and `design.md` if rationale changes |
| `../../adr/ADR-001-mvp-cli-architecture.md` | architecture owner | `internal/repo` owns root and `.zelma/` paths | Update ADR or create accepted ADR first |
| `../../engineering/architecture.md` | engineering boundary owner | centralized root detection; CLI error expectations | Update architecture first |
| `../../ops/config.md` | config owner | default registry path under detected root | Update config first for env overrides |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `cmd/zelma/main.go` | Binary entrypoint | Existing command path for future diagnostics | Keep command exit behavior through `internal/cli.Run` |
| `internal/cli/cli.go` | Minimal CLI runner | Place where unsupported-repo diagnostics can become visible when commands require repo context | Keep stdout/stderr discipline from architecture |
| `go.mod` | Go module root for implementation repo | Current codebase has no dependencies for resolver yet | Prefer standard library unless Git probing needs a small helper |
| `memory-bank/adr/ADR-001-mvp-cli-architecture.md` | Accepted architecture | Defines `internal/repo` boundary | Mirror package ownership |
| `memory-bank/engineering/architecture.md` | Error/config rules | Requires centralized root detection and agent-friendly errors | Reuse failure handling language |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Git worktree root discovery | `REQ-01`, `REQ-02`, `SC-01`, `CHK-01`, `SOL-01`, `SOL-02`, `CTR-01` | none | Unit tests with temp Git repository and nested directories | `go test ./...` | none yet | none | `none` |
| Unsupported repo diagnostic | `REQ-03`, `SC-02`, `NEG-01`, `CHK-02`, `SOL-03`, `CTR-02`, `FM-01` | none | Unit test for resolver error and CLI/error rendering where command surface exists | `go test ./...` | none yet | End-to-end CLI command may wait until a repo-requiring command exists | `none` |
| Non-scope guard | `NS-01`, `NS-03`, `NS-04`, `INV-02`, `INV-03` | none | Static search/code review | `rg -n "instances.json|zellij|\\.gitignore" internal cmd` | none yet | review remains manual for mutation semantics | `none` |

## Open Questions / Ambiguities

No unresolved `OQ-*` blockers remain for this plan.

| Resolved question | Resolution owner | Result |
| --- | --- | --- |
| Which repo marker defines supported repo? | `decision-log.md#dl-001-supported-repo-boundary`, `design.md#accepted-local-decisions` | Use Git worktree root. |
| Whether to support non-Git directories in FT-005 | `TRD-01`, `STOP-01` | Defer outside FT-005 unless human-approved scope changes. |
| Whether to add environment override for root or registry path | `../../ops/config.md`, `AG-01` | Do not add new override in FT-005. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Go toolchain available through repo-standard command environment | all steps | `go test ./...` cannot run |
| test fixtures | Tests can create temporary directories and initialize Git metadata | `STEP-02`, `STEP-03` | Root discovery tests cannot construct a Git worktree |
| external commands | No live `zellij`, Codex or registry file is required | all steps | Implementation tries to call external runtime |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `SD-01` | Git worktree is accepted as supported repo boundary | all steps | yes |
| `PRE-02` | `SD-02` | `internal/repo` is the owner of root discovery | `STEP-01`, `STEP-02` | yes |
| `PRE-03` | `INV-02`, `INV-03` | No registry or `.gitignore` mutation in this feature | `STEP-04` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `REQ-02`, `SOL-01`, `SOL-02`, `CTR-01` | `internal/repo` resolver API and tests | agent | `PRE-01`, `PRE-02` |
| `WS-2` | `REQ-03`, `SOL-03`, `CTR-02`, `FM-01` | typed unsupported-repo error and CLI diagnostic path | agent | `WS-1` |
| `WS-3` | `NS-01`, `NS-03`, `NS-04`, `INV-02`, `INV-03` | non-scope guard evidence | agent | `WS-1`, `WS-2` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Need to support non-Git directories or a new env override | `WS-1`, `WS-2` | Would change accepted `SD-01` or configuration ownership | Human approval plus updated `brief.md`/`design.md`/`ops/config.md` |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `SOL-01`, `SD-02` | Add `internal/repo` package API | `internal/repo` | resolver interface/function | `CHK-01` | `EVID-01` | code review + tests in `STEP-02` | `PRE-01`, `PRE-02` | none | implementation requires registry/zellij state |
| `STEP-02` | agent | `REQ-01`, `REQ-02`, `SOL-01`, `SOL-02`, `CTR-01` | Implement and test nested Git worktree root detection | `internal/repo` | passing resolver tests | `CHK-01` | `EVID-01` | `go test ./...` | `STEP-01` | none | temp Git fixture cannot be created |
| `STEP-03` | agent | `REQ-03`, `SOL-03`, `CTR-02`, `FM-01`, `FM-02` | Implement typed unsupported-repo failure and diagnostic mapping | `internal/repo`, `internal/cli` if needed | error contract and tests | `CHK-02` | `EVID-02` | `go test ./...` | `STEP-02` | none | CLI command surface is insufficient to expose diagnostic |
| `STEP-04` | agent | `NS-01`, `NS-03`, `NS-04`, `INV-02`, `INV-03` | Confirm no registry, `.gitignore` or zellij side effects | `cmd`, `internal` | review note | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` | `rg -n "instances.json|zellij|\\.gitignore" internal cmd` | `STEP-03` | `AG-01` if scope expansion is needed | side-effect behavior appears necessary |

## Parallelizable Work

- `PAR-01` Resolver implementation and documentation review can be reviewed
  separately after `design.md` is stable.
- `PAR-02` `WS-2` should wait for `WS-1` because diagnostics depend on typed
  resolver errors.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `CHK-01`, `SOL-01`, `SOL-02` | Nested directory resolves to stable normalized Git worktree root | `EVID-01` |
| `CP-02` | `STEP-03`, `CHK-02`, `SOL-03` | Outside-repo failure is distinguishable and agent-friendly | `EVID-02` |
| `CP-03` | `STEP-04`, `INV-02`, `INV-03` | Non-scope side effects absent | `EVID-01`, `EVID-02` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Git worktree probing differs across normal repos and linked worktrees | `SC-01` may pass only for simple repos | Include linked-worktree behavior if local fixture support is practical; otherwise record follow-up evidence gap before Done |
| `ER-02` | CLI has no repo-requiring subcommand yet | End-to-end diagnostic may be premature | Test resolver error directly and wire CLI diagnostic when first repo-requiring command lands |
| `ER-03` | Scope pressure to initialize `.zelma/` or edit `.gitignore` | FT-005 would absorb FT-031 or registry scope | Stop through `AG-01` and update feature ownership |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `AG-01`, `SD-01` | Non-Git support becomes required | Stop implementation and update upstream docs with human approval | Resolver remains Git-only or unmerged |
| `STOP-02` | `INV-02`, `INV-03` | Implementation needs registry IO or `.gitignore` mutation | Stop and move behavior to owning feature | FT-005 contains only resolver behavior |
| `STOP-03` | `FM-02` | Filesystem probing failure cannot be distinguished from unsupported repo | Stop until error taxonomy is explicit in design | No ambiguous diagnostic shipped |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Simplify-review verdict and non-scope review note | implementer | PR/commit notes or `artifacts/ft-005/` | `CP-03` |

## Готово для приемки

- All workstreams complete or explicitly stopped through `STOP-*`.
- `go test ./...` passes for resolver and diagnostic coverage.
- Static non-scope search does not show registry, `.gitignore` or zellij side effects introduced by FT-005.
- Any linked-worktree gap is either covered by automated tests or documented before `delivery_status: done`.
- Final acceptance follows `brief.md` `Verify`.
