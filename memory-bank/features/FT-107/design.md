---
title: "FT-107: Design"
doc_kind: feature
doc_function: canonical
purpose: "Solution-space owner for evidence-led documentation status reconciliation in FT-107."
derived_from:
  - brief.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_107_scope
  - ft_107_acceptance_criteria
  - ft_107_evidence_contract
  - implementation_sequence
---

# FT-107: Design

## Design Pack

| Artifact | Role | Owns |
| --- | --- | --- |
| `design.md` | Feature-local solution owner | `SOL-*`, `SD-*`, reconciliation rules |
| `decision-log.md` | FPF decision record | rationale and evidence for closed questions |

## Context

The task is a documentation reconciliation, not a new product or runtime
design. It still needs an explicit method because roadmap, epic and feature
documents have different ownership and stale statements must not become a new
source of product requirements.

## C4 Applicability

| C4 ID | Decision | Trigger / reason | Artifact |
| --- | --- | --- | --- |
| `C4-00` | `not required` | No runtime, deployable, interface or data boundary changes. | `none` |

## Selected Solution

- `SOL-01` Treat tags, changelog, merged PR history and closed GitHub issues as delivery evidence; do not infer delivery from an old document label.
- `SOL-02` Keep product intent/status in roadmap, execution sequencing in `execution-order.md`, initiative boundary in EP-008, and slice lifecycle in each feature brief.
- `SOL-03` Describe the current supervisor exactly as local zellij launch/poll/review/fix/re-review/cleanup simulation; route real GitHub PR/CI/merge gates to open issue #111.

## Alternatives Considered

| Alternative ID | Option | Why not selected |
| --- | --- | --- |
| `ALT-01` | Rewrite the roadmap as a full feature backlog | Conflicts with roadmap ownership and issue non-goal. |
| `ALT-02` | Preserve draft statuses until a new runtime change | Contradicts closed issues, merged PRs and releases. |
| `ALT-03` | State that real GitHub gates are implemented because the supervisor simulates merge | Conflicts with current CLI help and issue #111. |

## Accepted Local Decisions

- `SD-01` The established legacy value `delivery_status: implemented` is retained for already delivered packages; it is not redefined as a new lifecycle taxonomy in this feature.
- `SD-02` A feature lifecycle status is corrected only when a closed issue or merged PR directly maps to that feature package.

## Invariants

- `INV-01` No corrected document claims runtime behaviour absent from code/help.
- `INV-02` No roadmap line duplicates detailed feature acceptance or implementation steps.
- `INV-03` Open work retains an explicit dependency and next decision.

## Failure Modes

- `FM-01` A historical candidate mapping is mistaken for the current feature ID; mitigate by using the package's own title and direct delivery evidence.
- `FM-02` A simulated merge is documented as a GitHub merge; mitigate by retaining the CLI-help limitation and issue #111 route.

## Traceability

| Requirement ID | Solution refs | Contracts / invariants | Failure refs |
| --- | --- | --- | --- |
| `REQ-01`, `REQ-02` | `SOL-01`, `SOL-02`, `SD-01` | `INV-02`, `INV-03` | `FM-01` |
| `REQ-03` | `SOL-03` | `INV-01` | `FM-02` |
| `REQ-04` | `SOL-01`, `SD-02` | `INV-01` | `FM-01` |
