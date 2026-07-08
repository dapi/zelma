---
title: "UC-008: Параллельная доставка несколькими агентами"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий запуска и контроля нескольких independent issue agents в zellij."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-008: Параллельная доставка несколькими агентами

## Goal

Supervisor запускает несколько независимых issue agents, контролирует их progress, избегает конфликтов worktree/lock и последовательно вливает готовые результаты.

## Primary Actor

Shipping supervisor.

## Trigger

Есть набор GitHub issues, часть которых можно выполнять параллельно.

## Preconditions

- Для задач определены зависимости и параллельные группы.
- Zellij поддерживает несколько pane/tab.
- Git worktree strategy или startup delay снижает конфликты lock.

## Main Flow

1. Supervisor строит execution order с parallel groups.
2. Supervisor запускает task agents с задержкой между стартами.
3. Supervisor poll-ит каждую pane не реже одного раза в минуту.
4. Готовые PR проходят review/fix/CI/merge loop.
5. После merge supervisor подтягивает `main` и запускает следующую доступную задачу.

## Alternate Flows / Exceptions

- `ALT-01` Задача завершилась без кода: supervisor запускает implementation prompt перед review loop, если issue требует implementation.
- `EX-01` Git lock занят: startup или retry policy повторяет операцию до timeout.
- `EX-02` Worktree уже существует: supervisor переиспользует или явно очищает его по безопасному правилу.

## Postconditions

- Все задачи группы либо merged, либо blocked с evidence.
- Нет дублирующих agents по одной issue.

## Business Rules

- `BR-01` Параллельно запускаются только независимые задачи.
- `BR-02` Между стартами parallel agents должна быть пауза или retry на lock.
- `BR-03` Сценарий должен иметь e2e-тест multi-pane orchestration with simulated completions.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-040` |
| ADR | `none` |
