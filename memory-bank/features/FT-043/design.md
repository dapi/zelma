---
title: "FT-043: Command Arg Codex Session Evidence Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for resolving CodexSessionRef from privacy-safe command argv evidence."
derived_from:
  - brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../FT-020/design.md
  - ../FT-021/design.md
status: active
audience: humans_and_agents
---

# FT-043: Command Arg Codex Session Evidence Design

## Selected Design

Add a command evidence parser in `internal/codex`. The parser receives the
observed pane command string, identifies whether it is a Codex command, then
extracts only safe UUID evidence:

1. `codex resume <uuid>` or `node .../codex resume <uuid>` resolves
   `CodexSessionRef` with source `argv_resume`.
2. `CODEX_EXTERNAL_SESSION_UUID=<uuid>` or `External session UUID: <uuid>`
   resolves `CodexSessionRef` with source `argv_external_session_uuid`.
3. If neither source is present, command evidence is insufficient and detection
   may still fall back to `session_meta` lookup.

`detection` stores only the UUID in the candidate record. The raw command,
developer instructions, prompts and arbitrary args are not persisted.

## Source Semantics

| Source | Meaning | Confidence |
| --- | --- | --- |
| `argv_resume` | User explicitly resumed a Codex session by UUID. | strong |
| `argv_external_session_uuid` | Wrapper supplied a stable external UUID for this Codex process. | strong external reference |

`argv_external_session_uuid` is not claimed to be the internal Codex
`session_meta.payload.session_id`. It is still a valid `CodexSessionRef` because
the domain model permits an identifier or reference to the Codex runtime session.

## Detection Rules

- Command evidence is evaluated only after the command identifies Codex.
- Non-Codex commands with UUID-looking args are ignored.
- `argv_resume` takes precedence over external UUID evidence.
- If command evidence resolves a UUID, CLI enrichment skips the expensive
  `session_meta` directory scan for that candidate.
- Registry state transition remains unchanged: a record becomes `active` only
  when `zellij_session`, `zellij_pane`, `codex_session` and normalized
  `opened_path` are all present.

## Supported Command Shapes

```text
codex resume 019f3d81-b070-7a91-9a6f-9f50f1cba355
codex --dangerously-bypass-approvals-and-sandbox --search resume 019f3d81-b070-7a91-9a6f-9f50f1cba355
node /path/to/bin/codex resume 019f3d81-b070-7a91-9a6f-9f50f1cba355
env CODEX_EXTERNAL_SESSION_UUID=019f3d81-b070-7a91-9a6f-9f50f1cba355 codex --search
codex -c "developer_instructions='External session UUID: 019f3d81-b070-7a91-9a6f-9f50f1cba355.'"
```

## Invariants

- Raw argv is never stored in `.zelma/sessions.json`.
- The parser returns `insufficient_evidence` for invalid UUIDs.
- The parser rejects UUIDs from commands that do not identify Codex.
- Detection does not read Codex transcript content for argv evidence.

## Verification

- `CHK-01`: `internal/codex` tests cover direct, node and env/developer
  instruction UUID extraction.
- `CHK-02`: `internal/detection` tests assert `ClassifyPane` and
  `DetectCandidates` propagate the UUID.
- `CHK-03`: `internal/cli` test asserts `sessions detect` writes an `active`
  record from `codex resume <uuid>` without session log evidence.
