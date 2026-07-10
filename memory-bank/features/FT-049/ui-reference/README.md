---
title: "FT-049: TUI Reference"
doc_kind: feature-support
doc_function: reference
purpose: "Support reference для TUI states, controls and low-fidelity mockups. Не владеет requirements, selected design или execution plan."
derived_from:
  - ../brief.md
  - ../design.md
status: active
audience: humans_and_agents
must_not_define:
  - ft_049_scope
  - ft_049_selected_design
  - ft_049_acceptance_criteria
  - implementation_sequence
---

# FT-049: TUI Reference

Этот support doc фиксирует UI surface для review. Canonical requirements и
acceptance остаются в `../brief.md`; selected solution и contracts остаются в
`../design.md`.

## Screen Map

| UI ID | State | Canonical refs | Purpose |
| --- | --- | --- | --- |
| `UI-01` | Live sessions present | `REQ-03`, `REQ-05`, `SOL-03` | Show live/active work first and keep selection on that group by default |
| `UI-02` | Mixed active and non-active records | `REQ-03`, `REQ-04`, `SOL-03` | Keep stale/non-active records secondary |
| `UI-03` | Empty live state | `REQ-04`, `REQ-09`, `SC-03` | State that no live sessions are running and expose recovery/non-active context |
| `UI-04` | Degraded status | `REQ-09`, `FM-01` | Surface status backend recovery hints |
| `UI-05` | Focus result / failure | `REQ-08`, `FM-04` | Show transient outcome while keeping monitor running |

## Controls

| UI ID | Control | Behavior | Canonical refs |
| --- | --- | --- | --- |
| `UI-10` | Up/down navigation | Move selection over visible rows without reordering groups | `REQ-07`, `SD-04` |
| `UI-11` | Manual refresh key | Request one bounded provider refresh | `REQ-06`, `CTR-04` |
| `UI-12` | Focus key | Focus selected live/active session id or show guarded failure | `REQ-08`, `SD-05` |
| `UI-13` | Toggle/filter for non-active records | Show/hide secondary records without changing default live-first grouping | `REQ-04`, `TRD-03` |
| `UI-14` | Quit key | Exit the monitor without registry mutation | `CON-01`, `INV-03` |

## Low-Fidelity Mockups

### `UI-01`: Live Sessions Present

```text
zelma monitor                         live 2  stale 1  degraded no

LIVE
> 2  active  live  /repo/api        zelma-main tab_6 terminal_75  codex:abc
  5  active  live  /repo/web        zelma-main tab_7 terminal_81  codex:def

OTHER
  3  stale   unreachable  /repo/old  zelma-main tab_2 terminal_44  -

status: refreshed 12:30:14
```

### `UI-03`: Empty Live State

```text
zelma monitor                         live 0  stale 2  degraded no

LIVE
  No live zelma sessions.

OTHER
> 3  stale  unreachable  /repo/old   zelma-main tab_2 terminal_44  -
  4  closed unknown      /repo/docs  -          -     -            -

status: use refresh to re-check zellij reachability
```

### `UI-04`: Degraded Status

```text
zelma monitor                         live 0  unknown 3  degraded yes

LIVE
  Live state unavailable.

OTHER
> 2  active  unknown  /repo/api  zelma-main tab_6 terminal_75  codex:abc
  5  active  unknown  /repo/web  zelma-main tab_7 terminal_81  codex:def

recovery: status backend could not inspect live zellij state: <reason>
```

## UI Traceability

| UI ID | Brief refs | Design refs |
| --- | --- | --- |
| `UI-01` | `REQ-03`, `REQ-05`, `SC-01` | `SOL-02`, `SOL-03`, `CTR-03`, `INV-02` |
| `UI-02` | `REQ-03`, `REQ-04`, `SC-02` | `SOL-03`, `TRD-03`, `INV-02` |
| `UI-03` | `REQ-04`, `REQ-09`, `SC-03` | `SOL-02`, `SOL-03`, `INV-04` |
| `UI-04` | `REQ-09`, `SC-07` | `SOL-02`, `TRD-02`, `FM-01` |
| `UI-05` | `REQ-08`, `SC-06`, `NEG-02` | `SOL-05`, `CTR-05`, `INV-05`, `FM-04` |
| `UI-10`-`UI-14` | `REQ-06`, `REQ-07`, `REQ-08` | `SD-04`, `SD-05`, `CTR-04`, `CTR-05` |
