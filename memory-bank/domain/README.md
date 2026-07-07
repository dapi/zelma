---
title: Domain Documentation Index
doc_kind: domain
doc_function: index
purpose: Навигация по domain-level документации шаблона. Читать для фиксации предметной модели, ubiquitous language, бизнес-правил, состояний, событий и bounded contexts.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Domain Documentation Index

Каталог `memory-bank/domain/` хранит предметную модель `zelma`: язык домена,
сущности, правила, состояния, события и bounded contexts вокруг управления
Codex-сессиями в `zellij panes`.

Этот слой описывает то, что должно оставаться истинным независимо от текущей
продуктовой инициативы или технической реализации: что такое `zelma session`,
какие свойства она содержит, как `.zelma/sessions.json` становится локальным
реестром, какие переходы состояния допустимы и где проходят границы между CLI,
`zellij`, Codex и skills.

Domain-документы не определяют market positioning, product metrics, UI design system, concurrency pattern, deployment config или implementation sequence.

## На Какие Вопросы Отвечает Domain

- Какие понятия существуют в предметной области и что они означают?
- Какие сущности, value objects, actors или aggregates важны для reasoning?
- Какие бизнес-правила и инварианты нельзя нарушать?
- Какие состояния и переходы допустимы?
- Какие domain events являются бизнес-значимыми фактами?
- Где проходят bounded contexts и language boundaries?

## Граница С `product/`

| Layer | Отвечает на вопросы | Не отвечает на вопросы |
| --- | --- | --- |
| `product/` | Зачем существует продукт, для кого он, какие outcomes и metrics важны | Какие domain entities, states, invariants и events существуют |
| `domain/` | Что истинно в предметной области и какие правила обязана соблюдать система | Почему именно эта аудитория приоритетна, как продукт позиционируется, какой roadmap выбран |

Пример:

- Product: "Пользователь должен видеть все Codex panes текущего репозитория
  через `zelma sessions list`".
- Domain: "`active` `zelma session` должна иметь `zellij session`, `zellij
  pane`, `codex session` и normalized opened path".

## Граница С Engineering

- `domain/context-map.md` описывает business bounded contexts и language ownership.
- `engineering/architecture.md` описывает code/module boundaries, runtime patterns, concurrency, error handling и configuration ownership.
- Если документ отвечает на вопрос "какое бизнес-правило истинно?", он принадлежит `domain/`.
- Если документ отвечает на вопрос "как это безопасно реализовать в системе?", он принадлежит `engineering/`.

## Аннотированный Индекс

- [Glossary](glossary.md) — ubiquitous language, термины, запрещенные двусмысленности и canonical names.
- [Domain Model](model.md) — ключевые domain concepts, relationships, ownership и model notes.
- [Domain Rules](rules.md) — бизнес-правила, инварианты, policies и rule ownership.
- [States](states.md) — lifecycle states, allowed transitions и terminal states.
- [Events](events.md) — domain events как бизнес-значимые факты и их минимальный contract.
- [Context Map](context-map.md) — bounded contexts, upstream/downstream relations и language boundaries.
