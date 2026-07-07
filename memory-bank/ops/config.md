---
title: Configuration Guide
doc_kind: ops
doc_function: canonical
purpose: Ownership-модель конфигурации zelma: registry location, external binaries, env contract и compatibility notes.
derived_from:
  - ../dna/governance.md
  - ../engineering/zellij-integration.md
status: active
audience: humans_and_agents
---

# Configuration Guide

Этот документ не обязан перечислять все переменные окружения подряд. Его задача:
объяснить, где живет canonical schema конфигурации `zelma` и какие runtime
contracts важны для CLI, registry и external binaries.

## Configuration Architecture

Текущая модель конфигурации:

### File Layout

```text
<repo-root>/
└── .zelma/
    └── sessions.json
```

### Ownership Rules

1. Registry schema belongs to the future Go `registry` package.
2. Registry default path is `.zelma/sessions.json` under detected repo root.
3. External binary paths default to `PATH` lookup.
4. Environment overrides are allowed only after they are documented here and
   covered by tests.
5. No secrets are required for MVP.

## Naming Convention For Env Vars

| Setting | Env variable |
| --- | --- |
| zellij binary path override | `ZELMA_ZELLIJ_BIN` |
| codex binary path override | `ZELMA_CODEX_BIN` |
| registry path override | `ZELMA_REGISTRY_PATH` |

Rules:

- env vars use prefix `ZELMA_`;
- no nested separator is defined yet;
- every override must have CLI-visible diagnostics;
- registry path override must remain repo/workspace explicit and must not create
  hidden global state accidentally.

## Documenting Important Variables

Если проекту нужен справочник ключевых переменных, не перечисляй все подряд. Сфокусируйся на значимых runtime contracts.

| Variable | Description | Default | Owner |
| --- | --- | --- | --- |
| `ZELMA_ZELLIJ_BIN` | Optional path/name for zellij executable | `zellij` via `PATH` | `zellij-adapter` |
| `ZELMA_CODEX_BIN` | Optional path/name for Codex executable | `codex` via `PATH` | `codex-adapter` |
| `ZELMA_REGISTRY_PATH` | Optional registry file path for tests/recovery | `.zelma/sessions.json` | `registry` |

## Secrets

- Никогда не вставляй реальные значения секретов в репозиторий.
- Документируй только способ их хранения, выдачи и rotation policy.
- Если часть конфигурации приходит из secret manager, это должно быть написано явно.
- MVP `zelma` does not require secrets.

## Compatibility

| Dependency | Current local probe | Minimum supported version | Notes |
| --- | --- | --- | --- |
| Go | pinned by `.mise.toml` as `1.25.11`; plain `go` may be absent from shell `PATH` | `1.25.11` for scaffold | Use `mise install` and `mise exec -- go ...` unless shell activates mise shims |
| zellij | `0.44.0` on `2026-07-07` | likely `0.44.0` | `list-panes --json --all` and returned pane IDs are core MVP assumptions |
| Codex CLI | not probed yet | `unknown` | Needed for create/detect Codex identity |

## Adoption Checklist

- [x] описан schema-owner конфигурации
- [x] задокументирована naming convention
- [x] перечислены ключевые runtime/env contracts
- [x] описан secret handling
- [ ] после scaffold уточнить zellij/Codex minimum versions
