---
title: "UC-005: Восстановление после ошибок agent-сессии"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарии recovery для поврежденного registry, недоступного zellij, stale pane и частичных запусков."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-005: Восстановление после ошибок agent-сессии

## Goal

Агент получает точную диагностику и следующий безопасный шаг, когда session registry, zellij state или запуск Codex оказались в неконсистентном состоянии.

## Primary Actor

Supervising agent.

## Trigger

Команда `setup`, `detect`, `list`, `create` или `cleanup` возвращает ошибку или inconsistent state.

## Preconditions

- Ошибка воспроизводится через CLI.
- Данные пользователя не должны быть потеряны автоматически.

## Main Flow

1. Агент запускает команду с `--json`.
2. Zelma классифицирует ошибку и возвращает structured diagnostic.
3. Агент выполняет предложенный recovery command или безопасно останавливает workflow.
4. После recovery агент повторяет исходную команду.

## Alternate Flows / Exceptions

- `ALT-01` Ошибка устраняется `zelma setup`: команда обновляет `.gitignore` и
  проверяет repo root без неявного создания registry-файла.
- `ALT-02` Ошибка устраняется `instances detect`: registry синхронизируется с live pane.
- `EX-01` Ошибка требует человека: diagnostic явно помечает manual action.

## Postconditions

- Агент знает, можно ли retry-ить автоматически.
- Registry и zellij state не повреждаются recovery-командами.

## Business Rules

- `BR-01` Diagnostics должны быть machine-readable и human-readable.
- `BR-02` Автоматический recovery не должен удалять данные без explicit confirm.
- `BR-03` Сценарий должен иметь e2e-тесты для основных классов ошибок.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-037` |
| ADR | `none` |
