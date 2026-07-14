---
title: "FT-008: Registry Validation And Recovery Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for registry validation diagnostics, machine-readable error codes and non-destructive recovery hints."
derived_from:
  - brief.md
  - ../../domain/rules.md
  - ../../engineering/architecture.md
status: active
audience: humans_and_agents
---

# FT-008: Registry Validation And Recovery Design

## Selected Design

Registry validation lives in `internal/registry` and returns typed diagnostics
instead of plain text errors. The diagnostic contract is:

- `code` - stable machine-readable error code;
- `path` - registry field, record path or file path when available;
- `message` - short human-readable problem;
- `recovery_hint` - safe next action.

Diagnostics are exposed through `registry.DiagnosticError`, so callers can use
`errors.As` to branch on `Diagnostic.Code` without parsing text.

## Error Codes

| Code | Meaning | Recovery boundary |
| --- | --- | --- |
| `registry_invalid_json` | File is not valid JSON or cannot be decoded as a registry object. | Do not write; restore valid JSON first. |
| `registry_trailing_data` | Valid registry object has extra bytes after it. | Remove trailing data manually. |
| `registry_unknown_field` | JSON includes fields outside schema v1. | Remove unsupported field or migrate later. |
| `registry_missing_required_field` | A required root/session field is absent. | Restore the missing field manually. |
| `registry_unsupported_version` | `version` is not schema v1. | Use v1 or a future migration command. |
| `registry_invalid_field` | A field is empty, unsupported, relative or non-normalized. | Correct the field manually. |
| `registry_duplicate_instance` | Two active records describe the same pane and identity. | Remove the duplicate active record. |
| `registry_conflicting_instance` | Two active records claim the same pane with different identity data. | Inspect and keep one authoritative active record. |
| `registry_read_failed` | Registry file cannot be read. | Inspect path and filesystem permissions. |

## Recovery Contract

FT-008 does not perform destructive repair. `registry.DiagnoseFile(path)` reads
and validates the file, returning diagnostics only. Invalid JSON or invalid
schema state must leave the original file bytes unchanged.

Mutating commands must treat any `DiagnosticError` from registry read/validation
as a stop condition until a future explicit migration or repair command exists.

## Verification

- Unit tests assert codes, paths and non-empty recovery hints for parse,
  required-field, version, invalid-field, duplicate and conflict scenarios.
- File-level regression test verifies invalid JSON diagnostics do not mutate
  `instances.json`.
