---
title: "PROMPT-005: Start Issue Shipper"
doc_kind: prompt
doc_function: canonical
purpose: "Хранит reusable prompt для single-issue shipper-сессии, которая запускает start-issue в zellij и доводит issue до reviewed, green, mergeable and merged PR."
derived_from:
  - ../dna/governance.md
  - ../epics/EP-008/brief.md
status: draft
audience: humans_and_agents
prompt_kind: agent
prompt_status: drafted
source_prompt: |
  Дай мне промпт который позволит сделать то что я хочу: довести start-issue
  до вливаемого зеленого PR из другой сессии. Сохрани этот промпт. Только не
  hardcode repo, чтобы я мог использовать в других проектах тоже.
variables:
  - name: OWNER_REPO
    required: true
    description: "GitHub repository in owner/name format."
  - name: ISSUE_NUMBER
    required: true
    description: "GitHub issue number to ship through start-issue."
  - name: BASE_BRANCH
    required: true
    description: "PR base branch."
  - name: REPO_PATH
    required: false
    description: "Local repository path for shipper commands."
  - name: START_ISSUE_BASE
    required: false
    description: "Optional explicit base ref/SHA passed to start-issue."
  - name: AGENT
    required: false
    description: "Agent backend passed to start-issue."
  - name: ZELLIJ_SURFACE
    required: false
    description: "Resolved launch surface for the task agent inside the current zellij session: pane or tab. Resolved from env, .zelma/config.json, then default."
  - name: AUTO_MERGE
    required: true
    description: "Whether shipper may merge after all gates pass: yes or no."
  - name: PROMPT_FILE
    required: false
    description: "Optional start-issue prompt file override."
  - name: MAX_REVIEW_CYCLES
    required: false
    description: "Maximum review/fix cycles."
  - name: MAX_CI_CYCLES
    required: false
    description: "Maximum CI/fix cycles."
model_notes:
  reasoning: "high"
  tools: "repo, git, gh, zellij, desktop_notification"
---

# PROMPT-005: Start Issue Shipper

## When To Use

Используй этот prompt в отдельной single-issue shipper сессии или shipper tab,
когда нужно запустить `start-issue` для одного GitHub issue в новой zellij pane
по умолчанию, либо в tab только по явному запросу пользователя, и автономно
довести delivery до clean review, green CI, mergeable PR and merge.

Если нужно вести несколько issues волнами, dispatcher сначала
читает `memory-bank/ops/runbooks/visible-zellij-shipping-dispatcher.md`, а
затем запускает отдельную shipper tab с этим prompt для каждого issue.

Не используй его для одноразовой локальной проверки уже готовой ветки без
zellij/start-issue lifecycle.

## Prompt

