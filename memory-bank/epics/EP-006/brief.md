---
title: "EP-006: Brief Agent Skill Pack"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для Codex skill pack, который управляет zelma через стабильный CLI contract."
derived_from:
  - ../../product/roadmap.md
  - ../../product/context.md
  - ../../engineering/architecture.md
status: active
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-006: Brief Agent Skill Pack

## Проблема

`zelma` ориентирована на agent-first workflows, но skills должны быть thin
wrappers над CLI, а не вторым независимым implementation path.

## Результат

Codex skills вызывают `zelma instances list/create/detect`, понимают stable JSON
output и дают agents recovery guidance без дублирования domain logic.

## Набросок Объема

- `EP-006-REQ-01` Создать skill command wrappers.
- `EP-006-REQ-02` Проверять compatibility с machine-readable CLI output.
- `EP-006-REQ-03` Документировать agent usage contract.
- `EP-006-REQ-04` Описать recovery flows для common failures.

## Что Не Входит

- `EP-006-NS-01` Нет прямой работы skills с zellij в обход `zelma`.
- `EP-006-NS-02` Нет альтернативного registry parser внутри skills.
- `EP-006-NS-03` Нет UI/interactive wizard как первичного интерфейса.

## Briefs Фич

- [FT-023: Skill Command Wrappers](../../features/FT-023/README.md)
- [FT-024: Machine-Readable Output Compatibility Tests](../../features/FT-024/README.md)
- [FT-025: Skill Docs](../../features/FT-025/README.md)
- [FT-026: Agent Recovery Flows](../../features/FT-026/README.md)

## Заметки О Готовности

- Работа над skills начинается после стабилизации CLI output contract, чтобы
  избежать лишних переделок.
