---
title: "EP-001: Go CLI Foundation"
doc_kind: epic
doc_function: index
purpose: "Навигация по epic package для Go CLI foundation: binary scaffold, Cobra command tree, agent-first help и output/error contract."
derived_from:
  - ../../dna/governance.md
  - brief.md
  - charter.md
status: active
audience: humans_and_agents
---

# EP-001: Go CLI Foundation

## О разделе

Этот epic package управляет foundational delivery для `zelma` CLI: создать
запускаемый Go binary, заложить Cobra command tree и зафиксировать
agent-first help/output contract до работы с registry и live `zellij`.

## Аннотированный индекс

- [Brief](brief.md)
  Читать, когда нужен компактный набросок problem/outcome и список feature
  briefs под epic.

- [Charter](charter.md)
  Читать, когда нужно понять problem, outcome, scope/non-scope и acceptance
  boundaries epic.

- [Roadmap](roadmap.md)
  Читать, когда нужно понять waves, dependencies, gates и first slice.

- [Subissues](subissues.md)
  Читать, когда нужно увидеть candidate/accepted delivery features под epic.

- [Risks](risks.md)
  Читать, когда нужно проверить epic-level риски и mitigations.

- [Decision Log](decision-log.md)
  Читать, когда нужно увидеть локальные решения epic, не требующие отдельного ADR.

## Delivery Packages

- [FT-001: Go Module Scaffold](../../features/FT-001/README.md)
- [FT-002: Cobra Command Tree](../../features/FT-002/README.md)
- [FT-003: Agent-First Help Templates](../../features/FT-003/README.md)
- [FT-004: Output And Error Contract Tests](../../features/FT-004/README.md)
