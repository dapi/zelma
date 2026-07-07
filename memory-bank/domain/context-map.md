---
title: Domain Context Map
doc_kind: domain
doc_function: canonical
purpose: Каноничное место для bounded contexts, upstream/downstream relations, language ownership и business integration boundaries.
derived_from:
  - ../dna/governance.md
  - glossary.md
  - model.md
status: active
audience: humans_and_agents
canonical_for:
  - bounded_contexts
  - domain_context_map
---

# Domain Context Map

Этот документ фиксирует business bounded contexts. Он не описывает runtime deployment, package layout или service topology, если они не совпадают с domain boundary.

## Bounded Contexts

| Context | Owns language / rules for | Upstream contexts | Downstream contexts | Must not know |
| --- | --- | --- | --- | --- |
| `Session Registry` | `ZelmaSession`, `.zelma/sessions.json`, schema, uniqueness, lifecycle state | Product Context | CLI Experience, Skill Integration, Reconciliation | Internal details of Codex transcripts or zellij UI layout |
| `Zellij Integration` | Observing/creating `zellij sessions` and `zellij panes` | External `zellij` runtime | Session Registry, CLI Experience, Reconciliation | Codex conversation semantics |
| `Codex Runtime Identification` | Evidence that a pane is running Codex and mapping to `CodexSessionRef` | External Codex runtime/logs/processes | Detection, Session Registry, Skill Integration | zellij layout semantics beyond pane/process evidence |
| `CLI Experience` | User-facing commands, arguments, output and errors | Session Registry, Zellij Integration, Codex Runtime Identification | Skill Integration, humans | Private persistence details beyond stable output/schema |
| `Skill Integration` | Codex skill wrappers and agent-facing workflows | CLI Experience | Codex agents | Direct undocumented mutation of registry |
| `Reconciliation` | Comparing registry records with live runtime state | Session Registry, Zellij Integration, Codex Runtime Identification | CLI Experience, Skill Integration | Product positioning or marketing claims |

## Context Relationships

| Relationship ID | Upstream | Downstream | Contract | Notes |
| --- | --- | --- | --- | --- |
| `REL-01` | Session Registry | CLI Experience | Read/write domain API or internal module contract | CLI must not hand-edit JSON outside registry rules |
| `REL-02` | Zellij Integration | Session Registry | Observed runtime facts: session refs, pane refs, cwd/process evidence | Facts must be revalidatable |
| `REL-03` | Codex Runtime Identification | Detection | Codex evidence and `CodexSessionRef` resolution | Conservative failure is preferred over false positive |
| `REL-04` | CLI Experience | Skill Integration | Stable commands and machine-readable output | Skills remain thin wrappers |
| `REL-05` | Reconciliation | CLI Experience | State verdicts: active/stale/candidate/closed | List should present verdicts without surprise mutation unless policy says otherwise |

## Shared Kernel / Published Language

- `SK-01` Shared value objects: `ZellijSessionRef`, `ZellijPaneRef`,
  `CodexSessionRef`, `OpenedPath`, `SessionOrigin`.
- `SK-02` Shared states: `candidate`, `active`, `stale`, `closed`, `archived`.
- `PL-01` Published CLI language uses `zelma session`, `zellij session`,
  `zellij pane`, `codex session`, `opened path`.
- `PL-02` Machine-readable output must preserve stable field names for required
  `ZelmaSession` properties once defined.

## Boundary Rules

- Context владеет своими domain facts и public contracts.
- Другой context не должен читать или менять internal state в обход published boundary.
- Если technical module boundary отличается от domain boundary, объясни это в [`../engineering/architecture.md`](../engineering/architecture.md).
- `zellij` and Codex are external systems; `zelma` records references to them but
  does not own their lifecycle except when a supported command explicitly
  creates or closes a pane.
- Skills must cross the boundary through CLI/public API, not by constructing
  private registry JSON independently.

## Open Boundary Questions

- `OQ-01` Should Codex session identification live entirely in CLI code, or as a
  separate adapter with independent tests and fixtures?
- `OQ-02` Should reconciliation happen inside `sessions list`, or only in an
  explicit future `sessions sync` command?
- `OQ-03` Should `.zelma/sessions.json` be manually editable with validation, or
  treated as CLI-owned state with recovery commands?
