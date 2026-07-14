---
title: "FT-002: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-002. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical problem или solution facts."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_002_scope
  - ft_002_selected_design
  - ft_002_acceptance_criteria
  - ft_002_blocker_state
---

# FT-002: Implementation Plan

## Цель текущего плана

Реализовать Cobra command tree для `zelma setup` и
`zelma instances list/create/detect` как routed stubs, сохранив отсутствие
registry и live `zellij` side effects.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem / verify owner | `REQ-*`, `NS-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` or ADR first |
| `../../adr/ADR-001-mvp-cli-architecture.md` | accepted architecture owner | Go CLI, Cobra, command/application/adapter separation | Update ADR before changing architecture constraints |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `cmd/zelma/main.go` | Thin binary entrypoint | Delegates args/stdout/stderr to `internal/cli.Run` | Keep entrypoint thin per `CTR-03` |
| `internal/cli/cli.go` | Current CLI package | `Run` currently returns 0 and owns the likely command tree location | Add Cobra command construction here or adjacent internal CLI files |
| `go.mod` | Go module declaration | Module exists as `github.com/dapi/zelma`, Go version `1.25` | Add Cobra dependency through normal Go module tooling |
| `memory-bank/features/FT-001/implementation-plan.md` | Prior scaffold execution pattern | Shows accepted local checks for test/build/side-effect boundary | Mirror grounded plan structure |
| `memory-bank/adr/ADR-001-mvp-cli-architecture.md` | Architecture baseline | Requires Cobra and handler/adapters separation | Keep registry and zellij adapters out of FT-002 |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Cobra route availability | `REQ-01`, `REQ-02`, `SC-01`, `CHK-01`, `SOL-01`, `SOL-02`, `SOL-03`, `CTR-01` | none in current scaffold | Add Go tests for `setup --help` and `instances list/create/detect --help` route execution | `go test ./...` | none configured | none | `none` |
| Stub diagnostics | `REQ-03`, `SC-02`, `CHK-02`, `SOL-04`, `SD-01`, `CTR-02` | none in current scaffold | Add Go tests for deterministic non-implemented stub behavior | `go test ./...` | none configured | Exact final output contract beyond stub diagnostics belongs to FT-004 | `none` |
| Side-effect boundary | `REQ-03`, `NEG-01`, `CHK-02`, `INV-01`, `INV-02`, `FM-01` | FT-001 static check only | Keep FT-002 code free of registry/zellij adapters; verify by tests and static search | `go test ./...`; `rg -n "zellij|instances.json|\\.zelma" cmd internal` | none configured | Static review remains acceptable because no adapter exists in this slice | `none` |
| Binary build | `ASM-01`, `CTR-03` | FT-001 scaffold builds | Ensure Cobra dependency and command tree compile | `go build ./cmd/zelma` | none configured | none | `none` |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Exact final agent-first help copy | Owned by FT-003, not FT-002 | none | Do not assert final help copy; assert route availability only |
| `OQ-02` | Broader output/error contract for all commands | Owned by FT-004 | none | Keep FT-002 diagnostics deterministic but narrow to non-implemented stubs |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Go toolchain available in `PATH` | all steps | `go version` fails |
| dependency | Network or existing module cache can fetch `github.com/spf13/cobra` if missing | `STEP-02` | `go get` / `go mod tidy` cannot resolve Cobra |
| test | `go test ./...` from repo root is authoritative local test command | `CHK-01`, `CHK-02` | command fails or skips changed package unexpectedly |
| build | `go build ./cmd/zelma` from repo root must succeed | `CTR-03` | binary build fails |
| side-effect review | Static search covers forbidden runtime strings until adapters exist | `CHK-02` | search reveals registry or zellij behavior introduced in `cmd` or `internal` |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-01` | FT-001 scaffold exists and Go can run tests/build | all steps | yes |
| `PRE-02` | `../../adr/ADR-001-mvp-cli-architecture.md` | ADR-001 remains `decision_status: accepted` | `STEP-02`, `STEP-03` | yes |
| `PRE-03` | `design.md` `C4-00` | No C4 artifact required for this local command routing change | `STEP-02` | no |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`, `SOL-01`, `SOL-02`, `CTR-01`, `CTR-03` | Root command and `setup` route exist | agent | `PRE-01`, `PRE-02` |
| `WS-2` | `REQ-02`, `SOL-03`, `CTR-01` | `instances` command group and subcommand routes exist | agent | `WS-1` |
| `WS-3` | `REQ-03`, `SOL-04`, `SD-01`, `CTR-02`, `INV-01`, `INV-02` | Stubs return deterministic diagnostics without side effects | agent | `WS-1`, `WS-2` |
| `WS-4` | `CHK-01`, `CHK-02` | Tests and verification evidence cover routing and side-effect boundary | agent | `WS-1`, `WS-2`, `WS-3` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Need to implement registry writes, `.gitignore` mutation, live `zellij`, or finalized help/output contracts | `STEP-02` through `STEP-05` | Would exceed FT-002 scope or another feature owner | User confirmation plus updated upstream owner document |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `PRE-01` | Confirm scaffold and environment | repo root | discovery note / command output | `CHK-01` | `EVID-01` | `go test ./...` | none | none | Go toolchain unavailable |
| `STEP-02` | agent | `REQ-01`, `SOL-01`, `SOL-02`, `CTR-01`, `CTR-03` | Add Cobra root command and `setup` route | `internal/cli/*`, `go.mod`, `go.sum` | command construction code | `CHK-01` | `EVID-01` | Go command tests | `PRE-01`, `PRE-02` | `AG-01` if behavior expands | Cobra architecture is no longer accepted |
| `STEP-03` | agent | `REQ-02`, `SOL-03`, `CTR-01` | Add `instances` group and `list/create/detect` routes | `internal/cli/*` | routed subcommands | `CHK-01` | `EVID-01` | Go command tests | `STEP-02` | `AG-01` if behavior expands | command names conflict with upstream docs |
| `STEP-04` | agent | `REQ-03`, `SOL-04`, `SD-01`, `CTR-02`, `INV-01`, `INV-02` | Add deterministic non-implemented stub behavior | `internal/cli/*` | stub handlers | `CHK-02` | `EVID-02` | Go tests plus side-effect search | `STEP-03` | `AG-01` if real behavior is needed | tests require registry or zellij |
| `STEP-05` | agent | `CHK-01`, `CHK-02` | Run final verification | repo root | verification output | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` | `go test ./...`; `go build ./cmd/zelma`; `rg -n "zellij|instances.json|\\.zelma" cmd internal` | `STEP-04` | none | local checks cannot produce trustworthy evidence |

## Parallelizable Work

- `PAR-01` Route tests for `setup` and `sessions` subcommands can be written in
  parallel with command construction if expected route names are taken from
  `brief.md`.
- `PAR-02` Side-effect static review should run after code changes, not in
  parallel with them, because it verifies the final changed surface.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-02`, `STEP-03`, `CHK-01`, `SOL-01`, `SOL-02`, `SOL-03` | All requested routes exist and help routes execute | `EVID-01` |
