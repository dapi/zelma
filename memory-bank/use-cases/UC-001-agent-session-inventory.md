---
title: "UC-001: Инвентаризация agent-сессий"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий получения machine-readable картины текущих zelma/codex pane для supervising agent."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-001: Инвентаризация agent-сессий

## Goal

Supervising agent получает актуальный список zelma-сессий, их zellij session/pane, codex session, workspace path и live-state, чтобы решить, какие задачи уже запущены и где нужно вмешательство.

## Primary Actor

Supervising agent.

## Trigger

Агент начинает управление работами или делает очередной poll активных pane.

## Preconditions

- Команда запускается внутри репозитория с установленным `zelma`.
- `.zelma/instances.json` может существовать или отсутствовать.
- Zellij может быть доступен или недоступен; ответ должен оставаться понятным агенту.

## Main Flow

1. Агент вызывает `zelma instances list --live --json`.
2. Zelma читает registry и сверяет записи с zellij.
3. Zelma возвращает JSON со списком сессий, статусами, идентификаторами pane и путями.
4. Агент использует результат для маршрутизации дальнейших действий.

## Alternate Flows / Exceptions

- `ALT-01` Registry пустой: возвращается пустой список без ошибки.
- `EX-01` Zellij недоступен: JSON содержит диагностический статус, пригодный для автоматического recovery.

## Postconditions

- Registry не меняется.
- Агент получает данные без необходимости парсить human output.

## Business Rules

- `BR-01` JSON output является главным контрактом сценария.
- `BR-02` Live-проверка не должна удалять записи из registry.
- `BR-03` Сценарий должен иметь e2e-тест с fake/realistic zellij adapter.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-033` |
| ADR | `none` |
