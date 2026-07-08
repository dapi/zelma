---
title: "FT-041: Environment Smoke Diagnostics E2E"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для e2e-покрытия smoke diagnostics окружения."
derived_from:
  - ../../use-cases/UC-009-environment-smoke-diagnostics.md
status: draft
delivery_status: planned
milestone: milestone-1
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-041: Environment Smoke Diagnostics E2E

## Что

### Проблема

Перед agentic delivery нужно быстро доказать, что repo и окружение готовы: `.zelma`, `.gitignore`, zellij visibility и базовые команды работают предсказуемо.

### Результат

E2E-тест fresh repo проверяет `zelma setup --json`, идемпотентность `.gitignore`, `sessions list --json` и `sessions detect --json`.

### Объем Работ

- `REQ-01` Подготовить fresh temp repo fixture.
- `REQ-02` Проверить создание `.zelma` и добавление `.zelma` в `.gitignore`.
- `REQ-03` Проверить повторный setup без дублей.
- `REQ-04` Проверить базовые `list` и `detect` diagnostics.

### Что Не Входит

- `NS-01` Нет установки zellij/codex на машине пользователя.

## Проверка

### Критерии Готовности

- `EC-01` `.gitignore` содержит одну строку `.zelma`.
- `EC-02` Повторный setup идемпотентен.
- `EC-03` Недоступный zellij возвращает actionable diagnostic.

### Обязательное E2E-Покрытие

Feature считается готовой только после e2e-теста fresh repo setup plus repeated setup.
