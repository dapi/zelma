---
title: Engineering Documentation Index
doc_kind: engineering
doc_function: index
purpose: Навигация по engineering-level документации zelma.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Engineering Documentation Index

Каталог `memory-bank/engineering/` содержит инженерные правила реализации
`zelma`: архитектурные boundaries, testing policy, coding style, git workflow и
границы автономии агента. Runtime-код пока не создан, но стек реализации выбран:
Go CLI с zellij integration через внешний `zellij` binary.

- [Engineering Architecture Patterns](architecture.md) — code/module boundaries, runtime patterns, concurrency, error handling и configuration ownership. Domain bounded contexts живут отдельно в [`../domain/context-map.md`](../domain/context-map.md).
- [Zellij Integration Research](zellij-integration.md) — актуальные zellij CLI/API surfaces, Go library candidates и правила первого `zellij-adapter`.
- [Codex Runtime Identification Design](codex-runtime-identification.md) — правила evidence, ambiguity policy и extraction design для `CodexSessionRef`.
- [Codex Skill Contract](skill-contract.md) — agent-facing contract для Codex skills поверх `zelma instances list/create/detect/focus/cleanup`, machine-readable outputs, recovery и boundaries.
- [Frontend Engineering](frontend.md) — UI surfaces, frontend stack, component boundaries, design system integration и i18n.
- [Testing Policy](testing-policy.md) — правила тестирования, обязательные automated tests, sufficient coverage. Отвечает на вопрос: когда feature обязана иметь test cases и когда допустим manual-only verify.
- [Autonomy Boundaries](autonomy-boundaries.md) — границы автономии агента: автопилот, супервизия, эскалация. Отвечает на вопрос: что агент может делать сам, а где должен остановиться и спросить.
- [Coding Style](coding-style.md) — конвенции оформления кода, tooling и правила локальной сложности.
- [Git Workflow](git-workflow.md) — git-конвенции: commits, ветки, PR и optional worktrees.
- [ADR](../adr/README.md) — instantiated Architecture Decision Records проекта.
