---
title: "FT-019: Metadata Source Inventory"
doc_kind: feature-support
doc_function: evidence
purpose: "Inventory of usable Codex metadata sources, confidence and downstream ownership."
derived_from:
  - brief.md
  - design.md
  - ../../engineering/codex-runtime-identification.md
status: draft
audience: humans_and_agents
---

# FT-019: Metadata Source Inventory

## Inventory

| Source | Confidence | FT-019 status | Safe use |
| --- | --- | --- | --- |
| `zellij_pane_command` | weak | usable | Establish Codex pane candidacy from command entrypoint only. |
| `zellij_pane_cwd` | weak | usable | Keep candidates equal to or inside the target repo root. |
| `process_argv` | strong | not probed | Future direct `resume <uuid>` evidence after explicit correlation. |
| `codex_home_env` | medium | present when `CODEX_HOME` is set | Locate a non-default Codex home path. |
| `codex_home_default` | medium | fallback | Locate default `~/.codex` when `CODEX_HOME` is unset. |
| `session_log_directory` | medium | present/missing | Observe `<codex home>/sessions` and count `.jsonl` files only. |
| `session_meta_record` | medium | candidate source | Future parser may read first `session_meta` record only. |

## Confidence Notes

- `process_argv` is strong only for a correlated `codex resume <uuid>` process.
  FT-019 does not probe it because raw argv can contain user prompt text.
- `session_meta_record` is medium until FT-020 proves parser behavior and
  matching rules. A file with matching cwd is not enough by itself to create an
  active registry record.
- `zellij_pane_command` and `zellij_pane_cwd` remain weak sources because they
  identify a pane candidate, not a Codex session identity.

## Runtime Evidence

`internal/codex.DiscoverMetadataSources` is the runtime inventory surface. Its
tests cover:

- `CODEX_HOME` discovery;
- default `~/.codex` fallback;
- `.jsonl` session log file counting without file reads;
- required confidence and privacy fields for every source.
