---
title: "FT-020: Session Evidence Parser Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for parsing privacy-safe Codex session evidence into CodexSessionRef."
derived_from:
  - brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../../engineering/testing-policy.md
status: draft
audience: humans_and_agents
---

# FT-020: Session Evidence Parser Design

## Selected Design

FT-020 adds a parser contract in `internal/codex` that reads the first
non-empty Codex JSONL record and accepts it only when it is `session_meta`.
The parser extracts a minimal `CodexSessionRef` from `payload.session_id`, with
`payload.id` as fallback, and returns `insufficient_evidence` when required
identity fields are absent or invalid.

## Contracts

| Contract ID | Contract | Owner |
| --- | --- | --- |
| `CTR-01` | `CodexSessionRef.session_id` is a UUID extracted from `session_meta`. | `internal/codex` |
| `CTR-02` | Parser result is either `resolved` with a ref or `insufficient_evidence` with a reason. | `internal/codex` |
| `CTR-03` | Parser output never includes prompts, assistant responses, tool payloads or arbitrary JSONL content. | `internal/codex` |

## Invariants

- `INV-01` The parser does not write `.zelma/sessions.json`.
- `INV-02` The parser does not promote candidate sessions to `active`.
- `INV-03` Relative or empty `cwd` values are not emitted as safe metadata.

## Verification

- `CHK-01`: Go fixture tests assert valid `session_meta` returns a
  `CodexSessionRef` and partial metadata returns `insufficient_evidence`.
- `CHK-02`: Go privacy scan tests assert synthetic private content does not
  appear in serialized parser output.
