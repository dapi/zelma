---
title: "FT-015: Codex Launch Contract Design"
doc_kind: feature
doc_function: canonical
purpose: "Selected design for resolving and validating the Codex launch contract used by managed create."
derived_from:
  - brief.md
  - ../../engineering/architecture.md
  - ../../engineering/zellij-integration.md
  - ../../ops/config.md
status: draft
audience: humans_and_agents
---

# FT-015: Codex Launch Contract Design

## Selected Design

The Codex launch contract lives in `internal/codex`. It is a typed contract,
not a shell string:

- resolved Codex executable;
- argv: `--cd <opened_path>`;
- working directory: `<opened_path>`;
- opened path: normalized absolute path equal to or inside the current repo
  root.

`zelma sessions create --dry-run` resolves this contract and prints it without
creating a `zellij` pane or writing `.zelma/sessions.json`. Plain
`zelma sessions create` performs the same preflight and then stops with a clear
pending-zellij diagnostic until the FT-016/FT-017 create path lands.

## Path Policy

If no path is passed, the opened path is the detected repository root. This
avoids depending on the caller's shell cwd when the command is launched from a
nested directory.

If a path is passed, it must be an existing directory equal to or inside the
detected repository root. Relative paths are resolved by normal CLI filesystem
rules before validation, then normalized through symlink resolution.

## Binary Policy

`ZELMA_CODEX_BIN` may point to the Codex executable. If it is not set, `codex`
is resolved through `PATH`.

The resolved executable path is used in the contract so a later zellij launch
does not depend on shell aliases, functions or a different interactive-shell
startup path.

## Failure Modes

| Failure | Result |
| --- | --- |
| Opened path is missing, not a directory or outside repo root | `codex_invalid_input`; no registry write |
| Codex binary is missing or not executable | `codex_missing_binary`; no registry write |
| zellij pane creation is requested after successful preflight | Explicit pending diagnostic; no registry write |

Actual zellij pane creation, launch failure normalization, Codex session
identity parsing and registry promotion remain outside FT-015.

## Verification

- `CHK-01`: `internal/codex` unit tests assert the command vector and working
  directory for the resolved opened path.
- `CHK-02`: CLI tests use a missing fake Codex binary and assert a recoverable
  diagnostic without creating `.zelma/sessions.json`.
