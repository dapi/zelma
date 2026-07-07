---
title: Marketing And Positioning
doc_kind: product
doc_function: canonical
purpose: Каноничное место для positioning, messaging, go-to-market channels, competitive alternatives и launch constraints.
derived_from:
  - ../dna/governance.md
  - context.md
  - customers.md
status: active
audience: humans_and_agents
canonical_for:
  - product_positioning
  - product_messaging
  - go_to_market_context
---

# Marketing And Positioning

Этот документ фиксирует, как продукт объясняется рынку, customers и internal stakeholders. Он не заменяет PRD и не определяет implementation scope.

## Positioning

| Audience | Current alternative | Product difference | Proof |
| --- | --- | --- | --- |
| `SEG-01` | Ручной `zellij` layout, pane names, shell history, личная память | `zelma` создает repo-local registry и дает CLI-инвентарь Codex panes | MVP target; evidence pending |
| `SEG-02` | Custom scripts вокруг `zellij`/Codex | `zelma` задает общий domain contract и skill-friendly CLI | MVP target; evidence pending |
| `SEG-03` | Неформальные team conventions | `.zelma/sessions.json` и команды `sessions *` делают workflow повторяемым | Assumption; evidence pending |

## Messaging

- `MSG-01` `zelma` turns Codex panes in `zellij` into manageable repo-local sessions.
- `MSG-02` Create sessions when you want control; detect them when you already
  started manually.
- `MSG-03` One CLI contract for humans and Codex skills.

## Channels

| Channel | Audience | Goal | Constraint | Owner |
| --- | --- | --- | --- | --- |
| `README.md` | Developers evaluating the repo | Activation | Must stay concise and match implemented CLI | Maintainer |
| `memory-bank/product/roadmap.md` | Maintainers and agents | Alignment | Must not pretend features are already implemented | Maintainer |
| Codex skill docs | Agentic workflow users | Retention / repeat use | Depends on first skill implementation | Maintainer |

## Competitive Alternatives

- `ALT-01` Manual `zellij` pane management.
- `ALT-02` Ad hoc shell aliases or scripts.
- `ALT-03` `tmux`-specific workflows outside current product scope.
- `ALT-04` Doing nothing and relying on Codex session logs after the fact.

## Launch Constraints

- `LC-01` До MVP launch должны работать `sessions create`, `sessions detect` и
  `sessions list` на поддерживаемой версии `zellij`.
- `LC-02` Нельзя обещать надежный detect без явного списка поддерживаемых Codex и
  `zellij` versions.
- `LC-03` Нельзя заявлять multi-machine или GUI support до отдельного roadmap
  решения.
