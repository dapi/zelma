---
title: "FT-019: Metadata Privacy Boundary"
doc_kind: feature-support
doc_function: evidence
purpose: "Privacy boundary for Codex metadata source discovery."
derived_from:
  - brief.md
  - design.md
  - ../../engineering/codex-runtime-identification.md
  - ../../engineering/testing-policy.md
status: draft
audience: humans_and_agents
---

# FT-019: Metadata Privacy Boundary

## Allowed Metadata

- Source ID, availability status, confidence and privacy class.
- `CODEX_HOME` or default Codex home path.
- Presence of `<codex home>/sessions`.
- Count of `.jsonl` files below the sessions directory.
- Documented future-safe fields from the first `session_meta` record:
  `session_id`, `id`, `cwd`, `cli_version` and `timestamp`.

## Explicitly Excluded

- Full Codex JSONL transcripts.
- Conversation items, user prompts and assistant responses.
- Tool input/output content.
- Full process argv and environment variables.
- Any parser output that claims a `CodexSessionRef`.

## Test Evidence

`internal/codex` has a synthetic JSONL fixture test that writes private message
content after a `session_meta` line, runs source discovery and serializes the
inventory. The assertion verifies that the private content is absent from the
inventory output while the excluded content classes remain documented.
