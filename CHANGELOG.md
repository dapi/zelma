# Changelog

## Unreleased

### Next Roadmap Candidates

- Follow-up hardening from real Codex/zellij usage.
- Future autonomous issue shipping supervisor improvements.
- Further dashboard/status backend refinement for multi-agent work.

## v0.4.0 - Observation, monitor and safe messaging

### Added

- `zelma sessions send` for sending a message to a verified Codex session without echoing message bodies.
- `zelma sessions buffer <id> --json` for bounded read-only zellij pane screen observation.
- `zelma sessions transcript <id> --json` for bounded read-only Codex transcript event observation.
- `zelma monitor` live terminal monitor for observing active sessions and recovery hints.
- Skill client support for send, buffer and transcript commands.
- Product documentation for safe session messaging and live monitor workflows.

### Fixed

- Session transcript observation is more tolerant of real Codex transcript shapes.
- Monitor refresh avoids overlapping update cycles.
- `sessions list` shows candidate sessions in the default human output.
- Safe session send validation now handles dash-prefixed message arguments and target command argument separation correctly.

### Documentation

- Moved the distributable `zelma` Codex skill to the repository root for direct install from GitHub.

## v0.3.0 - Dashboard and supervisor expansion

### Added

- `zelma status` dashboard snapshot command for session state and recovery hints.
- `zelma supervisor start-issue` orchestration flow for issue-driven agent work.
- Supervisor launch, polling, review and cleanup state in machine-readable output.
- Multi-agent delivery e2e coverage for supervisor flows.
- Dashboard backend that reconciles registry state with live zellij status.
- New help and output contracts for the dashboard and supervisor surfaces.

### Fixed

- Session detection and recovery paths now tolerate more real-world zellij and Codex evidence shapes.
- E2E coverage now exercises the managed agent launch, recovery and handoff flows more directly.

### Documentation

- Synchronized help output and release docs with the implemented dashboard and supervisor commands.

## v0.2.0 - Session operations hardening

### Added

- Repo-local numeric `zelma session` IDs in registry and CLI output.
- `zelma sessions focus <id>` with table and JSON output.
- Zellij tab metadata capture for sessions when available.
- `zelma sessions detect --explain` evidence reporting for text and JSON output.
- One-pass Codex session evidence lookup for detect.
- Codex session ID extraction from safe `codex resume <uuid>` argv evidence.
- Optional PID-correlated process evidence path for pane/session resolution.
- Docker zellij e2e target covering setup, create, live list and manual detect.
- Environment smoke diagnostics e2e for fresh repositories.

### Fixed

- Detect now handles node-wrapped Codex pane commands.
- Detect skips missing zellij sessions instead of failing the full scan.

### Documentation

- Added agentic use cases for inventory, manual adoption, managed launch,
  recovery, cleanup, handoff, parallel delivery and environment diagnostics.
- Added the visible zellij shipping dispatcher runbook.
- Synchronized feature and epic statuses with the implemented CLI baseline.

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
