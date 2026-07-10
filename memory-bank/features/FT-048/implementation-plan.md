---
title: "FT-048: Implementation Plan"
doc_kind: feature
doc_function: derived
purpose: "Execution-план реализации FT-048 без переопределения canonical problem и solution facts."
derived_from:
  - brief.md
  - design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_048_scope
  - ft_048_selected_design
  - ft_048_acceptance_criteria
  - ft_048_blocker_state
---

# План имплементации

## Цель текущего плана

Создать и проверить repo-local Codex skill package `./` так, чтобы
Codex agents могли управлять `zelma` sessions через публичный CLI contract, а
не через прямой доступ к zellij или registry internals.

## Grounding / Support References

| Document | Role in this plan | Facts reused | Conflict action |
| --- | --- | --- | --- |
| `brief.md` | canonical problem / verify owner | `REQ-*`, `SC-*`, `NEG-*`, `CHK-*`, `EVID-*` | Update `brief.md` first |
| `design.md` | solution owner | `SOL-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` | Update `design.md` first |
| `../../engineering/skill-contract.md` | command / recovery contract | public `zelma` commands, JSON modes and recovery expectations | Update canonical engineering doc only if the actual contract changes |
| `../../engineering/architecture.md` | boundary baseline | skills call public CLI/API, not internals | Escalate if implementation needs a new boundary |

## Current State / Reference Points

| Path / module | Current role | Why relevant | Reuse / mirror |
| --- | --- | --- | --- |
| `../../engineering/skill-contract.md` | Canonical skill command and recovery contract | FT-048 skill instructions must route to it | Mirror command names, JSON modes and recovery boundaries concisely |
| `../../../internal/skills/` | Existing Go wrappers and recovery implementation | Shows current agent-facing command wrapper behavior | Do not import or expose internals from the skill package; reference public CLI only |
| `../../../skills/` | Target skill package root | Currently absent before FT-048 | Create `./` |
| Local installed skill examples under Codex/agent skill directories | Packaging examples | Show `SKILL.md` frontmatter and `agents/openai.yaml` metadata shape | Mirror only metadata structure, not unrelated skill content |
| `../../features/FT-023/` and `../../features/FT-026/` | Upstream skill wrapper and recovery features | Provide historical boundaries for wrappers/recovery | Keep FT-048 as distribution package, not wrapper redesign |

## Test Strategy

| Test surface | Canonical refs | Existing coverage | Planned automated coverage | Required local suites / commands | Required CI suites / jobs | Manual-only gap / justification | Manual-only approval ref |
| --- | --- | --- | --- | --- | --- | --- | --- |
| Skill frontmatter and metadata files | `REQ-01`, `REQ-02`, `REQ-07`, `SC-01`, `SC-05`, `SOL-01`, `SOL-02` | No repo-local skill files | Static review and optional skill validator if available | Skill validator if present; otherwise manual frontmatter/metadata review | Same static validation if CI has it | Validator availability may be environment-specific | `AG-01` only if validator is absent and reviewer requires manual acceptance |
| Intent-to-command routing | `REQ-03`, `REQ-05`, `SC-02`, `SC-04`, `CTR-01`-`CTR-07` | `skill-contract.md` documents commands | Static comparison against `skill-contract.md` | Manual/static review; `rg` for expected command strings | N/A unless docs lint exists | Manual review is acceptable because `SKILL.md` is documentation/instruction artifact | `none` |
| Safety boundary | `REQ-04`, `REQ-06`, `NEG-01`, `NEG-02`, `INV-01`-`INV-03` | Upstream docs state boundary | Static search for direct `zellij`, direct registry parsing and ungated cleanup confirm guidance | `rg -n "zellij|sessions.json|cleanup --confirm" skills/zelma memory-bank/features/FT-048` | N/A unless docs lint exists | N/A | `none` |
| Repo checks | `CHK-05` | Existing Go and memory-bank checks | Run required local suites | `go test ./...`; `python3 scripts/check_memory_bank_index.py`; `git diff --check` | Same suites in CI when available | N/A | `none` |

