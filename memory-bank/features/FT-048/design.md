---
title: "FT-048: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space документ для распространяемого Codex skill package `zelma`."
derived_from:
  - brief.md
  - ../../engineering/skill-contract.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_048_scope
  - ft_048_acceptance_criteria
  - ft_048_evidence_contract
  - implementation_sequence
---

# FT-048: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `../../engineering/skill-contract.md` | Canonical command / recovery contract | Public `zelma` commands, JSON modes, recovery expectations and skill boundaries |

## Context

FT-048 turns the already documented `zelma` skill contract into a distributable
Codex skill package. The design problem is not a new CLI capability; it is the
agent-facing packaging and instruction surface that lets Codex discover the
skill and route user intents to existing `zelma` commands safely.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C1` | The feature adds a Codex-facing integration surface: a Codex agent discovers `skills/zelma/SKILL.md` and invokes the `zelma` CLI inside a repository. | C1 context table below |

### C4 Artifact

| Element ID | Element | Type | Direction | Boundary / responsibility |
| --- | --- | --- | --- | --- |
| `C4-E01` | Codex agent | External actor / agent runtime | Reads skill instructions; invokes commands | Must follow `SKILL.md` and user intent |
| `C4-E02` | `skills/zelma/SKILL.md` | Skill package artifact | Instructs agent which `zelma` command to run | Owns concise agent-facing routing only |
| `C4-E03` | `zelma` CLI | System under feature | Receives public commands and returns JSON/diagnostics | Owns registry, zellij adapter and command behavior |
| `C4-E04` | zellij / `.zelma/sessions.json` | External/runtime internals relative to skill | Accessed only through `zelma` CLI | Not directly called or parsed by the skill |

## Selected Solution

- `SOL-01` Create `skills/zelma/SKILL.md` as the single distributable skill
  instruction artifact. It closes `REQ-01` through `REQ-06` by making intent
  routing and safety boundaries discoverable without duplicating large docs.
- `SOL-02` Add `skills/zelma/agents/openai.yaml` with UI metadata only. Local
  Codex skill examples show this path for display metadata, and issue 87 allows
  it when appropriate.
- `SOL-03` Keep install/development notes outside `SKILL.md` when extra repo
  guidance is needed, so the skill stays concise and canonical command details
  remain in `../../engineering/skill-contract.md`.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Put all command and recovery details directly in `SKILL.md` | Rejected because issue 87 says not to include large duplicated reference docs, and `../../engineering/skill-contract.md` already owns the detailed contract. |
| `ALT-02` | Ship only `SKILL.md` and omit OpenAI UI metadata | Rejected for this package because local Codex skill examples support `agents/openai.yaml`, and metadata improves discoverability without changing behavior. |
| `ALT-03` | Add a new wrapper script or CLI surface for the skill | Rejected because issue 87 and `CON-02` require the skill to invoke only the public `zelma` CLI. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Keep `SKILL.md` concise and refer to canonical docs for detail | Reduces drift between skill package and existing command contract | Review must verify the concise routing still covers every required intent |
| `TRD-02` | Include metadata-only `agents/openai.yaml` | Matches local Codex skill packaging examples | Requires validation that metadata does not redefine behavior |

## Accepted Local Decisions

- `SD-01` `skills/zelma/SKILL.md` is the behavioral instruction owner for the
  distributable skill; `../../engineering/skill-contract.md` remains the
  command/recovery contract owner.
- `SD-02` `skills/zelma/agents/openai.yaml` is appropriate because local Codex
  skill examples use it for display metadata and it does not add runtime
  behavior.
- `SD-03` The skill may mention `cleanup --confirm --json` only as a command
  allowed after explicit user intent to remove stale records.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | User asks to list known sessions | Codex agent / `zelma sessions list --json` | Read registry through CLI; do not parse `.zelma/sessions.json` directly |
| `CTR-02` | User asks for live status | Codex agent / `zelma sessions list --live --json` | Read-only live reachability check through CLI |
| `CTR-03` | User asks to create a managed Codex pane | Codex agent / `zelma sessions create [path] --json` | Mutating create remains inside `zelma`; dry-run may be used to preview inputs |
| `CTR-04` | User asks to detect/adopt manual panes | Codex agent / `zelma sessions detect --json` | Detection and registry upsert remain CLI-owned |
| `CTR-05` | User asks to focus session id | Codex agent / `zelma sessions focus <id> --json` | Focus by positive numeric repo-local id |
| `CTR-06` | User asks to cleanup stale sessions | Codex agent / `zelma sessions cleanup --json`; `--confirm` only with explicit removal intent | Preview is default; destructive confirm is gated |
| `CTR-07` | Command fails or reports partial state | `zelma` CLI / Codex agent | Preserve stdout/stderr diagnostics and use recovery expectations from skill contract |

## Invariants

- `INV-01` The skill calls only `zelma` commands and never instructs agents to
  call `zellij` directly.
- `INV-02` The skill does not instruct agents to read or parse
  `.zelma/sessions.json` directly.
- `INV-03` Cleanup confirmation requires explicit user intent.
- `INV-04` `agents/openai.yaml` stays metadata-only and does not redefine
  command routing, recovery or acceptance semantics.

## Failure Modes

- `FM-01` Skill description misses a required trigger phrase; Codex may not
  discover the skill for issue 87 acceptance examples.
- `FM-02` Skill guidance drifts from `../../engineering/skill-contract.md`;
  agents may run wrong commands or omit JSON mode.
- `FM-03` Cleanup guidance is too permissive; stale records may be removed
  without explicit user intent.
- `FM-04` Metadata file becomes a second behavior definition; reviewers may see
  conflicting trigger or command semantics.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add skill package files | `brief.md` and `design.md` active | Remove `skills/zelma/` before merge if validation fails |
| `RB-02` | Add minimal install/development docs if needed | Skill files exist and validation path is known | Revert doc additions if they duplicate canonical contract |
| `RB-03` | Verify and merge | All `CHK-*` pass | Keep feature branch unmerged until skill package is corrected |

## ADR / External Design Dependencies

| Artifact | Current status | Used for | Rule |
| --- | --- | --- | --- |
| `../../engineering/skill-contract.md` | `active` | CLI command, JSON and recovery contract | Skill package must route to this contract and not replace it |
| `../../engineering/architecture.md` | `active` | Boundary that skills call public CLI/API | Skill must not bypass `zelma` |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `C4-01`, `SD-01` | `INV-01`, `INV-02` | `FM-01`, `RB-01` |
| `REQ-02` | `SOL-01`, `TRD-01` | `CTR-01`-`CTR-06` | `FM-01`, `FM-02` |
| `REQ-03` | `SOL-01`, `SD-01` | `CTR-01`-`CTR-07`, `INV-01` | `FM-02`, `RB-03` |
| `REQ-04` | `SOL-01`, `SD-01` | `INV-01`, `INV-02` | `FM-02`, `RB-03` |
| `REQ-05` | `SOL-01`, `TRD-01` | `CTR-07` | `FM-02`, `RB-03` |
| `REQ-06` | `SOL-01`, `SD-03` | `CTR-06`, `INV-03` | `FM-03`, `RB-03` |
| `REQ-07` | `SOL-02`, `TRD-02`, `SD-02` | `INV-04` | `FM-04`, `RB-01` |
| `REQ-08` | `SOL-03`, `TRD-01` | `INV-01`, `INV-02` | `RB-02` |
