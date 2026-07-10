---
title: "FT-101: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "Feature-local журнал решений для FT-101. Фиксирует FPF-решения из issue #101 и review-improve без владения scope, selected design или execution sequencing."
derived_from:
  - brief.md
  - design.md
  - implementation-plan.md
  - ../../flows/feature-flow.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_101_scope
  - ft_101_selected_design
  - ft_101_acceptance_criteria
  - implementation_sequence
---

# FT-101: Decision Log

Этот журнал фиксирует решения, принятые до и во время подготовки feature
package. Он не подменяет canonical owners: scope и verify живут в `brief.md`,
solution facts живут в `design.md`, execution sequencing живет в
`implementation-plan.md`.

## DL-001: Command Shape

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | Какой public command должен отправлять сообщения в Codex session? |
| FPF frame | Boundary classification + target identity |

### Available Facts

- Issue #101 requires a public CLI flow for sending text to a known
  `zelma session`.
- Issue #101 says the target is likely numeric repo-local session id.
- `../FT-045/brief.md` already defines positive numeric repo-local
  `ZelmaSessionID`.
- Direct zellij calls would bypass `zelma` readiness checks and skill contract.

### Decision

Use guarded public CLI commands:

```bash
zelma sessions send <id> [message] --json
zelma sessions send <id> --stdin --json
```

`<id>` is the repo-local numeric `ZelmaSessionID`. Do not support fuzzy matching
by path, pane, title or Codex session id in this delivery unit.

### Rationale

Numeric `ZelmaSessionID` is the existing repo-local user-facing identity. Other
selectors would cross external runtime/domain boundaries and create ambiguity
before a write action.

### Human Gate

None.

## DL-002: Message Source Policy

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | How does `send` choose between argument text and STDIN? |
| FPF frame | Input contract separation |

### Available Facts

- Issue #101 requires both message text passed as CLI argument and message text
  read from STDIN.
- Issue #101 expects tests rejecting empty or missing message input unless
  explicit supported behavior is defined.
- Prompt content and final Codex submit action have different semantics.

### Decision

The message source is exclusive:

- argument message XOR `--stdin`;
- both sources present -> `conflicting_message_sources`;
- no source present -> `missing_message`;
- empty message -> `empty_message`, unless a later design explicitly
  introduces an allow-empty flag;
- multiline STDIN is allowed.

Message content and submit action are separate concepts. A trailing newline in
the input is message content policy; the final Codex submit action is
controlled by `zelma`.

### Rationale

Exclusive sources make CLI behavior deterministic and testable. Rejecting empty
messages avoids accidental submit-only commands in a feature whose purpose is
text delivery.

### Human Gate

None.

## DL-003: Runtime Readiness Gate

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | What must be true before any text is sent? |
| FPF frame | Evidence graph + safety gate |

### Available Facts

- Issue #101 says registry record alone is insufficient.
- A pane can still exist after Codex exits, leaving an ordinary shell.
- `../../domain/rules.md` requires active sessions to include zellij session,
  zellij pane, Codex session and normalized opened path.
- `../../engineering/codex-runtime-identification.md` requires conservative
  Codex evidence and prefers false negatives over false positives.

### Decision

`send` is allowed only after live revalidation of the target record. The target
must satisfy:

- registry record exists and `state == active`;
- zellij session is reachable;
- recorded tab/pane exists;
- target pane is a terminal pane;
- target pane identity matches the registry record;
- live evidence still indicates Codex runtime in that pane;
- live Codex evidence is compatible with recorded `codex_session` and
  `opened_path`.

If the pane still exists but Codex has exited and left a shell, `zelma` must
refuse to send and report not ready.

### Rationale

The write action is only safe when live evidence still connects the registry
record to the intended Codex runtime. Pane reachability without Codex evidence
is not enough.

### Human Gate

None.

## DL-004: Not-Ready Diagnostics

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | Which reason codes and recovery hints should not-ready send failures use? |
| FPF frame | Diagnostic taxonomy + boundary discipline |

### Available Facts

- Existing JSON diagnostics use `code`, `retryable`, `manual_action_required`,
  `recovery_hint` and `next_command`.
