---
title: "FT-043: Command Arg Codex Session Evidence"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для определения CodexSessionRef из безопасных argv evidence."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-005/brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../FT-020/brief.md
  - ../FT-021/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-043: Command Arg Codex Session Evidence

## Что

### Проблема

`instances detect` остается в `candidate`, когда несколько `session_meta` файлов
имеют один `opened_path`. При этом live Codex pane может уже содержать явную
identity в command argv, например `codex resume <uuid>` или wrapper-injected
external session UUID.

### Результат

`instances detect` извлекает UUID из безопасных command evidence и регистрирует
record как `active`, если остальные active invariants известны.

### Объем Работ

- `REQ-01` Извлекать UUID из `codex resume <uuid>` для direct Codex command.
- `REQ-02` Поддержать npm/node entrypoint form, где command выглядит как
  `node .../codex resume <uuid>`.
- `REQ-03` Поддержать explicit external session UUID evidence из
  `CODEX_EXTERNAL_SESSION_UUID=<uuid>` или `External session UUID: <uuid>` в
  command args.
- `REQ-04` Не сохранять raw argv, prompt text или developer instructions в
  registry/output.
- `REQ-05` Не сканировать Codex session logs для candidate, у которого
  `codex_session` уже извлечен из command evidence.

### Что Не Входит

- `NS-01` Нет чтения live process environment через PID/procfs.
- `NS-02` Нет ручной disambiguation-команды.
- `NS-03` Нет утверждения, что external session UUID равен внутреннему Codex
  `session_meta.payload.session_id`.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Argv evidence влияет на identity, privacy и transition в `active`. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` `codex resume <uuid>` переводит detected pane в `active`.
- `EC-02` `node .../codex resume <uuid>` извлекает тот же UUID.
- `EC-03` Wrapper-injected external UUID извлекается без сохранения raw argv.
- `EC-04` Non-Codex command with UUID is rejected.

### Обязательное Покрытие

- Unit tests для command evidence parser.
- Detection tests для переноса argv UUID в candidate.
- CLI regression test для `instances detect` active summary без session log
  match.
