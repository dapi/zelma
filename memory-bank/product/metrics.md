---
title: Product Metrics
doc_kind: product
doc_function: canonical
purpose: Каноничное место для product success metrics, baselines, targets, measurement ownership и instrumentation constraints.
derived_from:
  - ../dna/governance.md
  - context.md
status: active
audience: humans_and_agents
canonical_for:
  - product_metrics
  - success_measurement
---

# Product Metrics

Этот документ фиксирует метрики продукта и правила их измерения. Feature-level checks и test evidence остаются в feature package; здесь живут только product-level outcomes и measurement contract.

## North Star

| Metric ID | Metric | Why it matters | Current baseline | Target | Review cadence |
| --- | --- | --- | --- | --- | --- |
| `NSM-01` | Session inventory correctness | Главная ценность `zelma` — надежно знать, какие Codex-сессии существуют в `zellij` для текущего репозитория | `unknown` | >= 95% live Codex panes представлены корректными `active` records после `sessions detect` | Каждый MVP milestone |

## Product Metrics

| Metric ID | Metric | Owner | Baseline | Target | Measurement method | Source |
| --- | --- | --- | --- | --- | --- | --- |
| `MET-01` | Detection precision | Maintainer | `unknown` | >= 98% detected panes действительно содержат Codex | Integration fixtures + manual zellij checks | Test run |
| `MET-02` | Detection recall | Maintainer | `unknown` | >= 95% live Codex panes текущего repo обнаруживаются | Integration fixtures + manual zellij checks | Test run |
| `MET-03` | Create success rate | Maintainer | `unknown` | >= 99% на поддерживаемых zellij/Codex версиях | CLI integration tests | CI/local |
| `MET-04` | Registry write safety | Maintainer | `unknown` | 0 corrupted `sessions.json` в тестах concurrent/failed writes | Fault-injection tests | CI/local |
| `MET-05` | List latency | Maintainer | `unknown` | < 500 ms для типичного repo-local registry | Benchmark/integration test | CI/local |

## Guardrails

| Guardrail ID | Metric | Why it must not regress | Threshold | Response |
| --- | --- | --- | --- | --- |
| `GR-01` | Non-Codex pane takeover rate | Ошибочный takeover может нарушить пользовательскую работу в терминале | 0 known cases | Отключить unsafe detect path, добавить stricter evidence |
| `GR-02` | Duplicate active records for same pane | Дубликаты делают registry ненадежным | 0 в supported workflows | Исправить idempotency/keying before release |
| `GR-03` | Backward-incompatible registry schema changes | `.zelma/sessions.json` должен переживать обновления CLI | Только через documented migration | Добавить versioning/migration или отложить change |

## Instrumentation Constraints

- `ICON-01` До появления telemetry canonical source для метрик — local test
  output, fixtures и manual validation notes.
- `ICON-02` Runtime evidence от `zellij` и Codex может отличаться между
  версиями; тестовые fixtures должны фиксировать поддерживаемые версии.
- `ICON-03` Не собирать содержимое Codex conversation для продуктовых метрик.
  Достаточно metadata о panes, process/session refs и paths.

## Metric Change Policy

- Не меняй definition метрики внутри feature package без обновления этого документа или upstream PRD.
- Если feature вводит новую локальную метрику, держи ее в feature package до тех пор, пока она не станет shared product metric.
