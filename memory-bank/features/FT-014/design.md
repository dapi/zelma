---
title: "FT-014: Detect Upsert Idempotency Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for idempotent `zelma sessions detect` candidate registry upsert."
derived_from:
  - brief.md
  - ../../domain/rules.md
  - ../../domain/states.md
  - ../../engineering/architecture.md
  - ../../engineering/codex-runtime-identification.md
status: draft
audience: humans_and_agents
---

# FT-014: Detect Upsert Idempotency Design

## Selected Design

`zelma sessions detect` remains conservative. It reads live zellij sessions and
panes through `internal/zellij`, classifies panes through the FT-013
`internal/detection.ClassifyPane` contract, converts `candidate` verdicts into
unresolved registry records, then writes through `internal/registry`.

The command does not create panes, remove stale records or promote unresolved
candidates to `active`.

## Upsert Rules

The registry match key is `(zellij_session, zellij_pane)`.

| Case | Result |
| --- | --- |
| No existing record for the pane key | Append one `candidate` record. |
| Existing candidate record | Fill only missing candidate evidence, such as empty `opened_path`. |
| Existing active record | Preserve the existing record unchanged and do not append a duplicate candidate. |
| Existing stale/closed/archived record only | Preserve the historical record and append a new `candidate` record. |
| Detected pane lacks Codex command or repo-local cwd evidence | Skip it. |

Existing non-empty `codex_session` and `opened_path` values are never overwritten
by unresolved detect evidence.

## CLI Contract

Default output is a compact summary:

```text
added=1 unchanged=0 skipped=0
```

`--json` returns the same counters as a stable object with `added`,
`unchanged` and `skipped` fields.

## Verification

- First detect with a Codex-like pane creates one candidate record and reports
  `added=1`.
- Repeating detect for the same pane keeps one registry record and reports
  `unchanged=1`.
- An existing active record for the same pane key is preserved and not replaced
  by unresolved candidate evidence.
