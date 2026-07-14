---
title: "FT-101: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-101 без переопределения canonical problem и solution facts."
derived_from:
  - brief.md
  - design.md
  - decision-log.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_101_scope
  - ft_101_selected_design
  - ft_101_acceptance_criteria
  - ft_101_blocker_state
---

# План имплементации

## Цель текущего плана

Добавить `zelma instances send` и обновить skill contract так, чтобы сообщение
можно было безопасно доставить в существующую active Codex session только после
строгой live-readiness проверки, без утечки prompt body в output/diagnostics.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` first |
| `decision-log.md` | accepted FPF decisions | `DL-001`-`DL-009` from issue #101 and review-improve decisions | Record new local decision before changing design |
| `../../engineering/skill-contract.md` | canonical skill command/recovery contract | Existing command table, JSON diagnostics and skill boundaries | Update contract together with `SKILL.md` |
| `../../engineering/codex-runtime-identification.md` | Codex evidence baseline | Safe evidence, ambiguity and privacy rules | Update only if FT-101 discovers a reusable evidence rule change |
| `../../engineering/zellij-integration.md` | zellij adapter baseline | Existing write and pane-listing automation surfaces | Update only if selected zellij mechanism changes shared adapter guidance |
| `../../use-cases/UC-011-send-message-to-codex-session.md` | product scenario owner | Trigger, preconditions, main flow and exceptions | Update use case if stable scenario semantics change |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `../../../internal/cli/cli.go` | Cobra command tree, help text, JSON diagnostics, current session commands | New `instances send` command belongs here | Mirror list/create/focus command structure and JSON diagnostic shape |
| `../../../internal/cli/cli_test.go` | CLI help/output/behavior tests | Send needs command/help snapshots, argument errors and privacy tests | Add focused send tests near existing session command tests |
| `../../../internal/cli/machine_readable_compat_test.go` | JSON compatibility and diagnostic tests | Send JSON and error diagnostics are agent-facing | Add send success/error compatibility examples |
| `../../../internal/zellij/zellij.go` and `adapter.go` | Zellij adapter control methods | Existing `WriteChars` can be reused or wrapped by `SendTextToPane` | Keep raw zellij command details behind adapter |
| `../../../internal/zellij/zellij_test.go` | Adapter command construction tests | Must prove explicit pane targeting and submit behavior | Add deterministic delivery tests |
| `../../../internal/live/reconcile.go` | Live reachability snapshot | Useful contrast: not enough for send readiness | Do not treat live reachability alone as readiness |
| `../../../internal/detection/` and `../../../internal/codex/` | Codex command/evidence parsing | Readiness needs compatible Codex evidence without prompt leakage | Reuse safe evidence helpers |
| `../../../internal/registry/` | Registry schema, states and id lookup | Send targets numeric `ZelmaInstanceID` and state gates | Reuse read/validate helpers; no schema change expected |
| `../../../internal/skills/client.go` and tests | Skill wrapper over public CLI JSON | Needs send wrapper only through `zelma instances send` | Follow existing runJSON/error handling patterns; handle stdin if wrapper supports it |
| `../../../SKILL.md` | Repo-local Codex skill instructions | Must route send-message intent safely | Add send intent, boundaries and not-ready recovery guidance |
| `../../../internal/e2e/` | End-to-end fake zellij harness | Required for wrong/focused pane and shell-not-Codex scenarios | Add or extend e2e only if unit/CLI tests cannot prove behavior |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| CLI command and message input | `REQ-01`-`REQ-05`, `SC-01`, `SC-02`, `NEG-02` | No send command | Argument/STDIN success tests and source-policy error tests | `go test ./internal/cli` | Go test job if configured | None | `none` |
| Readiness gate | `REQ-06`, `REQ-07`, `SC-03`, `NEG-03`, `NEG-04`, `SOL-03`, `INV-01`-`INV-03` | Duplicate-create guard has partial live Codex match; `live.Reconcile` checks only pane reachability | Unit tests for active/stale/candidate/unreachable/non-terminal/non-Codex/identity mismatch/ambiguous targets and zero adapter calls | `go test ./internal/cli ./internal/detection ./internal/codex` | Go test job if configured | None | `none` |
| Zellij delivery adapter | `REQ-08`, `SC-04`, `SOL-05`, `SD-08`, `CTR-07`, `CTR-10`, `INV-04` | Existing `WriteChars` tests cover one write-chars call | Adapter tests for explicit pane targeting, `message + "\n"` construction and submit metadata separation | `go test ./internal/zellij` | Go test job if configured | None | `none` |
| JSON and privacy diagnostics | `REQ-09`, `REQ-10`, `NEG-05`, `INV-05`, `INV-06` | Existing JSON diagnostics for other commands | Machine-readable compatibility tests and sentinel body no-leak checks | `go test ./internal/cli ./internal/e2e` as needed | Go test job if configured | None | `none` |
| Skill contract and wrapper | `REQ-11`, `SC-06`, `NEG-06`, `SOL-07`, `INV-07` | Skill covers list/create/detect/focus/cleanup only | Static checks and `internal/skills` tests for send command args and no direct zellij/registry parsing | `go test ./internal/skills`; `rg` static checks | Go test job if configured | Static review is acceptable for `SKILL.md` text | `none` |
| Repository documentation checks | `CHK-07` | Memory-bank checker exists | Run required repo-level commands | `go test ./...`; `python3 scripts/check_memory_bank_index.py`; `git diff --check`; project-name typo check from `AGENTS.md` | Same if CI configured | None | `none` |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `none` | No unresolved question remains for document-ready execution. | `DL-009` selects the FT-101 zellij binding. | `none` | Reopen `design.md` first if implementation evidence contradicts `DL-009`. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- |
| setup | Commands run from repository worktree root | `STEP-01`-`STEP-10` | Repo root, registry path or memory-bank links resolve incorrectly |
| test | Go tests and memory-bank checks are available locally | `STEP-08`, `STEP-09` | Cannot produce `EVID-07` |
| zellij | Runtime tests use fake zellij fixtures; no live user pane is required | `STEP-04`, `STEP-06`, `STEP-08` | Tests would depend on current user focus/session state |
| privacy | Tests use synthetic sentinel message bodies only | `STEP-02`, `STEP-07`, `STEP-08` | Real prompt content appears in fixture/output |
| access / network / secrets | No secrets or network are required for implementation | all steps | Any external service requirement is out of scope and must stop |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `SD-01`, `CTR-01` | Numeric `ZelmaInstanceID` remains available in registry/session JSON | `STEP-01`, `STEP-03` | yes |
| `PRE-02` | `SD-03`, `INV-01` | Readiness implementation has access to registry and live zellij pane facts before delivery | `STEP-03`, `STEP-04` | yes |
| `PRE-03` | `SD-06`, `INV-05` | Message body must be treated as private data in all outputs | `STEP-02`, `STEP-05`, `STEP-07` | yes |
| `PRE-04` | `SD-08`, `CTR-10` | Adapter mechanism is deterministic under fake zellij tests | `STEP-04`, `STEP-08` | yes |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`-`REQ-05`, `SOL-01`, `SOL-02` | CLI command, message source parsing and help output | agent | `PRE-01`, `PRE-03` |
| `WS-2` | `REQ-06`, `REQ-07`, `SOL-03`, `SOL-04` | Send readiness service and diagnostics | agent | `PRE-01`, `PRE-02` |
| `WS-3` | `REQ-08`, `SOL-05`, `SD-08`, `CTR-10` | Adapter delivery method and tests | agent | `PRE-04`, `WS-2` |
| `WS-4` | `REQ-09`, `REQ-10`, `SOL-06` | Success JSON and failure diagnostics | agent | `WS-1`, `WS-2`, `WS-3` |
| `WS-5` | `REQ-11`, `SOL-07` | Skill contract, root skill and wrapper tests updated | agent | `WS-1`, `WS-4` |
| `WS-6` | `CHK-01`-`CHK-07` | Verification evidence produced | agent | `WS-1`-`WS-5` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `none` | No pre-approved manual gate is required for the selected document-ready plan. | `none` | If implementation evidence contradicts `SD-08`, stop through `STOP-02` and update `design.md` / `decision-log.md` before continuing. | `none` |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `REQ-02`, `SOL-01` | Register `instances send` command, help route and JSON flag | `internal/cli/cli.go`, `internal/cli/cli_test.go` | Command skeleton and help snapshots | `CHK-01`, `CHK-07` | `EVID-01`, `EVID-07` | `go test ./internal/cli` | `PRE-01` | `none` | Command shape requires non-numeric target |
| `STEP-02` | agent | `REQ-03`-`REQ-05`, `SOL-02`, `CTR-02`-`CTR-04` | Implement message source parsing and early diagnostics | `internal/cli/cli.go`, CLI tests | Source parser and tests | `CHK-01`, `CHK-02`, `CHK-03` | `EVID-01`, `EVID-02`, `EVID-03` | `go test ./internal/cli` | `PRE-03` | `none` | Empty-message behavior needs product change |
| `STEP-03` | agent | `REQ-06`, `REQ-07`, `SOL-03`, `SOL-04` | Implement readiness service and reason codes | `internal/cli/cli.go` or new focused internal package; tests | Ready/not-ready target result | `CHK-03`, `CHK-04` | `EVID-03`, `EVID-04` | `go test ./internal/cli ./internal/detection ./internal/codex` | `PRE-01`, `PRE-02` | `none` | Existing evidence helpers cannot distinguish shell from Codex |
| `STEP-04` | agent | `REQ-08`, `SOL-05`, `SD-08`, `CTR-10` | Add delivery adapter method using explicit-pane `write-chars` with `message + "\n"` when `submit=true` | `internal/zellij/zellij.go`, adapter tests | `SendTextToPane` or equivalent wrapper over adapter | `CHK-05` | `EVID-05` | `go test ./internal/zellij` | `PRE-04`, `STEP-03` | `none` | deterministic fake-zellij tests cannot prove the selected binding |
| `STEP-05` | agent | `REQ-09`, `REQ-10`, `SOL-06` | Wire send success JSON and structured diagnostics without body echo | `internal/cli/cli.go`, compatibility tests | JSON output and diagnostics | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` | `go test ./internal/cli` | `STEP-02`, `STEP-03`, `STEP-04` | `none` | Adapter diagnostics include raw message body |
| `STEP-06` | agent | `REQ-08`, `SC-04` | Prove explicit pane targeting independent of focus | CLI/adapter tests, fake zellij fixture | Focus-independent targeting test | `CHK-04`, `CHK-05` | `EVID-04`, `EVID-05` | `go test ./internal/cli ./internal/zellij` | `STEP-03`, `STEP-04` | `none` | Test requires real focus state |
| `STEP-07` | agent | `REQ-11`, `SOL-07` | Update skill contract, root skill and optional wrapper API | `memory-bank/engineering/skill-contract.md`, `SKILL.md`, `internal/skills/` | Skill send routing docs/wrapper | `CHK-06`, `CHK-07` | `EVID-06`, `EVID-07` | `go test ./internal/skills`; `rg` static checks | `STEP-05` | `none` | Skill needs direct zellij fallback |
| `STEP-08` | agent | `CHK-01`-`CHK-06` | Run focused automated suites and privacy/static checks | Relevant Go packages and docs | Focused test output | `CHK-01`-`CHK-06` | `EVID-01`-`EVID-06` | Focused `go test` commands and `rg` checks | `STEP-01`-`STEP-07` | `none` | Any test reveals design contradiction |
| `STEP-09` | agent | `CHK-07` | Run full repo checks | Full repo | Repo verification output | `CHK-07` | `EVID-07` | `go test ./...`; `python3 scripts/check_memory_bank_index.py`; `git diff --check`; project-name typo check from `AGENTS.md` | `STEP-08` | `none` | Required check fails outside FT-101 scope |
| `STEP-10` | agent | `RB-03` | Update docs/evidence and final acceptance notes | `memory-bank/features/FT-101/*`, evidence artifacts if used | Final feature handoff state | `CHK-07` | `EVID-07` | Review docs against implemented behavior | `STEP-09` | `none` | Delivered behavior diverges from `design.md` |

## Parallelizable Work

- `PAR-01` `STEP-01` help wiring and `STEP-02` message-source parser tests can
  be developed together because both live in CLI and share source policy.
- `PAR-02` `STEP-04` adapter command-construction tests can be prepared while
  `STEP-03` readiness tests are refined, but adapter calls must not be wired
  into CLI until readiness passes.
- `PAR-03` `STEP-07` documentation wording can be drafted after `STEP-01`, but
  final skill contract must wait for JSON/diagnostic shape from `STEP-05`.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02`, `CHK-01`, `CHK-02` | CLI source policy works for argument, STDIN and invalid inputs | `EVID-01`, `EVID-02` |
| `CP-02` | `STEP-03`, `STEP-04`, `CHK-04`, `CHK-05` | Readiness gate and adapter delivery semantics are independently tested | `EVID-04`, `EVID-05` |
| `CP-03` | `STEP-05`, `STEP-06`, `CHK-03` | Success/failure outputs are private and target the recorded pane | `EVID-03` |
| `CP-04` | `STEP-07`, `CHK-06` | Skill contract routes send intent only through public CLI | `EVID-06` |
| `CP-05` | `STEP-08`, `STEP-09`, `CHK-07` | Focused and full repo checks pass | `EVID-07` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Readiness code reuses live reachability alone | Could type into shell after Codex exit | Tests must include pane exists but command/evidence is non-Codex | `CHK-04` missing shell-not-Codex case |
| `ER-02` | Adapter diagnostics include raw payload | Prompt/privacy leak | Sentinel tests and adapter error normalization review | `NEG-05` fails |
| `ER-03` | `--stdin` tests are hard with current `Run` signature | STDIN path may be under-tested | Introduce testable input reader seam in CLI, scoped to send command | `CHK-01` cannot simulate stdin |
| `ER-04` | Zellij `write-chars` newline submit behavior differs by version | Message may not submit deterministically | Keep mechanism behind adapter, cover selected args in fake-zellij tests and stop through `STOP-02` if implementation evidence contradicts `SD-08` | Adapter tests cannot model behavior or local zellij evidence contradicts selected binding |
| `ER-05` | Skill docs drift from engineering contract | Agents may call unsafe fallback | Update `SKILL.md`, `skill-contract.md` and wrapper tests in same workstream | `CHK-06` mismatch |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `INV-01`, `FM-04`, `FM-05` | Implementation cannot prove target is still Codex before delivery | Stop, record human gate, do not wire adapter write | Command absent or disabled; no send shipped |
| `STOP-02` | `SD-08`, `CTR-10` | Selected `write-chars` newline binding cannot be tested deterministically or contradicts implementation evidence | Stop, update `design.md` and `decision-log.md`; do not switch to another zellij mechanism inside the plan silently | Adapter method not used by CLI |
| `STOP-03` | `INV-05`, `FM-07` | Message body appears in any output or diagnostic | Stop and redesign output/error wrapping before continuing | No send behavior accepted |
| `STOP-04` | `INV-07`, `NEG-06` | Skill requires direct zellij or registry parsing fallback | Stop and update design/contract before shipping | Skill send intent omitted rather than unsafe |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-08` | Readiness discovery note comparing `live.Reconcile` and send readiness | implementer / reviewer | `artifacts/ft-101/plan/readiness-discovery/` | `CP-02` |
| `EVID-09` | Simplify-review verdict for readiness/delivery code | implementer / reviewer | `artifacts/ft-101/plan/simplify-review/` | `CP-05` |

## Готово для приемки

- Все workstreams завершены или явно остановлены через `STOP-*`.
- Все checkpoints имеют evidence.
- Required local suites зелёные, а CI не противоречит local verify.
- No manual-only gaps remain for the document-ready plan.
- `SKILL.md`, `../../engineering/skill-contract.md` and implemented CLI do not
  diverge.
- Финальная приемка идет по `brief.md` `Verify`, а не по этому checklist.
