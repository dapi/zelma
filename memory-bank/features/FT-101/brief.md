---
title: "FT-101: Safe Message Sending To Codex Sessions"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для delivery-единицы, добавляющей безопасную отправку сообщения в существующую Codex session через публичный `zelma` CLI."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../../domain/states.md
  - ../../engineering/architecture.md
  - ../../engineering/skill-contract.md
  - ../../engineering/zellij-integration.md
  - ../../engineering/codex-runtime-identification.md
  - ../../use-cases/UC-011-send-message-to-codex-session.md
  - ../FT-045/brief.md
  - ../FT-047/brief.md
  - ../FT-048/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-101: Safe Message Sending To Codex Sessions

## What

### Problem

GitHub issue #101 описывает разрыв в публичном управлении `zelma sessions`:
агент или человек может list/create/detect/focus существующие Codex panes, но не
может доставить follow-up message в известную managed Codex session через
публичный `zelma` command.

Из-за этого callers вынуждены использовать ad hoc `zellij` commands. Такой путь
обходит readiness checks, numeric `ZelmaSessionID`, skill boundary и защиту от
опасного случая, когда прежняя pane еще существует, но Codex уже вышел и в pane
остался обычный shell.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Public send flow availability | Нет `zelma sessions send` command | `zelma sessions send <id> [message] --json` и `zelma sessions send <id> --stdin --json` доступны | CLI help, CLI tests and machine-readable compatibility checks |
| `MET-02` | Runtime safety before send | Callers can bypass `zelma` through direct zellij input | Send refuses unless target active record is live-revalidated as the intended Codex runtime | Readiness-gate tests for active/stale/candidate/unreachable/non-Codex/ambiguous targets |
| `MET-03` | Prompt privacy in diagnostics | Direct zellij errors may include typed payload | Success/failure output and adapter diagnostics never echo message body | Static and runtime diagnostic tests with sentinel message body |
| `MET-04` | Skill contract routing | `SKILL.md` has no send-message intent | Skill routes send-message intents only through public `zelma sessions send` commands | Static skill checks against `SKILL.md` and `../../engineering/skill-contract.md` |

### Scope

- `REQ-01` Add guarded public CLI command
  `zelma sessions send <id> [message] --json` and
  `zelma sessions send <id> --stdin --json`.
- `REQ-02` Use only repo-local positive numeric `ZelmaSessionID` as target
  selector.
- `REQ-03` Support message text passed as a single CLI argument.
- `REQ-04` Support message text read from STDIN for longer prompts and
  multiline messages.
- `REQ-05` Enforce exclusive message source policy: argument XOR `--stdin`;
  missing, conflicting or empty message sources fail with stable diagnostics.
- `REQ-06` Before any adapter write, live-revalidate that the registry target is
  an active record and still points to the intended live terminal pane running
  compatible Codex runtime evidence.
- `REQ-07` Reject stale, candidate, closed, archived, unreachable, ambiguous,
  non-terminal, non-Codex and identity-mismatched targets without sending.
- `REQ-08` Deliver message content to the recorded zellij pane explicitly,
  independent of currently focused pane, then perform the final submit action
  controlled by `zelma`.
- `REQ-09` Define success JSON with target/session identity and message metadata
  only, without echoing the message body.
- `REQ-10` Define agent-readable failure diagnostics using existing JSON shape:
  `code`, `retryable`, `manual_action_required`, `recovery_hint` and
  `next_command`.
- `REQ-11` Update Codex skill instructions and engineering skill contract so
  send-message intents route only through `zelma`, not direct `zellij` or direct
  `.zelma/sessions.json` parsing.

### Non-Scope

- `NS-01` Do not support fuzzy matching by path, pane title, zellij pane id,
  zellij session name or Codex session id.
- `NS-02` Do not send messages to arbitrary panes outside the repo-local
  `zelma sessions` registry.
- `NS-03` Do not treat focus as a prerequisite; send must target the recorded
  pane explicitly.
- `NS-04` Do not silently re-detect, repair ambiguous identity or promote
  candidates and then send in the same command.
- `NS-05` Do not introduce an allow-empty-message mode in this delivery unit.
- `NS-06` Do not expose zellij-specific delivery details in CLI output or skill
  instructions.
- `NS-07` Do not create a second implementation path inside `SKILL.md` or
  `internal/skills`.

### Constraints / Assumptions

- `ASM-01` GitHub issue #101 is the tracker source for this delivery unit and
  is still open as of 2026-07-10.
- `ASM-02` Issue comment "Decisions from FPF review" provides accepted local
  decisions `DL-001` through `DL-008` in `decision-log.md`; review-improve
  cycle 1 adds `DL-009` for the selected zellij binding.
- `ASM-03` `ZelmaSessionID` is already repo-local, positive and numeric through
  `../FT-045/brief.md`.
- `ASM-04` `zelma sessions focus <id>` already proves a public command can
  address a registry-backed pane by numeric session id, but focus is not a send
  prerequisite.
