---
title: Product Vision
doc_kind: product
doc_function: canonical
purpose: Каноничное место для долгосрочного направления продукта, strategic bets, experience principles и product non-goals.
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
canonical_for:
  - product_vision
  - product_strategy_principles
---

# Product Vision

Этот документ фиксирует устойчивое направление продукта. Он должен помогать принимать решения между competing features, но не заменяет PRD, roadmap или domain rules.

## Product Promise

`zelma` обещает разработчику понятный control plane над Codex-сессиями,
запущенными в `zellij`: пользователь может создать сессию, обнаружить уже
запущенную вручную сессию и увидеть инвентарь активных Codex panes текущего
репозитория.

Долгосрочно `zelma` должна стать тонким, предсказуемым coordination layer для
agentic development: CLI остается источником команд, `.zelma/instances.json`
остается локальным реестром, а Codex skills используют эти контракты для
управления параллельными рабочими сессиями.

## Strategic Bets

| Bet ID | Bet | Why now | Evidence | Review cadence |
| --- | --- | --- | --- | --- |
| `BET-01` | Локальный registry-first подход лучше скрытой глобальной автоматики | Пользователю нужна прозрачность и versionable reasoning вокруг repo-local Codex работы | Product prompt `2026-07-07` | После MVP `instances create/list/detect` |
| `BET-02` | CLI должен быть canonical interface для skills | Skills проще поддерживать, если они вызывают те же команды, что и человек | Архитектурное предположение | При проектировании первого skill |
| `BET-03` | Авто-detect важен не меньше controlled create | Пользователи уже создают panes вручную и не будут всегда начинать через `zelma` | Product prompt `2026-07-07` | После реализации detection fixtures |

## Experience Principles

- `XP-01` Explicit over magical: каждая управляемая сессия должна быть видна в
  `.zelma/instances.json` и объяснима через `instances list`.
- `XP-02` Idempotent by default: повторный `instances detect` не должен плодить
  дубликаты или менять unrelated panes.
- `XP-03` Human and agent parity: команда, которой пользуется skill, должна быть
  пригодна для ручного запуска человеком.
- `XP-04` Minimal takeover: `zelma` управляет только panes, которые она создала
  или корректно распознала как Codex panes.
- `XP-05` Agent-first help: `zelma` и `zelma help` сначала помогают агенту
  выбрать безопасную следующую команду, а уже затем объясняют продукт человеку.
  Help text должен быть коротким, структурированным, без marketing prose и с
  copy-ready командами.

## Product Non-Goals

- `PNG-01` Не заменять `zellij` и не реализовывать собственный terminal
  multiplexer.
- `PNG-02` Не заменять Codex CLI и не управлять содержимым Codex-диалога как
  conversational memory manager.
- `PNG-03` Не строить распределенный multi-machine orchestration layer в MVP.
- `PNG-04` Не оптимизировать первый scope под GUI; первичный surface — CLI и
  Codex skills.

## Decision Rules

- Если две инициативы дают сопоставимый impact, приоритет получает та, которая
  улучшает `WF-01`, `WF-02` или `WF-03` из [`context.md`](context.md).
- Если изменение затрагивает свойства `zelma instance`, сначала обнови
  [`../domain/model.md`](../domain/model.md) и [`../domain/rules.md`](../domain/rules.md).
- Если skill требует возможности, которой нет в CLI, сначала спроектируй CLI
  contract, затем добавляй skill wrapper.
- Если меняется команда или flag, одновременно обнови agent-first help contract
  и tests, которые проверяют `zelma help` / command help.
- Если automation может случайно повлиять на non-Codex pane, выбирай более
  консервативное поведение и требуй явного подтверждения или evidence.

## Source Documents

- Исходное описание продукта от пользователя в текущей рабочей сессии
  `2026-07-07`.