- Issue #101 requires clear diagnostic/recovery hint.
- Skill boundary forbids direct zellij fallback.

### Decision

Use specific reason codes where possible:

- `session_not_found`
- `pane_not_found`
- `pane_not_terminal`
- `session_state_not_active`
- `runtime_unreachable`
- `codex_runtime_missing`
- `codex_identity_mismatch`
- `runtime_ambiguous`
- `target_not_ready`

Recovery hints must stay within public `zelma` commands, such as
`zelma sessions list --live --json`, `zelma sessions detect --json --explain`,
or `zelma sessions focus <id> --json`. Do not suggest direct `zellij`
commands.

### Rationale

Specific codes allow agents to stop or recover safely without guessing. Public
`zelma` next commands preserve the skill and adapter boundaries.

### Human Gate

None.

## DL-005: Zellij Mechanism Abstraction

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | How should zellij delivery be exposed inside `zelma`? |
| FPF frame | Adapter boundary + implementation options |

### Available Facts

- Issue #101 lists `write-chars`, `write`, `send-keys` and `paste` as possible
  zellij delivery mechanisms.
- `../../engineering/architecture.md` requires zellij details behind
  `zellij-adapter`.
- Existing `internal/zellij` already exposes `WriteChars` for supervisor
  orchestration.

### Decision

Do not expose zellij-specific details in the CLI or skill contract. Implement an
adapter capability shaped like `SendTextToPane(pane, text, submit=true)`.

The implementation should research and select among:

- `zellij action write-chars --pane-id <pane>`;
- `zellij action write --pane-id <pane>`;
- `zellij action send-keys --pane-id <pane> Enter`;
- `zellij action paste`.

Initial preferred model: write/paste the payload to the explicit pane id, then
perform a separate submit action. The selected mechanism must be covered by
deterministic adapter tests.

Mechanism selection for FT-101 is refined and closed by `DL-009`; this entry
remains the accepted adapter-boundary decision.

### Rationale

The public feature needs stable delivery semantics, not a public zellij API.
Keeping the mechanism behind an adapter lets tests prove target and submit
behavior without leaking zellij details into skills.

### Human Gate

None for the abstraction. `DL-009` selects the current mechanism. If
implementation evidence later contradicts `DL-009`, `implementation-plan.md`
`STOP-02` requires updating `design.md` and this decision log before continuing.

## DL-006: JSON Output And Prompt Privacy

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | What should success/failure output reveal? |
| FPF frame | Privacy boundary + evidence minimization |

### Available Facts

- Issue #101 expects diagnostics that preserve enough information for agents to
  recover without leaking prompt contents unnecessarily.
- `../../engineering/codex-runtime-identification.md` privacy rules forbid
  storing or logging prompts, transcripts or raw argv.
- Existing JSON diagnostics already provide structured failure fields.

### Decision

Success JSON should not echo the message body. It should include target/session
identity and message metadata only, for example source, byte count and line
count.

Failure diagnostics must follow existing agent-readable shape: `code`,
`retryable`, `manual_action_required`, `recovery_hint`, `next_command`.

Diagnostics and logs must not leak prompt contents. Avoid echoing message text
in CLI output, zellij adapter errors, or skill-level recovery responses.

### Rationale

The message body is user/agent prompt content, not operational evidence.
Metadata is sufficient for acceptance and recovery.

### Human Gate

None.

## DL-007: Skill Contract

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | How should Codex skill instructions send messages? |
| FPF frame | Role/function separation |

### Available Facts

- `SKILL.md` currently routes list/create/detect/focus/cleanup through public
  `zelma` CLI.
- Issue #101 requires updating the Codex skill instructions and preserving the
  boundary against direct `zellij` or registry parsing.

### Decision

Update `SKILL.md` with a send-message intent that routes only through:

```bash
zelma sessions send <id> [message] --json
zelma sessions send <id> --stdin --json
```

The skill must not call `zellij` directly and must not parse
`.zelma/sessions.json` directly. On not-ready diagnostics, the skill should
stop and present the recovery hint rather than attempting manual terminal input.

### Rationale

The skill is an agent-facing wrapper over public CLI, not a second runtime
adapter. Keeping the send path in `zelma` ensures readiness, diagnostics and
privacy checks are applied.

