---
title: "FT-005: Repo Root Resolver"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для определения корня репозитория, относительно которого `zelma` работает с repo-local `.zelma/` и `.gitignore`."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-002/brief.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: active
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-005: Repo Root Resolver

## Что

### Проблема

Все команды, которые работают с repo-local `.zelma/` или `.gitignore`, должны
одинаково понимать, какой каталог является корнем проекта. Иначе
`.zelma/sessions.json`, `.zelma/` и `.gitignore` могут относиться к разным
каталогам, чем ожидают агенты и пользователи.

### Результат

`zelma` стабильно определяет repo root из вложенного каталога, нормализует путь
для downstream filesystem operations и явно сообщает ошибку, если команда
запущена вне поддерживаемого проекта.

### Объем Работ

- `REQ-01` Определить правила поиска repo root.
- `REQ-02` Нормализовать путь repo root для последующих repo-local filesystem operations.
- `REQ-03` Вернуть agent-friendly ошибку вне repo.

### Что Не Входит

- `NS-01` Нет чтения или записи `.zelma/sessions.json`.
- `NS-02` Нет multi-repo/global registry.
- `NS-03` Нет zellij integration.
- `NS-04` Нет изменения `.gitignore`; это scope `zelma setup`.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Repo root resolution становится filesystem contract для всех команд. | `design.md` |

## Проверка

- `SC-01` Из вложенного каталога команда находит один и тот же repo root.
- `SC-02` Вне repo команда завершает работу с понятной диагностикой.

### Negative / Edge Coverage

- `NEG-01` Запуск из каталога без поддерживаемого repo marker не должен
  создавать repo-local state или принимать текущий каталог за repo root.

### Traceability Matrix

| Requirement | Acceptance scenarios | Checks | Evidence |
| --- | --- | --- | --- |
| `REQ-01` | `SC-01`, `SC-02`, `NEG-01` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-02` | `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `SC-02`, `NEG-01` | `CHK-02` | `EVID-02` |

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | unit tests с temp directories | stable normalized root | `artifacts/ft-005/verify/chk-01/` |
| `CHK-02` | `REQ-03` | CLI/error test вне repo | agent-friendly error | `artifacts/ft-005/verify/chk-02/` |

### Доказательства

- `EVID-01` Test output для repo root cases.
- `EVID-02` Captured diagnostic для запуска вне repo.
