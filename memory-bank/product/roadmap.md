---
title: Product Roadmap
doc_kind: product
doc_function: canonical
purpose: Каноничное место для product themes, bets, horizons и dependencies без превращения roadmap в feature backlog.
derived_from:
  - ../dna/governance.md
  - context.md
  - vision.md
  - metrics.md
  - ../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
canonical_for:
  - product_roadmap
  - product_themes
---

# Product Roadmap

Этот документ описывает направление и sequencing продуктовых тем. Он не должен становиться списком всех feature packages: delivery-единицы живут в [`../features/README.md`](../features/README.md), а инициативы — в [`../prd/README.md`](../prd/README.md).

Практический порядок запуска GitHub issues и параллельных волн реализации
зафиксирован в [`execution-order.md`](execution-order.md).

## Horizons

| Horizon | Theme | Intended outcome | Candidate PRD / Feature | Dependency | Status |
| --- | --- | --- | --- | --- | --- |
| `now` | Project bootstrap and docs | Репозиторий содержит memory-bank, product/domain model и первичный roadmap | `unknown` | none | active |
| `now` | Go CLI skeleton and command routing | Есть запускаемый `zelma` binary с группой `sessions` и predictable errors | `unknown` | Go toolchain; Cobra | draft |
| `now` | `zelma setup` repo initialization | CLI готовит репозиторий к работе zelma и добавляет `.zelma` в `.gitignore` | `FT-031` | Go CLI skeleton; repo root resolver | draft |
| `now` | Session registry foundation | `.zelma/sessions.json` имеет versioned schema, atomic writes, validation и duplicate prevention | `unknown` | Domain model/rules | draft |
| `now` | `zelma sessions list` | Пользователь видит известные `zelma sessions` текущего repo | `unknown` | Registry foundation | draft |
| `now` | `zelma sessions create` | CLI создает `zellij pane`, запускает Codex и регистрирует active session | `unknown` | Registry foundation; zellij adapter; Codex launch contract | draft |
| `now` | `zelma sessions detect` | CLI обнаруживает вручную созданные Codex panes и идемпотентно регистрирует их | `unknown` | Registry foundation; zellij introspection; Codex identification | draft |
| `next` | Reconciliation and stale handling | Registry отражает закрытые, moved или недоступные panes без разрушительных действий | `unknown` | create/list/detect MVP | idea |
| `next` | Codex skill pack | Skills вызывают CLI для create/list/detect и используют тот же domain contract | `unknown` | Stable CLI output/schema | idea |
| `next` | Test harness for zellij/Codex integration | Поддерживаемые workflows проверяются fixtures/integration tests, а не только manual QA | `unknown` | MVP commands | idea |
| `next` | Autonomous issue shipping supervisor | Пользователь задает issue, а supervisor доводит `start-issue` delivery до reviewed, mergeable PR with green CI and merge | `EP-008` | zellij control; GitHub CLI; stable prompt policy | draft |
| `later` | Session focus/attach helpers | Пользователь может быстро перейти к нужному pane/session из CLI | `unknown` | Reliable registry + zellij control | idea |
| `later` | Multi-repo and worktree ergonomics | `zelma` лучше работает с несколькими worktrees и repo roots | `unknown` | Real usage feedback | idea |
| `later` | Rich status surfaces | Машиночитаемый status, summaries или zellij status integration | `unknown` | Skill pack and reconciliation | idea |

## Candidate Delivery Slices

Эти slices пока не являются заведенными epics/features. При декомпозиции перенеси
их в [`../epics/README.md`](../epics/README.md), [`../prd/README.md`](../prd/README.md)
или [`../features/README.md`](../features/README.md).

| Slice ID | Candidate slice | Why it comes here | Notes |
| --- | --- | --- | --- |
| `SLICE-01` | Go CLI project scaffold | Без запускаемого binary нельзя проверять UX команд | Go stack выбран; CLI framework: Cobra |
| `SLICE-02` | Registry schema and persistence | Все команды зависят от `.zelma/sessions.json` | Нужны versioning и atomic writes |
| `SLICE-03` | Zellij adapter over CLI automation | `create` и `detect` требуют надежной работы с sessions/panes | Primary path: Go `os/exec` + `zellij --session ... action list-panes --json --all` |
| `SLICE-04` | Codex session identification | `zelma session` должна ссылаться на Codex session, не только на pane | Самая рискованная часть detect |
| `SLICE-05` | Sessions list UX | Самая простая команда для проверки registry и output contracts | Должен быть human + machine readable mode |
| `SLICE-06` | Managed create workflow | Первый end-to-end value path | Должен сохранять opened path |
| `SLICE-07` | Manual detect workflow | Критичный workflow для уже запущенных panes | Требует conservative matching |
| `SLICE-08` | Codex skills | Делает `zelma` полезной для agentic workflows | Skills должны быть thin wrappers |
| `SLICE-09` | `zelma setup` gitignore initialization | Repo-local state не должен случайно попадать в git | Команда должна идемпотентно добавлять `.zelma` в `.gitignore` |
| `SLICE-10` | Autonomous issue shipping supervisor | Убирает ручное наблюдение за `start-issue`, `/review`, PR, CI и merge | Prompt должен быть редактируемым и overridable через `.zelma` |

