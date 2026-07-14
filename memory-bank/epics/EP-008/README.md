---
title: "EP-008: Autonomous Issue Shipping Supervisor"
doc_kind: epic
doc_function: index
purpose: "Навигация по epic: implemented local supervisor lifecycle и remaining real GitHub PR/CI/merge gates."
derived_from:
  - ../../dna/governance.md
  - brief.md
status: active
audience: humans_and_agents
---

# EP-008: Autonomous Issue Shipping Supervisor

## О разделе

Implemented baseline запускает `start-issue` в zellij, наблюдает pane markers,
ведет локальный review/fix/re-review cycle и закрывает pane после **merge
simulation**. Он не читает реальное состояние GitHub PR/CI и не выполняет
GitHub merge. Эта remaining capability принадлежит open issue
[#111](https://github.com/dapi/zelma/issues/111) и требует отдельной merge
policy.

## Аннотированный индекс

- [Brief](brief.md)
  Читать, когда нужно понять implemented local scope и boundary future GitHub
  gates.

- [PROMPT-005: Start Issue Shipping Supervisor](../../prompts/PROMPT-005-start-issue-shipping-supervisor.md)
  Читать, когда нужно запустить generic supervisor-сессию без hardcoded repo.

## Delivery Evidence And Next Work

- [`FT-032`](../../features/FT-032/README.md): implemented supervisor command,
  launch surface resolution and zellij launch.
- [`FT-036`](../../features/FT-036/README.md): implemented local supervisor
  orchestration e2e, including merge simulation.
- [#111](https://github.com/dapi/zelma/issues/111): future real GitHub PR/CI
  gates and explicit merge policy.
