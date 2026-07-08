---
title: "FT-048: Distributable Codex Skill"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для delivery-единицы, создающей распространяемый Codex skill package для `zelma`."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../product/roadmap.md
  - ../../epics/EP-006/brief.md
  - ../../engineering/architecture.md
  - ../../engineering/skill-contract.md
  - ../../use-cases/UC-001-agent-session-inventory.md
  - ../../use-cases/UC-002-manual-pane-adoption.md
  - ../../use-cases/UC-003-managed-agent-launch.md
  - ../../use-cases/UC-005-agent-recovery.md
  - ../../use-cases/UC-006-stale-cleanup.md
  - ../../use-cases/UC-007-agent-handoff.md
  - ../FT-023/brief.md
  - ../FT-024/brief.md
  - ../FT-025/brief.md
  - ../FT-026/brief.md
  - ../FT-047/brief.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-048: Distributable Codex Skill

## What

### Problem

Issue 87 требует конкретный repo-local skill package artifact для `zelma`.
Сейчас проект уже описывает CLI/skill contract и Go wrappers, но в репозитории
нет `skills/zelma/SKILL.md`, который Codex может установить или обнаружить как
готовый skill.

Без такого artifact agent workflows остаются привязаны к документации и
внутренним wrappers, а не к распространяемому `SKILL.md`, который явно
маршрутизирует user intents к публичным `zelma` CLI commands.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Distributable skill package availability | `skills/zelma/` отсутствует | `skills/zelma/SKILL.md` существует с валидным `name` и `description` frontmatter | Static file review and skill validation if available |
| `MET-02` | Intent-to-command coverage | Contract exists only in docs / wrappers | Skill covers list, live status, create, detect, focus and cleanup intents | Review against `../../engineering/skill-contract.md` |
| `MET-03` | Safety boundary preservation | Boundary documented upstream | Skill instructions prohibit direct `zellij` calls, direct `.zelma/sessions.json` parsing and ungated cleanup confirm | Static review and `rg` checks |

### Scope

- `REQ-01` Create a repo-local Codex skill package at `skills/zelma/SKILL.md`
  with valid frontmatter containing `name` and `description`.
- `REQ-02` Make `description` trigger on requests to list `zelma` sessions,
  create a Codex pane with `zelma`, detect manual Codex panes, focus a
  numeric session id and cleanup stale `zelma` sessions.
- `REQ-03` Route each supported intent to the public `zelma` CLI command and
  JSON mode documented by `../../engineering/skill-contract.md`.
- `REQ-04` Preserve the skill boundary: use only `zelma`, not direct `zellij`
  calls, and do not directly read or parse `.zelma/sessions.json`.
- `REQ-05` Document safe recovery behavior from the skill contract, including
  preserving CLI diagnostics and using only safe `zelma` next commands.
- `REQ-06` Gate destructive stale cleanup: `zelma sessions cleanup --confirm
  --json` is allowed only after explicit user intent to remove stale records.
- `REQ-07` Add Codex UI metadata at `skills/zelma/agents/openai.yaml` if it is
  appropriate for discoverability and can stay consistent with local skill
  metadata examples.
- `REQ-08` Add installation/development notes in the appropriate repo doc if
  the final package needs repo-specific install or validation guidance.

### Non-Scope

- `NS-01` Do not add a second parser for `.zelma/sessions.json`.
- `NS-02` Do not call `zellij` directly from the skill.
- `NS-03` Do not introduce a new CLI surface in parallel with `zelma`.
- `NS-04` Do not run or recommend `zelma sessions cleanup --confirm --json`
  without explicit user intent to remove stale records.
- `NS-05` Do not include large duplicated reference docs inside `SKILL.md`;
  keep the skill concise and route to canonical repo docs when detail is needed.
- `NS-06` Do not change the underlying `zelma` CLI command contract in this
  delivery unit.

### Constraints / Assumptions

- `ASM-01` GitHub issue 87 is the tracker source for this delivery unit and
  names the proposed package path `skills/zelma/SKILL.md`.
- `ASM-02` Existing local Codex skills use `SKILL.md` frontmatter with `name`
  and `description`.
- `ASM-03` Existing local Codex skill metadata examples use
  `agents/openai.yaml` with `interface.display_name` and
  `interface.short_description`.
- `CON-01` `../../engineering/skill-contract.md` is the canonical command and
  recovery contract for this feature.
- `CON-02` The skill layer must remain a thin agent-facing wrapper over the
  public `zelma` CLI.

