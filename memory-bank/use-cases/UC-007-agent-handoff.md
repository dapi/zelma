---
title: "UC-007: Handoff между агентами"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий передачи контроля над активными zelma-сессиями новому агенту без устного контекста."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-007: Handoff между агентами

## Goal

Новый агент быстро восстанавливает картину активных работ по registry и live zellij state, чтобы продолжить supervision без повторного запуска задач.

## Primary Actor

Incoming agent.

## Trigger

Контекст предыдущего агента потерян, сессия Codex возобновлена или пользователь передает управление другому агенту.

## Preconditions

- `.zelma/instances.json` содержит сохраненные session metadata.
- Часть pane может быть live, stale или неизвестной.

## Main Flow

1. Incoming agent вызывает `zelma instances list --live --json`.
2. Zelma возвращает active/stale/candidate status по каждой записи.
3. Агент сопоставляет pane с issue/task metadata.
4. Агент продолжает poll, cleanup или recovery без повторного запуска уже активных задач.

## Alternate Flows / Exceptions

- `ALT-01` Registry отсутствует: агент продолжает через `zelma instances list --json`, который auto-detects по умолчанию.
- `ALT-02` Есть stale entries: агент применяет cleanup flow.
- `EX-01` Live state недоступен: агент получает diagnostic и не делает destructive action.

## Postconditions

- Новый агент понимает, какие pane уже работают.
- Повторные task agents не запускаются для уже активных задач.

## Business Rules

- `BR-01` Handoff должен опираться на JSON, а не на human prose.
- `BR-02` Duplicate launch для активной issue должен предотвращаться.
- `BR-03` Сценарий должен иметь e2e-тест handoff after registry reload.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-039` |
| ADR | `none` |
