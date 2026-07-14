---
title: "UC-002: Взятие вручную созданной pane под контроль"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует сценарий обнаружения codex pane, созданной человеком вне zelma, и записи ее в registry."
derived_from:
  - ../product/context.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-002: Взятие вручную созданной pane под контроль

## Goal

Zelma находит zellij pane, где уже запущен Codex, и добавляет ее в `.zelma/instances.json`, чтобы дальнейший supervisor мог управлять этой работой как обычной zelma-сессией.

## Primary Actor

Supervising agent.

## Trigger

Пользователь вручную создал pane и запустил Codex, после чего агент вызывает `zelma instances list --json`.

## Preconditions

- В zellij есть одна или несколько pane.
- Некоторые pane могут быть Codex-сессиями, но не все должны быть приняты автоматически.
- Registry доступен для чтения и записи.

## Main Flow

1. Агент вызывает `zelma instances list --json`.
2. Zelma получает список pane и классифицирует Codex candidates.
3. Zelma idempotent upsert записывает уверенные совпадения в registry.
4. Zelma возвращает JSON с добавленными, уже известными и пропущенными pane.

## Alternate Flows / Exceptions

- `ALT-01` Pane уже есть в registry: запись не дублируется.
- `ALT-02` Уверенности недостаточно: pane возвращается как candidate, но не становится active.
- `EX-01` Registry поврежден: команда возвращает recovery hint без потери данных.

## Postconditions

- Обнаруженные Codex pane доступны в `instances list`; standalone `instances detect` остается diagnostic/manual вариантом.
- Не-Codex pane не записываются как active instances.

## Business Rules

- `BR-01` Detect должен быть идемпотентным.
- `BR-02` При сомнении нельзя записывать pane как active.
- `BR-03` Сценарий должен иметь e2e-тест с несколькими pane и повторным запуском detect.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-034` |
| ADR | `none` |
