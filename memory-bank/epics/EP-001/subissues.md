---
title: "EP-001: Subissues"
doc_kind: epic
doc_function: subissue_registry
purpose: "Delivery subissue registry for Go CLI Foundation."
derived_from:
  - charter.md
  - roadmap.md
status: active
audience: humans_and_agents
must_not_define:
  - code_steps
---

# EP-001: Subissues

## Registry

| ID | Candidate issue title | Roadmap wave | Source slices / UC | Status | Feature package |
| --- | --- | --- | --- | --- | --- |
| `EP-001-SI-001` | Go module scaffold and empty `zelma` binary | `W1` | `SLICE-01`, `EP-001 REQ-01`, `EP-001 REQ-02` | accepted | [FT-001](../../features/FT-001/README.md) |
| `EP-001-SI-002` | Cobra command tree for `zelma instances` | `W2` | `SLICE-01`, `EP-001 REQ-03`, `EP-001 REQ-05` | candidate | TBD |
| `EP-001-SI-003` | Agent-first help templates | `W3` | `EP-001 REQ-04`, `XP-05` | candidate | TBD |
| `EP-001-SI-004` | Output and error contract tests | `W4` | `EP-001 REQ-06` | candidate | TBD |

## Creation Rules

- Create GitHub subissue only after scope is approved.
- Create `memory-bank/features/FT-<id>/` after the issue exists or after a
  stable internal feature ID is accepted.
- Link the feature package back to this epic and relevant source docs.
