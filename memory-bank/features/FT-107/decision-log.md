---
title: "FT-107: Decision Log"
doc_kind: feature-support
doc_function: reference
purpose: "FPF-backed decisions made during FT-107 review-improve cycles."
derived_from:
  - brief.md
  - design.md
  - implementation-plan.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_107_scope
  - ft_107_selected_design
  - ft_107_acceptance_criteria
  - implementation_sequence
---

# FT-107: Decision Log

## DL-001: Evidence-Led Status Reconciliation

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-14 |
| Review cycle | 1 |
| Closed question | What can establish delivery status without guessing? |
| FPF frame | B.5 canonical reasoning cycle: propose → analyze → test |

### Available Facts

- Tags `v0.1.0` through `v0.4.0` exist; `CHANGELOG.md` records the delivered
  CLI, release and supervisor outcomes.
- Closed GitHub issues map directly to the historical feature packages; merged
  PRs #93, #96, #98, #99, #105 and #106 provide repository evidence.
- `zelma supervisor start-issue --help` says it simulates the lifecycle and
  does not merge GitHub PRs; open issue #111 owns real PR/CI gates.

### Decision

Use those evidence carriers to correct only the product, epic and feature
statements they directly support. Preserve remaining work as future work with
its dependency and next decision.

### Rationale

FPF abduction proposed that stale documentation, rather than missing runtime
work, caused the conflict. Deduction predicts that direct release/tracker/help
evidence must agree with every corrected statement. The collected tags, merged
PRs, issue states and help text corroborate that prediction, so no human gate
is needed.

### Human Gate

None.

## Verification Evidence

| Evidence ID | Result |
| --- | --- |
| `EVID-01` | Tags v0.1.0–v0.4.0, `CHANGELOG.md`, and closed-issue inventory were reviewed. |
| `EVID-02` | `go run ./cmd/zelma supervisor start-issue --help` confirms lifecycle simulation and states that it does not merge GitHub PRs. |
| `EVID-03` | Merged PR history confirms #93, #96, #98, #99, #105 and #106 delivery evidence. |
| `EVID-04` | `python3 scripts/check_memory_bank_index.py` and `git diff --check` passed after the final edits. |

## DL-002: One Open Issue Per Remaining Roadmap Line

| Field | Value |
| --- | --- |
| Status | accepted |
| Date | 2026-07-14 |
| Review cycle | 2 |
| Closed question | How should roadmap satisfy a current status, dependency and next decision for each remaining delivery? |
| FPF frame | B.5 deduction from the acceptance criterion |

### Available Facts

- Issue #107 requires every remaining roadmap line to have a current status,
  dependency and next decision.
- Open issues #108–#113 have distinct titles and problem scopes.
- The first reconciliation grouped #110, #112 and #113 into one line, leaving
  more than one possible next decision.

### Decision

Give each open issue its own roadmap line and add an explicit `Next decision`
column. Keep the completed baseline aggregated because it is historical context,
not a remaining delivery unit.

### Rationale

The deduction is direct: a line with several independent issue scopes cannot
have one unambiguous next decision. Splitting only the remaining lines removes
that ambiguity without turning the roadmap into a duplicate feature backlog.

### Human Gate

None.
