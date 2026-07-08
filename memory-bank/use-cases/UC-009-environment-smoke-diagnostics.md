---
title: "UC-009: Smoke-диагностика окружения"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий быстрой проверки готовности repo/zellij/codex/registry перед agentic работой."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-009: Smoke-диагностика окружения

## Goal

Агент перед запуском работ проверяет, что repo root, `.gitignore`, zellij, Codex command и registry доступны и дают предсказуемые diagnostics.

## Primary Actor

Setup agent.

## Trigger

Первый запуск zelma в репозитории, fresh clone или подготовка перед parallel delivery.

## Preconditions

- Репозиторий доступен локально.
- У агента есть shell-доступ к CLI.

## Main Flow

1. Агент вызывает `zelma setup --json`.
2. Zelma проверяет repo root и идемпотентно добавляет `.zelma` в `.gitignore`.
3. Агент вызывает `zelma sessions list --json`.
4. Агент вызывает `zelma sessions detect --json`.
5. Агент получает summary готовности и warnings.

## Alternate Flows / Exceptions

- `ALT-01` `.gitignore` уже содержит `.zelma`: setup не дублирует строку.
- `EX-01` Zellij недоступен: diagnostic объясняет, какие команды невозможны.
- `EX-02` Repo root не найден: команда возвращает actionable error.

## Postconditions

- Репозиторий подготовлен к zelma workflow без неявного создания registry-файла.
- Агент знает, можно ли запускать create/detect/supervisor.

## Business Rules

- `BR-01` Setup должен быть идемпотентным.
- `BR-02` `.zelma` должен быть исключен из git.
- `BR-03` `zelma setup` не должен создавать `.zelma/sessions.json`; готовность
  registry проверяется отдельным `sessions list --json`.
- `BR-04` Сценарий должен иметь e2e-тест fresh repo setup plus repeated setup.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-041` |
| ADR | `none` |