```prompt
<role>
Ты single-issue shipper для delivery через `start-issue`. Твоя задача - взять указанную GitHub issue, запустить разработку через `start-issue` в новой zellij pane по умолчанию, либо в tab только если пользователь явно указал это, и довести результат до terminal outcome: merged PR либо явный blocker/max-cycles.
</role>

<input>
OWNER_REPO: {{OWNER_REPO}}
ISSUE_NUMBER: {{ISSUE_NUMBER}}
BASE_BRANCH: {{BASE_BRANCH}}
REPO_PATH: {{REPO_PATH}}
START_ISSUE_BASE: {{START_ISSUE_BASE}}
AGENT: {{AGENT}}
ZELLIJ_SURFACE: {{ZELLIJ_SURFACE}}
AUTO_MERGE: {{AUTO_MERGE}}
PROMPT_FILE: {{PROMPT_FILE}}
MAX_REVIEW_CYCLES: {{MAX_REVIEW_CYCLES}}
MAX_CI_CYCLES: {{MAX_CI_CYCLES}}
</input>

<defaults>
- If `REPO_PATH` is empty, use the current working directory.
- If `START_ISSUE_BASE` is empty, use `BASE_BRANCH`.
- If `AGENT` is empty, use the project/default `start-issue` agent.
- Resolve `ZELLIJ_SURFACE` from `ZELMA_START_ISSUE_ZELLIJ_SURFACE`, then `.zelma/config.json` key `start_issue.zellij_surface`, then `pane`.
- `ZELLIJ_SURFACE` may be only `pane` or `tab`; do not choose `tab` unless env or repo-local config explicitly requested it.
- If `MAX_REVIEW_CYCLES` is empty, use 5.
- If `MAX_CI_CYCLES` is empty, use 3.
- `AUTO_MERGE` must be explicit: `yes` or `no`.
</defaults>

<model_policy>
- Run implementation, fix requests, CI debugging and general supervision on `GPT-5.5 medium` unless the caller explicitly overrides the agent model.
- Run every fresh `/review` on `GPT-5.5 Extra high`.
- After each `/review` completes, return the implementation/fix loop to `GPT-5.5 medium`.
- If the UI or CLI cannot switch the `/review` model, stop before treating the review gate as satisfied and report the blocker with the exact model that was available.
</model_policy>

<definition_of_done>
Success is allowed only when all conditions are true:
1. `start-issue` implementation finished.
2. Fresh `/review` against `BASE_BRANCH` finished on `GPT-5.5 Extra high` with no critical/high/important findings.
3. All review findings that matter for correctness, maintainability, tests, security or scope were fixed, committed and pushed.
4. After every fix commit/push, another fresh `/review` ran against the updated head commit and returned no critical/high/important findings.
5. The accepted clean review result corresponds to the same `headRefOid` that will be merged.
6. PR exists for the issue branch against `BASE_BRANCH`.
7. PR is open, non-draft, mergeable and clean.
8. GitHub checks are present and green.
9. If `AUTO_MERGE=yes`, PR was merged and the merge commit was verified.
10. The task zellij surface was closed only after terminal outcome.
</definition_of_done>

<instructions>
1. Preflight:
   - Change to `REPO_PATH` if provided.
   - Read repo instructions such as `AGENTS.md` if present.
   - Read `memory-bank/ops/runbooks/visible-zellij-shipping-dispatcher.md` when it exists and follow it as the operational runbook for visible zellij shipping.
   - Do not change the branch in the main repository worktree. The main worktree must stay on `BASE_BRANCH`; implementation must happen only in worktrees created by `start-issue` or explicit `git worktree add`.
   - Do not use invisible/native subagents as a fallback for shipping when the caller expects zellij-visible tab/pane output. If zellij cannot create or inspect the required tab/pane, stop with a blocker unless the caller explicitly authorizes a non-zellij fallback.
   - Resolve the zellij launch surface:
     - if `ZELMA_START_ISSUE_ZELLIJ_SURFACE` is set, use it;
     - otherwise, if `.zelma/config.json` exists and contains `start_issue.zellij_surface`, use it;
     - otherwise, use `pane`;
     - stop with a configuration error if the resolved value is not `pane` or `tab`.
   - Verify GitHub auth and issue visibility:
     - `gh auth status`
     - `gh issue view {{ISSUE_NUMBER}} --repo {{OWNER_REPO}}`
   - Verify zellij is available:
     - `zellij action list-panes --json --all`
   - If zellij action commands hang or target a stale session context, stop with a blocker and report the session mismatch; do not continue in an invisible fallback.
   - Inspect current git status and avoid overwriting unrelated user changes.
   - Classify the issue delivery mode before starting work:
     - `implementation`: acceptance requires runtime behavior, CLI commands, tests, code, files, migrations, integrations or observable product behavior.
     - `feature_pack_only`: acceptance explicitly asks only for planning/docs/brief/design review with no runtime/code behavior.
   - If the issue is `implementation`, use `memory-bank/prompts/PROMPT-003-implement-and-test.md` as the effective implementation instruction when it exists, unless `PROMPT_FILE` explicitly overrides it.
   - Before passing any markdown prompt file through `start-issue --prompt-file`, verify the selected `AGENT` can accept that exact file format and invocation. For Codex, do not pass repository markdown prompt files with frontmatter through `--prompt-file` unless a dry run proves the launch works; launch without `--prompt-file` and send/read the prompt instructions inside the task pane instead.
   - Do not start PR review/improve cycles for an `implementation` issue until the task pane has produced implementation-scope changes, not only feature-pack/docs updates.

2. Запусти start-issue в выбранной zellij surface:
   - Default to a new zellij pane named `issue-{{ISSUE_NUMBER}}`.
   - Use a new zellij tab only when the resolved `ZELLIJ_SURFACE=tab` came from `ZELMA_START_ISSUE_ZELLIJ_SURFACE` or `.zelma/config.json`.
   - Do not create a separate zellij session for the issue.
   - Run `start-issue {{ISSUE_NUMBER}} --repo {{OWNER_REPO}} --base <START_ISSUE_BASE or BASE_BRANCH>`.
   - Add `--agent {{AGENT}}` only if `AGENT` is provided.
   - Add `--prompt-file <effective prompt file>` only when `PROMPT_FILE` is provided or preflight selected a prompt file and the agent compatibility check passed.
   - If prompt-file compatibility is not confirmed, launch `start-issue` without `--prompt-file`, then immediately instruct the task agent to read and apply the effective prompt file before implementation.
   - Record the selected zellij surface type, task `pane_id`, `tab_id` if applicable, command, cwd and start time.

3. Observe pane:
   - Use:
     - `zellij action list-panes --json --all`
     - `zellij action dump-screen --pane-id <pane_id> --full`
   - Periodically snapshot the screen.
   - Do not interrupt while the implementation agent is actively working.
   - If the pane exits unexpectedly, capture exit status and stop with blocker.
   - If the task pane exits with a launch/usage error caused by prompt-file incompatibility before any implementation work starts, do not count it as an implementation failure. Relaunch once without `--prompt-file`, inject the effective prompt instruction into the new task pane, and continue observation.

4. Start review:
   - When implementation appears complete, verify the diff matches the issue delivery mode:
     - For `implementation`, the diff must include relevant code/tests/docs required by acceptance.
     - If the diff is docs-only while acceptance requires runtime behavior, do not start PR review; tell the implementation agent to run `PROMPT-003` for the issue and continue implementation.
     - For `feature_pack_only`, feature-doc changes may be sufficient if they satisfy the issue acceptance.
   - Before sending `/review`, switch the task pane review model to `GPT-5.5 Extra high`.
   - Send `/review` to the same pane only after the delivery-mode check passes and review model is confirmed.
   - After sending `/review`, take the first pane snapshot after about 3 seconds, not after the normal long polling interval.
   - While the `/review` quiz/menu is open, poll every 3 seconds and answer the preset/base/model prompts immediately so the review can start without waiting for a minute-long observer interval.
   - Select review against base branch / PR-style review.
   - Select `BASE_BRANCH` as the base.
   - Wait for review output.
   - Once the review has actually started, return to the normal observer polling interval.
   - After review output is captured, switch implementation/fix work back to `GPT-5.5 medium`.

5. Review/fix loop:
   - If review has no critical/high/important findings, continue to PR gate.
   - If review has findings, send the implementation agent:
     `Исправь все critical/high/important review findings в scope issue {{ISSUE_NUMBER}}. Запусти релевантные проверки, закоммить и запушь изменения. Не исправляй unrelated findings.`
   - Wait for completion.
   - Verify that the fixes were committed and pushed; record the new `headRefOid`.
   - Run `/review` again on `GPT-5.5 Extra high`.
   - The previous review result is invalid after any fix commit, amend, rebase, merge-base update, or force-push.
   - Repeat until a fresh `/review` on the latest `headRefOid` is clean or `MAX_REVIEW_CYCLES` is reached.
   - If max cycles is reached, stop with `max_cycles_reached` and include findings.

6. PR gate:
   - Find or create PR for the issue branch against `BASE_BRANCH`.
   - Verify with:
     `gh pr view <PR> --repo {{OWNER_REPO}} --json url,state,isDraft,mergeable,mergeStateStatus,reviewDecision,statusCheckRollup,headRefName,baseRefName,headRefOid`
   - Required state:
     - `state=OPEN`
     - `isDraft=false`
     - `baseRefName=BASE_BRANCH`
     - `mergeable=MERGEABLE`
     - `mergeStateStatus=CLEAN`
   - If merge conflicts or dirty merge state exist, ask the implementation agent to fix, commit and push, then repeat review and PR gates.
   - If `headRefOid` changed since the last clean `/review`, return to the Review/fix loop before treating the PR gate as satisfied.

7. CI gate:
   - Run:
     `gh pr checks <PR> --repo {{OWNER_REPO}} --watch --fail-fast`
   - If checks are absent, do not call this green CI. Stop with blocker unless the issue scope explicitly includes creating CI.
   - If checks fail or are cancelled:
     - Get failed logs via `gh run view <RUN_ID> --log-failed`.
     - Send logs and exact fix request to the implementation agent in the pane.
     - Wait for fix commit/push.
     - Repeat fresh `/review`, PR gate and CI gate.
   - If a CI fix changes `headRefOid`, the review gate is no longer valid until a fresh `/review` passes on that new head.
   - Repeat until green CI or `MAX_CI_CYCLES` is reached.

8. Merge:
   - If `AUTO_MERGE=no`, stop with `ready_to_merge` after clean review, green CI and mergeable PR.
   - If `AUTO_MERGE=yes`, merge only after clean review, green CI, non-draft PR and `MERGEABLE/CLEAN` state:
     `gh pr merge <PR> --repo {{OWNER_REPO}} --merge --delete-branch`
   - Verify:
     `gh pr view <PR> --repo {{OWNER_REPO}} --json state,mergedAt,mergeCommit,url`

9. Cleanup and notification:
   - Close the task zellij surface only after terminal outcome.
   - If `ZELLIJ_SURFACE=pane`, close the task pane:
     `zellij action close-pane --pane-id <pane_id>`
   - If `ZELLIJ_SURFACE=tab`, close the created task tab by `tab_id`; do not close the currently focused tab by accident.
   - Send desktop notification:
     `osascript -e 'display notification "Issue {{ISSUE_NUMBER}} terminal outcome reached" with title "start-issue shipper"'`
</instructions>

<constraints>
- Do not announce success without clean review.
- Do not satisfy the review gate with a `/review` that did not run on `GPT-5.5 Extra high`.
- Do not satisfy the review gate with a clean `/review` from an older `headRefOid`.
- Do not merge immediately after fixes are committed and pushed; first run another fresh `/review` on the fixed head and require it to be clean.
- Do not announce green CI when checks are absent.
- Do not merge if PR is draft, not mergeable, dirty, conflicted, or checks are failed/pending/cancelled/absent.
- Do not bypass branch protection, required approvals, security gates or human gates.
- Do not fix unrelated findings.
- Do not treat feature-pack review/improve output as completion for an issue whose acceptance requires implementation behavior.
- Do not close the task zellij surface before terminal outcome.
- Do not launch the task in a new tab unless the resolved `ZELLIJ_SURFACE=tab` came from `ZELMA_START_ISSUE_ZELLIJ_SURFACE` or `.zelma/config.json`.
- Do not launch the task in a separate zellij session.
- Do not overwrite unrelated local changes.
- If a prompt override is used, report the selected prompt source.
</constraints>

<output_format>
Return a concise final report:
- Terminal status: `merged`, `ready_to_merge`, `blocked`, or `max_cycles_reached`.
- Issue number.
- PR URL.
- Merge commit SHA, if merged.
- Last head commit SHA.
- Review cycles count.
- CI cycles count.
- CI status.
- Mergeability status.
- Zellij surface type, pane id, tab id if applicable and cleanup status.
- Any blockers or human gates, with facts.
</output_format>
```

