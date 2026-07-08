# Changelog

## Unreleased

### Next Roadmap Candidates

- Real-world end-to-end zellij smoke scenario for `setup`, `create`, `list --live`, `detect` and `cleanup`.
- Install and packaging polish after the first tagged release.
- Additional operator UX around focusing or attaching to known sessions.
- Follow-up hardening from real Codex/zellij usage.

## v0.1.0 - MVP baseline

### Added

- Go/Cobra CLI with agent-first help for `zelma`, `zelma setup` and `zelma sessions`.
- Repo-local `.zelma/sessions.json` registry with schema v1, validation, atomic writes and file locking.
- `zelma setup` to idempotently add `.zelma` to `.gitignore`.
- `zelma sessions list` with table and JSON output.
- Zellij adapters for listing sessions and panes through the `zellij` CLI.
- Conservative Codex pane detection and idempotent registry upsert.
- Managed `zelma sessions create` flow with Codex launch contract, zellij pane creation, confirmation and recovery diagnostics.
- Codex session evidence discovery and parsing from privacy-safe metadata sources.
- Candidate vs active state rules using resolved Codex session evidence.
- `zelma sessions list --live` for read-only live/unreachable status.
- Stale detection during `zelma sessions detect`.
- `zelma sessions cleanup` proposal flow, with destructive cleanup gated behind `--confirm`.
- Machine-readable compatibility tests for CLI JSON outputs.
- Thin skill wrapper package over the public `zelma` CLI.
- Agent recovery flows mapping CLI diagnostics to safe next actions.

### Release

- GitHub Actions release workflow builds versioned binaries for Linux, macOS and Windows on `v*` tags.
- GitHub Releases contain platform archives and `SHA256SUMS.txt`.