## Open Questions / Ambiguities

| Open Question ID | Question | Why unresolved | Blocks | Default action / escalation owner |
| --- | --- | --- | --- | --- |
| `OQ-01` | Is a skill validation command available in this environment? | Issue 87 requires running it only if it exists; no repo-local script was identified during package review. | `CHK-04` evidence shape | Search local environment before final verification; if absent, document manual validation result. |

## Environment Contract

| Area | Contract | Used by | Failure symptom |
| --- | --- | --- | --- |
| setup | Commands run from the repository worktree root | `STEP-01`-`STEP-07` | Relative paths or memory-bank checks resolve wrong files |
| test | Required local checks are `go test ./...`, `python3 scripts/check_memory_bank_index.py` and `git diff --check` | `STEP-06`, `STEP-07` | Verification cannot support `CHK-05` |
| access / network / secrets | No secrets or external services are required for skill package creation; GitHub issue link is informational | `STEP-01`-`STEP-07` | Any implementation path requiring secrets or network must stop and be redesigned |

## Preconditions

| Precondition ID | Canonical ref | Required state | Used by steps | Blocks start |
| --- | --- | --- | --- | --- |
| `PRE-01` | `CON-01`, `SD-01` | `../../engineering/skill-contract.md` remains active and is the command owner | `STEP-02`, `STEP-04`, `STEP-05` | yes |
| `PRE-02` | `SD-02`, `INV-04` | `agents/openai.yaml` remains metadata-only | `STEP-03`, `STEP-04` | no |
| `PRE-03` | `OQ-01` | Validator availability is checked before final evidence | `STEP-06` | no |

## Workstreams

| Workstream | Implements | Result | Owner | Dependencies |
| --- | --- | --- | --- | --- |
| `WS-1` | `REQ-01`-`REQ-07`, `SOL-01`, `SOL-02`, `CTR-01`-`CTR-07` | `SKILL.md` and metadata exist | agent | `PRE-01` |
| `WS-2` | `REQ-08`, `SOL-03` | Minimal install/development notes are added only if needed | agent | `WS-1` |
| `WS-3` | `CHK-01`-`CHK-05` | Verification evidence is produced | agent | `WS-1`, `WS-2`, `PRE-03` |

## Approval Gates

| Approval Gate ID | Trigger | Applies to | Why approval is required | Approver / evidence |
| --- | --- | --- | --- | --- |
| `AG-01` | No skill validator exists and reviewer requires tool-backed validation | `STEP-06` | The issue permits manual validation when no validator exists, but reviewer may require an explicit acceptance note | Human reviewer note or PR approval |

## Порядок работ

| Step ID | Actor | Implements | Goal | Touchpoints | Artifact | Verifies | Evidence IDs | Check command / procedure | Blocked by | Needs approval | Escalate if |
| --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- | --- |
| `STEP-01` | agent | `REQ-01`, `SOL-01` | Create `./` package skeleton | `./` | Directory with `SKILL.md` | `CHK-01` | `EVID-01` | Static file review | `PRE-01` | `none` | Package path conflicts with existing non-skill content |
| `STEP-02` | agent | `REQ-02`, `REQ-03`, `REQ-05`, `SOL-01`, `CTR-01`-`CTR-07` | Write concise `SKILL.md` intent routing and recovery guidance | `SKILL.md` | Skill instructions | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` | Compare with `skill-contract.md` | `PRE-01` | `none` | Existing CLI contract lacks a required command |
| `STEP-03` | agent | `REQ-07`, `SOL-02`, `SD-02` | Add OpenAI UI metadata | `agents/openai.yaml` | Metadata file | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` | Static metadata review | `PRE-02` | `none` | Metadata requires behavioral fields not covered by design |
| `STEP-04` | agent | `REQ-04`, `REQ-06`, `INV-01`-`INV-04` | Review and tighten safety boundaries | `SKILL.md`, `agents/openai.yaml` | Boundary-safe instructions | `CHK-03` | `EVID-03` | `rg` static searches and manual verdict | `STEP-02`, `STEP-03` | `none` | Direct zellij or registry parsing appears necessary |
| `STEP-05` | agent | `REQ-08`, `SOL-03` | Add minimal install/development notes only if needed | Repo docs determined during implementation | Doc update or no-op note | `CHK-04` | `EVID-04` | Static doc review | `STEP-02` | `none` | Required note would duplicate full skill contract |
| `STEP-06` | agent | `CHK-04` | Check for and run skill validation if available | Local environment | Validator output or manual validation note | `CHK-04` | `EVID-04` | Search/run validator; otherwise document absence | `PRE-03`, `STEP-01` | `AG-01` if reviewer requires | Validator is present but fails for unknown format |
| `STEP-07` | agent | `CHK-05` | Run required repo checks | Full repo | Test/check output | `CHK-05` | `EVID-05` | `go test ./...`; `python3 scripts/check_memory_bank_index.py`; `git diff --check` | `STEP-01`-`STEP-06` | `none` | A failing check reveals scope/design contradiction |

