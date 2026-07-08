---
title: Packages Фич Index
doc_kind: feature
doc_function: index
purpose: Навигация по instantiated package фичиs. Читать, чтобы найти существующую delivery-единицу или понять, где создавать новую.
derived_from:
  - ../dna/governance.md
  - ../flows/feature-flow.md
status: active
audience: humans_and_agents
---

# Packages Фич Index

Каталог `memory-bank/features/` хранит instantiated package фичиs вида `FT-XXX/`.

## Правила

- Каждый package создается по правилам из [`../flows/feature-flow.md`](../flows/feature-flow.md).
- Bootstrap package начинается с `README.md` и `brief.md`; после `Problem Ready` в него добавляется `design.md`, если `brief.md` фиксирует `Design required: yes`; `implementation-plan.md` появляется после готовности нужных upstream owners.
- Для bootstrap и downstream-документов используй шаблоны из [`../flows/templates/feature/`](../flows/templates/feature/).
- Если работа требует roadmap, risk register и нескольких delivery subissues, сначала создай или обнови epic package в [`../epics/README.md`](../epics/README.md).
- По умолчанию feature ссылается на общий product context из [`../product/context.md`](../product/context.md), а при изменении предметных правил также на соответствующие документы из [`../domain/README.md`](../domain/README.md).
- Если feature реализует или существенно меняет устойчивый сценарий проекта, она должна ссылаться на соответствующий `UC-*` из [`../use-cases/README.md`](../use-cases/README.md).
- В шаблонном репозитории этот каталог может быть пустым. Это нормально.

## Именование

- Базовый формат: `FT-XXX/`
- Вместо `XXX` используй идентификатор, принятый в проекте: issue id, ticket id или другой стабильный ключ
- Один package = одна delivery-единица

## Созданные Packages Фич

- [FT-001: Go Module Scaffold](FT-001/README.md) — первый delivery slice
  `EP-001`: Go module scaffold and empty `zelma` binary without registry or
  zellij side effects.
- [FT-002: Cobra Command Tree](FT-002/README.md) — `zelma sessions` command
  group and routed command stubs.
- [FT-003: Agent-First Help Templates](FT-003/README.md) — help output optimized
  for agents first and humans second.
- [FT-004: Output And Error Contract Tests](FT-004/README.md) — contract tests
  for predictable CLI output and diagnostics.
- [FT-005: Repo Root Resolver](FT-005/README.md) — consistent repository root
  detection for registry commands.
- [FT-006: Sessions Schema V1](FT-006/README.md) — versioned
  `.zelma/sessions.json` schema.
- [FT-007: Atomic Registry Writes And Lock](FT-007/README.md) — safe writes and
  concurrency guard for registry persistence.
- [FT-008: Registry Validation And Recovery](FT-008/README.md) — validation,
  diagnostics and recovery behavior for invalid registry state.
- [FT-009: Sessions List Output](FT-009/README.md) — human and JSON output for
  known sessions.
- [FT-010: Zellij Adapter ListSessions](FT-010/README.md) — read zellij sessions
  through the Go adapter.
- [FT-011: Zellij Adapter ListPanes](FT-011/README.md) — read zellij panes and
  metadata through the Go adapter.
- [FT-012: Zellij JSON Fixture Tests](FT-012/README.md) — fixture coverage for
  supported zellij JSON shapes.
- [FT-013: Codex Pane Candidate Classifier](FT-013/README.md) — conservative
  classification of panes that may be running Codex.
- [FT-014: Detect Upsert Idempotency](FT-014/README.md) — idempotent registry
  updates for detected sessions.
- [FT-015: Codex Launch Contract](FT-015/README.md) — command/process contract
  for launching Codex in a managed pane.
- [FT-016: Zellij Run New-Pane Adapter](FT-016/README.md) — create panes via
  zellij CLI adapter.
- [FT-017: Create Confirmation And Reconciliation](FT-017/README.md) — confirm
  created panes before active registry writes.
- [FT-018: Create Failure Recovery Hints](FT-018/README.md) — agent-friendly
  diagnostics for partial create failures.
- [FT-019: Codex Metadata Source Discovery](FT-019/README.md) — identify usable
  Codex metadata sources.
- [FT-020: Session Evidence Parser](FT-020/README.md) — parse privacy-safe
  evidence for Codex session references.
- [FT-021: Candidate Vs Active State Rules](FT-021/README.md) — state rules for
  uncertain versus confirmed sessions.
- [FT-022: Privacy-Safe Evidence Fixtures](FT-022/README.md) — fixture corpus
  without private conversation content.
- [FT-023: Skill Command Wrappers](FT-023/README.md) — Codex skill wrappers over
  the `zelma` CLI.
- [FT-024: Machine-Readable Output Compatibility Tests](FT-024/README.md) —
  compatibility tests for agent-facing JSON output.
- [FT-025: Skill Docs](FT-025/README.md) — skill usage documentation.
- [FT-026: Agent Recovery Flows](FT-026/README.md) — skill-level recovery flows
  for common failures.
- [FT-027: Sessions List Live](FT-027/README.md) — live reconciliation view for
  listed sessions.
- [FT-028: Stale Detection](FT-028/README.md) — stale session detection rules.
- [FT-029: Cleanup Remove Proposal](FT-029/README.md) — explicit cleanup/remove
  proposal flow.
- [FT-030: Lifecycle State Tests](FT-030/README.md) — lifecycle state test suite.
- [FT-031: Zelma Setup Gitignore](FT-031/README.md) — команда `zelma setup`
  идемпотентно добавляет `.zelma` в `.gitignore`.
- [FT-032: Supervisor Command And Zellij Launch](FT-032/README.md) — запуск
  `start-issue` в zellij pane по умолчанию или tab по явному env/config override.
- [FT-033: Agent Session Inventory E2E](FT-033/README.md) — e2e-покрытие
  сценария инвентаризации active/stale zelma-сессий через `sessions list --live --json`.
- [FT-034: Manual Pane Adoption E2E](FT-034/README.md) — e2e-покрытие
  обнаружения вручную созданных Codex pane через `sessions detect --json`.
- [FT-035: Managed Agent Launch E2E](FT-035/README.md) — e2e-покрытие
  `sessions create --json` от запуска pane до записи registry.
- [FT-036: Issue Supervisor Orchestration E2E](FT-036/README.md) — e2e-покрытие
  `start-issue` supervisor flow для GitHub issue.
- [FT-037: Agent Recovery E2E](FT-037/README.md) — e2e-покрытие
  recovery diagnostics для типовых ошибок registry/zellij/create.
- [FT-038: Stale Cleanup E2E](FT-038/README.md) — e2e-покрытие
  proposal/confirm cleanup flow для stale registry entries.
- [FT-039: Agent Handoff E2E](FT-039/README.md) — e2e-покрытие
  восстановления картины active work новым агентом.
- [FT-040: Multi-Agent Parallel Delivery E2E](FT-040/README.md) — e2e-покрытие
  запуска и supervision нескольких independent issue agents.
- [FT-041: Environment Smoke Diagnostics E2E](FT-041/README.md) — e2e-покрытие
  `setup`, `.gitignore`, `list` и `detect` для fresh repo.
- [FT-042: Agent Dashboard Status Backend](FT-042/README.md) — milestone-2
  status backend для dashboard/agent UI поверх session registry и zellij state.
