---
title: "FT-019: Codex Metadata Source Discovery"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для поиска надежных источников metadata, связывающих zellij pane и Codex session."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-005/brief.md
  - ../../engineering/codex-runtime-identification.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-019: Codex Metadata Source Discovery

## Что

### Проблема

Нельзя стабильно ссылаться на Codex session, пока не понятно, какие metadata
источники доступны: process info, session logs, env или другие runtime traces.

### Результат

Задокументирован и протестирован discovery path для доступных metadata sources
с оценкой надежности и privacy constraints.

### Объем Работ

- `REQ-01` Найти candidate sources Codex session metadata.
- `REQ-02` Зафиксировать confidence для каждого source.
- `REQ-03` Отделить допустимые metadata от приватного content.

### Что Не Входит

- `NS-01` Нет parser implementation.
- `NS-02` Нет хранения conversation content.
- `NS-03` Нет обязательной зависимости от нестабильного private API.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Identity source selection влияет на privacy и reliability. | `design.md` |

## Проверка

- `SC-01` Discovery report перечисляет usable metadata sources.
- `SC-02` Privacy-sensitive content explicitly excluded.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | discovery doc/review | sources and confidence recorded | `artifacts/ft-019/verify/chk-01/` |
| `CHK-02` | `REQ-03` | privacy review | content excluded | `artifacts/ft-019/verify/chk-02/` |

### Доказательства

- `EVID-01` Metadata source inventory.
- `EVID-02` Privacy boundary review note.
