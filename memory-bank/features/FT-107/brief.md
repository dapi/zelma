---
title: "FT-107: Synchronize Roadmap And Delivery Statuses"
doc_kind: feature
doc_function: canonical
purpose: "Canonical brief для сверки roadmap, execution order, epic и feature delivery statuses после v0.4 без изменения runtime."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/roadmap.md
  - ../../product/execution-order.md
  - ../../epics/EP-008/brief.md
status: active
delivery_status: done
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-107: Synchronize Roadmap And Delivery Statuses

## What

### Problem

Canonical product documents still label delivery work released in v0.1–v0.4 as
future, and EP-008 states real GitHub PR/CI/merge behaviour that the current
supervisor command explicitly does not provide. This gives maintainers and
agents an incorrect planning baseline.

### Outcome

| Metric ID | Metric | Baseline | Target | Measurement method |
| --- | --- | --- | --- | --- |
| `MET-01` | Status conflicts for closed delivery work in scoped canonical docs | Present | None | Cross-check against tags, changelog, merged PRs and closed issues |

### Scope

- `REQ-01` Mark v0.1–v0.4 delivered product outcomes as implemented in the roadmap and retain only future work in the next horizon.
- `REQ-02` Replace completed execution waves with an implemented baseline and route future delivery to current open issues.
- `REQ-03` Make EP-008 distinguish implemented local supervisor lifecycle from the unimplemented real GitHub PR/CI/merge gates.
- `REQ-04` Align feature-package lifecycle metadata for closed, merged delivery work in this scope with the evidence.

### Non-Scope

- `NS-01` No runtime, CLI, GitHub workflow or release-process change.
- `NS-02` No duplicate feature backlog inside the roadmap.

### Constraints / Assumptions

- `ASM-01` Git tags v0.1.0–v0.4.0, `CHANGELOG.md`, merged PR history and closed GitHub issues are delivery evidence.
- `CON-01` Roadmap remains owner of product intent; feature briefs remain owner of delivery scope.
- `CON-02` Open issue #111 is the current owner for real GitHub PR/CI/merge gates.

## Design Requirement Decision

| Decision | Reason | Downstream owner |
| --- | --- | --- |
| `Design required: yes` | Reconciliation needs an explicit source-of-truth and status-mapping decision across product, epic and feature documents. | `design.md` |

## Verify

### Exit Criteria

- `EC-01` No scoped canonical document represents a released delivery outcome as draft, planned or future.
- `EC-02` EP-008 and roadmap separate local supervisor simulation from open GitHub gate work.
- `EC-03` Memory-bank index and whitespace checks pass.

### Traceability matrix

| Requirement ID | Problem refs | Acceptance refs | Checks | Evidence IDs |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-01` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-01`, `CON-01` | `EC-01` | `CHK-01` | `EVID-01` |
| `REQ-03` | `ASM-01`, `CON-02` | `EC-02` | `CHK-02` | `EVID-02` |
| `REQ-04` | `ASM-01` | `EC-01` | `CHK-03` | `EVID-03` |

### Acceptance Scenarios

- `SC-01` Maintainer reads the roadmap after v0.4 and sees delivered baseline separated from remaining work.
- `SC-02` Agent reads EP-008 and learns that real GitHub gates remain an open delivery, not an implemented supervisor capability.
- `SC-03` Reviewer follows each corrected feature status to merged/closed delivery evidence.

### Checks

| Check ID | Covers | How to check | Expected result | Evidence path |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `SC-01` | Review roadmap and execution order against releases and closed issues | Released baseline is explicit; remaining work has status, dependency and next decision | `git history + docs` |
| `CHK-02` | `EC-02`, `SC-02` | Compare EP-008 text with `zelma supervisor start-issue --help` and issue #111 | Local simulation and future GitHub gates are distinct | `CLI help + issue #111` |
| `CHK-03` | `EC-01`, `SC-03` | Review affected feature frontmatter against merged PR/closed issue evidence | No status contradicts delivery evidence | `git history + GitHub issues` |
| `CHK-04` | `EC-03` | Run `python3 scripts/check_memory_bank_index.py` and `git diff --check` | Both pass | command output |

### Evidence

- `EVID-01` Release tags, changelog and closed-issue inventory used for product-document reconciliation.
- `EVID-02` Current supervisor help and open issue #111 used for the EP-008 boundary.
- `EVID-03` Merged PRs #93, #96, #98, #99, #105 and #106 used for feature status reconciliation.
- `EVID-04` Memory-bank index and diff-check output.
