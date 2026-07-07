---
title: Engineering Architecture Patterns
doc_kind: engineering
doc_function: canonical
purpose: Каноничное место для архитектурных правил реализации: code/module boundaries, runtime patterns, concurrency, error handling и configuration ownership.
derived_from:
  - ../dna/governance.md
  - ../domain/context-map.md
  - ../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
---

# Engineering Architecture Patterns

Этот документ задает ожидаемые архитектурные правила реализации. Предметные bounded contexts описаны в [`../domain/context-map.md`](../domain/context-map.md); здесь фиксируй, как они отражаются в code modules, services, queues, adapters и configuration ownership.

## Implementation Stack

- Primary language: Go.
- Primary product surface: CLI binary `zelma`.
- Primary zellij integration: external `zellij` CLI invoked through Go
  `os/exec`, with JSON parsing where zellij provides JSON output.
- CLI framework: `github.com/spf13/cobra`.
- First registry persistence choices: Go `encoding/json`,
  `github.com/google/renameio/v2` for atomic file replacement and
  `github.com/gofrs/flock` for cross-process registry locking.
- Zellij plugin/WASM integration is not MVP. Revisit via ADR after CLI
  create/detect/list works.

Detailed zellij integration research and links live in
[`zellij-integration.md`](zellij-integration.md).
Codex runtime identification and `CodexSessionRef` extraction rules live in
[`codex-runtime-identification.md`](codex-runtime-identification.md).

## Module Boundaries

| Module / Layer | Owns | Must not depend on directly |
| --- | --- | --- |
| `cli` | Command routing, arguments, user-facing output and exit codes | Private JSON write details except through registry API |
| `registry` | `.zelma/sessions.json` schema, validation, atomic reads/writes, migrations | `zellij` command execution or Codex process probing |
| `zellij-adapter` | Creating panes and reading `zellij` session/pane facts | Registry persistence details |
| `codex-adapter` | Identifying Codex runtime/session refs from pane/process/log evidence | Registry persistence details or `zellij` UI assumptions |
| `detection` | Combining zellij facts and Codex evidence into candidates/session verdicts | Direct JSON mutation outside `registry` |
| `reconciliation` | Comparing registry records with live runtime state and producing state verdicts | Product positioning or skill-specific behavior |
| `skills` | Codex skill wrappers around stable CLI commands/output | Private modules or undocumented registry schema |

Минимальные правила:

- модуль владеет своим state и публичными контрактами;
- межмодульные зависимости проходят через явно названный API, event или adapter;
- UI, jobs и интеграции не должны читать чужие внутренние детали в обход owner-модуля.
- `skills` must call CLI/public API rather than hand-writing `.zelma/sessions.json`.
- `registry` must not shell out to `zellij` or inspect processes; it validates and persists facts passed by adapters.
- `zellij-adapter` must expose typed methods such as `ListSessions`,
  `ListPanes`, `RunCodexPane` and keep raw command execution details private.
- `codex-adapter` must expose typed evidence and verdict inputs rather than raw
  Codex transcript contents or unredacted process argv.

## CLI Help And Output Contract

`zelma` is agent-first, human-second. This applies especially to `zelma`,
`zelma help`, command help and error hints.

Requirements:

- Customize Cobra help/usage templates; do not ship the default generic Cobra
  help if it hides the agent workflow.
- Top-level `zelma` and `zelma help` must start with an agent-oriented command
  map: what to run for list/create/detect, when each command is safe, and which
  flags produce machine-readable output.
- Help examples must be copy-ready and non-interactive by default.
- Put human explanatory prose after the agent quickstart, not before it.
- Do not rely on color, ANSI styling or terminal width for meaning.
- Keep stdout/stderr discipline: successful machine-readable output goes to
  stdout; diagnostics, warnings and recovery hints go to stderr unless the
  command is explicitly help text.
- Any command intended for skills must provide a stable `--json` output mode
  before a skill depends on it.
- Error messages should include the next safe command when possible, for example
  `zelma sessions detect --json` or `zelma sessions list --json`.

## Concurrency And Critical Sections

- Writes to `.zelma/sessions.json` must be atomic: validate next state, write to
  a temporary file, then replace the registry file.
- Concurrent writes need an explicit lock strategy before multiple mutating
  commands are supported in parallel. Until then, mutating commands should fail
  clearly or serialize access if a lock exists.
- `sessions detect` must be idempotent. Re-processing the same live pane should
  update an existing record or no-op, never append a duplicate active record.
- External side effects happen before registry commit only when rollback is
  either unnecessary or represented by a clear stale/candidate state. For
  example, if pane creation succeeds but Codex identity cannot be resolved, the
  command must not write an invalid active session.

## Failure Handling And Error Tracking

Зафиксируй единый подход:

- CLI errors should include command, repo root, registry path when relevant, and
  the external command that failed without dumping Codex conversation contents.
- Missing `zellij` is a dependency error, not an empty session list.
- A pane that cannot be proven to contain Codex is a detection verdict, not an
  internal exception.
- Corrupt `sessions.json` is a registry integrity error; mutating commands must
  stop before writing until recovery/migration behavior is defined.
- Machine-readable output must remain parseable on success. Diagnostics belong
  on stderr or in structured error fields once the CLI contract is defined.

## Configuration Ownership

Документируй не все переменные окружения подряд, а ownership-модель конфигурации:

- Registry location defaults to `.zelma/sessions.json` under repo root.
- Repo root detection must be centralized; commands should not each invent their
  own root discovery behavior.
- Supported `zellij` and Codex versions belong in [`../ops/config.md`](../ops/config.md)
  or a future compatibility document once implementation begins.
- Any environment variable or config file that changes registry location,
  command behavior or external binary path must be documented in
  [`../ops/config.md`](../ops/config.md).
