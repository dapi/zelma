---
title: "EP-001: Go CLI Foundation"
doc_kind: epic
doc_function: canonical
purpose: "Фиксирует intent, scope и acceptance boundaries для foundational Go CLI epic."
derived_from:
  - ../../flows/epic-flow.md
  - ../../product/roadmap.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_issue_ids_not_approved
---

# EP-001: Go CLI Foundation

## Problem

`zelma` пока существует как product/domain/engineering documentation, но не
имеет запускаемого CLI. Без binary невозможно проверить agent-first help,
command tree, output/error contract и базовую developer workflow.

## Outcome

Есть installable/testable Go CLI skeleton:

- binary называется `zelma`;
- top-level command и `instances` command group существуют;
- `zelma`, `zelma help`, `zelma instances help` выводят agent-first help;
- пустые `instances list/create/detect` существуют как command stubs с
  predictable errors или placeholder behavior;
- local Go test command проходит.

## Stakeholder Channels

| Channel | ID / URL | Purpose |
| --- | --- | --- |
| GitHub repository | https://github.com/dapi/zelma | Source code, issues, PRs |
| Product roadmap | ../../product/roadmap.md | Epic/source roadmap alignment |
| ADR-001 | ../../adr/ADR-001-mvp-cli-architecture.md | Accepted architecture constraints |

## Scope

- `REQ-01` Create Go module and repository-local build/test commands.
- `REQ-02` Create `cmd/zelma` entrypoint and internal package layout compatible
  with ADR-001.
- `REQ-03` Use Cobra for command tree.
- `REQ-04` Implement top-level and `sessions` help as agent-first output.
- `REQ-05` Add command stubs for `instances list`, `instances create`,
  `instances detect`.
- `REQ-06` Add tests that lock help/output contract enough to prevent accidental
  fallback to generic Cobra help.

## Non-Scope

- `NS-01` No `.zelma/instances.json` schema or persistence implementation.
- `NS-02` No live `zellij` command execution.
- `NS-03` No Codex session identification.
- `NS-04` No GitHub Actions release pipeline.
- `NS-05` No Codex skill implementation.

## Source / Evidence Boundaries

| Source | Authority | Refresh rule |
| --- | --- | --- |
| ../../adr/ADR-001-mvp-cli-architecture.md | Accepted architecture decision | Update only through ADR lifecycle |
| ../../engineering/architecture.md | Engineering contract | Update when module/output boundaries change |
| ../../engineering/testing-policy.md | Testing policy | Update when test surfaces change |
| ../../product/roadmap.md | Product sequencing | Update when epic/features change |

## Acceptance

| Criterion | Check |
| --- | --- |
| Go scaffold exists | `go test ./...` can run once Go toolchain is available |
| Agent-first help exists | Help tests assert command map appears before prose |
| Command tree exists | `zelma instances list/create/detect --help` are routed by Cobra |
| No external side effects | Feature packages under this epic do not call live `zellij` unless explicitly moved to later epic |

## Handoff

Delivery work must be created as separate `memory-bank/features/FT-<id>/`
packages. First package: [FT-001](../../features/FT-001/README.md).
