---
title: "FT-017: Create Confirmation And Reconciliation Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for confirming a created zellij pane before writing a create registry record."
derived_from:
  - brief.md
  - ../../domain/rules.md
  - ../../domain/states.md
  - ../../engineering/architecture.md
  - ../../engineering/codex-runtime-identification.md
status: active
audience: humans_and_agents
---

# FT-017: Create Confirmation And Reconciliation Design

## Selected Design

`zelma instances create` now performs the managed create workflow through three
ordered steps:

1. Resolve the Codex launch contract from FT-015.
2. Create a command pane through `zellij.RunPane` from FT-016.
3. Read panes from the same zellij session and confirm that the returned pane id
   exists, has Codex command evidence and reports the requested opened path.

The registry write happens only after step 3 succeeds. Confirmation produces an
unresolved `candidate` registry record with `zellij_session`, `zellij_pane` and
normalized `opened_path` because this feature does not implement CodexSessionRef
extraction. This preserves the active-state invariant from `domain/rules.md`:
active records require zellij session, zellij pane, Codex session and opened
path.

## Target Session

The create path targets `ZELMA_ZELLIJ_SESSION` when it is set, otherwise
`zelma-main`. Session bootstrap is outside this feature; if the target zellij
session does not exist, the existing zellij adapter failure path is returned and
the registry is not changed.

## Confirmation Rules

The created pane is confirmed when all conditions are true:

- `RunPane` returned a terminal typed pane id such as `terminal_7`;
- `ListPanes` for the same zellij session contains that typed pane id;
- pane is terminal and non-exited;
- observed command executable matches the configured Codex launch binary;
- observed pane cwd normalizes to the requested opened path.

If the pane cannot be confirmed after a successful create call, the command
returns `created=1 registered=0 skipped=1` and does not create
`.zelma/instances.json`. Recovery hints and reason-code detail are owned by
FT-018.

## Registry Reconciliation

Confirmed create evidence is written through the existing candidate upsert path.
This keeps create consistent with detect:

| Existing record for pane key | Create result |
| --- | --- |
| none | append one `candidate` record |
| candidate | keep one candidate record and fill missing evidence |
| active | preserve the active record and report registered |
| closed/stale/archived only | append a new `candidate` record |

The create path does not clean up the zellij pane if registry write fails.

## CLI Contract

Default output:

```text
created=1 registered=1 skipped=0
```

`--json` returns the same counters as JSON. `--dry-run` keeps the FT-015 launch
contract output and does not create panes or write the registry.

## Verification

- Unit tests in `internal/create` cover launch request construction, pane
  confirmation and unconfirmed-pane skip behavior.
- CLI tests cover confirmed create writing a candidate record and unconfirmed
  create leaving the registry absent.
