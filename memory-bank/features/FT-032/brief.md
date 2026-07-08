---
title: "FT-032: Supervisor Command And Zellij Launch"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для запуска start-issue supervisor task agent в zellij pane по умолчанию или tab по явному env/config override."
derived_from:
  - ../../epics/EP-008/brief.md
  - ../../ops/config.md
  - ../../prompts/PROMPT-005-start-issue-shipping-supervisor.md
status: draft
delivery_status: planned
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-032: Supervisor Command And Zellij Launch

## Что

### Проблема

Supervisor для `start-issue` должен запускать task agent в zellij так, чтобы
типовой workflow не переключал пользователя в новую tab и не мешал вводу в
основной рабочей вкладке. При этом некоторым пользователям все еще нужен launch
в отдельной tab, поэтому surface должен быть управляемым через явный override.

### Результат

Supervisor выбирает launch surface по deterministic precedence: env, затем
repo-local `.zelma/config.json`, затем default `pane`. По умолчанию `start-issue`
запускается в новой zellij pane текущей session. Tab используется только при
явном значении `tab`.

### Объем Работ

- `REQ-01` Добавить supervisor launch entrypoint для запуска `start-issue <issue>` в текущей zellij session.
- `REQ-02` Поддержать launch surface `pane` и `tab`.
- `REQ-03` Сделать `pane` default launch surface.
- `REQ-04` Резолвить launch surface с приоритетом `ZELMA_START_ISSUE_ZELLIJ_SURFACE`, затем `.zelma/config.json` key `start_issue.zellij_surface`, затем default `pane`.
- `REQ-05` Валидировать launch surface и останавливать запуск с agent-friendly configuration error для значений кроме `pane` и `tab`.
- `REQ-06` Записывать в run state выбранный surface, source настройки, command, cwd, `pane_id` и `tab_id`, если применимо.
- `REQ-07` Не создавать отдельную zellij session для issue.

### Что Не Входит

- `NS-01` Нет полного review/fix loop orchestration.
- `NS-02` Нет PR, CI, merge или notification gates.
- `NS-03` Нет реализации prompt override merge beyond existing prompt selection contract.
- `NS-04` Нет глобального daemon.
- `NS-05` Нет отдельной zellij session per issue.
- `NS-06` Нет попытки гарантировать, что zellij tab launch не переключит focus; tab является явным user opt-in.

### Ограничения И Предположения

- `ASM-01` `.zelma/config.json` optional и описан в `../../ops/config.md`.
- `ASM-02` Supervisor запускается из существующей zellij session или имеет доступ к zellij CLI control surface.
- `CON-01` Env override должен иметь более высокий приоритет, чем repo-local config.
- `CON-02` Default должен оставаться `pane`, чтобы типовой запуск не открывал новую tab.
- `CON-03` Launch должен оставаться в текущей zellij session, а не создавать per-issue session.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | Фича меняет CLI/env/config contract и zellij integration behavior. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` Без env и config supervisor запускает task agent в zellij pane.
- `EC-02` При `.zelma/config.json` со `start_issue.zellij_surface=tab` supervisor выбирает tab.
- `EC-03` При `ZELMA_START_ISSUE_ZELLIJ_SURFACE=pane` env переопределяет config `tab`.
- `EC-04` Invalid surface value возвращает configuration error до zellij launch.
- `EC-05` Ни один launch path не создает отдельную zellij session для issue.

### Матрица Трассировки

| ID требования | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- |
| `REQ-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `EC-01`, `EC-02`, `SC-01`, `SC-02` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-03` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-04` | `EC-02`, `EC-03`, `SC-02`, `SC-03` | `CHK-02`, `CHK-03` | `EVID-02`, `EVID-03` |
| `REQ-05` | `EC-04`, `SC-04` | `CHK-04` | `EVID-04` |
| `REQ-06` | `SC-01`, `SC-02` | `CHK-05` | `EVID-05` |
| `REQ-07` | `EC-05`, `SC-01`, `SC-02` | `CHK-06` | `EVID-06` |

### Сценарии Приемки

- `SC-01` Agent запускает supervisor без env/config; task agent стартует в новой pane named `issue-<id>`.
- `SC-02` Repo содержит `.zelma/config.json` с `start_issue.zellij_surface` равным `tab`; task agent стартует в новой tab named `issue-<id>`.
- `SC-03` Repo config просит `tab`, но env `ZELMA_START_ISSUE_ZELLIJ_SURFACE=pane`; task agent стартует в pane.
- `SC-04` Env или config содержит invalid surface; supervisor не вызывает zellij и возвращает понятную ошибку.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `REQ-01`, `REQ-02`, `REQ-03` | fake zellij/supervisor launch test without env/config | invoked pane launch command | `artifacts/ft-032/verify/chk-01/` |
| `CHK-02` | `REQ-02`, `REQ-04` | fake repo config test with `tab` | invoked tab launch command | `artifacts/ft-032/verify/chk-02/` |
| `CHK-03` | `REQ-04` | env-over-config test | env value wins over config | `artifacts/ft-032/verify/chk-03/` |
| `CHK-04` | `REQ-05` | invalid env/config value test | configuration error before zellij launch | `artifacts/ft-032/verify/chk-04/` |
| `CHK-05` | `REQ-06` | run-state assertion | surface, source, command, cwd and ids recorded | `artifacts/ft-032/verify/chk-05/` |
| `CHK-06` | `REQ-07`, `NS-05` | command invocation/static assertion | no zellij session creation path | `artifacts/ft-032/verify/chk-06/` |

### Доказательства

- `EVID-01` Default pane launch test output.
- `EVID-02` Config-driven tab launch test output.
- `EVID-03` Env precedence test output.
- `EVID-04` Invalid config diagnostic test output.
- `EVID-05` Run-state contract test output.
- `EVID-06` No per-issue session launch assertion.
