---
title: Visible Zellij Shipping Dispatcher
doc_kind: ops
doc_function: runbook
purpose: Типовой алгоритм запуска issue shipping через видимые zellij tab/pane, git worktree и start-issue.
derived_from:
  - ../README.md
  - ../../prompts/PROMPT-005-start-issue-shipping-supervisor.md
  - ../../product/execution-order.md
status: active
audience: humans_and_agents
---

# Visible Zellij Shipping Dispatcher

## Summary

Этот runbook описывает штатный процесс agentic delivery, где пользователь видит
каждый shipper и task agent в zellij.

Иерархия процесса:

1. Dispatcher остается координатором волны и не реализует issue сам.
2. Dispatcher запускает отдельного single-issue shipper в новой
   zellij tab или pane.
3. Single-issue shipper работает по `PROMPT-005` и внутри своей zellij surface
   запускает `start-issue`.
4. `start-issue` создает отдельный git worktree, feature branch и запускает
   task agent в новой zellij pane.
5. Single-issue shipper доводит PR до clean review, green CI, merge и cleanup.

## Safety Notes

- Не меняй branch в основном репозитории `~/code/zelma`.
- Основной репозиторий должен оставаться на `main`.
- Любая implementation работа выполняется только в worktree, созданном
  `start-issue` или явным `git worktree add`.
- Не используй invisible/native subagents для shipping, если пользователь
  ожидает видеть работу в zellij.
- Если zellij action API не работает или зависает, остановись с blocker и
  не переходи на невидимый fallback без явного разрешения пользователя.
- Перед параллельными стартами делай паузу 15 секунд, чтобы снизить риск
  `git/config.lock`.

## Preflight

1. Проверь основной worktree:

   ```bash
   cd ~/code/zelma
   git status --short --branch
   git branch --show-current
   git pull --ff-only origin main
   ```

2. Если branch не `main` или есть tracked changes, остановись. Не делай
   implementation до решения, куда перенести изменения.
3. Untracked локальные артефакты не трогай без явного запроса пользователя.
4. Проверь zellij:

   ```bash
   zellij list-sessions --no-formatting
   zellij action list-panes --json --all
   ```

5. Если `zellij action ...` зависает, проверь session mismatch. Не запускай
   invisible fallback.

## Launch One Issue

1. Dispatcher создает видимую zellij tab для single-issue shipper:

   ```bash
   zellij action new-tab --name shipper-<issue> --cwd ~/code/zelma -- \
     codex --dangerously-bypass-approvals-and-sandbox
   ```

2. В shipper tab передай prompt `PROMPT-005` с переменными:

   ```text
   OWNER_REPO: dapi/zelma
   ISSUE_NUMBER: <issue>
   BASE_BRANCH: main
   REPO_PATH: ~/code/zelma
   AGENT: codex
   ZELLIJ_SURFACE: pane
   AUTO_MERGE: yes
   MAX_REVIEW_CYCLES: 5
   MAX_CI_CYCLES: 3
   ```

3. Single-issue shipper запускает:

   ```bash
   start-issue <issue> --repo dapi/zelma --base main --agent codex
   ```

4. `start-issue` обязан создать worktree и task agent pane. Если task agent
   не появился в zellij, это blocker.

## Observe And Gate

- Dispatcher poll-ит shipper/task panes не реже одного раза в минуту.
- Single-issue shipper не начинает review, если implementation issue завершился docs-only
  или без runtime/test изменений, требуемых acceptance.
- `/review` запускается на `GPT-5.5 Extra high`.
- Первый poll после `/review` делается примерно через 3 секунды, чтобы быстро
  пройти quiz/menu.
- После любого fix commit/push нужен новый fresh `/review` на новом head.
- PR merge допустим только при clean review, green CI, open non-draft PR,
  `MERGEABLE/CLEAN`.

## Shipper Acceptance Criteria

Работа single-issue shipper принимается только если:

1. Implementation выполнялась в отдельном `git worktree`; основной
   `~/code/zelma` оставался на `main`.
2. PR создан против правильной base branch и содержит только in-scope изменения.
3. Для implementation issue есть runtime/code/test/docs изменения по acceptance;
   docs-only результат не принимается, если issue не `feature_pack_only`.
4. Релевантные локальные проверки запущены и зафиксированы.
5. Fresh `/review` выполнен на `GPT-5.5 Extra high` по последнему `headRefOid`.
6. Все `critical/high/important` review findings исправлены или явно вынесены в
   blocker/human gate.
7. После каждого fix commit/push выполнен новый fresh `/review`.
8. Последний review чистый для актуального head.
9. GitHub checks присутствуют и green; отсутствующие checks не считаются green.
10. PR non-draft, mergeable, clean и без conflicts.
11. Если `AUTO_MERGE=yes`, PR merged и merge commit verified.
12. Issue закрыт через PR automation или explicit close с PR/commit evidence.
13. Task pane закрыта только после terminal outcome.
14. Shipper tab/pane закрывается только после финального отчета.
15. Финальный отчет содержит terminal status, issue, PR URL, last head SHA,
    merge SHA если есть, review cycles, CI status, mergeability, checks и
    blockers/human gates.

## Parallel Waves

1. Стартуй только независимые issues из текущей волны.
2. Для каждого issue dispatcher создает отдельную shipper tab.
3. Между стартами shipper tabs выдерживай паузу 15 секунд.
4. После merge каждого PR:
   - закрывай task pane после terminal outcome;
   - закрывай shipper tab только после отчета;
   - в основном `~/code/zelma` выполняй `git pull --ff-only origin main`;
   - запускай следующую доступную задачу.

## Current Milestone-1 Sequence

1. Сначала: `#72`.
2. Затем параллельно: `#64`, `#65`, `#69`.
3. Затем: `#66`.
4. Затем параллельно: `#68`, `#70`.
5. Затем: `#67`.
6. Затем: `#71`.

## Escalation

Остановись и сообщи пользователю факты, если:

- zellij action API зависает или не может создать tab/pane;
- основной worktree не на `main` или содержит tracked changes;
- `start-issue` не создает worktree/task pane;
- review model нельзя переключить на `GPT-5.5 Extra high`;
- CI checks отсутствуют, красные или недоступны после допустимых retries.
