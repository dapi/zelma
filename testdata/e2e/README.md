# E2E Test Data

This directory is mounted read-only by `make test-e2e`.

The Docker runner creates its synthetic repository and Codex session metadata
at runtime so it can keep each run isolated.

`fake-codex.go` builds into the mounted `/test/codex` executable used by the
runner. It writes one synthetic `session_meta` JSONL record and then stays alive
so zellij can report the pane as a live Codex command.
