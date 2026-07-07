---
title: Zelma Documentation Index
doc_kind: project
doc_function: index
purpose: Корневая навигация по memory-bank проекта zelma. Читать сначала, чтобы понять продукт, предметную область, правила документации и будущий delivery flow.
derived_from:
  - dna/principles.md
  - dna/governance.md
status: active
audience: humans_and_agents
---

# Zelma Documentation Index

Каталог `memory-bank/` содержит durable knowledge layer проекта `zelma`.
Главный фокус текущей версии документации: зафиксировать продуктовую цель,
предметную модель `zelma sessions` и roadmap до того, как проект будет разбит
на epics и feature packages.

`zelma` управляет Codex-сессиями внутри `zellij panes`: создает управляемые
сессии, обнаруживает вручную запущенные Codex panes и хранит локальный реестр в
`.zelma/sessions.json`.

## Аннотированный индекс

- [`product/README.md`](product/README.md)
  Читать, когда нужно: понять зачем существует `zelma`, для кого она создается,
  какие workflows важны и каким roadmap движется проект.

- [`domain/README.md`](domain/README.md)
  Читать, когда нужно: зафиксировать язык `zelma sessions`, свойства записи,
  правила `.zelma/sessions.json`, lifecycle, события и bounded contexts.

- [`prd/README.md`](prd/README.md)
  Читать, когда нужно: описать продуктовую инициативу между общим product context и downstream feature packages.

- [`epics/README.md`](epics/README.md)
  Читать, когда нужно: вести крупную инициативу через roadmap, decision log, risks и набор связанных delivery subissues.

- [`use-cases/README.md`](use-cases/README.md)
  Читать, когда нужно: зарегистрировать устойчивый пользовательский или операционный сценарий проекта.

- [`prompts/README.md`](prompts/README.md)
  Читать, когда нужно: найти или завести reusable prompt-документ с исходной формулировкой и copyable улучшенной версией.

- [`ops/README.md`](ops/README.md)
  Читать, когда нужно: описать локальную разработку, окружения, релизы, конфигурацию и runbooks.

- [`engineering/README.md`](engineering/README.md)
  Читать, когда нужно: задать architecture patterns, frontend rules, testing policy, coding style, git workflow и границы автономии агента.

- [`dna/README.md`](dna/README.md)
  Читать, когда нужно: проверить SSoT rules, frontmatter contract и governance-правила документации.

- [`flows/README.md`](flows/README.md)
  Читать, когда нужно: создать epic/feature package, провести инициативу по lifecycle gates или использовать шаблон.

- [`adr/README.md`](adr/README.md)
  Читать, когда нужно: найти или завести Architecture Decision Record.

- [`features/README.md`](features/README.md)
  Читать, когда нужно: понять, где живут instantiated feature packages.
