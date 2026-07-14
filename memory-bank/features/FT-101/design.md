---
title: "FT-101: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space документ для безопасной отправки сообщения в существующую Codex session через `zelma instances send`."
derived_from:
  - brief.md
  - decision-log.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../../domain/states.md
  - ../../engineering/architecture.md
  - ../../engineering/skill-contract.md
  - ../../engineering/zellij-integration.md
  - ../../engineering/codex-runtime-identification.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_101_scope
  - ft_101_acceptance_criteria
  - ft_101_evidence_contract
  - implementation_sequence
---

# FT-101: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `decision-log.md` | FPF decision record | Accepted local decisions imported from issue #101 and review-improve decisions |
| `../../engineering/skill-contract.md` | Canonical skill command / recovery contract | Public skill-facing command contract after FT-101 updates |
| `../../engineering/codex-runtime-identification.md` | Codex evidence baseline | Evidence sources, ambiguity and prompt privacy rules |
| `../../engineering/zellij-integration.md` | Zellij adapter baseline | Existing zellij automation surfaces and adapter rules |

## Context

FT-101 adds a controlled write path into a live terminal pane. This is riskier
than `instances focus`: writing a prompt into an ordinary shell or wrong pane can
execute unintended shell input. The selected design therefore separates target
selection, message source parsing, runtime readiness and delivery.

Current code already has pieces that can be reused:

- `internal/cli` has Cobra command wiring, JSON diagnostics and numeric
  session-id parsing patterns.
- `internal/zellij` has `WriteChars` used by supervisor orchestration.
- `findLiveActiveSessionForOpenedPath` contains a partial live Codex match used
  by duplicate-create guard.
