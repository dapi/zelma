---
title: "FT-006: Sessions Registry Schema V1"
doc_kind: feature-support
doc_function: reference
purpose: "Reference contract for `.zelma/sessions.json` schema v1 fields and fixtures."
derived_from:
  - brief.md
  - ../../domain/model.md
  - ../../domain/rules.md
  - ../../domain/states.md
status: active
audience: humans_and_agents
---

# FT-006: Sessions Registry Schema V1

This support reference documents the implemented schema contract. The canonical
feature scope and acceptance criteria remain in [brief.md](brief.md).

## File Shape

`.zelma/sessions.json` v1 is a JSON object:

```json
{
  "version": 1,
  "sessions": []
}
```

## Fields

| Field | Type | Required | Meaning |
| --- | --- | --- | --- |
| `version` | integer | yes | Schema version. FT-006 supports only `1`. |
| `sessions` | array | yes | Registry records for known `zelma sessions`. Empty is valid. |
| `sessions[].id` | integer | yes | Positive repo-local `ZelmaSessionID`, unique within this registry. Old records without `id` are backfilled by the CLI. |
| `sessions[].zellij_session` | string | yes | External `zellij session` reference. |
| `sessions[].zellij_tab` | string | no | External `zellij tab` reference observed from `list-panes`, formatted as `tab_<id>`. |
| `sessions[].zellij_tab_name` | string | no | Human-readable zellij tab name observed with the pane. |
| `sessions[].zellij_pane` | string | yes | External `zellij pane` reference inside the zellij session. |
| `sessions[].codex_session` | string | yes | Codex session reference known to `zelma`. |
| `sessions[].opened_path` | string | yes | Normalized absolute path opened in the pane. |
| `sessions[].state` | string | yes | One of `candidate`, `active`, `stale`, `closed`, `archived`. |

## Validation

- Unknown JSON fields are rejected.
- `version` must be `1`.
- `sessions` must be present, and may be empty.
- Session record fields must be present.
- `id` must be a positive integer and unique within the registry.
- Registry records created before `id` existed are accepted; the CLI assigns
  positive IDs in memory and writes them on the next mutating registry update.
- `zellij_session` and `zellij_pane` must be non-empty.
- `zellij_tab` and `zellij_tab_name` are optional for backward compatibility
  with registry records created before tab metadata was stored.
- `opened_path` must be non-empty, absolute and normalized for every persisted
  registry record, including `candidate`.
- `codex_session` must be non-empty for `active`, `stale`, `closed` and
  `archived` records.
- `candidate` records may keep `codex_session` empty because Codex identity
  evidence is incomplete.
- `state` must match a domain state from
  [../../domain/states.md](../../domain/states.md).
- Two `active` records must not use the same
  `(zellij_session, zellij_pane)` pair.

## Fixtures

Fixtures live under `internal/registry/testdata/`:

| Fixture | Purpose |
| --- | --- |
| `empty.json` | Valid v1 registry with no sessions. |
| `minimal.json` | Smallest useful registry with one active session. |
| `representative.json` | Multiple records preserving zellij, Codex, path and state references. |

## Non-Scope

FT-006 does not define atomic writes, live `zellij` reconciliation, CLI list
output, migrations, or recovery behavior for corrupt registry files.
