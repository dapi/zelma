---
title: "FT-001: Go Module Scaffold"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для первого delivery slice: создать Go module scaffold и пустой `zelma` binary без registry/zellij side effects."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../product/roadmap.md
  - ../../epics/EP-001/charter.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
  - ../../engineering/architecture.md
  - ../../engineering/testing-policy.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-001: Go Module Scaffold

## What

### Problem

У проекта `zelma` есть принятая архитектура MVP CLI и roadmap, но нет Go module,
entrypoint и проверяемого binary skeleton. Без scaffold нельзя начать
реализацию command tree, help output и последующих registry/zellij features.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Go module exists | no `go.mod` | `go test ./...` can discover module packages | local command |
| `MET-02` | Binary entrypoint exists | no `cmd/zelma` | `go build ./cmd/zelma` succeeds | local command |
| `MET-03` | Scope remains side-effect free | no code | no runtime `.zelma/` writes and no live `zellij` invocation | code review + tests |

### Scope

- `REQ-01` Create `go.mod` for the repository.
- `REQ-02` Create `cmd/zelma/main.go` entrypoint.
- `REQ-03` Create minimal internal package layout aligned with ADR-001 without
  implementing registry or zellij behavior.
- `REQ-04` Ensure `go test ./...` and `go build ./cmd/zelma` are the canonical
  checks for this slice.
- `REQ-05` Keep CLI behavior minimal enough that `FT-002` can own Cobra command
  tree and `FT-003` can own agent-first help templates.

### Non-Scope

- `NS-01` No Cobra command tree beyond what is strictly necessary to compile, if
  any.
- `NS-02` No `sessions list/create/detect` behavior.
- `NS-03` No `.zelma/sessions.json` schema, read or write behavior.
- `NS-04` No live `zellij` execution.
- `NS-05` No Codex session identification.
- `NS-06` No GitHub Actions or release packaging.

### Constraints / Assumptions

- `ASM-01` Go toolchain will be installed before implementation starts.
- `CON-01` Package layout must remain compatible with
  [ADR-001](../../adr/ADR-001-mvp-cli-architecture.md).
- `CON-02` This feature must not introduce runtime side effects.
- `DEC-01` Cobra is selected by ADR-001, but full Cobra command tree is owned by
  the next feature unless scaffold requires a minimal dependency.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: no` | Architecture is already accepted in ADR-001 and this feature only creates scaffold. No new integration contract, schema or runtime side effect is selected here. | `none` |

## Verify

### Exit Criteria

- `EC-01` `go.mod` exists and declares the project module.
- `EC-02` `cmd/zelma/main.go` exists and builds.
- `EC-03` `go test ./...` succeeds.
- `EC-04` implementation does not call `zellij` or write `.zelma/sessions.json`.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-01`, `CON-01` | `EC-02`, `SC-01` | `CHK-02` | `EVID-02` |
| `REQ-03` | `CON-01`, `CON-02` | `EC-04`, `SC-02` | `CHK-03` | `EVID-03` |
| `REQ-04` | `ASM-01` | `EC-03` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-05` | `DEC-01`, `CON-02` | `EC-04`, `SC-02` | `CHK-03` | `EVID-03` |

### Acceptance Scenarios

- `SC-01` A developer or agent with Go installed can run `go test ./...` and
  `go build ./cmd/zelma` from repo root.
- `SC-02` The scaffold establishes package boundaries without attempting to
  access live `zellij`, Codex or `.zelma/sessions.json`.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-03`, `SC-01` | `go test ./...` | Command exits 0 | `artifacts/ft-001/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-01` | `go build ./cmd/zelma` | Command exits 0 | `artifacts/ft-001/verify/chk-02/` |
| `CHK-03` | `EC-04`, `SC-02` | code review / `rg -n "zellij|sessions.json|\\.zelma" cmd internal` | No runtime invocation/write behavior in scaffold | `artifacts/ft-001/verify/chk-03/` |

### Test matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-001/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-001/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-001/verify/chk-03/` |

### Evidence

- `EVID-01` Captured output for `go test ./...`.
- `EVID-02` Captured output for `go build ./cmd/zelma`.
- `EVID-03` Review note or command output proving no runtime side effects.

### Evidence contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Test output | implementer | `artifacts/ft-001/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Build output | implementer | `artifacts/ft-001/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Side-effect review note | implementer / reviewer | `artifacts/ft-001/verify/chk-03/` | `CHK-03` |