- `live.Reconcile` only proves session/pane reachability; it is not sufficient
  for send readiness because it does not prove current Codex identity.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C3` | The feature changes collaboration between CLI, registry, zellij adapter, Codex evidence and skill integration inside the `zelma` CLI container. It does not add a new runtime/deployable container. | C3 component table below |

### C4 Artifact

| Element ID | Component | Responsibility | Collaborates with | Boundary |
| --- | --- | --- | --- | --- |
| `C4-E01` | CLI `instances send` command | Parse `<id>`, message source flags, `--json`, stdout/stderr and diagnostics | `C4-E02`, `C4-E05` | Public user/agent command surface |
| `C4-E02` | Send readiness service | Resolve registry record and revalidate live target before any write | `registry`, `zellij-adapter`, Codex evidence helpers | Must not perform delivery |
| `C4-E03` | Registry module | Read repo-local session records and validate state/id facts | `C4-E02` | Owns `.zelma/instances.json`, not live zellij probing |
| `C4-E04` | Zellij adapter | List live instances/panes and deliver text to explicit pane | `C4-E02`, `C4-E05` | Owns zellij CLI details |
| `C4-E05` | Delivery adapter method | Send payload and submit action to explicit pane | `C4-E04` | Does not decide target readiness |
| `C4-E06` | Skill contract / `SKILL.md` | Route agent send intent to public CLI | `C4-E01` | Must not call zellij or parse registry directly |

## Selected Solution

- `SOL-01` Add `zelma instances send <id> [message] --json` and
  `zelma instances send <id> --stdin --json` as the only public send surface.
  The command uses repo-local numeric `ZelmaInstanceID`, matching `FT-045` and
  avoiding fuzzy target selection.
- `SOL-02` Implement exclusive message-source parsing in the CLI command:
  positional message XOR `--stdin`; no source, both sources or empty message
  return stable diagnostics before registry or zellij work.
- `SOL-03` Add a send readiness service that accepts a target id and returns
  either a ready target or a structured not-ready reason. It must read the
  registry, require `state == active`, list live zellij sessions/panes, require
  the recorded terminal pane and validate Codex runtime evidence compatible
  with recorded `codex_session` and `opened_path`.
- `SOL-04` Reuse existing live Codex matching logic where it is already safe,
  but promote it into a send-specific readiness path stricter than
  `live.Reconcile`; do not use live reachability alone as send permission.
- `SOL-05` Add zellij adapter delivery as a method shaped like
  `SendTextToPane(pane, text, submit=true)`. For FT-101 the selected adapter
  mechanism is one explicit-pane
  `zellij action write-chars --pane-id <pane> <text + "\n">` invocation when
  `submit=true`. The adapter API keeps message body and submit semantics
  separate, while the zellij binding serializes submit as the final newline.
- `SOL-06` Return success JSON containing target/session identity and metadata
  such as source, byte count and line count. Do not echo the message body.
- `SOL-07` Extend `../../engineering/skill-contract.md`, root `SKILL.md` and
  `internal/skills` wrapper/tests so agents route send-message intents through
  `zelma instances send` and stop on not-ready diagnostics.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Let skills or callers send direct `zellij action write-chars` | Rejected by issue #101 and `CON-03`; it bypasses readiness, JSON diagnostics and prompt privacy controls. |
| `ALT-02` | Target by `opened_path`, pane title, zellij pane id or Codex session id | Rejected by `DL-001`; fuzzy or external identifiers create ambiguity and weaken repo-local session ownership. |
| `ALT-03` | Reuse `instances focus <id>` then type into focused pane | Rejected because focus is UI state, can race with humans/agents and is explicitly not a prerequisite in issue #101. |
| `ALT-04` | Treat `live.Reconcile` live status as enough readiness | Rejected because it checks only zellij session/pane reachability and can pass for a shell left in the same pane. |
| `ALT-05` | Auto-detect/repair ambiguous target and send in one command | Rejected by issue non-goal and `DL-003`; automatic repair before a write can guess wrong. |
| `ALT-06` | Allow empty messages by default | Rejected by `DL-002`; empty message needs a later explicit design if ever required. |
| `ALT-07` | Write payload, then issue a separate zellij `send-keys Enter` action | Deferred because local adapter/e2e coverage already supports `write-chars` with newline, while a second zellij action adds ordering/failure surface without current evidence of better safety. |
| `ALT-08` | Use zellij `paste` for all sends | Deferred because no existing local adapter pattern or test coverage proves paste behavior for this repo; it can be reconsidered if `write-chars` cannot support required payloads. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Require full live readiness every send | Prevents typing into stale/shell/wrong pane | Send may fail more often when evidence is temporarily unavailable |
| `TRD-02` | Use numeric session id only | Simple, stable target selection aligned with existing list/focus flows | Callers must list sessions before sending if they only know path or pane |
| `TRD-03` | Hide message body from all command output | Protects prompt privacy and avoids diagnostic leakage | Operators inspect byte/line metadata rather than echoed prompt text |
| `TRD-04` | Separate payload content from submit action | Makes newline/content semantics testable | Adapter implementation needs explicit tests for payload and submit ordering |
| `TRD-05` | Encode submit as trailing newline in one `write-chars` call | Reuses existing deterministic adapter surface and avoids a two-action partial failure mode | If future evidence shows newline submit is insufficient, update `design.md` before changing the adapter binding |

## Accepted Local Decisions

- `SD-01` Target selector is exactly repo-local numeric `ZelmaInstanceID`.
- `SD-02` Message source is exclusive: argument XOR `--stdin`; missing,
  conflicting or empty message fails before readiness checks.
- `SD-03` Readiness must prove active registry state, reachable zellij session,
  recorded terminal pane, matching pane identity, Codex runtime evidence and
  compatibility with recorded `codex_session` / `opened_path`. Active records
  whose Codex session was resolved from process/session metadata do not require
  the live launch command to repeat the UUID when the pane and opened path still
  match; an explicitly different live UUID remains a mismatch.
- `SD-04` Send diagnostics use specific reason codes where possible:
  `instance_not_found`, `pane_not_found`, `pane_not_terminal`,
  `instance_state_not_active`, `runtime_unreachable`, `codex_runtime_missing`,
  `codex_identity_mismatch`, `runtime_ambiguous`, `target_not_ready`.
- `SD-05` Public CLI and skill contract must not expose zellij-specific delivery
  mechanics; zellij commands remain adapter internals.
- `SD-06` Success JSON and failure diagnostics never echo message body.
- `SD-07` On not-ready diagnostics, skill guidance stops and presents public
  recovery hints instead of attempting manual terminal input.
- `SD-08` FT-101 selects zellij `write-chars` with the submitted payload encoded
  as `message + "\n"` for the first implementation. The CLI/success metadata
  counts only the original message body; the submit newline is adapter control
  data, not user message echo.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | CLI input `<id>` | Caller / `instances send` | Positive numeric repo-local id; no fuzzy fallback |
| `CTR-02` | CLI message argument | Caller / `instances send` | Single positional message source; may contain spaces as shell-provided one arg; empty is rejected |
| `CTR-03` | STDIN message | Caller / `instances send --stdin` | Reads full stdin, allows multiline content, rejects empty input |
| `CTR-04` | Message source diagnostic | `instances send` / caller | `conflicting_message_sources`, `missing_message`, `empty_message` before adapter write |
| `CTR-05` | Ready target | Send readiness service / delivery | Contains registry session plus validated live terminal pane identity |
| `CTR-06` | Not-ready diagnostic | Send readiness service / CLI | Stable reason code, no adapter write, public `zelma` recovery hint |
| `CTR-07` | Delivery request | CLI / zellij adapter | Explicit zellij session and pane id from ready target, message text and submit flag |
| `CTR-08` | Success JSON | CLI / caller | Includes id, zellij session/tab/pane, codex session, opened path, source, byte count, line count, submitted flag; excludes message body |
| `CTR-09` | Skill wrapper method | `internal/skills` / agents | Invokes `zelma instances send <id> ... --json` or stdin equivalent and parses JSON/diagnostics only |
| `CTR-10` | Zellij binding | Zellij adapter / zellij CLI | For `submit=true`, call `write-chars` once with `message + "\n"` against explicit recorded pane; do not expose this binding in CLI or skill output |

## Invariants

- `INV-01` No text is written to zellij before the readiness service returns a
  ready target.
- `INV-02` `candidate`, `stale`, `closed` and `archived` records are never send
  targets.
- `INV-03` A live pane with missing or incompatible Codex evidence is not ready,
  even if the zellij pane id still exists.
- `INV-04` Current focus does not influence delivery target.
- `INV-05` Message body is never included in stdout, stderr, structured
  diagnostics, adapter command diagnostics or skill recovery responses.
- `INV-06` Recovery hints and `next_command` values stay within public `zelma`
  commands.
- `INV-07` `SKILL.md` and `internal/skills` do not call `zellij` directly and
  do not parse `.zelma/instances.json` directly.

## Failure Modes

- `FM-01` Registry id is missing or inactive; send must stop with
  `instance_not_found` or `instance_state_not_active`.
- `FM-02` Zellij session or pane is missing; send must stop with
  `runtime_unreachable`, `instance_not_found` or `pane_not_found`.
- `FM-03` Pane exists but is plugin/non-terminal; send must stop with
  `pane_not_terminal`.
- `FM-04` Pane exists but Codex exited or command evidence no longer indicates
  Codex; send must stop with `codex_runtime_missing`.
- `FM-05` Live evidence resolves a different Codex session or opened path; send
  must stop with `codex_identity_mismatch`.
- `FM-06` Evidence is ambiguous; send must stop with `runtime_ambiguous` and
  avoid auto-repair.
- `FM-07` Adapter write/submit fails after readiness passes; diagnostics must
  preserve safe adapter context but not message body.
- `FM-08` Skill attempts direct zellij fallback; static checks must reject it.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add CLI/readiness/delivery behind explicit `instances send` command | `brief.md` and `design.md` active | Remove command registration and helper code before merge if readiness tests fail |
| `RB-02` | Update skill contract and root skill instructions | CLI contract tests pass locally | Revert skill updates if CLI surface is not shipped |
| `RB-03` | Verify and ship | All `CHK-*` pass | Keep feature branch unmerged; no runtime migration is required |

## ADR / External Design Dependencies

| Artifact | Current status | Used for | Rule |
| --- | --- | --- | --- |
| `../../engineering/architecture.md` | `active` | Module boundaries and public CLI/skill boundary | Feature must not bypass adapters or public CLI |
| `../../engineering/zellij-integration.md` | `active` | Zellij CLI adapter surfaces and targeting rules | Zellij command details stay in adapter tests |
| `../../engineering/codex-runtime-identification.md` | `active` | Runtime evidence, ambiguity and privacy rules | Readiness must not read/log prompt contents |
| `decision-log.md` | `active` | Accepted FPF decisions from issue #101 | New conflicting decisions must be recorded before changing design |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `C4-01`, `SD-01` | `CTR-01`, `CTR-08` | `RB-01` |
| `REQ-02` | `SOL-01`, `TRD-02`, `SD-01` | `CTR-01` | `FM-01` |
| `REQ-03` | `SOL-02`, `SD-02` | `CTR-02`, `CTR-04` | `RB-01` |
| `REQ-04` | `SOL-02`, `SD-02` | `CTR-03`, `CTR-04` | `RB-01` |
| `REQ-05` | `SOL-02`, `SD-02` | `CTR-04`, `INV-01` | `FM-01` |
| `REQ-06` | `SOL-03`, `SOL-04`, `TRD-01`, `SD-03` | `CTR-05`, `CTR-06`, `INV-01`-`INV-03` | `FM-01`-`FM-06`, `RB-01` |
| `REQ-07` | `SOL-03`, `SOL-04`, `SD-03`, `SD-04` | `CTR-06`, `INV-02`, `INV-03` | `FM-01`-`FM-06` |
| `REQ-08` | `SOL-05`, `TRD-04`, `TRD-05`, `SD-05`, `SD-08` | `CTR-07`, `CTR-10`, `INV-04` | `FM-07` |
| `REQ-09` | `SOL-06`, `TRD-03`, `SD-06` | `CTR-08`, `INV-05` | `FM-07` |
| `REQ-10` | `SOL-03`, `SOL-06`, `SD-04`, `SD-06` | `CTR-04`, `CTR-06`, `INV-05`, `INV-06` | `FM-01`-`FM-07` |
| `REQ-11` | `SOL-07`, `SD-07` | `CTR-09`, `INV-06`, `INV-07` | `FM-08`, `RB-02` |
