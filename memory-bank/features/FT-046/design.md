---
title: "FT-046: Pane PID Codex Session Evidence Design"
doc_kind: feature
doc_function: design
purpose: "Selected design for optional PID-correlated Codex session evidence in sessions detect."
derived_from:
  - brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../../engineering/zellij-integration.md
status: draft
audience: humans_and_agents
---

# FT-046: Pane PID Codex Session Evidence Design

## Decision

`sessions detect` keeps the existing evidence order:

1. command evidence from the zellij pane command;
2. indexed `session_meta` lookup by opened path;
3. optional PID-correlated process evidence.

The PID path is a fallback only. It is invoked only for a candidate pane whose
command evidence and `session_meta` lookup did not resolve a `CodexSessionRef`.

## Adapter Boundary

The selected boundary is `codex.PaneProcessEvidenceResolver`. It accepts a
transient `PaneProcessEvidenceInput` containing the zellij session, pane id,
opened path and optional pane PID. It returns `SessionEvidenceResult`, the same
safe evidence shape used by command and `session_meta` evidence.

Current zellij JSON does not expose a stable pane PID, so the production CLI
uses an unsupported resolver. If a zellij version or platform adapter provides
`pid` or `pane_pid`, `internal/zellij` parses it as transient `Pane.ProcessID`
and detection forwards it to the resolver without persisting it.

## Resolution Rules

PID-correlated evidence resolves only when exactly one live Codex process is
correlated with the pane PID and its process evidence yields a valid UUID. The
resulting source is `pid_correlated_process`.

Zero, multiple, stale or unsupported PID candidates remain unresolved
candidates. `sessions detect --explain` reports the fallback verdict and reason.

## Privacy

PID, raw argv, process environment and process tree details are not written to
`.zelma/sessions.json`. Explain output reports only verdict, source and redacted
reason. The resolver extracts only a safe UUID session ref.
