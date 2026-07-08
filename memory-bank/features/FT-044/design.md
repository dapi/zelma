---
title: "FT-044: Detect Evidence Explain And Indexed Lookup Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for one-pass Codex evidence indexing and detect explain output."
derived_from:
  - brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../../engineering/skill-contract.md
  - ../FT-021/design.md
status: active
audience: humans_and_agents
---

# FT-044: Detect Evidence Explain And Indexed Lookup Design

## Selected Design

`internal/codex` exposes `BuildSessionEvidenceIndex`, which walks
`$CODEX_HOME/sessions` once, parses only first-line `session_meta` evidence and
groups resolved refs by normalized `opened_path`. `FindSessionEvidenceForOpenedPath`
now builds that index and delegates to it, preserving the old public behavior
for single lookups.

CLI enrichment uses the index once per `sessions detect` run. Candidates that
already have a `codex_session` from stronger evidence skip session-log lookup.

## Explain Contract

`zelma sessions detect --explain` appends one line per detected candidate:

```text
candidate zellij_session=<name> zellij_tab=<tab> zellij_pane=<pane> evidence=<verdict> source=<source> codex_session=<uuid> opened_path=<path> reason="<reason>"
```

`zelma sessions detect --json --explain` adds:

```json
{
  "candidate_explanations": [
    {
      "zellij_session": "zelma",
      "zellij_tab": "tab_1",
      "zellij_pane": "terminal_7",
      "opened_path": "/workspace/zelma",
      "codex_session": "019f3d81-b070-7a91-9a6f-9f50f1cba355",
      "evidence_verdict": "resolved",
      "evidence_source": "command_argv"
    }
  ]
}
```

The field is omitted unless `--explain` is set.

## Invariants

- Default text and JSON detect output remain unchanged.
- The index stores only safe `session_meta` fields already accepted by FT-020.
- Multiple `session_meta` refs for one `opened_path` still return
  `insufficient_evidence`.
- Explain output must not include raw argv, prompts, assistant messages or
  arbitrary transcript content.

## Verification

- `CHK-01`: `internal/codex` tests verify one index can resolve multiple paths
  while preserving ambiguity.
- `CHK-02`: `internal/cli` tests verify text and JSON `--explain` output.
- `CHK-03`: Existing machine-readable compatibility tests verify default JSON
  output remains stable.
