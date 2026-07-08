---
title: "FT-046: Pane PID Codex Session Evidence"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для fallback-определения CodexSessionRef через PID Codex process, привязанный к zellij pane."
derived_from:
  - ../../epics/EP-005/brief.md
  - ../../engineering/codex-runtime-identification.md
  - ../../engineering/zellij-integration.md
  - ../FT-019/brief.md
  - ../FT-043/brief.md
  - ../FT-044/brief.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - selected_solution
---

# FT-046: Pane PID Codex Session Evidence

## Что

### Проблема

`sessions detect` может оставлять Codex pane в `candidate`, когда command
metadata из `zellij action list-panes --json --all` и `session_meta` lookup не
дают однозначный `CodexSessionRef`. Это особенно заметно, когда несколько
Codex processes работают в одном repo path: cwd совпадает, но session id нельзя
выбрать без более сильной связи между конкретной pane и конкретным process.

Текущий verified fact: `list-panes --json --all` показывает command/cwd, но не
pane PID. Если безопасная PID-корреляция доступна, она может стать fallback
evidence source для точного Codex session id. Если такой связи нет или она
неоднозначна, `zelma` должна оставлять запись candidate, а не угадывать.

### Результат

`sessions detect` использует PID-correlated process evidence только при
необходимости: после того как более простые sources не разрешили
`CodexSessionRef`. При успешной корреляции ровно одного live Codex process с
pane feature извлекает только безопасный session ref и переводит запись в
`active`; при ambiguity или unsupported platform оставляет объяснимый
candidate.

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Ambiguous same-repo Codex panes resolved by strong process evidence | Such panes remain `candidate` without manual action | Pane becomes `active` only when exactly one correlated Codex PID yields safe UUID evidence | Detection fixture/e2e with two Codex sessions in one repo |
| `MET-02` | Privacy leakage from process evidence | Raw argv/env may contain private prompt text | Registry and CLI output contain no raw argv, env, prompt text or transcript content | Unit tests and JSON snapshot assertions |

### Объем Работ

- `REQ-01` Добавить feature-level contract для optional correlation между
  zellij pane и ровно одним live Codex process PID.
- `REQ-02` Использовать PID-correlated evidence только как fallback после
  existing command evidence и `session_meta` index lookup.
- `REQ-03` Извлекать из process evidence только безопасный `CodexSessionRef`
  или ссылку на единственный безопасно подтвержденный Codex session file.
- `REQ-04` Возвращать `candidate_ambiguous` или `candidate_unresolved`, если
  pane/process correlation дает ноль, несколько или stale PID candidates.
- `REQ-05` Показывать PID fallback decision в `sessions detect --explain`
  без raw argv/env и без сохранения PID в registry.
- `REQ-06` Покрыть unsupported-platform behavior: отсутствие PID surface не
  считается ошибкой detect и не ломает существующие evidence paths.

### Что Не Входит

- `NS-01` Нет превращения PID самого по себе в `CodexSessionRef`.
- `NS-02` Нет хранения PID, raw argv, process environment или process tree в
  `.zelma/sessions.json`.
- `NS-03` Нет чтения Codex transcript content, user prompts, assistant answers
  или tool payloads.
- `NS-04` Нет ptrace/debugger/memory inspection и других invasive process
  techniques.
- `NS-05` Нет замены existing argv/session-file evidence order; PID path
  остается fallback, а не default.
- `NS-06` Нет широкого сканирования unrelated user processes вне candidate
  zellij panes.

### Constraints / Assumptions

- `ASM-01` Zellij может знать pane child PID internally, но текущий
  `list-panes --json --all` не обязан отдавать его как stable API.
- `ASM-02` На Unix-like systems безопасная correlation может быть platform-
  specific; unsupported platforms must degrade to existing candidate behavior.
- `CON-01` PID reuse делает PID ephemeral. Его нельзя сохранять как durable
  identity и нужно валидировать только в пределах текущего detect run.
- `CON-02` Process argv/env can contain private prompt text. Parser должен
  извлекать только UUID/source metadata и redacted explanation.