### Human Gate

None.

## DL-008: Test Obligations

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | issue FPF review |
| Closed question | What coverage is required before FT-101 can be accepted? |
| FPF frame | Assurance criteria |

### Available Facts

- Issue #101 calls the failure mode risky because prompt text could be typed
  into an ordinary shell or wrong pane.
- Existing project testing policy requires automated coverage for changed
  behavior and contracts where realistic.

### Decision

Required coverage:

- CLI argument input;
- CLI STDIN input;
- conflicting/missing/empty message source errors;
- explicit pane targeting independent of currently focused pane;
- adapter command construction for selected zellij mechanism;
- payload and submit/newline behavior;
- rejection of stale, candidate, unreachable and ambiguous records;
- rejection when a live pane contains shell instead of Codex;
- no adapter call before readiness gate passes;
- diagnostics do not include prompt body;
- skill/static checks prove send-message intent uses public `zelma`, not direct
  `zellij`.

### Rationale

The tests must cover both acceptance and high-risk negative cases. Unit tests
alone are acceptable only where they deterministically prove the same safety
property with fake registry/zellij evidence.

### Human Gate

None.

## DL-009: Select `write-chars` Newline Binding For Submit

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-10 |
| Review cycle | 1 |
| Closed question | Should FT-101 leave submit mechanism open, or select a concrete zellij binding before implementation? |
| FPF frame | B.5 reasoning cycle + C.24 tool-call planning + B.3 assurance |

### Available Facts

- Issue #101 lists `write-chars`, `write`, `send-keys` and `paste` as possible
  mechanisms and requires deterministic adapter tests for the selected
  mechanism.
- `internal/zellij` already exposes `WriteChars`.
- `internal/zellij/zellij_test.go` already verifies
  `write-chars --pane-id terminal_7 "/review\n"` command construction.
- `internal/e2e/issue_supervisor_orchestration_test.go` already uses fake
  zellij fixtures that observe `write-chars --pane-id ... /review` calls.
- `brief.md` / `design.md` require explicit-pane targeting, no message body in
  output, and separation between user message body and submit action.

### FPF Resolution

- B.5: Abductive hypothesis: use the existing `write-chars` adapter surface and
  encode submit as a trailing newline. Deductive consequence: adapter tests can
  prove exact args, explicit pane id, body-plus-submit construction and no
  public exposure of zellij mechanics. Inductive evidence target: run the
  deterministic adapter/CLI tests required by `CHK-05`.
- C.24: Treat zellij invocation as a planned tool call with a hard readiness
  gate before execution; choose the already modeled tool call when assurance
  and budget are better than adding a new two-call mechanism.
- B.3: Assurance is scoped to the typed claim "FT-101 can deterministically
  construct the selected zellij call under fake-zellij tests", not to a broader
  claim that all zellij input methods are equivalent.

### Alternatives

| Alternative | Fit | Rejection / selection reason |
| --- | --- | --- |
| `write-chars` with `message + "\n"` | selected | Existing adapter/test/e2e patterns support explicit pane targeting and deterministic command construction. |
| Payload write plus separate `send-keys Enter` | deferred | Adds a two-action partial failure surface and lacks local adapter/test evidence of better safety. |
| `paste` | deferred | No existing local adapter pattern or tests prove paste behavior for FT-101. |
| Leave mechanism open in `implementation-plan.md` | rejected | Active `design.md` should own selected solution facts; leaving this open would push a solution decision into execution. |

### Decision

FT-101 selects one explicit-pane
`zellij action write-chars --pane-id <pane> <message + "\n">` invocation as the
first implementation binding for `SendTextToPane(..., submit=true)`.

The adapter API and JSON metadata still treat message body and submit action as
separate concepts: byte/line counts apply to the original message body, while
the final newline is adapter control data.

### Rationale

This uses the smallest existing, deterministic adapter surface that satisfies
the safety contract. It avoids adding a second zellij action whose partial
failure semantics would need additional design. If implementation evidence
contradicts this binding, the plan must stop and update `design.md` /
`decision-log.md` before selecting another mechanism.

### Human Gate

None.
