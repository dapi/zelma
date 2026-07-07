---
title: "EP-001: Decision Log"
doc_kind: epic
doc_function: decision_log
purpose: "Local epic decisions for Go CLI Foundation that do not require a global ADR."
derived_from:
  - charter.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
status: active
audience: humans_and_agents
---

# EP-001: Decision Log

## Decisions

| Decision ID | Decision | Status | Rationale | Escalate to ADR when |
| --- | --- | --- | --- | --- |
| `EP-001-DEC-001` | First delivery feature is `FT-001 Go Module Scaffold` | accepted | It proves Go toolchain, repo layout and binary entrypoint before side effects | Scope expands beyond scaffold |
| `EP-001-DEC-002` | `FT-001` must not invoke live `zellij` | accepted | Keeps first slice deterministic and safe | A command needs runtime zellij facts |
