---
title: "FT-101: Safe Message Sending To Codex Sessions"
doc_kind: feature
doc_function: index
purpose: "Навигация по feature package для безопасной отправки сообщения в существующую Codex session через публичный `zelma` CLI."
derived_from:
  - ../../dna/governance.md
  - brief.md
status: active
audience: humans_and_agents
---

# FT-101: Safe Message Sending To Codex Sessions

## О разделе

Каталог фиксирует feature package для GitHub issue #101. Сначала читай
`brief.md`, затем `design.md`, `implementation-plan.md` и `decision-log.md`.

## Аннотированный индекс

- [`brief.md`](brief.md)
  Читать, когда нужно понять problem space, scope, non-scope и canonical
  verify contract команды `zelma instances send`.
  Отвечает на вопрос: что должно быть доставлено и как это принимается.

- [`design.md`](design.md)
  Читать, когда нужно проверить selected design, readiness gate, contracts,
  invariants, failure modes и C4/component boundary.
  Отвечает на вопрос: как feature безопасно отправляет текст только в
  подтвержденную live Codex pane.

- [`implementation-plan.md`](implementation-plan.md)
  Читать, когда нужно выполнить реализацию по шагам, test surfaces и
  checkpoints.
  Отвечает на вопрос: как провести изменения от текущего состояния к приемке.

- [`decision-log.md`](decision-log.md)
  Читать, когда нужно увидеть принятые FPF-решения из issue #101 и решения
  review-improve.
  Отвечает на вопрос: почему выбраны конкретные command shape, readiness gate и
  message-source policy.
