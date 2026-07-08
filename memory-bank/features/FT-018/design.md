---
title: "FT-018: Create Failure Recovery Hints Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for create failure reason codes, retryability and recovery hints."
derived_from:
  - brief.md
  - ../FT-017/design.md
  - ../../engineering/architecture.md
status: draft
audience: humans_and_agents
---

# FT-018: Create Failure Recovery Hints Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | Create failure reason codes, retryability policy and recovery hint semantics. |

## Context

FT-017 intentionally left unconfirmed create panes as
`created=1 registered=0 skipped=1`. FT-018 turns those partial outcomes into agent-facing diagnostics
because a created zellij pane may now exist even when `.zelma/sessions.json` is
unchanged.

The design keeps upstream adapter contracts intact: `codex`, `zellij` and
`registry` diagnostics remain their own canonical error contracts. Create adds a
workflow-level wrapper that agents can use without parsing adapter-specific
messages.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | The change is an in-process CLI error contract and does not add runtime containers or cross-system topology. | `none` |

## Selected Design

- `SOL-01` `internal/create` owns a `DiagnosticError` for create workflow
  failures. The diagnostic carries `code`, optional upstream `cause`,
  `retryable`, human-readable message, recovery hint and partial create summary.
- `SOL-02` CLI create preflight errors from `codex` are wrapped as create reason
  codes before they reach stderr. The original Codex diagnostic remains
  available as the cause code.
- `SOL-03` `LaunchAndConfirm` returns create diagnostics for pane launch,
  confirmation read and unconfirmed-pane failures. Unconfirmed panes are no
  longer reported as successful skipped output.
- `SOL-04` Registry write failures after a confirmed pane include the partial
  summary so an agent can see that zellij side effects may already exist.

## Reason Codes

| Code | Typical cause | Retryable | Recovery direction |
| --- | --- | --- | --- |
| `create_codex_invalid_input` | Opened path is invalid for the repository. | `false` | Fix input path. |
| `create_codex_missing_binary` | Codex CLI is missing or not executable. | `false` | Fix environment with Codex install or `ZELMA_CODEX_BIN`. |
| `create_pane_launch_failed` | zellij run failed before a pane was confirmed. | `true` for zellij command failure, otherwise `false` | Inspect zellij availability/session and retry only when environment is fixed. |
| `create_pane_unconfirmed` | Returned pane cannot be proven to be the requested Codex pane. | `false` | Run detect and inspect zellij before retrying create. |
| `create_confirmation_failed` | Pane was created, but list-panes confirmation failed. | `false` | Run detect and inspect zellij before retrying create. |
| `create_registry_write_failed` | Confirmed pane could not be written to `.zelma/sessions.json`. | `true` only for registry lock contention | Inspect registry/filesystem; run detect before retrying create when a pane may already exist. |

`create_invalid_request` is reserved for internal caller contract errors.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | stderr diagnostic string | `zelma sessions create` / human or agent caller | Includes create reason code, retryability and recovery hint on create failures. |
| `CTR-02` | `create.DiagnosticError` | `internal/create` / CLI and tests | Preserves upstream cause through `errors.As` / `errors.Is` and includes original cause detail in stderr; callers must not parse free text for retryability. |
| `CTR-03` | partial `Summary` | create diagnostics / agent caller | Non-zero summary means side effects may already exist and destructive cleanup is not automatic. |

## Invariants

- `INV-01` Create diagnostics must never imply that `zelma` cleaned up a zellij
  pane.
- `INV-02` Machine-readable success output stays parseable; diagnostics remain
  on stderr for failed commands.
- `INV-03` Adapter behavior is not changed by this feature.

## Failure Modes

- `FM-01` Missing Codex stops before zellij and registry side effects, with
  `retryable=false` and fix-environment guidance.
- `FM-02` Unconfirmed pane stops before registry write, with detect and zellij
  inspection guidance.
- `FM-03` Registry write failure may leave a live created pane without a registry
  record, so the diagnostic points agents to detect before retrying create.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | CLI diagnostic contract enabled by default | Unit and CLI tests cover reason codes, hints and retryability. | Revert create diagnostic wrapper and restore FT-017 skipped-output behavior. |

## Verification

- `CHK-01`: CLI tests assert stderr includes create reason codes and recovery
  hints for missing Codex, unconfirmed pane and registry write failure.
- `CHK-02`: `internal/create` tests assert retryable classification for launch,
  unconfirmed, confirmation and preflight failures.
