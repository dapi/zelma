---
title: "UC-004: Supervisor orchestration для GitHub issue"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий запуска issue shipping agent через supervisor/start-issue в zellij."
derived_from:
  - ../product/context.md
  - ../prompts/PROMPT-005-start-issue-shipping-supervisor.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-004: Supervisor orchestration для GitHub issue

## Goal

Shipping supervisor запускает task agent для конкретного GitHub issue, наблюдает его pane и доводит работу до PR/merge/cleanup по утвержденному процессу.

## Primary Actor

Shipping supervisor.

## Trigger

Пользователь или план запускает работу по GitHub issue через `start-issue`.

## Preconditions

- Issue существует и доступен через GitHub CLI/API.
- Zellij доступен для запуска task agent.
- Prompt supervisor описывает review/fix/merge loop.

## Main Flow

1. Supervisor запускает task agent в zellij pane или tab согласно настройке.
2. Task agent выполняет issue shipping prompt.
3. Supervisor регулярно poll-ит pane, фиксирует completion и инициирует review/fix loop.
4. После merge supervisor подтягивает `main`, закрывает отработанную pane и запускает следующие задачи по плану.

## Alternate Flows / Exceptions

- `ALT-01` Issue уже выполнен: supervisor верифицирует состояние и не запускает лишнюю реализацию.
- `EX-01` Review нашел замечания: fix loop повторяется до чистого review.
- `EX-02` CI сломался: supervisor запускает корректирующий цикл и не merge-ит до green.

## Postconditions

- Issue закрыт или явно отмечен blocked с доказательствами.
- Отработанная zellij pane закрыта, registry очищен или обновлен.

## Business Rules

- `BR-01` Review/fix loop завершается только после review без замечаний.
- `BR-02` Supervisor должен poll-ить активные pane не реже одного раза в минуту.
- `BR-03` Сценарий должен иметь e2e-тест supervisor happy path и review-fix повтор.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-036` |
| ADR | `none` |
