---
title: "FT-002: Decision Log"
doc_kind: feature
doc_function: canonical
purpose: "Локальный журнал решений FT-002. Фиксирует FPF-обоснования, review-improve решения и human gates без переопределения brief, design или implementation plan."
derived_from:
  - brief.md
  - design.md
  - implementation-plan.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_002_scope
  - ft_002_acceptance_criteria
  - ft_002_execution_sequence
---

# FT-002: Decision Log

## DL-001: Promote Feature Package To Problem/Solution/Plan Ready

- Status: accepted
- Date: 2026-07-07
- Review cycle: 1
- FPF frame: bounded contexts + evidence graph.
- Question: Can FT-002 move from draft docs to an active feature package without
  a human gate?
- Available facts:
  - Issue 2 names `FT-002: Дерево команд Cobra` and lists `zelma setup`,
    `zelma sessions list/create/detect`, no registry writes, no `.gitignore`
    changes and no `zellij` calls.
  - `brief.md` has matching `REQ-*`, `NS-*`, `SC-*`, `CHK-*` and
    `Design required: yes`.
  - ADR-001 is `decision_status: accepted` and already selects Go + Cobra.
  - FT-001 scaffold exists in `cmd/zelma/main.go`, `internal/cli/cli.go` and
    `go.mod`.
- Decision: Promote `README.md` and `brief.md` to `status: active`, create
  `design.md`, and create `implementation-plan.md`.
- Rationale: The problem boundary, accepted architecture and current scaffold
  are all documented. The missing work is document completeness and
  traceability, not an unresolved product decision.
- Human gate: none.

## DL-002: Stub Behavior Is Deterministic Non-Implemented Diagnostics

- Status: accepted
- Date: 2026-07-07
- Review cycle: 1
- FPF frame: options + assurance.
- Question: What predictable stub behavior can FT-002 define without taking
  ownership from later features?
- Available facts:
  - Issue 2 requires stubs to return predictable diagnostics without side
    effects.
  - EP-001 charter allows `sessions list/create/detect` stubs with predictable
    errors or placeholder behavior.
  - `brief.md` excludes `.zelma/sessions.json`, `.gitignore` and live `zellij`
    behavior.
  - FT-003 owns agent-first help templates; FT-004 owns broader output/error
    contract tests.
- Options:
  - Success placeholder output: easy for demos, but can mislead automation into
    believing real behavior ran.
  - Deterministic non-implemented diagnostics: communicates route existence
    while preserving no-side-effects and later output ownership.
- Decision: Use deterministic non-implemented diagnostics for command stubs.
- Rationale: This closes FT-002 acceptance without inventing registry/zellij
  behavior or finalizing help/output templates.
- Human gate: none.

## DL-003: C4 Artifact Not Required

- Status: accepted
- Date: 2026-07-07
- Review cycle: 1
- FPF frame: boundary classification.
- Question: Does FT-002 require a C4 diagram or separate architecture artifact?
- Available facts:
  - ADR-001 already owns Go CLI architecture and command/application/adapter
    separation.
  - FT-002 only reserves command routes inside the existing Go CLI scaffold.
  - The feature excludes registry persistence, live `zellij`, schemas,
    queues, security boundary changes and new deployables.
- Decision: Record `C4-00: not required` in `design.md`.
- Rationale: The change is inside one existing CLI container and does not cross
  runtime or integration boundaries that would require C1-C4 modeling.
- Human gate: none.

## DL-004: Verify Side-Effect Boundary Through Tests And Static Review

- Status: accepted
- Date: 2026-07-07
- Review cycle: 1
- FPF frame: assurance path.
- Question: How should FT-002 verify no registry or `zellij` side effects before
  those adapters exist?
- Available facts:
  - Current code has only `cmd/zelma/main.go`, `internal/cli/cli.go` and
    `go.mod`.
  - FT-002 excludes `.zelma/sessions.json` and live `zellij`.
  - FT-001 used `rg -n "zellij|sessions.json|\\.zelma" cmd internal` as a
    side-effect boundary check.
- Decision: Use Go tests for routing/stub behavior and static search/code
  review for forbidden side-effect paths.
- Rationale: With no registry or zellij adapter in scope, static review plus
  deterministic tests is sufficient and avoids inventing fake integration
  infrastructure.
- Human gate: none.