## Proposed Epics And Features

Эти epics/features являются roadmap proposal, а не открытыми GitHub issues.
После подтверждения scope создай instantiated epic package в
[`../epics/README.md`](../epics/README.md), затем delivery feature packages в
[`../features/README.md`](../features/README.md).

| Epic ID | Epic | Outcome | Candidate features | Dependencies |
| --- | --- | --- | --- | --- |
| `EP-001` | Go CLI Foundation | Есть installable/testable `zelma` binary с agent-first help и command tree | `FT-001` Go module scaffold; `FT-002` Cobra command tree including `zelma setup`; `FT-003` agent-first help templates; `FT-004` output/error contract tests | Go toolchain |
| `EP-002` | Registry And Repo State | `.zelma/sessions.json` безопасно хранит known sessions текущего repo, а repo setup исключает `.zelma` из git | `FT-005` repo root resolver; `FT-006` schema v1; `FT-007` atomic writes + lock; `FT-008` registry validation/recovery; `FT-009` `sessions list --json/table`; `FT-031` `zelma setup` gitignore initialization | `EP-001` |
| `EP-003` | Zellij Read Integration And Detect | `sessions detect` находит manual Codex panes без unsafe takeover | `FT-010` zellij adapter `ListSessions`; `FT-011` zellij adapter `ListPanes`; `FT-012` fixture tests for zellij JSON; `FT-013` Codex pane candidate classifier; `FT-014` detect upsert/idempotency | `EP-001`, `EP-002` |
| `EP-004` | Managed Create Workflow | `sessions create` создает Codex pane и регистрирует active session | `FT-015` Codex launch contract; `FT-016` zellij run/new-pane adapter; `FT-017` create confirmation/reconciliation; `FT-018` create failure recovery hints | `EP-002`, `EP-003` |
| `EP-005` | Codex Session Identity | Active records получают надежный `CodexSessionRef` | `FT-019` identify available Codex metadata sources; `FT-020` session log/process evidence parser; `FT-021` candidate vs active state rules; `FT-022` privacy-safe evidence fixtures; `FT-043` command argv session evidence; `FT-044` detect evidence explain and indexed lookup | `EP-003`, `EP-004` |
| `EP-006` | Agent Skill Pack | Codex skills управляют `zelma` через стабильный CLI contract | `FT-023` skill command wrappers; `FT-024` machine-readable output compatibility tests; `FT-025` skill docs; `FT-026` agent recovery flows | `EP-001`-`EP-004` |
| `EP-007` | Reconciliation And Lifecycle | Registry отражает stale/closed sessions без разрушительных действий | `FT-027` `sessions list --live`; `FT-028` stale detection; `FT-029` cleanup/remove proposal; `FT-030` lifecycle state tests | `EP-002`, `EP-003` |
| `EP-008` | Autonomous Issue Shipping Supervisor | Supervisor запускает `start-issue` для issue и доводит delivery до clean review, green CI, mergeable PR, merge и notification | `FT-032` supervisor launch; `FT-033` editable prompt and `.zelma` override; `FT-034` pane observation; `FT-035` review/fix loop; `FT-036` PR/CI gate; `FT-037` merge cleanup notification | `EP-001`, zellij control, GitHub CLI |

## First Epic Recommendation

Начать с `EP-001 Go CLI Foundation`.

Первый feature slice: `FT-001` + `FT-002` + минимальная часть `FT-003`, чтобы
получить binary, `zelma help`, `zelma sessions help` и пустые команды
`list/create/detect`. Это дает проверяемый skeleton без side effects в zellij и
без записи registry.

## Roadmap Rules

- Roadmap theme описывает product intent, а не implementation plan.
- Если тема требует нескольких delivery slices, создай PRD и перечисли downstream features там.
- Если тема меняет предметную модель, сначала обнови [`../domain/model.md`](../domain/model.md), [`../domain/rules.md`](../domain/rules.md) или [`../domain/context-map.md`](../domain/context-map.md).
- Не заводи feature package только ради roadmap line item. Feature появляется,
  когда есть owner, scope, acceptance criteria и test strategy.

## Open Bets

- `BET-01` Надежная идентификация Codex session из `zellij pane` возможна без
  brittle parsing конкретной версии Codex.
- `BET-02` Repo-local `.zelma/sessions.json` достаточно для MVP, без глобального
  daemon или background watcher.
- `OQ-01` Какой packaging/release flow выбрать для Go CLI.
- `OQ-02` Какой machine-readable output contract нужен для skills: JSON flag,
  stable stdout schema или отдельный command namespace.
- `OQ-03` Нужен ли lock file для concurrent writes в `.zelma/`.