| `CP-02` | `STEP-04`, `CHK-02`, `SOL-04`, `INV-01`, `INV-02` | Stub execution is deterministic and has no registry/zellij side effects | `EVID-02` |
| `CP-03` | `STEP-05`, `CTR-03` | Full local test/build verification passes | `EVID-01`, `EVID-02` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Tests lock down default Cobra help wording too tightly | Collides with FT-003 help template ownership | Assert route availability and command identity only | Help snapshot assertions appear |
| `ER-02` | Stub code starts real setup/session behavior | Violates `REQ-03`, `NS-01`, `NS-02`, `NS-04` | Stop via `AG-01`; move behavior to owning feature | `.zelma`, `instances.json`, `.gitignore` or `zellij` logic appears |
| `ER-03` | Cobra dependency cannot be fetched | Blocks implementation | Stop with environment evidence; do not hand-edit vendored dependency | `go get` / `go mod tidy` fails due to environment |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `PRE-01` | Go toolchain unavailable | Stop before code changes | FT-001 scaffold remains unchanged |
| `STOP-02` | `NS-01`, `NS-02`, `NS-04`, `AG-01` | Implementation requires registry, `.gitignore` or live `zellij` behavior | Stop and update owning feature/design first | Routed-stub scope remains planned |
| `STOP-03` | `SD-03`, `ER-01` | FT-002 requires finalized help/output contract | Stop and route to FT-003 or FT-004 | Route availability remains the FT-002 target |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Simplify review verdict for command tree implementation | implementer / reviewer | PR notes or `artifacts/ft-002/verify/simplify-review/` | `CP-03` |

## Готово для приемки

- All workstreams complete or stopped through `STOP-*`.
- `go test ./...` passes.
- `go build ./cmd/zelma` passes.
- Side-effect check confirms no registry, `.gitignore` or live `zellij`
  behavior in FT-002 implementation.
- Tests do not over-own FT-003 help templates or FT-004 broader output/error
  contracts.
- Final acceptance follows `brief.md` `Verify`.
