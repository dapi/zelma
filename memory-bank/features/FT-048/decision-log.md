---
title: "FT-048: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "Feature-local журнал решений для FT-048. Фиксирует FPF-обоснования review-improve без владения scope, selected design или execution sequencing."
derived_from:
  - brief.md
  - design.md
  - implementation-plan.md
  - ../../flows/feature-flow.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_048_scope
  - ft_048_selected_design
  - ft_048_acceptance_criteria
  - implementation_sequence
---

# FT-048: Decision Log

Этот журнал фиксирует решения, принятые во время подготовки feature package. Он
не подменяет canonical owners: scope и verify живут в `brief.md`, solution facts
живут в `design.md`, execution sequencing живет в `implementation-plan.md`.

## DL-001: Promote FT-048 To Feature Package

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-08 |
| Review cycle | 1 |
| Closed question | Можно ли создать `FT-048` package без human gate? |
| FPF frame | Bounded contexts + evidence graph |

### Available Facts

- GitHub issue 87 names `FT-048: Distributable Codex Skill` and requires a
  repo-local skill package, proposed path `SKILL.md`.
- `../../engineering/skill-contract.md` is active and already owns the
  command/recovery contract for `zelma` skills.
- `../../epics/EP-006/brief.md` and `../../product/roadmap.md` frame skills as
  thin wrappers over the stable CLI contract.
- The repository had no `memory-bank/features/FT-048/` package before this
  review.

### Decision

Create `memory-bank/features/FT-048/` with `README.md`, `brief.md`,
`design.md`, `implementation-plan.md` and this `decision-log.md`.

### Rationale

The missing feature package is a documentation/process gap, not an unresolved
product decision. Upstream facts are sufficient to define problem space, design
surface and execution plan for the distributable skill artifact.

### Human Gate

None.

## DL-002: Design Required For Skill Package

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-08 |
| Review cycle | 1 |
| Closed question | Нужен ли отдельный `design.md` для FT-048? |
| FPF frame | Boundary classification + role/function separation |

### Available Facts

- Issue 87 requires a Codex-installable/discoverable skill artifact with
  trigger behavior and safe cleanup semantics.
- `../../flows/feature-flow.md` requires `design.md` when a feature changes an
  integration contract or requires explicit solution reasoning.
- `../../engineering/architecture.md` states that skills must call public CLI /
  API rather than hand-writing `.zelma/instances.json`.
- `../../engineering/skill-contract.md` owns command routing and recovery
  expectations, but does not itself create the repo-local package artifact.

### Decision

Set `Design required: yes` in `brief.md` and make `design.md` the owner of
skill packaging, C4 context, local decisions, contracts, invariants and failure
modes.

### Rationale

The feature crosses the boundary between repo documentation/CLI contract and
Codex skill discovery. Putting selected packaging decisions in `brief.md` or
`implementation-plan.md` would mix problem, solution and execution ownership.

### Human Gate

None.

## DL-003: Include Metadata-Only `agents/openai.yaml`

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-08 |
| Review cycle | 1 |
| Closed question | Is `agents/openai.yaml` appropriate for FT-048? |
| FPF frame | Evidence graph + options comparison |

### Available Facts

- Issue 87 explicitly says to add `agents/openai.yaml` if
  appropriate for Codex skill UI metadata.
- Local installed Codex/agent skill examples include `agents/openai.yaml` with
  `interface.display_name` and `interface.short_description`.
- The acceptance criteria require `SKILL.md` behavior and do not require
  behavioral semantics in `openai.yaml`.

### Alternatives

| Alternative | Fit | Rejection / selection reason |
| --- | --- | --- |
| Add metadata-only `agents/openai.yaml` | selected | Matches local skill package examples and improves UI discoverability without changing behavior. |
| Omit metadata | rejected | Would still satisfy minimum `SKILL.md` acceptance, but leaves the issue's explicit optional metadata opportunity unused despite available local evidence. |
| Put command routing into metadata | rejected | Would create a second behavior owner and conflict with `SKILL.md` / `skill-contract.md`. |

### Decision

Include `agents/openai.yaml` as metadata-only UI information.

### Rationale

The available evidence supports the path and shape. Restricting it to metadata
keeps command routing and safety behavior in the correct owners.

### Human Gate

None.

## DL-004: C4 System Context Is Sufficient

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-08 |
| Review cycle | 1 |
| Closed question | What C4 level is required for FT-048? |
| FPF frame | Boundary classification |

### Available Facts

- `../../flows/feature-flow.md` requires C1 when an interaction with an external
  actor/system or trust boundary changes.
- FT-048 creates a Codex-facing integration surface that an external Codex
  agent reads before invoking the `zelma` CLI.
- FT-048 does not add a new runtime/deployable container, queue, storage engine
  or internal component topology.

### Decision

Use a C1 System Context table in `design.md` and do not create deeper C2/C3/C4
artifacts.

### Rationale

C1 captures the relevant actor/system boundary: Codex agent -> skill package ->
`zelma` CLI -> zellij/registry internals through CLI only. Deeper diagrams would
invent runtime structure not changed by this feature.

### Human Gate

None.

## DL-005: Session Activity Display Is Follow-Up Scope

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-08 |
| Review cycle | implementation |
| Closed question | Should FT-048 document or implement "what a session is doing" in `zelma instances list`? |
| FPF frame | Boundary classification + evidence graph |

### Available Facts

- The user asked to handle "показывать чем занята сессия" cautiously: document
  it only if public CLI already exposes such a field; otherwise do not add a
  new CLI surface in #87 and record it as separate follow-up / human gate.
- `../../engineering/skill-contract.md` documents public session fields:
  identifiers, `opened_path`, `state` and optional `live_status`.
- `../../../internal/skills/client.go` models the same public JSON fields and
  has no activity/task summary field.
- Issue 87 scope is the distributable Codex skill package, not a new
  `instances list` output contract.

### Decision

FT-048 will not add or promise a new session activity/task field. The skill may
tell agents to use only public fields currently exposed by `zelma` JSON and not
infer current work beyond those fields.

### Rationale

Adding an activity field would change the CLI/output contract and requires a
separate feature or issue. Documenting it as if it already existed would violate
the public CLI boundary and mislead agents.

### Human Gate

Follow-up issue required if product wants first-class session activity display.