## Parallelizable Work

- `PAR-01` `STEP-03` can run after `STEP-01` while `STEP-02` text is still being
  refined, because metadata is behavior-free.
- `PAR-02` `STEP-04` must run after both `STEP-02` and `STEP-03`; it is the
  boundary review for the whole package.

## Checkpoints

| Checkpoint ID | Refs | Condition | Evidence IDs |
| --- | --- | --- | --- |
| `CP-01` | `STEP-01`, `STEP-02`, `STEP-03`, `CHK-01` | Skill files exist with expected frontmatter and metadata | `EVID-01` |
| `CP-02` | `STEP-04`, `CHK-02`, `CHK-03` | Routing matches skill contract and boundaries are preserved | `EVID-02`, `EVID-03` |
| `CP-03` | `STEP-06`, `STEP-07`, `CHK-04`, `CHK-05` | Validation and required local checks are recorded | `EVID-04`, `EVID-05` |

## Execution Risks

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `ER-01` | Skill text drifts from `skill-contract.md` | Agents may call unsupported or unsafe commands | Keep command mapping compact and review against `CTR-*` | `CHK-02` mismatch |
| `ER-02` | No validator exists | `CHK-04` is manual-only | Record manual validation explicitly as issue 87 allows | Validator search returns no command |
| `ER-03` | Install notes duplicate canonical contract | Documentation drift | Add only package location / development note or skip docs | `STEP-05` finds no repo-specific install need |

## Stop Conditions / Fallback

| Stop ID | Related refs | Trigger | Immediate action | Safe fallback state |
| --- | --- | --- | --- | --- |
| `STOP-01` | `INV-01`, `INV-02`, `NEG-01` | Implementation appears to require direct `zellij` or direct registry parsing | Stop and update `design.md`; do not ship package | Feature docs active, no skill package merge |
| `STOP-02` | `INV-03`, `NEG-02` | Cleanup confirm cannot be clearly gated by explicit user intent | Stop and revise skill guidance before verification | Skill package remains unaccepted |
| `STOP-03` | `OQ-01`, `AG-01` | Existing validator fails and failure semantics are unclear | Stop and escalate with validator output | No manual override without reviewer approval |

## Plan-local Evidence

| Evidence ID | Artifact | Producer | Path contract | Reused by checkpoints |
| --- | --- | --- | --- | --- |
| `EVID-09` | Validator availability note | implementer | `artifacts/ft-048/plan/validator-availability/` | `CP-03` |
| `EVID-10` | Simplify-review verdict for skill text | implementer / reviewer | `artifacts/ft-048/plan/simplify-review/` | `CP-02` |

## Готово для приемки

- Все workstreams завершены или явно остановлены через `STOP-*`.
- Все checkpoints имеют evidence.
- Required local suites зелёные, а CI не противоречит local verify.
- Manual-only validation gap закрыт через documented result или approved
  `AG-01`.
- Финальная приемка идет по `brief.md` `Verify`, а не по этому checklist.
