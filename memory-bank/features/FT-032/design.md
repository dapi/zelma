---
title: "FT-032: Supervisor Command And Zellij Launch Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for resolving supervisor zellij launch surface and starting start-issue in the current zellij session."
derived_from:
  - brief.md
  - ../../ops/config.md
  - ../../engineering/zellij-integration.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_032_scope
  - ft_032_acceptance_criteria
  - ft_032_evidence_contract
  - implementation_sequence
---

# FT-032: Supervisor Command And Zellij Launch Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `../../ops/config.md` | Config contract owner | `.zelma/config.json`, env variable naming and precedence |
| `../../engineering/zellij-integration.md` | Zellij surface reference | Supported zellij CLI automation surfaces |

## Context

FT-032 is the first EP-008 slice that turns the supervisor launch policy into an
implementation contract. The central design question is how to choose pane vs
tab deterministically while preserving the user's current zellij session and
avoiding a per-issue session model.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | The feature stays inside the CLI container and external zellij CLI integration already has an architecture owner; no new deployable/runtime boundary is introduced. | `none` |

## Selected Solution

- `SOL-01` Add a supervisor launch config resolver that reads
  `ZELMA_START_ISSUE_ZELLIJ_SURFACE`, then `.zelma/config.json`
  `start_issue.zellij_surface`, then returns default `pane`.
- `SOL-02` Represent resolved surface as a typed value with allowed values
  `pane` and `tab`, plus source metadata: `env`, `config`, or `default`.
- `SOL-03` Launch `pane` with zellij's current-session pane creation surface,
  naming the task surface `issue-<id>` and running `start-issue` with the
  selected repo/base/agent/prompt arguments.
- `SOL-04` Launch `tab` only when the typed resolved surface is `tab`, naming the
  tab `issue-<id>` and running the same `start-issue` command.
- `SOL-05` Persist run state with selected surface, source, command, cwd,
  `pane_id` and `tab_id` when available.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Separate zellij session per issue | Rejected by product decision; it fragments the user's working context and is explicitly out of scope. |
| `ALT-02` | Tab as default launch surface | Rejected because new tabs can steal focus and interfere with user input in the current tab. |
| `ALT-03` | CLI flag only, no repo config | Rejected because project-level preference should be stable and shareable through `.zelma/config.json`; env remains the highest-priority local override. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Default to pane | Minimizes accidental focus disruption for typical delivery work | Pane layout can become crowded for large epics |
| `TRD-02` | Allow tab via env/config | Preserves user choice for workflows that benefit from full tab space | Tab launch may still focus the new tab; this is explicit opt-in |
| `TRD-03` | Env overrides config | Enables one-off local control without editing repo files | Misconfigured shell env can surprise users until diagnostics show source |

## Accepted Local Decisions

- `SD-01` `ZELMA_START_ISSUE_ZELLIJ_SURFACE` is the env override name because it
  follows the documented `ZELMA_` prefix and scopes the setting to start-issue
  supervisor behavior.
- `SD-02` `.zelma/config.json` uses nested key `start_issue.zellij_surface` to
  keep future supervisor settings grouped without introducing a broad config
  schema now.
- `SD-03` Invalid env or config values are configuration errors and must stop
  before any zellij launch side effect.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | `ZELMA_START_ISSUE_ZELLIJ_SURFACE` -> resolved surface | shell env -> supervisor config resolver | Trim whitespace; allowed values are `pane` and `tab`; env wins over config. |
| `CTR-02` | `.zelma/config.json` `start_issue.zellij_surface` -> resolved surface | repo-local config -> supervisor config resolver | Optional file and optional key; absent value falls through to default `pane`. |
| `CTR-03` | resolved surface -> zellij launch command | supervisor launch service -> zellij CLI | `pane` uses pane launch; `tab` uses tab launch; neither path creates a new zellij session. |
| `CTR-04` | launch result -> run state | supervisor launch service -> supervisor observation/cleanup | State includes surface type, source, command, cwd, `pane_id` and `tab_id` when known. |

## Invariants

- `INV-01` Default launch surface is always `pane`.
- `INV-02` Env has higher priority than `.zelma/config.json`.
- `INV-03` Only `pane` and `tab` are accepted surface values.
- `INV-04` FT-032 launch never creates a per-issue zellij session.
- `INV-05` Zellij launch side effects happen only after surface config validation succeeds.

## Failure Modes

- `FM-01` Env value is invalid: return configuration error that names
  `ZELMA_START_ISSUE_ZELLIJ_SURFACE` and lists allowed values.
- `FM-02` `.zelma/config.json` is unreadable or malformed: return configuration
  error before zellij launch; do not silently ignore a broken repo-local config.
- `FM-03` Config value is invalid: return configuration error that names
  `.zelma/config.json` and `start_issue.zellij_surface`.
- `FM-04` Zellij launch fails: return zellij diagnostic with the selected surface
  and no run state marked as started.
- `FM-05` Tab launch focuses the new tab: accepted behavior because tab is
  explicit opt-in; pane remains the default mitigation.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add config resolver and launch command tests | FT-032 design accepted | Remove supervisor launch command and keep PROMPT-005 docs as manual guidance |
| `RB-02` | Enable tab override | Default pane launch tests pass | Remove `tab` acceptance while keeping default pane path |

## ADR / External Design Dependencies

| Artifact | Current status | Used for | Rule |
| --- | --- | --- | --- |
| `../../adr/ADR-001-mvp-cli-architecture.md` | `accepted` | CLI + zellij adapter baseline | Supervisor launch must stay consistent with explicit zellij CLI automation boundaries. |
| `../../ops/config.md` | `active` | Env/config precedence | Implementation must not introduce undocumented config keys or env names. |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-03`, `SOL-04`, `C4-00` | `CTR-03`, `INV-04` | `FM-04`, `RB-01` |
| `REQ-02` | `SOL-02`, `SOL-03`, `SOL-04` | `CTR-03`, `INV-03` | `FM-01`, `FM-03` |
| `REQ-03` | `SOL-01`, `TRD-01` | `CTR-02`, `INV-01` | `RB-01` |
| `REQ-04` | `SOL-01`, `TRD-03`, `SD-01`, `SD-02` | `CTR-01`, `CTR-02`, `INV-02` | `FM-01`, `FM-02`, `FM-03` |
| `REQ-05` | `SOL-02`, `SD-03` | `INV-03`, `INV-05` | `FM-01`, `FM-03` |
| `REQ-06` | `SOL-05` | `CTR-04` | `FM-04` |
| `REQ-07` | `ALT-01` | `INV-04` | `RB-01` |