No unresolved blocking problem-space decisions remain after `decision-log.md`
entries `DL-001` through `DL-003`.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature creates a Codex-facing package and metadata surface, maps intents to commands, and must preserve safety boundaries from the skill contract. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` `skills/zelma/SKILL.md` exists, is concise and has valid `name` /
  `description` frontmatter.
- `EC-02` The skill description and body cover list, live status, create,
  detect, focus and stale cleanup intents.
- `EC-03` Every covered intent routes to the correct public `zelma` command and
  JSON mode.
- `EC-04` Skill instructions preserve the no-direct-`zellij`,
  no-direct-registry-parser and explicit-cleanup-confirm boundaries.
- `EC-05` Required local checks pass or unavailable validation commands are
  explicitly documented.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `ASM-02` | `EC-01`, `SC-01` | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` |
| `REQ-02` | `ASM-01` | `EC-02`, `SC-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-03` | `CON-01` | `EC-03`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-04` | `CON-02` | `EC-04`, `SC-03`, `NEG-01` | `CHK-03` | `EVID-03` |
| `REQ-05` | `CON-01` | `EC-04`, `SC-04` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` |
| `REQ-06` | `CON-01` | `EC-04`, `NEG-02` | `CHK-03` | `EVID-03` |
| `REQ-07` | `ASM-03` | `EC-01`, `SC-05` | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-04` |
| `REQ-08` | `ASM-01` | `EC-05`, `SC-06` | `CHK-04`, `CHK-05` | `EVID-04`, `EVID-05` |

### Acceptance Scenarios

- `SC-01` A Codex agent discovers `skills/zelma/SKILL.md`; the frontmatter
  `description` matches requests such as "list zelma sessions", "create a
  Codex pane with zelma", "detect manual Codex panes", "focus zelma session 2"
  and "cleanup stale zelma sessions".
- `SC-02` For each supported intent, the skill instructs the agent to call the
  corresponding `zelma` command with `--json` or documented safe variant.
- `SC-03` The skill tells agents to use `zelma` as the only runtime interface
  and not to call `zellij` or parse `.zelma/sessions.json` directly.
- `SC-04` When a command fails or returns an incomplete state, the skill keeps
  CLI diagnostics visible and follows recovery actions from
  `../../engineering/skill-contract.md`.
- `SC-05` If UI metadata is present, `skills/zelma/agents/openai.yaml` remains
  metadata-only and does not redefine skill behavior.
- `SC-06` Repo docs explain install/development notes only where needed and do
  not duplicate the full skill contract.

### Negative / Edge Scenarios

- `NEG-01` A review finds direct `zellij` command guidance or direct
  `.zelma/sessions.json` parsing in the skill; the feature must be rejected.
- `NEG-02` A cleanup path recommends `cleanup --confirm` without explicit user
  intent to remove stale records; the feature must be rejected.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `SC-01`, `SC-05` | Static review of `skills/zelma/SKILL.md` and optional `skills/zelma/agents/openai.yaml` | Required files exist with expected frontmatter / metadata fields | `artifacts/ft-048/verify/chk-01/` |
| `CHK-02` | `EC-03`, `SC-02`, `SC-04` | Compare skill intent table and recovery guidance with `../../engineering/skill-contract.md` | Intent routing and recovery commands match the contract | `artifacts/ft-048/verify/chk-02/` |
| `CHK-03` | `EC-04`, `NEG-01`, `NEG-02` | Static search for forbidden direct access guidance and cleanup confirm gating | No direct `zellij` or direct registry parser path; cleanup confirm is explicitly gated | `artifacts/ft-048/verify/chk-03/` |
| `CHK-04` | `EC-05` | Run available skill validation command if present; otherwise record manual validation | Validation passes or manual validation gap is documented | `artifacts/ft-048/verify/chk-04/` |
| `CHK-05` | `EC-05` | Run `go test ./...`, `python3 scripts/check_memory_bank_index.py` and `git diff --check` | All required local checks pass | `artifacts/ft-048/verify/chk-05/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01`, `EVID-04` | `artifacts/ft-048/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-048/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-048/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-048/verify/chk-04/` |
| `CHK-05` | `EVID-05` | `artifacts/ft-048/verify/chk-05/` |

### Evidence

- `EVID-01` File/frontmatter review result for `skills/zelma/SKILL.md` and
  optional OpenAI UI metadata.
- `EVID-02` Intent-to-command and recovery mapping review result.
- `EVID-03` Boundary/static-search review result for forbidden guidance and
  cleanup confirm gating.
- `EVID-04` Skill validation command result or documented manual validation.
- `EVID-05` Required local check output for `go test ./...`,
  `python3 scripts/check_memory_bank_index.py` and `git diff --check`.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Review note or validator output | implementer / reviewer | `artifacts/ft-048/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Mapping review note | implementer / reviewer | `artifacts/ft-048/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Static search output and verdict | implementer / reviewer | `artifacts/ft-048/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Skill validator output or manual validation note | implementer / reviewer | `artifacts/ft-048/verify/chk-04/` | `CHK-04` |
| `EVID-05` | Local command output summary | implementer / CI | `artifacts/ft-048/verify/chk-05/` | `CHK-05` |