- `ASM-05` The zellij adapter already has a `WriteChars` capability used by
  supervisor orchestration; FT-101 may reuse or wrap it only after the new
  readiness gate passes.
- `CON-01` `../../domain/rules.md` `DR-01`, `DR-04` and `DR-08` require active
  sessions to have complete identity evidence and prevent control of panes
  without Codex proof.
- `CON-02` `../../domain/states.md` `SI-03` forbids using stale records for
  destructive pane control without revalidation.
- `CON-03` `../../engineering/architecture.md` requires zellij integration
  through `zellij-adapter` and skill integration through public CLI, not direct
  internals.
- `CON-04` `../../engineering/codex-runtime-identification.md` privacy rules
  forbid storing or logging Codex prompts, transcript content or raw argv.
- `CON-05` Direct zellij recovery hints are not allowed in send diagnostics or
  skill guidance; next steps must stay within public `zelma` commands.

No unresolved blocking problem-space or solution-space decisions remain after
`decision-log.md` entries `DL-001` through `DL-009`.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | The feature changes the public CLI contract, zellij adapter control surface, skill contract, runtime safety gate and prompt privacy boundary. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` `zelma sessions send <id> [message] --json` succeeds for a live,
  active, Codex-verified target and returns JSON without message body.
- `EC-02` `zelma sessions send <id> --stdin --json` accepts multiline STDIN and
  returns JSON metadata without message body.
- `EC-03` Missing, conflicting and empty message inputs return stable JSON
  diagnostics and do not call the zellij adapter.
- `EC-04` Send refuses stale, candidate, closed, archived, missing,
  unreachable, ambiguous, non-terminal, non-Codex and identity-mismatched
  targets before any adapter write.
- `EC-05` Delivery targets the recorded zellij session/pane, not the currently
  focused pane.
- `EC-06` Payload delivery and final submit behavior are deterministic and
  covered by adapter tests.
- `EC-07` Failure diagnostics and adapter errors do not leak message content and
  provide only public `zelma` recovery hints.
- `EC-08` `SKILL.md` and `../../engineering/skill-contract.md` route
  send-message intent through `zelma sessions send` only and keep the
  no-direct-zellij/no-direct-registry boundary.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01` | `EC-01`, `EC-02`, `SC-01`, `SC-02` | `CHK-01`, `CHK-06` | `EVID-01`, `EVID-06` |
| `REQ-02` | `ASM-03` | `EC-01`, `SC-01`, `NEG-01` | `CHK-01`, `CHK-03` | `EVID-01`, `EVID-03` |
| `REQ-03` | `ASM-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `ASM-01` | `EC-02`, `SC-02` | `CHK-01` | `EVID-01` |
| `REQ-05` | `ASM-02` | `EC-03`, `NEG-02` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` |
| `REQ-06` | `CON-01`, `CON-02`, `CON-03` | `EC-04`, `SC-03`, `NEG-03`, `NEG-04` | `CHK-03`, `CHK-04`, `CHK-07` | `EVID-03`, `EVID-04`, `EVID-07` |
| `REQ-07` | `CON-01`, `CON-02` | `EC-04`, `NEG-03`, `NEG-04` | `CHK-03`, `CHK-04`, `CHK-07` | `EVID-03`, `EVID-04`, `EVID-07` |
| `REQ-08` | `ASM-05`, `CON-03` | `EC-05`, `EC-06`, `SC-04` | `CHK-04`, `CHK-05`, `CHK-07` | `EVID-04`, `EVID-05`, `EVID-07` |
| `REQ-09` | `CON-04` | `EC-01`, `EC-02`, `EC-07`, `NEG-05` | `CHK-01`, `CHK-03`, `CHK-07` | `EVID-01`, `EVID-03`, `EVID-07` |
| `REQ-10` | `CON-05` | `EC-03`, `EC-04`, `EC-07`, `SC-05` | `CHK-02`, `CHK-03`, `CHK-07` | `EVID-02`, `EVID-03`, `EVID-07` |
| `REQ-11` | `CON-03`, `CON-05` | `EC-08`, `SC-06`, `NEG-06` | `CHK-06`, `CHK-07` | `EVID-06`, `EVID-07` |

### Acceptance Scenarios

- `SC-01` A caller runs `zelma sessions send 2 "please continue" --json`;
  `zelma` validates session id `2`, verifies the target is the intended live
  Codex pane, writes to that pane and returns success metadata without the
  message text.
- `SC-02` A caller pipes a multiline prompt into
  `zelma sessions send 2 --stdin --json`; `zelma` preserves the multiline
  message content for delivery, appends/performs the controlled submit action
  once, and returns metadata without the message text.
- `SC-03` A registry record exists and the pane still exists, but Codex has
  exited and left a shell; `zelma` refuses with a not-ready diagnostic and no
  adapter write.
