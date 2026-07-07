---
title: "FT-002: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space документ для FT-002. Фиксирует выбранный подход к Cobra command tree, routing contracts, local decisions и side-effect boundaries без переопределения problem space или execution plan."
derived_from:
  - brief.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_002_scope
  - ft_002_acceptance_criteria
  - ft_002_evidence_contract
  - implementation_sequence
---

# FT-002: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, feature-local `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `../../adr/ADR-001-mvp-cli-architecture.md` | Accepted architecture decision | Go CLI, Cobra, command/application/adapter separation |

## Context

`brief.md` requires `zelma setup` and `zelma sessions list/create/detect` to be
routed through Cobra while preserving the no-side-effects boundary. ADR-001
already chooses Go + Cobra and states that CLI handlers must not call `zellij`
or write the registry directly. FT-003 owns agent-first help templates, so this
design stabilizes route existence and stub behavior without finalizing help copy.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | FT-002 changes only command routing inside the existing Go CLI container from FT-001 and ADR-001. It does not add a new container, datastore, queue, external integration, security boundary, registry schema, or live `zellij` path. | `none` |

## Selected Solution

- `SOL-01` Build the root `zelma` Cobra command inside `internal/cli` and keep
  `cmd/zelma/main.go` as the thin binary entrypoint from FT-001.
- `SOL-02` Add a routed `setup` stub as a top-level Cobra command to reserve the
  command surface for FT-031 without implementing repo initialization behavior.
- `SOL-03` Add `sessions` as a Cobra command group with routed `list`, `create`
  and `detect` stub subcommands.
- `SOL-04` Stub command handlers return deterministic non-implemented
  diagnostics and avoid registry and `zellij` side effects.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Implement command parsing with Go standard library only | Rejected by ADR-001, which accepted Cobra for the nested command tree. |
| `ALT-02` | Implement registry or live `zellij` behavior behind the stubs now | Conflicts with `REQ-03`, `NS-01` and `NS-02` in `brief.md`. |
| `ALT-03` | Defer `setup` until FT-031 | Conflicts with issue 2 and `REQ-01`, which include routed `zelma setup`; FT-031 owns filesystem behavior, not route reservation. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Route command stubs before real behavior | Downstream features can depend on stable entrypoints early | Stubs must clearly avoid pretending real behavior exists |
| `TRD-02` | Keep help template customization out of FT-002 | Preserves FT-003 ownership of agent-first help templates | FT-002 tests should assert route availability, not final help copy |

## Accepted Local Decisions

- `SD-01` Stub execution returns deterministic non-implemented diagnostics
  because issue 2 requires predictable diagnostics without side effects, and
  EP-001 allows predictable errors or placeholder behavior.
- `SD-02` `zelma setup` exists in the command tree in FT-002, but any `.zelma`
  or `.gitignore` mutation remains owned by FT-031.
- `SD-03` FT-002 command routing tests may rely on Cobra behavior and local
  command constructors, but must not require live `zellij`, registry files or
  finalized agent-first help templates.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | CLI args `setup`, `sessions list`, `sessions create`, `sessions detect` | `internal/cli` / user or agent | Args route to existing Cobra commands; `--help` produces routed command output. |
| `CTR-02` | Stub execution output and exit status | Stub handlers / user or agent | Output is deterministic non-implemented diagnostics; handlers do not perform registry writes or live `zellij` calls. |
| `CTR-03` | Root binary wiring | `cmd/zelma/main.go` / `internal/cli` | Binary delegates args/stdout/stderr to CLI package and does not own command definitions. |

## Invariants

- `INV-01` Command handlers in FT-002 do not call live `zellij`.
- `INV-02` Command handlers in FT-002 do not create, read or write
  `.zelma/sessions.json`.
- `INV-03` FT-002 does not finalize help templates beyond proving command
  route availability.

## Failure Modes

- `FM-01` A routed stub accidentally performs registry or `zellij` behavior.
  Mitigation: cover `REQ-03` through `CHK-02` and keep adapters out of the
  FT-002 change surface.
- `FM-02` Tests over-specify Cobra help copy and collide with FT-003.
  Mitigation: assert route availability and command identity, not final
  agent-first wording.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add command tree as normal CLI code | FT-001 scaffold builds and ADR-001 remains accepted | Revert FT-002 command tree changes; scaffold from FT-001 remains valid |

## ADR / External Design Dependencies

| Artifact | Current status | Used for | Rule |
| --- | --- | --- | --- |
| `../../adr/ADR-001-mvp-cli-architecture.md` | `accepted` | Go + Cobra and command/application/adapter separation | If Cobra or layer separation changes, update ADR before changing this design |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `SOL-02`, `SOL-03`, `TRD-01`, `C4-00`, `SD-02` | `CTR-01`, `CTR-03`, `INV-03` | `FM-02`, `RB-01` |
| `REQ-02` | `SOL-03`, `TRD-01`, `C4-00` | `CTR-01`, `INV-03` | `FM-02`, `RB-01` |
| `REQ-03` | `SOL-04`, `TRD-01`, `C4-00`, `SD-01`, `SD-03` | `CTR-02`, `INV-01`, `INV-02` | `FM-01`, `RB-01` |
