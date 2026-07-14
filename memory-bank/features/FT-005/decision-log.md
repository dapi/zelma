---
title: "FT-005: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "Feature-local журнал решений для FT-005. Фиксирует закрытые вопросы review-improve без владения scope, selected design или execution sequencing."
derived_from:
  - brief.md
  - ../../flows/feature-flow.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_005_scope
  - ft_005_selected_design
  - ft_005_acceptance_criteria
  - implementation_sequence
---

# FT-005: Decision Log

Этот журнал фиксирует решения, принятые во время подготовки feature package. Он
не подменяет canonical owners: scope и verify живут в `brief.md`, solution facts
живут в `design.md`, execution sequencing живет в `implementation-plan.md`.

## DL-001: Supported Repo Boundary

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-07 |
| Closed question | Какой marker определяет "поддерживаемый repo" для FT-005? |
| FPF frame | Bounded context + evidence graph + propose/analyze/test reasoning |

### Available Facts

- GitHub issue 5 задает цель: одинаково определять repo root для команд,
  работающих с `.zelma/` и `.gitignore`.
- `brief.md` ограничивает scope правилами поиска root, нормализацией path и
  agent-friendly ошибкой вне repo; чтение/запись `.zelma/instances.json` не
  входит в scope.
- `../../adr/ADR-001-mvp-cli-architecture.md` назначает `internal/repo`
  centralized owner-ом repo root и `.zelma/` paths.
- `../../engineering/architecture.md` требует централизованного repo root
  detection и default registry location `.zelma/instances.json` under repo root.
- `../../ops/config.md` фиксирует default layout `<repo-root>/.zelma/instances.json`
  и запрещает недокументированные environment overrides.
- `../../features/FT-031/brief.md` использует тот же resolver для `.gitignore`
  setup, а `.zelma/` может еще не существовать до setup.

### Alternatives

| Alternative | Fit | Rejection / selection reason |
| --- | --- | --- |
| Ascend to Git worktree root | selected | Единственный marker из доступных фактов, который одновременно объясняет `.gitignore`, repo-local `.zelma/` и работу из вложенного каталога. |
| Ascend to existing `.zelma/` | rejected | Не подходит для `zelma setup`, потому что `.zelma/` может еще отсутствовать. |
| Ascend to `go.mod` | rejected | Описывает implementation repo, а не произвольный пользовательский repo, где `zelma` управляет sessions. |
| Use current working directory as root | rejected | Противоречит acceptance: из вложенного каталога должен находиться тот же repo root. |

### Decision

FT-005 treats a supported repo as a Git worktree. Resolver ascends from the
starting directory to the Git worktree root, returns a normalized absolute path,
and reports an agent-friendly unsupported-repo error when no Git worktree root is
found.

### Consequences

- `design.md` owns the selected solution and concrete contract.
- `implementation-plan.md` must ground implementation in a future `internal/repo`
  package and tests with temporary Git repositories.
- Non-Git project support, global registry behavior and environment overrides
  stay out of FT-005.
