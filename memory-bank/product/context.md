---
title: Product Context
doc_kind: product
doc_function: canonical
purpose: Каноничное project-wide описание продукта, проблемного пространства и top-level outcomes. Читать перед PRD, use cases и feature briefs, чтобы не повторять общий контекст в каждой delivery-единице.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
canonical_for:
  - project_product_context
  - product_problem_space
  - top_level_outcomes
must_not_define:
  - domain_model
  - domain_invariants
  - implementation_sequence
  - architecture_decision
---

# Product Context

Этот документ фиксирует общий продуктовый контекст проекта. Downstream-документы должны ссылаться на него, а не переписывать один и тот же background каждый раз.

PRD, если он нужен, уточняет отдельную инициативу относительно уже зафиксированного project-wide контекста.

## Boundary With PRD And Domain

- `product/context.md` — общий для всего проекта контекст: продукт, пользователи, ключевые product workflows, top-level outcomes и устойчивые product constraints.
- `prd/PRD-XXX-short-name.md` — инициативный слой: какая именно продуктовая проблема берется в работу сейчас, для каких пользователей и с каким scope.
- `domain/` — предметная модель: language, entities, states, invariants, events и bounded contexts, которые должны оставаться истинными независимо от текущей инициативы.
- Если новый документ просто повторяет общий фон проекта и не вводит initiative-specific scope, PRD создавать не нужно.

## Product Context

`zelma` — CLI-утилита и набор Codex skills для управления Codex-сессиями,
запущенными в panes `zellij`. Продукт рассчитан на разработчиков и agentic
workflows, где в одном репозитории одновременно существует несколько Codex
сессий, каждая открыта в своем pane и выполняет отдельную задачу.

Без `zelma` пользователь вручную помнит, в каком pane находится какая Codex
сессия, какой путь был открыт при запуске и какие panes уже можно считать
управляемыми. Это плохо масштабируется при параллельной работе, handoff между
агентами и повторном входе в контекст после перерыва.

`zelma` должна сделать этот слой явным: создать или обнаружить Codex pane,
связать его с `zellij session`, `zellij pane`, `codex session` и открытым путем,
а затем сохранить эту связь в `.zelma/sessions.json` в корне репозитория.

Граница продукта: `zelma` управляет учетом и запуском Codex-сессий в `zellij`;
она не заменяет Codex, не является терминальным мультиплексором и не берет на
себя управление содержимым Codex-диалога.

## Core Product Workflows

- `WF-01` Managed create: пользователь запускает `zelma sessions create`, CLI
  создает новый `zellij pane`, стартует в нем Codex и регистрирует `zelma
  session` в `.zelma/sessions.json`.
- `WF-02` Manual detect: пользователь вручную создал `zellij pane`, запустил в
  нем Codex, затем запускает `zelma sessions detect`; CLI обнаруживает pane и
  добавляет запись в реестр без дублирования уже известных сессий.
- `WF-03` Session inventory: пользователь запускает `zelma sessions list` и
  видит актуальный список `zelma sessions` текущего репозитория с привязкой к
  `zellij session`, pane, Codex session и path.
- `WF-04` Skill-driven management: Codex skill вызывает `zelma` CLI, чтобы
  создавать, обнаруживать или перечислять сессии без ручного разбора panes.
- `WF-05` Agent-first discovery: агент запускает `zelma` или `zelma help` без
  дополнительного контекста и сразу получает краткую, task-oriented карту
  команд, пригодную для выбора следующего действия. Человеческое объяснение
  допускается, но идет после агент-ориентированных подсказок.

Если workflow становится устойчивым canonical scenario с trigger, preconditions, main flow и postconditions, заведи отдельный `UC-*` в [`../use-cases/README.md`](../use-cases/README.md).

## Top-Level Outcomes

Подробные definitions и ownership метрик фиксируй в [`metrics.md`](metrics.md). Здесь оставь только краткий executive summary.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Доля live Codex panes текущего репозитория, корректно отраженных в `.zelma/sessions.json` | `unknown` | >= 95% после `sessions detect` | Fixtures + manual validation against `zellij` state |
| `MET-02` | Успешность `sessions create` | `unknown` | >= 99% в поддерживаемой среде | CLI integration tests |
| `MET-03` | Время до ответа `sessions list` | `unknown` | < 500 ms для обычного локального реестра | CLI benchmark / integration test |

## Product Constraints

- `PCON-01` `zellij` является обязательной runtime-зависимостью для управления
  panes; без него `zelma` может только читать/валидировать локальный реестр.
- `PCON-02` `.zelma/sessions.json` является repo-local source of truth для
  `zelma sessions`; глобальный кросс-репозиторный реестр не входит в первый
  scope.
- `PCON-03` `sessions detect` должен быть идемпотентным и не должен брать под
  контроль panes, в которых Codex не запущен.
- `PCON-04` CLI и skills должны использовать одну предметную модель; skill не
  должен обходить CLI-контракты прямой записью несовместимого JSON.
- `PCON-05` Help output для `zelma` и `zelma help` оптимизируется для агентов в
  первую очередь и для человека во вторую: сначала точные команды, условия,
  machine-readable flags и recovery hints; затем поясняющий текст.

Domain-level invariants и state rules фиксируй в [`../domain/rules.md`](../domain/rules.md) и [`../domain/states.md`](../domain/states.md).

## Source Documents

- Исходное описание продукта от пользователя в текущей рабочей сессии
  `2026-07-07`.
- Других upstream strategy docs пока нет.
