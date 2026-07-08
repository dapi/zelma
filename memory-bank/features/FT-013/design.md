---
title: "FT-013: Codex Pane Candidate Classifier Design"
doc_kind: feature-design
doc_function: canonical
purpose: "Selected design for conservative Codex pane candidate classification."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
  - ../../engineering/zellij-integration.md
  - ../../engineering/codex-runtime-identification.md
status: draft
audience: humans_and_agents
---

# FT-013: Codex Pane Candidate Classifier Design

## Selected Design

Candidate classification lives in `internal/detection`. The package exposes a
pure classifier over `zellij.Pane` facts and a caller-supplied absolute repo
root. It does not shell out, read Codex metadata, write `.zelma/sessions.json`
or produce a `CodexSessionRef`.

The first implementation returns two verdicts:

| Verdict | Meaning | Registry effect |
| --- | --- | --- |
| `candidate` | Pane has enough weak evidence to be shown as a Codex pane candidate. | None in FT-013. |
| `unknown` | Pane lacks required candidate evidence or metadata is unsafe/incomplete. | None. |

Each verdict includes stable reason codes for agent review.

## Candidate Policy

A pane is `candidate` only when all evidence below is present:

- pane kind is terminal, not plugin;
- pane is not exited;
- `pane_command` identifies `codex` as the executable;
- `pane_cwd` is an absolute path equal to or inside the current repo root.

All other combinations return `unknown`. Partial metadata, relative paths,
non-Codex commands, plugin panes and panes outside the repo root are not
promoted.

## Reason Codes

Positive evidence:

- `terminal_pane`
- `codex_command`
- `cwd_inside_repo`

Conservative rejection reasons:

- `non_terminal_pane`
- `pane_exited`
- `missing_command`
- `command_not_codex`
- `missing_cwd`
- `cwd_outside_repo`
- `invalid_repo_root`
- `invalid_cwd`

## Boundaries

This classifier intentionally stops before active session resolution. Future
detect/upsert work can store `candidate` records, but `active` remains forbidden
until Codex identity rules produce a valid `CodexSessionRef`.

## Verification

- `CHK-01`: fixture test parses `zellij list-panes` output and returns
  `candidate` with reason codes for an explicit Codex command.
- `CHK-02`: partial metadata fixture returns `unknown` with conservative reason
  codes.
- Additional unit tests cover plugin panes, exited panes, command false
  positives, paths outside the repo and invalid path metadata.
