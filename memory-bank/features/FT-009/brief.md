---
title: "FT-009: Sessions List Output"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для `zelma sessions list` как первого read surface над registry."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-002/brief.md
  - ../../engineering/architecture.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-009: Sessions List Output

## Что

### Проблема

После появления registry пользователю и агенту нужен простой read command,
чтобы увидеть known zelma sessions текущего repo.

### Результат

`zelma sessions list` показывает зарегистрированные sessions в человекочитаемом
виде и в стабильном JSON mode для агентов.

### Объем Работ

- `REQ-01` Читать registry текущего repo.
- `REQ-02` Выводить table/default representation для человека.
- `REQ-03` Выводить stable JSON representation для агентов.
- `REQ-04` Обрабатывать empty registry без ошибки.

### Что Не Входит

- `NS-01` Нет live проверки zellij panes.
- `NS-02` Нет detect/create behavior.
- `NS-03` Нет stale cleanup.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Output format и JSON mode являются agent-facing contract. | `design.md` |

## Проверка

- `SC-01` Empty registry выводит пустой список с exit code 0.
- `SC-02` Non-empty registry выводит zellij/codex/path fields в JSON.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-04` | CLI tests с temp registry | empty list завершается успешно | `artifacts/ft-009/verify/chk-01/` |
| `CHK-02` | `REQ-02`, `REQ-03` | golden output tests | table and JSON match contract | `artifacts/ft-009/verify/chk-02/` |

### Доказательства

- `EVID-01` Empty registry test output.
- `EVID-02` Golden output test output.
