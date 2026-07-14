---
title: "EP-001: Brief Go CLI Foundation"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для Go CLI Foundation: зачем нужен запускаемый skeleton и какие feature briefs его реализуют."
derived_from:
  - ../../product/roadmap.md
  - charter.md
  - roadmap.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-001: Brief Go CLI Foundation

## Проблема

`zelma` описана в memory-bank, но пока не имеет запускаемого CLI. Без binary
нельзя проверить command routing, agent-first help, machine-readable output и
developer workflow.

## Результат

Есть Go CLI skeleton с binary `zelma`, командой `setup`, группой `sessions`,
command stubs `list/create/detect`, agent-first help и contract tests для
базового output.

## Набросок Объема

- `EP-001-REQ-01` Создать Go module и buildable `cmd/zelma`.
- `EP-001-REQ-02` Завести Cobra command tree для `zelma setup` и `zelma instances`.
- `EP-001-REQ-03` Сделать help output пригодным для агентов в первую очередь.
- `EP-001-REQ-04` Зафиксировать predictable output/error contract тестами.

## Что Не Входит

- `EP-001-NS-01` Нет registry persistence и `.zelma/instances.json`.
- `EP-001-NS-02` Нет live `zellij` integration.
- `EP-001-NS-03` Нет Codex session identity.

## Briefs Фич

- [FT-001: Go Module Scaffold](../../features/FT-001/README.md)
- [FT-002: Cobra Command Tree](../../features/FT-002/README.md)
- [FT-003: Agent-First Help Templates](../../features/FT-003/README.md)
- [FT-004: Output And Error Contract Tests](../../features/FT-004/README.md)

## Заметки О Готовности

- `charter.md`, `roadmap.md`, `subissues.md`, `risks.md` и `decision-log.md`
  уже созданы для этого epic.
- `FT-001` имеет implementation plan; остальные feature briefs пока draft.
