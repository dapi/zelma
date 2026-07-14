---
title: "FT-001: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-001. Фиксирует discovery context, шаги, риски и test strategy без переопределения canonical problem и scope."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_001_scope
  - ft_001_selected_design
  - ft_001_acceptance_criteria
  - ft_001_blocker_state
---

# FT-001: Implementation Plan

## Цель текущего плана

Создать минимальный Go scaffold для `zelma`, достаточный для сборки binary и
запуска `go test ./...`, не заходя в command tree, registry или live `zellij`
integration.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| ../../adr/ADR-001-mvp-cli-architecture.md | architecture owner | Go CLI, Cobra, module boundaries | Update ADR or create new ADR first |
| ../../engineering/architecture.md | engineering boundary owner | `cmd/zelma`, `internal/*` boundaries | Update architecture first |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `README.md` | Product entrypoint | Names binary and CLI expectations | Keep command names aligned |
| `memory-bank/adr/ADR-001-mvp-cli-architecture.md` | Accepted architecture | Defines Go/Cobra/adapter boundaries | Mirror package layout |
| `memory-bank/engineering/testing-policy.md` | Test policy | Defines Go test expectations after scaffold | Use `go test ./...` |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Go module discovery | `REQ-01`, `CHK-01` | none | Go package tests compile | `go test ./...` | none yet | CI not configured in FT-001 | `none` |
| Binary build | `REQ-02`, `CHK-02` | none | Build command succeeds | `go build ./cmd/zelma` | none yet | CI not configured in FT-001 | `none` |
| Side-effect boundary | `REQ-03`, `REQ-05`, `CHK-03` | docs only | Static search/code review | `rg -n "zellij|instances.json|\\.zelma" cmd internal` | none yet | Review remains manual until behavior exists | `none` |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Exact Go module path | GitHub repo exists as `github.com/dapi/zelma` | `STEP-01` | Use `github.com/dapi/zelma` |
| `OQ-02` | Whether Cobra dependency is introduced in FT-001 or FT-002 | Brief keeps command tree out of FT-001 | none | Avoid Cobra until FT-002 unless needed to compile |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Go toolchain available in `PATH` | all steps | `go version` fails |
| test | `go test ./...` from repo root | `CHK-01` | command cannot discover module or tests fail |
| build | `go build ./cmd/zelma` from repo root | `CHK-02` | binary build fails |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `ASM-01` | Go toolchain installed | all steps | yes |
| `PRE-02` | `CON-01` | ADR-001 remains accepted | `STEP-02`, `STEP-03` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01` | `go.mod` | agent | `PRE-01` |
| `WS-2` | `REQ-02`, `REQ-03` | `cmd/zelma/main.go` and minimal internal package | agent | `WS-1` |
| `WS-3` | `REQ-04`, `REQ-05` | passing tests/build and side-effect check | agent | `WS-1`, `WS-2` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | Need to add behavior outside scaffold | `WS-2` | Would change accepted feature scope | User confirmation or updated brief |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01` | Initialize Go module | `go.mod` | module file | `CHK-01` | `EVID-01` | `go test ./...` | `PRE-01` | none | module path conflicts with repo |
| `STEP-02` | agent | `REQ-02`, `REQ-03` | Add minimal binary entrypoint | `cmd/zelma/main.go`, optional `internal/*` | buildable package | `CHK-02` | `EVID-02` | `go build ./cmd/zelma` | `STEP-01` | none | command tree scope appears necessary |
| `STEP-03` | agent | `REQ-05` | Confirm no side-effect scope | `cmd`, `internal` | review note | `CHK-03` | `EVID-03` | `rg -n "zellij|instances.json|\\.zelma" cmd internal` | `STEP-02` | `AG-01` if behavior needed | side effects are required |
| `STEP-04` | agent | `REQ-04` | Run final checks | repo root | command outputs | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01`, `EVID-02`, `EVID-03` | all checks from brief | `STEP-03` | none | Go toolchain unavailable |

## Parallelizable Work

- `PAR-01` No meaningful parallel code work in this small scaffold slice.
- `PAR-02` Documentation updates can be reviewed separately from code once
  scaffold paths are known.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `CHK-01` | Go module discovered and tests run | `EVID-01` |
| `CP-02` | `STEP-02`, `CHK-02` | Binary builds | `EVID-02` |
| `CP-03` | `STEP-03`, `CHK-03` | No runtime side effects introduced | `EVID-03` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Go is not installed | Cannot implement/verify | Stop before code changes or install Go outside feature scope | `go version` fails |
| `ER-02` | Feature expands into Cobra/help | FT-001 becomes too broad | Stop and create/activate FT-002 or FT-003 | help/command tree decisions appear |
| `ER-03` | Side-effect code appears early | Violates non-scope | Move to later feature | `zellij` or `.zelma` runtime code in scaffold |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `PRE-01` | Go toolchain unavailable | Do not create unverified scaffold | Feature remains planned |
| `STOP-02` | `NS-03`, `NS-04` | Need registry or zellij behavior | Stop and open next feature scope | Scaffold-only package unchanged |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Implementation notes if scope changes | implementer | PR/commit notes or `artifacts/ft-001/` | `CP-03` |

## Готово для приемки

- All workstreams complete or stopped through `STOP-*`.
- `go test ./...` passes.
- `go build ./cmd/zelma` passes.
- Side-effect check confirms no `zellij` or `.zelma/instances.json` behavior.
- Final acceptance follows `brief.md` `Verify`.
