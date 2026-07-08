---
title: "UC-006: Очистка stale-сессий после завершения задачи"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий обнаружения и удаления stale registry entries после завершения или закрытия pane."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-006: Очистка stale-сессий после завершения задачи

## Goal

Агент находит записи `.zelma/sessions.json`, которые больше не соответствуют live zellij pane, и удаляет их только через явное подтверждение.

## Primary Actor

Supervising agent.

## Trigger

Task agent завершил работу, pane закрыта или zellij session была перезапущена.

## Preconditions

- Registry содержит одну или несколько записей.
- Live zellij state может отличаться от registry.

## Main Flow

1. Агент вызывает `zelma sessions detect --json` или `zelma sessions list --live --json`.
2. Zelma помечает отсутствующие pane как stale.
3. Агент вызывает `zelma sessions cleanup --json` для proposal.
4. Агент вызывает `zelma sessions cleanup --confirm --json`, если proposal ожидаем.
5. Zelma удаляет stale entries и возвращает итоговый diff.

## Alternate Flows / Exceptions

- `ALT-01` Stale entries нет: cleanup возвращает empty proposal.
- `EX-01` Registry lock занят: команда возвращает retryable diagnostic.

## Postconditions

- Registry не содержит завершенные pane.
- Live sessions не удаляются.

## Business Rules

- `BR-01` Cleanup без confirm не меняет registry.
- `BR-02` Cleanup должен быть идемпотентным.
- `BR-03` Сценарий должен иметь e2e-тест proposal plus confirm.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-038` |
| ADR | `none` |
