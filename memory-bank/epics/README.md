---
title: Epics Index
doc_kind: epic
doc_function: index
purpose: "Навигация по instantiated epic packages. Читать, когда инициатива крупнее одной feature и должна исполняться через roadmap и набор связанных subissues."
derived_from:
  - ../dna/governance.md
  - ../flows/epic-flow.md
  - ../flows/feature-flow.md
status: active
audience: humans_and_agents
---

# Epics Index

Каталог `memory-bank/epics/` хранит instantiated epic packages вида `EP-XXX/`.

## Правила

- Epic описывает крупное проектное изменение, которое нельзя безопасно реализовать одной delivery-feature.
- Epic владеет intent, roadmap, декомпозицией, decision log, рисками и реестром subissues.
- Epic не владеет code-level execution: реализация идёт через отдельные `memory-bank/features/FT-<issue>/` packages.
- Каждый delivery subissue должен ссылаться на соответствующие epic artifacts и project-level `UC-*`, если меняет устойчивый сценарий.
- Правила создания и ведения epic packages живут в [`../flows/epic-flow.md`](../flows/epic-flow.md).

## Именование

- Базовый формат: `EP-XXX/`
- Вместо `XXX` используй стабильный идентификатор инициативы: issue id, project id или другое устойчивое имя
- Один epic = одна крупная программа/инициатива с несколькими delivery-slices

## Package Layers

| Layer | Files | Purpose |
| --- | --- | --- |
| Intake | `brief.md` | Ранний набросок problem/outcome и ссылки на candidate features до полной готовности epic |
| Intent | `charter.md`, source refs, stakeholder channels | Зачем существует epic, что входит/не входит, какие facts уже подтверждены |
| Governance | `roadmap.md`, `decision-log.md`, `risks.md`, `subissues.md` | Как исполнять epic, какие решения приняты, какие риски и subissues управляются |
| Knowledge | `design.md`, `specs/**`, `diagrams/**`, linked `UC-*` | Нормализованные требования, bounded contexts, сценарии, контракты и audit trail |
| Execution Handoff | future `memory-bank/features/FT-<issue>/` | Конкретные code changes, тесты, rollout/backout для одного approved delivery issue |

Knowledge-файлы опциональны. Если они создаются как Markdown внутри epic package, они должны быть индексированы из package `README.md` или owner-документа и следовать правилам frontmatter из [`../flows/epic-flow.md`](../flows/epic-flow.md).

## Созданные Epic Packages

- [EP-001: Go CLI Foundation](EP-001/README.md) — активный foundational epic
  для Go CLI: binary scaffold, дерево команд Cobra, agent-first help и
  output/error contract.
- [EP-002: Registry And Repo State](EP-002/README.md) — активный package для
  repo-local registry sessions, `.zelma/instances.json` и setup behavior.
- [EP-003: Zellij Read Integration And Detect](EP-003/README.md) — активный
  package для zellij introspection и консервативного обнаружения ручных panes.
- [EP-004: Managed Create Workflow](EP-004/README.md) — активный package для
  создания Codex panes через `zelma instances create`.
- [EP-005: Codex Session Identity](EP-005/README.md) — активный package для
  надежных ссылок на Codex session.
- [EP-006: Agent Skill Pack](EP-006/README.md) — активный package для Codex
  skills, построенных поверх стабильного CLI contract.
- [EP-007: Reconciliation And Lifecycle](EP-007/README.md) — активный package
  для обработки stale/live lifecycle и cleanup proposals.
- [EP-008: Autonomous Issue Shipping Supervisor](EP-008/README.md) — draft
  brief для supervisor-агента, который ведет issue через `start-issue`,
  `/review`, PR, CI, merge и notification.
