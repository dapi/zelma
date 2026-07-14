---
title: "UC-010: Dashboard/status backend для agent-сессий"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует milestone-2 сценарий предоставления агрегированного статуса zelma-сессий для dashboard или внешних agent UI."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-010: Dashboard/status backend для agent-сессий

## Goal

Dashboard или внешний agent UI получает агрегированный status snapshot по zelma-сессиям, zellij pane, tasks, stale state и recovery hints без прямого парсинга CLI help или registry internals.

## Primary Actor

Dashboard agent.

## Trigger

Milestone-2 вводит визуальный или programmatic status layer поверх текущих CLI команд.

## Preconditions

- Milestone-1 стабилизировал session registry, live list, detect, cleanup и supervisor flows.
- У backend есть доступ к repo-local `.zelma` и zellij adapter.

## Main Flow

1. Dashboard вызывает status backend endpoint или command.
2. Backend агрегирует registry, live zellij state и task metadata.
3. Backend возвращает normalized status model.
4. Dashboard отображает active/stale/blocked/completed sessions и suggested actions.

## Alternate Flows / Exceptions

- `ALT-01` Backend запускается без dashboard: status остается machine-readable.
- `EX-01` Zellij недоступен: snapshot содержит degraded status и recovery hints.

## Postconditions

- Внешний UI не зависит от внутренних форматов registry.
- Agentic workflow получает единый status source.

## Business Rules

- `BR-01` Это milestone-2 сценарий и не должен блокировать milestone-1 CLI.
- `BR-02` Status model должен быть версионирован.
- `BR-03` Сценарий должен иметь e2e-тест backend snapshot with active and stale instances.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-042` |
| ADR | `none` |
