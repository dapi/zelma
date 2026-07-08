---
title: "FT-019: Codex Metadata Source Discovery Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for privacy-safe inventory of Codex metadata sources."
derived_from:
  - brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../../engineering/testing-policy.md
status: draft
audience: humans_and_agents
---

# FT-019: Codex Metadata Source Discovery Design

## Selected Design

FT-019 adds a source inventory contract, not a `CodexSessionRef` parser.
`internal/codex.DiscoverMetadataSources` reports which metadata sources are
usable, present, missing or intentionally not probed, and annotates each source
with confidence and privacy constraints.

The inventory is intentionally useful for FT-020 without doing FT-020's work:
it records that `session_meta` records are a candidate source, but it does not
open JSONL records, parse UUIDs or promote registry records to `active`.

## Source IDs

| Source ID | Confidence | Status policy | Privacy boundary |
| --- | --- | --- | --- |
| `zellij_pane_command` | weak | usable through existing zellij adapter | executable token only; raw args excluded |
| `zellij_pane_cwd` | weak | usable through existing zellij adapter | normalized cwd is allowed metadata |
| `process_argv` | strong | not probed in FT-019 | full argv excluded; parser/correlation deferred to FT-020 |
| `codex_home_env` | medium | present when `CODEX_HOME` is set | path only |
| `codex_home_default` | medium | present fallback when `CODEX_HOME` is unset | default path only |
| `session_log_directory` | medium | present when `<codex home>/sessions` exists | directory presence and JSONL count only |
| `session_meta_record` | medium | present when JSONL files exist | safe fields documented; no parsing in FT-019 |

## Contracts

| Contract ID | Contract | Owner |
| --- | --- | --- |
| `CTR-01` | Every source has stable `id`, `status`, `confidence`, `privacy` and `safe_fields`. | `internal/codex` |
| `CTR-02` | Discovery may count `.jsonl` session log files but must not read JSONL content. | `internal/codex` |
| `CTR-03` | Sources with possible private content must list explicit `excluded_fields`. | `internal/codex` |
| `CTR-04` | Parser/downstream work remains marked with `downstream_feature: FT-020`. | `internal/codex` |

## Invariants

- `INV-01` FT-019 does not extract, validate or persist `CodexSessionRef`.
- `INV-02` FT-019 does not parse Codex conversation records.
- `INV-03` Raw process argv, prompts, assistant responses and tool content do
  not appear in the inventory output.
- `INV-04` `sessions detect` still writes unresolved `candidate` records until
  FT-020/FT-021 implement parser and state transition rules.

## Verification

- `CHK-01`: `internal/codex` unit tests assert `CODEX_HOME`, default
  `~/.codex`, session directory and `session_meta` source discovery.
- `CHK-02`: `internal/codex` unit tests assert synthetic private JSONL content
  is not present in serialized inventory output and that sensitive sources list
  exclusions.
