---
title: "EP-001: Roadmap"
doc_kind: epic
doc_function: roadmap
purpose: "Execution waves, dependencies, gates and stop rules for Go CLI Foundation."
derived_from:
  - charter.md
status: active
audience: humans_and_agents
must_not_define:
  - code_steps
  - final_database_schema
  - production_rollout_dates
---

# EP-001: Roadmap

## Waves

| Wave | Target | Depends on | Exit gate |
| --- | --- | --- | --- |
| `W1` | Go module scaffold and empty binary | Go toolchain | `go test ./...` runs; `zelma` binary can be built |
| `W2` | Cobra command tree | `W1` | `zelma`, `zelma sessions`, and three subcommands route |
| `W3` | Agent-first help templates | `W2` | help/output contract tests pass |
| `W4` | Output/error stubs | `W2`, `W3` | command stubs return predictable status and diagnostics |

## First Slice Recommendation

Start with [FT-001: Go Module Scaffold](../../features/FT-001/README.md).

This slice should create the Go module and minimal binary only. It should avoid
registry and `zellij` integration so architecture can be verified without side
effects.

## Handoff Gates

| Gate | Required evidence |
| --- | --- |
| `GATE-01` Feature scope ready | `FT-001/brief.md` active and linked from `subissues.md` |
| `GATE-02` Implementation ready | `FT-001/implementation-plan.md` active before code changes begin |
| `GATE-03` Epic next wave ready | `FT-001` done or explicitly stopped; next feature package created |

## Stop Rules

- Stop if implementation requires live `zellij`; move that scope to `EP-003` or
  `EP-004`.
- Stop if help/output contract needs an architectural decision that conflicts
  with ADR-001.
- Stop if Go toolchain is unavailable and scaffold cannot be verified locally.
