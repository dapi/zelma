---
title: "FT-021: Candidate Vs Active State Rules Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for candidate-to-active state transition rules."
derived_from:
  - brief.md
  - ../../domain/states.md
  - ../../domain/rules.md
status: active
audience: humans_and_agents
---

# FT-021: Candidate Vs Active State Rules Design

## Selected Design

`zelma sessions create` and `zelma sessions detect` both write registry records
through the same detected-record upsert policy. A record becomes `active` only
when all active invariants are known:

- `zellij_session`
- `zellij_pane`
- `codex_session`
- normalized absolute `opened_path`

Detected pane evidence alone remains `candidate`. Codex session identity is
resolved from privacy-safe `session_meta` evidence produced by FT-020. The
lookup only reads the first record of `.jsonl` session logs and accepts evidence
for a pane when exactly one parsed `session_meta.payload.cwd` matches the pane's
`opened_path`.

## Contracts

| Contract ID | Contract | Owner |
| --- | --- | --- |
| `CTR-01` | Full evidence promotes detected records to `active`. | Session Registry |
| `CTR-02` | Missing, invalid or ambiguous Codex evidence keeps records `candidate`. | Session Registry |
| `CTR-03` | Existing `active` records for the same zellij pane are not overwritten by weaker detected evidence. | Session Registry |
| `CTR-04` | `sessions detect` output reports active/candidate counts in addition to added/unchanged/skipped. | CLI |

## Invariants

- `INV-01` `active` records satisfy domain rule `DR-01`.
- `INV-02` Candidate records remain visible in `sessions list` and keep their
  `state` field as `candidate`.
- `INV-03` Ambiguous session logs are not used to promote a pane.
- `INV-04` Lookup output does not store conversation content.

## Verification

- `CHK-01`: Go tests cover full evidence promotion to `active`.
- `CHK-02`: Go tests cover partial evidence remaining `candidate` and detect
  output showing active/candidate counts.