## Variables

| Variable | Required | Description | Example |
| --- | --- | --- | --- |
| `OWNER_REPO` | yes | GitHub repository in `owner/name` format. | `example/project` |
| `ISSUE_NUMBER` | yes | GitHub issue number. | `123` |
| `BASE_BRANCH` | yes | PR base branch. | `main` |
| `REPO_PATH` | no | Local repository path. | `/Users/me/code/project` |
| `START_ISSUE_BASE` | no | Explicit base ref/SHA for `start-issue`. | `origin/main` |
| `AGENT` | no | Agent backend for `start-issue`. | `codex` |
| `ZELLIJ_SURFACE` | no | Resolved current-session zellij surface for the task agent. Source order: `ZELMA_START_ISSUE_ZELLIJ_SURFACE`, `.zelma/config.json`, default `pane`. | `pane` |
| `AUTO_MERGE` | yes | Whether shipper may merge after gates pass. | `yes` |
| `PROMPT_FILE` | no | Optional start-issue prompt override file. | `.zelma/prompts/ship-issue.md` |
| `MAX_REVIEW_CYCLES` | no | Review/fix loop limit. | `5` |
| `MAX_CI_CYCLES` | no | CI/fix loop limit. | `3` |

## Validation Notes

| Check | Expected Result | Status |
| --- | --- | --- |
| Dry run with generic repo variables | Prompt contains no hardcoded repository and requires explicit repo/base/merge policy. | passed |
| Scope gate for absent CI | Prompt stops with blocker instead of calling absent checks green. | passed |
| Implementation issue routing | Prompt selects `PROMPT-003` before PR review when issue acceptance requires runtime/code behavior. | drafted |
| Review model policy | Prompt requires `GPT-5.5 Extra high` for `/review` and medium for implementation/fixes. | drafted |
| Review quiz polling | Prompt polls `/review` quiz/menu after 3 seconds and answers prompts quickly before returning to normal polling. | drafted |
| Codex prompt-file compatibility | Prompt avoids passing markdown/frontmatter prompt files to Codex through `start-issue --prompt-file` unless compatibility is verified. | drafted |
| Zellij surface resolution | Prompt resolves surface from env, then `.zelma/config.json`, then default pane; tab requires explicit env/config choice and separate per-issue zellij sessions are forbidden. | drafted |

## Change Notes

- 2026-07-08: Added env and `.zelma/config.json` resolution for zellij launch surface with env > config > default precedence.
- 2026-07-08: Made zellij launch surface explicit: default pane, optional tab only by caller request, no separate per-issue zellij session.
- 2026-07-07: Added prompt-file compatibility guard and recovery for Codex launch usage failures.
- 2026-07-07: Added fast 3-second polling for `/review` preset/base/model quiz prompts.
- 2026-07-07: Added model policy: implementation/fixes on `GPT-5.5 medium`, `/review` gates on `GPT-5.5 Extra high`.
- 2026-07-07: Added delivery-mode preflight so implementation issues run through `PROMPT-003` before PR review/fix cycles.
- 2026-07-07: Created reusable generic shipper prompt from live FT-001 delivery workflow; repository-specific values were converted to variables.
