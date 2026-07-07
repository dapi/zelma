---
title: "EP-001: Risks"
doc_kind: epic
doc_function: risk_register
purpose: "Epic-level risk register for Go CLI Foundation."
derived_from:
  - charter.md
status: active
audience: humans_and_agents
---

# EP-001: Risks

## Risk Register

| Risk ID | Risk | Impact | Mitigation | Trigger |
| --- | --- | --- | --- | --- |
| `RISK-01` | Go toolchain unavailable in local/CI environment | Scaffold cannot be verified | Document setup; fail early with clear precondition | `go version` fails |
| `RISK-02` | Default Cobra help leaks into product | Help becomes human-generic, not agent-first | Snapshot/contract tests for `zelma help` and command help | Help output changes without tests |
| `RISK-03` | Scope creeps into registry/zellij integration | First feature becomes too large and side-effectful | Keep FT-001 non-side-effecting; move integration to later epics | Code touches `.zelma/` runtime writes or invokes `zellij` |
| `RISK-04` | Early package layout fights future adapter boundaries | Refactor cost before useful behavior | Mirror ADR-001 module boundaries from scaffold | New package bypasses `internal/app` or adapters |
