---
title: "EP-008: Autonomous Issue Shipping Supervisor"
doc_kind: epic
doc_function: index
purpose: "Навигация по draft package epic для автономного orchestration вокруг start-issue: zellij pane, review/fix loops, PR, CI, merge и notification."
derived_from:
  - ../../dna/governance.md
  - brief.md
status: draft
audience: humans_and_agents
---

# EP-008: Autonomous Issue Shipping Supervisor

## О разделе

Draft package для supervisor-агента, который запускает разработку указанной
GitHub issue через `start-issue` в отдельной zellij pane/tab и доводит delivery
до отревьюшенного, mergeable PR с зеленым CI, исправленными review/CI issues,
запушенными коммитами и merge.

## Аннотированный индекс

- [Brief](brief.md)
  Читать, когда нужно понять scope autonomous issue shipping, prompt override
  model и candidate features.

- [PROMPT-005: Start Issue Shipping Supervisor](../../prompts/PROMPT-005-start-issue-shipping-supervisor.md)
  Читать, когда нужно запустить generic supervisor-сессию без hardcoded repo.

## Кандидатные Packages Фич

- `FT-032`: Supervisor command and zellij launch
- `FT-033`: Editable prompt template and `.zelma` override
- `FT-034`: Pane observation and completion detection
- `FT-035`: Review/fix loop orchestration
- `FT-036`: PR, mergeability and CI gate
- `FT-037`: Merge, cleanup and desktop notification
