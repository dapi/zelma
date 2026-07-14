---
title: "UC-003: Управляемый запуск новой agent-сессии"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий создания новой zellij pane с Codex через zelma и записи session metadata."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-003: Управляемый запуск новой agent-сессии

## Goal

Агент создает новую Codex-сессию через `zelma instances create`, получает подтвержденную pane и registry-запись, пригодную для последующего мониторинга.

## Primary Actor

Supervising agent.

## Trigger

Нужно запустить нового task agent в текущем репозитории или заданном workspace path.

## Preconditions

- Zellij CLI доступен.
- Codex command доступна в окружении.
- Repo root определен корректно.

## Main Flow

1. Агент вызывает `zelma instances create --json`.
2. Zelma запускает Codex в новой zellij pane.
3. Zelma подтверждает созданную pane через zellij adapter.
4. Zelma записывает session metadata в `.zelma/instances.json`.
5. Агент получает JSON с идентификаторами pane/session и workspace path.

## Alternate Flows / Exceptions

- `ALT-01` Пользователь указал path: pane открывается в этом path, если он валиден.
- `EX-01` Pane создана, но подтверждение не прошло: команда возвращает recovery hints.
- `EX-02` Codex command недоступна: registry не должен получить active запись.

## Postconditions

- Успешный запуск виден в `instances list --live --json`.
- Частичный сбой не оставляет ложную active-сессию без диагностики.

## Business Rules

- `BR-01` Registry пишется только после подтверждения pane.
- `BR-02` Вывод команды должен быть agent-first JSON.
- `BR-03` Сценарий должен иметь e2e-тест create-to-list.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-035` |
| ADR | `none` |
