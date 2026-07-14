---
title: Product Roadmap
doc_kind: product
doc_function: canonical
purpose: Каноничное место для product themes, bets, horizons и dependencies без превращения roadmap в feature backlog.
derived_from:
  - ../dna/governance.md
  - context.md
  - vision.md
  - metrics.md
  - ../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
canonical_for:
  - product_roadmap
  - product_themes
---

# Product Roadmap

Этот документ описывает направление и sequencing продуктовых тем. Он не должен становиться списком всех feature packages: delivery-единицы живут в [`../features/README.md`](../features/README.md), а инициативы — в [`../prd/README.md`](../prd/README.md).

Практический порядок запуска GitHub issues и параллельных волн реализации
зафиксирован в [`execution-order.md`](execution-order.md).

## Horizons

| Horizon | Theme | Intended outcome | Candidate PRD / Feature | Dependency | Status | Next decision |
| --- | --- | --- | --- | --- | --- | --- |
| `implemented` | v0.1 delivery baseline | CLI, registry, list/create/detect, lifecycle and initial skill/release flows shipped | `EP-001`–`EP-007`, `FT-001`–`FT-031` | Released in v0.1.0 | implemented | Monitor follow-ups through open issues. |
| `implemented` | v0.2 hardening baseline | Numeric IDs, focus, stronger identity evidence and zellij e2e diagnostics shipped | `FT-041`–`FT-047` | Released in v0.2.0 | implemented | Monitor follow-ups through open issues. |
| `implemented` | v0.3 supervisor and status baseline | Local supervisor launch/poll/review/fix/re-review/cleanup simulation and status backend shipped | `FT-032`, `FT-033`–`FT-040`, `FT-042` | Released in v0.3.0 | implemented | Decide real GitHub gates in #111. |
| `implemented` | v0.4 observation and messaging | Read-only buffer/transcript, monitor, safe send and distributable skill shipped | `FT-048`, `FT-049`, `FT-101` | Released in v0.4.0 | implemented | Prioritize follow-up from real usage. |
| `next` | Real GitHub PR/CI/merge gates | Supervisor verifies live PR/CI state and follows an explicit merge policy | [#111](https://github.com/dapi/zelma/issues/111) | Current local supervisor lifecycle | open | Decide merge policy and GitHub integration contract. |
| `next` | Ambiguous identity resolution | User can safely resolve an ambiguous Codex identity without unsafe auto-selection | [#108](https://github.com/dapi/zelma/issues/108) | Current identity evidence model | open | Decide manual-resolution UX and evidence threshold. |
| `next` | Worktree isolation | Sessions are isolated by worktree with strict ownership | [#109](https://github.com/dapi/zelma/issues/109) | Repo-local registry baseline | open | Decide registry/worktree identity model. |
| `next` | Compatibility matrix | Supported zellij/Codex versions and CI canaries are explicit | [#110](https://github.com/dapi/zelma/issues/110) | Existing zellij/Codex integrations | open | Define supported-version policy. |
| `next` | Event-driven lifecycle follow | Users can wait/follow instance lifecycle and handoff events | [#112](https://github.com/dapi/zelma/issues/112) | Current lifecycle/handoff flows | open | Define event source and follow semantics. |
| `next` | Current zellij session resolution | Supervisor uses the current zellij session instead of a hardcoded target | [#113](https://github.com/dapi/zelma/issues/113) | Current zellij integration | open | Decide discovery and fallback behaviour. |

## Delivered Baseline

`EP-001`–`EP-007` are implemented historical delivery packages. Their detailed
scope remains in feature briefs, not in this roadmap. EP-008 has an implemented
local supervisor baseline; its real GitHub PR/CI/merge extension is future work
owned by issue #111.

## Roadmap Rules

- Roadmap theme описывает product intent, а не implementation plan.
- Если тема требует нескольких delivery slices, создай PRD и перечисли downstream features там.
- Если тема меняет предметную модель, сначала обнови [`../domain/model.md`](../domain/model.md), [`../domain/rules.md`](../domain/rules.md) или [`../domain/context-map.md`](../domain/context-map.md).
- Не заводи feature package только ради roadmap line item. Feature появляется,
  когда есть owner, scope, acceptance criteria и test strategy.

## Open Bets

- `OQ-01` Для #111: какая explicit merge policy допустима для real GitHub gates?
- `OQ-02` Для #108: какой evidence threshold достаточен для manual identity resolution?
- `OQ-03` Для #110: какие zellij/Codex versions составляют supported matrix?
