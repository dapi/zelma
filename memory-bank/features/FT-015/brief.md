---
title: "FT-015: Codex Launch Contract"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для command/process contract запуска Codex в managed zellij pane."
derived_from:
  - ../../product/roadmap.md
  - ../../epics/EP-004/brief.md
  - ../../engineering/architecture.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-015: Codex Launch Contract

## Что

### Проблема

`instances create` должен запускать Codex предсказуемо: в нужном path, с
ожидаемой командой и без скрытых assumptions о пользовательском shell.

### Результат

Определен launch contract: какую команду запускает `zelma`, какой working
directory используется и какие failures считаются recoverable.

### Объем Работ

- `REQ-01` Определить command line для запуска Codex.
- `REQ-02` Определить working directory/opened path behavior.
- `REQ-03` Определить diagnostics при отсутствии Codex или ошибке launch.

### Что Не Входит

- `NS-01` Нет zellij pane creation.
- `NS-02` Нет Codex session identity parsing.
- `NS-03` Нет установки Codex CLI.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Launch contract влияет на user environment и create failure modes. | `design.md` |

## Проверка

- `SC-01` Команда launch формируется с ожидаемым path.
- `SC-02` Missing Codex дает понятную ошибку без registry write.

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02` | unit tests for command builder | expected command/path | `artifacts/ft-015/verify/chk-01/` |
| `CHK-02` | `REQ-03` | fake missing binary test | recoverable diagnostic | `artifacts/ft-015/verify/chk-02/` |

### Доказательства

- `EVID-01` Command builder test output.
- `EVID-02` Missing Codex diagnostic test output.
