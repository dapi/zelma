---
title: "FT-031: Zelma Setup Gitignore"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для `zelma setup`: идемпотентно добавить `.zelma` в `.gitignore` текущего repo."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-002/brief.md
  - ../../features/FT-005/brief.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-031: Zelma Setup Gitignore

## Что

### Проблема

Repo-local registry живет в `.zelma/`. Если этот каталог не добавлен в
`.gitignore`, пользователь или агент может случайно закоммитить session state.

### Результат

`zelma setup` находит repo root и идемпотентно добавляет `.zelma` в
repo-local `.gitignore`, не ломая существующее содержимое файла.

### Объем Работ

- `REQ-01` Найти repo root тем же способом, что и registry commands.
- `REQ-02` Создать `.gitignore`, если его нет.
- `REQ-03` Добавить `.zelma` как отдельную ignore-запись, если ее еще нет.
- `REQ-04` Повторный запуск `zelma setup` не должен дублировать `.zelma`.
- `REQ-05` Вернуть agent-friendly summary о том, изменен файл или уже был готов.

### Что Не Входит

- `NS-01` Нет создания `.zelma/sessions.json`.
- `NS-02` Нет изменения global gitignore.
- `NS-03` Нет удаления или переписывания существующих пользовательских правил.
- `NS-04` Нет zellij или Codex integration.

### Ограничения И Предположения

- `ASM-01` FT-005 repo root resolver доступен или реализуется раньше.
- `CON-01` Команда должна быть идемпотентной.
- `CON-02` `.zelma` должна попадать в repo-local `.gitignore`, а не только в diagnostics.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Команда меняет пользовательский `.gitignore`, поэтому нужен явный filesystem contract и failure behavior. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` Если `.gitignore` отсутствует, `zelma setup` создает его и добавляет `.zelma`.
- `EC-02` Если `.gitignore` уже содержит `.zelma`, повторный запуск не меняет файл.
- `EC-03` Если `.gitignore` содержит другие правила, они сохраняются.

### Матрица Трассировки

| ID требования | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `EC-02`, `SC-02` | `CHK-02` | `EVID-02` |
| `REQ-05` | `EC-01`, `EC-02` | `CHK-03` | `EVID-03` |

### Сценарии Приемки

- `SC-01` Agent запускает `zelma setup` в repo без `.gitignore`; команда создает файл с `.zelma`.
- `SC-02` Agent запускает `zelma setup` повторно; `.gitignore` остается без duplicate entries.
- `SC-03` Agent запускает `zelma setup` в repo с существующими ignore rules; правила сохраняются.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | CLI/filesystem test с temp repo | `.gitignore` содержит `.zelma` | `artifacts/ft-031/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | запустить setup дважды и сравнить файл | нет duplicate `.zelma` line | `artifacts/ft-031/verify/chk-02/` |
| `CHK-03` | `REQ-05` | CLI output assertion | summary сообщает changed или already configured | `artifacts/ft-031/verify/chk-03/` |

### Доказательства

- `EVID-01` Test output для отсутствующего `.gitignore`.
- `EVID-02` Output idempotency test.
- `EVID-03` Output CLI summary assertion.
