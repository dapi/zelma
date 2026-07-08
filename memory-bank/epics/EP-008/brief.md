---
title: "EP-008: Brief Autonomous Issue Shipping Supervisor"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для supervisor-агента, который автономно ведет issue через start-issue, review, PR, CI и merge."
derived_from:
  - ../../product/roadmap.md
  - ../../engineering/zellij-integration.md
  - ../../engineering/codex-runtime-identification.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---

# EP-008: Brief Autonomous Issue Shipping Supervisor

## Проблема

Ручной workflow вокруг `start-issue` требует постоянного наблюдения: нужно
запустить агента в zellij, дождаться реализации, вручную вызвать `/review`,
попросить исправить findings, создать или проверить PR, дождаться CI, проверить
mergeability, выполнить merge и закрыть pane. Это повторяемая agent-operations
работа, в которой легко потерять шаг или преждевременно объявить задачу
готовой.

## Результат

Появляется supervisor-команда, которая принимает GitHub issue и автономно
управляет lifecycle delivery: запускает `start-issue` в новой zellij pane по
умолчанию или в tab только по явному выбору пользователя,
наблюдает за pane, выполняет review/fix и CI/fix cycles, требует clean review,
mergeable PR и зеленый CI, пушит in-scope fixes через исполнительного агента,
вливает PR по явной policy и отправляет desktop notification о результате.

## Набросок Объема

- `EP-008-REQ-01` Запускать `start-issue <issue>` в новой zellij pane с
  контролируемым cwd, repo, base и agent prompt; tab разрешен только как явный
  launch-surface override пользователя.
- `EP-008-REQ-02` Хранить prompt как редактируемый template, а не hardcoded
  строку внутри supervisor.
- `EP-008-REQ-03` Поддерживать project-local override prompt через `.zelma`,
  например `.zelma/prompts/ship-issue.md`, и явный `--prompt-file`.
- `EP-008-REQ-04` Наблюдать за pane через zellij API/CLI, снимать screen
  snapshots, определять завершение фаз и отправлять команды в pane.
- `EP-008-REQ-05` Автоматически запускать `/review` после implementation phase и
  повторять review/fix loop до clean review или terminal gate.
- `EP-008-REQ-06` Проверять PR state: URL, draft flag, mergeability,
  mergeStateStatus, branch, pushed commits и review decision.
- `EP-008-REQ-07` Проверять CI через GitHub checks; если CI отсутствует,
  pending, cancelled или failed, не объявлять success без явного policy/gate.
- `EP-008-REQ-08` Выполнять CI/fix loop: передавать failed logs агенту,
  дожидаться fix commit/push и заново проверять review/CI gates.
- `EP-008-REQ-09` Вливать PR только когда policy разрешает auto-merge и gates
  выполнены: clean review, green CI, non-draft, mergeable/clean.
- `EP-008-REQ-10` После terminal outcome закрывать task zellij surface, сохранять
  state/log и отправлять desktop notification.

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

## Candidate Feature Briefs

- `FT-032` Supervisor command and zellij launch.
- `FT-033` Editable prompt template and `.zelma` override.
- `FT-034` Pane observation and completion detection.
- `FT-035` Review/fix loop orchestration.
- `FT-036` PR, mergeability and CI gate.
- `FT-037` Merge, cleanup and desktop notification.

## Заметки О Готовности

- Нужен product decision по default auto-merge policy: `off`, `on with flag` или
  `on for trusted repos`.
- Нужен design для state model: `.zelma/runs/issue-<id>.json`,
  resumability, timeouts и cycle limits.
- Нужна явная политика для repositories без CI: terminal blocker, allowed
  documented exception или automatic CI-bootstrap через отдельную feature.
