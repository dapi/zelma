---
title: Customers And Users
doc_kind: product
doc_function: canonical
purpose: Каноничное описание customer/user segments, jobs to be done, pains, evidence и assumptions.
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
canonical_for:
  - product_customers
  - user_segments
  - jobs_to_be_done
---

# Customers And Users

Этот документ описывает людей, команды или организации, для которых создается продукт. Он не определяет domain entities: если customer segment совпадает по названию с domain concept, различай product-смысл и domain-смысл явно.

## Segments

| Segment ID | Segment | Job To Be Done | Current Pain | Success Signal | Evidence |
| --- | --- | --- | --- | --- | --- |
| `SEG-01` | Solo developer using Codex in `zellij` | Вести несколько Codex-сессий в одном репозитории и быстро понимать, какая задача где выполняется | Pane state живет только в голове пользователя и в текущем `zellij` layout | `instances list` дает полный и понятный инвентарь | Product prompt `2026-07-07` |
| `SEG-02` | Agentic development power user | Запускать параллельные Codex-сессии для разных задач и передавать управление через CLI/skills | Нет общего registry и устойчивого contract для skills | Skill может создать/detect/list сессии без ручного разбора panes | Product prompt `2026-07-07`; unvalidated |
| `SEG-03` | Small engineering team adopting Codex workflows | Стандартизировать работу с Codex panes внутри репозитория | У каждого участника свои shell aliases и naming conventions | Команда использует одинаковые CLI-команды и `.zelma` layout | Assumption; unvalidated |

## Users And Actors

| Actor ID | Actor | Uses product how | Decision power | Notes |
| --- | --- | --- | --- | --- |
| `ACT-01` | Developer | Запускает `zelma instances create/list/detect` из shell внутри репозитория | End user / operator | Основной пользователь MVP |
| `ACT-02` | Codex agent | Вызывает `zelma` через skill для управления panes и сессиями | Operator within user-approved workspace | Не должен обходить CLI и писать несовместимый registry |
| `ACT-03` | Maintainer | Определяет schema, CLI contracts, release policy и поддержку платформ | Admin / builder | Может менять domain rules через docs + code changes |

Если actor становится участником устойчивого сценария, use case фиксируй в [`../use-cases/README.md`](../use-cases/README.md).

## Research Inputs

- Исходное описание продукта от пользователя в текущей рабочей сессии
  `2026-07-07`.
- Интервью, telemetry и usability studies пока отсутствуют.

## Assumptions

- `ASM-01` Пользователь уже находится внутри `zellij` при работе с `zelma`.
- `ASM-02` Основной сценарий выполняется из корня Git-репозитория или из
  подкаталога, для которого можно надежно определить repo root.
- `ASM-03` Codex session можно идентифицировать достаточно надежно через
  доступные локальные признаки: процесс, pane command, session log или CLI output.
- `ASM-04` Skills должны оставаться thin wrappers над CLI, а не отдельной
  системой управления state.

## Must Not Assume

- `NA-01` Нельзя считать, что все Codex panes были созданы через `zelma`.
- `NA-02` Нельзя считать, что `zellij pane id` стабилен после закрытия pane или
  рестарта `zellij` без проверки runtime state.
- `NA-03` Нельзя считать, что пользователь хочет автоматически закрывать,
  kill-ить или перезаписывать panes при detect/list.
- `NA-04` Нельзя молча записывать в `.zelma/instances.json` неполную активную
  сессию без явного state вроде `candidate` или `stale`.
