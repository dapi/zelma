---
title: "UC-011: Отправка сообщения в существующую Codex session"
doc_kind: use_case
doc_function: canonical
purpose: "Фиксирует устойчивый сценарий безопасной доставки follow-up message в managed Codex session через публичный `zelma` CLI."
derived_from:
  - ../product/context.md
  - ../domain/model.md
  - ../domain/rules.md
  - ../domain/states.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - architecture_decision
  - feature_level_test_matrix
---

# UC-011: Отправка сообщения в существующую Codex session

## Goal

Human или supervising agent доставляет follow-up message в уже известную
managed Codex session без прямого обращения к zellij и без риска напечатать
prompt в stale, non-Codex или wrong pane.

## Primary Actor

Human или supervising agent.

## Trigger

Нужно передать дополнительную инструкцию, ревью-команду или follow-up prompt в
уже созданную/обнаруженную `zelma session`.

## Preconditions

- Команда запускается внутри целевого repository worktree.
- `.zelma/sessions.json` содержит repo-local numeric `ZelmaSessionID`.
- Target session должна быть `active`.
- Zellij runtime доступен для live revalidation.
- Target pane должна быть terminal pane с совместимым Codex runtime evidence.

## Main Flow

1. Actor получает target id через публичный inventory flow, например
   `zelma sessions list --live --json`.
2. Actor вызывает `zelma sessions send <id> [message] --json` или передает
   multiline prompt через `zelma sessions send <id> --stdin --json`.
3. `zelma` валидирует target id и message source.
4. `zelma` revalidates registry record against live zellij/Codex evidence.
5. Если readiness проходит, `zelma` доставляет message в recorded pane and
   performs the controlled submit action.
6. `zelma` возвращает JSON с target/session identity и message metadata без
   echo message body.

## Alternate Flows / Exceptions

- `ALT-01` Long or multiline prompt: actor uses `--stdin`; multiline content is
  allowed.
- `EX-01` Missing/conflicting/empty message source: `zelma` returns a stable
  diagnostic and does not contact the delivery adapter.
- `EX-02` Target record is not active or missing: `zelma` refuses before any
  zellij write.
- `EX-03` Zellij session/pane is missing or unreachable: `zelma` returns a
  not-ready diagnostic with public `zelma` recovery hint.
- `EX-04` Pane exists but Codex has exited or identity evidence is incompatible:
  `zelma` refuses and does not type into the shell.
- `EX-05` Runtime evidence is ambiguous: `zelma` refuses and does not auto-repair
  or guess before sending.

## Postconditions

- On success, the intended live Codex session receives one submitted message.
- On failure, no message text is sent to zellij unless readiness had passed.
- CLI output/diagnostics do not leak the message body.
- Actor receives a machine-readable result or diagnostic suitable for safe
  recovery.

## Business Rules

- `BR-01` Send target is repo-local numeric `ZelmaSessionID`.
- `BR-02` Registry state alone is insufficient; live Codex readiness is required
  before delivery.
- `BR-03` Direct zellij commands and direct `.zelma/sessions.json` parsing are
  not valid skill paths for this scenario.
- `BR-04` Message body is private prompt content and must not be echoed in
  success or failure output.
- `BR-05` Ambiguous runtime identity must fail closed.

## Traceability

| Upstream / Downstream | References |
| --- | --- |
| PRD | `none` |
| Features | `FT-101` |
| ADR | `none` |
