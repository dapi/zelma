---
title: "EP-008: Brief Autonomous Issue Shipping Supervisor"
doc_kind: epic
doc_function: brief
purpose: "Brief для implemented local supervisor lifecycle и remaining real GitHub PR/CI/merge gates."
derived_from:
  - ../../product/roadmap.md
  - ../../engineering/zellij-integration.md
  - ../../engineering/codex-runtime-identification.md
status: active
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-008: Brief Autonomous Issue Shipping Supervisor

## Проблема

Ручной workflow вокруг `start-issue` требовал постоянного наблюдения: нужно
запустить агента в zellij, дождаться реализации, вручную вызвать `/review`,
попросить исправить findings, создать или проверить PR, дождаться CI, проверить
mergeability, выполнить merge и закрыть pane. Это повторяемая agent-operations
работа, в которой легко потерять шаг или преждевременно объявить задачу
готовой.

## Результат

Текущая supervisor-команда принимает issue, запускает `start-issue` в zellij,
наблюдает за pane и выполняет local review/fix/re-review/cleanup lifecycle по
structured markers. Реальная проверка GitHub PR, CI, mergeability и merge пока
не реализована: это follow-up [#111](https://github.com/dapi/zelma/issues/111).

## Набросок Объема

- `EP-008-REQ-01` Запускать `start-issue <issue>` в новой zellij pane с
  контролируемым cwd, repo, base и agent prompt; tab разрешен только как явный
  launch-surface override пользователя.
- `EP-008-REQ-02` Хранить prompt как редактируемый template, а не hardcoded
  строку внутри supervisor.
- `EP-008-REQ-03` Поддерживать project-local override prompt через `.zelma`,
  например `.zelma/prompts/ship-issue.md`, и явный `--prompt-file`.
- `EP-008-REQ-03A` Поддерживать repo-local launch surface config через
  `.zelma/config.json` с приоритетом `ZELMA_START_ISSUE_ZELLIJ_SURFACE`, затем
  config, затем default `pane`.
- `EP-008-REQ-04` Наблюдать за pane через zellij API/CLI, снимать screen
  snapshots, определять завершение фаз и отправлять команды в pane.
- `EP-008-REQ-05` Автоматически запускать `/review` после implementation phase и
  повторять review/fix loop до clean review или terminal gate.
- `EP-008-FUT-01` Проверять PR state: URL, draft flag, mergeability,
  mergeStateStatus, branch, pushed commits и review decision.
- `EP-008-FUT-02` Проверять CI через GitHub checks; если CI отсутствует,
  pending, cancelled или failed, не объявлять success без явного policy/gate.
- `EP-008-FUT-03` Выполнять CI/fix loop: передавать failed logs агенту,
  дожидаться fix commit/push и заново проверять review/CI gates.
- `EP-008-FUT-04` Вливать PR только когда policy разрешает auto-merge и gates
  выполнены: clean review, green CI, non-draft, mergeable/clean.
- `EP-008-REQ-06` После terminal local outcome закрывать task zellij surface и
  сохранять state/log. Desktop notification is not an implemented claim.

## Что Не Входит

- `EP-008-NS-01` Нет обхода branch protection, required approvals или security
  gates.
- `EP-008-NS-02` Нет объявления success при отсутствующем CI, если policy явно
  требует green CI.
- `EP-008-NS-03` Нет silent merge по умолчанию без явной auto-merge policy.
- `EP-008-NS-04` Нет исправления unrelated findings за пределами issue scope.
- `EP-008-NS-05` Нет глобального daemon как обязательной части MVP.
- `EP-008-NS-06` Нет отдельной zellij session per issue в default workflow.

## Prompt Override Model

Supervisor должен строить final prompt из управляемых слоев:

- built-in default template, versioned вместе с CLI;
- repository-level editable template, если project policy хочет зафиксировать
  единый workflow;
- runtime override в `.zelma/prompts/ship-issue.md` для локальной настройки без
  коммита в git;
- explicit `--prompt-file path`, который имеет наивысший приоритет.

Override через `.zelma` должен быть discoverable и безопасным: supervisor обязан
показывать, какой prompt source выбран, и сохранять ссылку/хэш выбранного prompt
в run state.

## Delivery Boundary

- `FT-032` and `FT-036` evidence the implemented local supervisor baseline.
- Future `EP-008-FUT-01`–`EP-008-FUT-04` are owned by open issue #111; no
  current feature package is asserted for them.

## Заметки О Готовности

- Нужен product decision по default auto-merge policy: `off`, `on with flag` или
  `on for trusted repos` before #111 can implement a real merge.
- Нужна явная политика для repositories без CI: terminal blocker, allowed
  documented exception или separate CI-bootstrap feature.