- `SC-04` The currently focused pane differs from the target; `zelma` writes to
  the recorded session/pane from the registry after readiness passes.
- `SC-05` When zellij or registry state prevents readiness, JSON diagnostics
  include a stable reason code and public `zelma` recovery hint.
- `SC-06` A Codex agent handling a user send-message intent follows `SKILL.md`
  and uses `zelma sessions send <id> ... --json` or
  `zelma sessions send <id> --stdin --json`, not direct `zellij`.

### Negative / Edge Scenarios

- `NEG-01` A caller tries to target by path, pane id, title or Codex session id;
  the command is rejected as invalid arguments or unsupported target selection.
- `NEG-02` A caller provides both argument text and `--stdin`, no message, or an
  empty message; the command returns `conflicting_message_sources`,
  `missing_message` or `empty_message`.
- `NEG-03` A stale, candidate, closed, archived or missing registry record is
  selected; `zelma` refuses before adapter write.
- `NEG-04` Live revalidation reports missing session, missing pane,
  non-terminal pane, missing Codex runtime, incompatible identity or ambiguity;
  `zelma` refuses before adapter write.
- `NEG-05` The message body contains a sentinel secret; stdout, stderr,
  diagnostics and adapter errors do not include it.
- `NEG-06` Skill instructions attempt direct `zellij`, direct registry parsing
  or manual terminal input fallback; the feature must be rejected.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-02`, `SC-01`, `SC-02` | CLI tests for argument and STDIN input with fake registry/zellij/Codex evidence | Success JSON includes target identity and metadata, no message body | `artifacts/ft-101/verify/chk-01/` |
| `CHK-02` | `EC-03`, `NEG-02` | CLI tests for conflicting, missing and empty message source errors | Stable diagnostics and zero adapter calls | `artifacts/ft-101/verify/chk-02/` |
| `CHK-03` | `EC-04`, `EC-07`, `NEG-01`-`NEG-05` | CLI/readiness tests for invalid targets and privacy sentinel | Send is refused before adapter write; diagnostics omit message body | `artifacts/ft-101/verify/chk-03/` |
| `CHK-04` | `EC-04`, `EC-05`, `SC-03`, `SC-04` | Unit tests for readiness service using fake registry and live zellij pane facts | Only active, terminal, Codex-compatible target passes | `artifacts/ft-101/verify/chk-04/` |
| `CHK-05` | `EC-05`, `EC-06`, `REQ-08` | Zellij adapter tests for selected delivery/submit command construction | Commands target explicit recorded pane and deterministic submit behavior | `artifacts/ft-101/verify/chk-05/` |
| `CHK-06` | `EC-08`, `SC-06`, `NEG-06` | Static checks and tests for `SKILL.md`, `../../engineering/skill-contract.md` and `internal/skills` wrapper | Send intent routes through public `zelma sessions send`; no direct `zellij` or registry parser path | `artifacts/ft-101/verify/chk-06/` |
| `CHK-07` | All exit criteria | Run `go test ./...`, `python3 scripts/check_memory_bank_index.py`, `git diff --check` and the project-name typo check from `AGENTS.md` | Required local checks pass; typo check has no accidental project-name misspelling beyond canonical docs that intentionally mention the typo check | `artifacts/ft-101/verify/chk-07/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-101/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-101/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-101/verify/chk-03/` |
| `CHK-04` | `EVID-04` | `artifacts/ft-101/verify/chk-04/` |
| `CHK-05` | `EVID-05` | `artifacts/ft-101/verify/chk-05/` |
| `CHK-06` | `EVID-06` | `artifacts/ft-101/verify/chk-06/` |
| `CHK-07` | `EVID-07` | `artifacts/ft-101/verify/chk-07/` |

### Evidence

- `EVID-01` CLI success test output for argument and STDIN send.
- `EVID-02` CLI message-source error test output and adapter-call assertion.
- `EVID-03` Readiness rejection and privacy sentinel test output.
- `EVID-04` Readiness service unit test output.
- `EVID-05` Zellij adapter command construction test output.
- `EVID-06` Skill contract/static check output.
- `EVID-07` Required repo check output.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | CLI test output | implementer / CI | `artifacts/ft-101/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Message-source diagnostic test output | implementer / CI | `artifacts/ft-101/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Rejection/privacy test output | implementer / CI | `artifacts/ft-101/verify/chk-03/` | `CHK-03` |
| `EVID-04` | Readiness unit test output | implementer / CI | `artifacts/ft-101/verify/chk-04/` | `CHK-04` |
| `EVID-05` | Adapter unit test output | implementer / CI | `artifacts/ft-101/verify/chk-05/` | `CHK-05` |
| `EVID-06` | Skill contract/static check output | implementer / CI | `artifacts/ft-101/verify/chk-06/` | `CHK-06` |
| `EVID-07` | Required repo check output | implementer / CI | `artifacts/ft-101/verify/chk-07/` | `CHK-07` |
