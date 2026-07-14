---
title: "FT-005: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space документ для FT-005. Фиксирует выбранный подход к repo root resolver, contracts и failure modes без переопределения problem space или execution contract."
derived_from:
  - brief.md
  - decision-log.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
  - ../../engineering/architecture.md
  - ../../ops/config.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_005_scope
  - ft_005_acceptance_criteria
  - ft_005_evidence_contract
  - implementation_sequence
---

# FT-005: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `ALT-*`, `TRD-*`, `C4-*`, `SD-*`, `CTR-*`, `INV-*`, `FM-*`, `RB-*` |
| `decision-log.md` | Review-improve decision evidence | FPF rationale for closed questions |
| `../../adr/ADR-001-mvp-cli-architecture.md` | Architecture baseline | Accepted module boundary: `internal/repo` owns repo root and `.zelma/` paths |

## Context

`brief.md` requires one root detection behavior for all commands that later touch
repo-local `.zelma/` or `.gitignore`. The accepted architecture already assigns
repo filesystem concerns to `internal/repo`, while `ops/config.md` fixes the
default registry path below the detected root.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-01` | `C3` | FT-005 introduces the `internal/repo` component inside the existing CLI container and defines its collaboration boundary with CLI/registry/setup callers. | Inline C3 table below |

### C4 Artifact

| C3 element | Responsibility | Collaborates with | Boundary |
| --- | --- | --- | --- |
| `internal/repo` | Detect and normalize repo root; derive repo-local paths when later features need them | `internal/cli`, future `internal/registry`, future setup command | Does not read/write `.zelma/instances.json`, does not call `zellij`, does not mutate `.gitignore` |
| `internal/cli` | Calls resolver for commands that require repo context and renders diagnostics | `internal/repo` | Does not implement independent root discovery |
| future `internal/registry` | Uses resolved root to locate `.zelma/instances.json` | `internal/repo` | Owns registry schema/read/write, not root discovery |

## Selected Solution

- `SOL-01` Repo root detection ascends from the starting directory to the Git
  worktree root. This closes `REQ-01` and aligns with `.gitignore` ownership in
  issue 5 and FT-031.
- `SOL-02` Resolver returns a normalized absolute repo root path for downstream
  filesystem operations. This closes `REQ-02`.
- `SOL-03` Resolver returns a typed unsupported-repo error that CLI surfaces as
  an agent-friendly diagnostic. This closes `REQ-03`.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Use existing `.zelma/` as root marker | `zelma setup` must work before `.zelma/` exists. |
| `ALT-02` | Use `go.mod` as root marker | This would bind user repositories to Go projects, while product docs describe repo-local sessions generally. |
| `ALT-03` | Treat current working directory as root | Fails `SC-01` because nested directories would produce different roots. |

## Trade-offs

| Trade-off ID | Decision | Benefit | Cost / Risk |
| --- | --- | --- | --- |
| `TRD-01` | Support Git worktrees as the FT-005 repo boundary | Matches `.gitignore`, repo-local state and nested-directory acceptance | Non-Git project support is deferred outside FT-005 |

## Accepted Local Decisions

- `SD-01` Git worktree root is the supported repo boundary for FT-005; the
  rationale is recorded in `decision-log.md#dl-001-supported-repo-boundary`.
- `SD-02` Root detection is centralized in `internal/repo` so CLI, registry and
  setup features do not duplicate filesystem discovery.

## Contracts

| Contract ID | Input / Output | Producer / Consumer | Semantics / Constraints |
| --- | --- | --- | --- |
| `CTR-01` | starting directory -> normalized absolute repo root | `internal/repo` -> CLI/registry/setup callers | Start path may be root or nested directory; result is stable for all paths inside the same Git worktree. |
| `CTR-02` | starting directory outside Git worktree -> unsupported-repo error | `internal/repo` -> CLI | Error is distinguishable from IO/internal failures and safe for agent diagnostics. |

## Invariants

- `INV-01` Commands must not implement a second root discovery algorithm outside
  `internal/repo`.
- `INV-02` Resolver must not create, read, write or validate `.zelma/instances.json`.
- `INV-03` Resolver must not mutate `.gitignore`; setup behavior belongs to FT-031.

## Failure Modes

- `FM-01` Outside a Git worktree, resolver returns unsupported-repo instead of
  silently treating cwd as root.
- `FM-02` If filesystem probing fails, resolver reports an internal/path error
  separately from unsupported-repo so agents can distinguish setup problems from
  broken IO.

## Rollout / Backout

| Stage ID | Stage | Entry condition | Backout |
| --- | --- | --- | --- |
| `RB-01` | Add resolver behind tests before wiring command behavior | `brief.md` and `design.md` active | Remove resolver wiring; package remains without repo-local side effects |

## ADR / External Design Dependencies

| Artifact | Current status | Used for | Rule |
| --- | --- | --- | --- |
| `../../adr/ADR-001-mvp-cli-architecture.md` | `accepted` | `internal/repo` owns repo root and `.zelma/` paths | Must remain accepted before implementation starts |

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure / rollout refs |
| --- | --- | --- | --- |
| `REQ-01` | `SOL-01`, `TRD-01`, `C4-01`, `SD-01`, `SD-02` | `CTR-01`, `INV-01` | `FM-01`, `RB-01` |
| `REQ-02` | `SOL-02`, `C4-01`, `SD-02` | `CTR-01`, `INV-01` | `FM-02`, `RB-01` |
| `REQ-03` | `SOL-03`, `C4-01` | `CTR-02` | `FM-01`, `FM-02`, `RB-01` |