- `DEC-01` Selected design должен выбрать concrete PID surface: direct zellij
  API, OS process tree correlation, open-file correlation или другой
  проверяемый adapter boundary.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: yes` | PID correlation меняет identity evidence, privacy boundary, platform behavior и failure modes. | `design.md` |

## Проверка

### Критерии Готовности

- `EC-01` Candidate с неоднозначным `session_meta` по cwd становится `active`,
  если ровно один live Codex PID надежно связан с pane и дает безопасный UUID.
- `EC-02` Если PID surface отсутствует, недоступен или неоднозначен,
  `sessions detect` оставляет candidate с explain reason.
- `EC-03` Registry после successful detect содержит `codex_session`, но не
  содержит PID, raw argv/env или process-tree details.
- `EC-04` Existing `sessions detect` behavior without `--explain` remains
  backward-compatible.

### Traceability Matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `DEC-01` | `EC-01`, `EC-02` | `CHK-01`, `CHK-02` | `EVID-01` |
| `REQ-02` | `CON-01`, `CON-02` | `EC-04` | `CHK-03` | `EVID-02` |
| `REQ-03` | `CON-02` | `EC-01`, `EC-03` | `CHK-01`, `CHK-04` | `EVID-01`, `EVID-03` |
| `REQ-04` | `ASM-02`, `CON-01` | `EC-02` | `CHK-02` | `EVID-02` |
| `REQ-05` | `CON-02` | `EC-02`, `EC-03` | `CHK-04` | `EVID-03` |
| `REQ-06` | `ASM-02` | `EC-02`, `EC-04` | `CHK-03` | `EVID-02` |

### Acceptance Scenarios

- `SC-01` Two Codex panes share one repo cwd; session file lookup is ambiguous,
  but one pane has a uniquely correlated Codex PID with safe UUID evidence, so
  `sessions detect` registers it as `active`.
- `SC-02` PID correlation finds multiple plausible Codex child processes for
  one pane; detect returns candidate with ambiguity explanation.
- `SC-03` Platform or zellij version does not expose a usable PID surface;
  detect behaves as before and explains that PID fallback was skipped.
- `SC-04` Process evidence includes private prompt-like text near the UUID;
  parser extracts only the allowed UUID/source and redacts everything else.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Detection unit/integration fixture with same-cwd sessions and one correlated Codex PID | Candidate becomes `active` with expected `codex_session` | `artifacts/ft-046/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-02` | Detection fixture with zero/multiple PID candidates | Candidate remains unresolved/ambiguous; no registry active write | `artifacts/ft-046/verify/chk-02/` |
| `CHK-03` | `EC-02`, `EC-04`, `SC-03` | CLI test with PID adapter unavailable | Existing detect output remains compatible; `--explain` reports fallback skipped | `artifacts/ft-046/verify/chk-03/` |
| `CHK-04` | `EC-03`, `SC-04` | Snapshot test for registry and JSON/text explain output | No PID, raw argv/env, prompt text or transcript content appears | `artifacts/ft-046/verify/chk-04/` |

### Test Matrix

| Check ID | Evidence IDs | Evidence path |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-046/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-046/verify/chk-02/` |
| `CHK-03` | `EVID-02` | `artifacts/ft-046/verify/chk-03/` |
| `CHK-04` | `EVID-03` | `artifacts/ft-046/verify/chk-04/` |

### Evidence

- `EVID-01` Fixture/test log showing unique PID-correlated Codex session
  resolution.
- `EVID-02` Fixture/test log showing ambiguity and unsupported-platform
  fallback without active write.
- `EVID-03` Registry and CLI output snapshots proving privacy exclusions.

### Evidence Contract

| Evidence ID | Artifact | Producer | Path contract | Reused by checks |
| --- | --- | --- | --- | --- |
| `EVID-01` | Detection fixture output or test log | test runner | `artifacts/ft-046/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Ambiguous/unsupported fallback output | test runner | `artifacts/ft-046/verify/chk-02/` and `artifacts/ft-046/verify/chk-03/` | `CHK-02`, `CHK-03` |
| `EVID-03` | Registry and CLI snapshot output | test runner | `artifacts/ft-046/verify/chk-04/` | `CHK-04` |
