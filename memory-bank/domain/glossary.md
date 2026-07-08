---
title: Domain Glossary
doc_kind: domain
doc_function: canonical
purpose: Каноничное место для ubiquitous language, domain terms, запрещенных двусмысленностей и naming decisions.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
canonical_for:
  - ubiquitous_language
  - domain_terms
---

# Domain Glossary

Этот документ фиксирует язык предметной области. Если термин здесь определен, downstream-документы используют это значение или явно объясняют исключение.

## Terms

| Term | Meaning | Context | Do not confuse with |
| --- | --- | --- | --- |
| `zelma` | CLI-утилита и набор Codex skills для управления Codex-сессиями в `zellij panes` | Product, CLI, skills | `zellij`, Codex CLI |
| `zelma session` | Управляемая запись о Codex-сессии, запущенной в конкретном `zellij pane`, с привязкой к `zellij session`, `zellij pane`, `codex session` и opened path | Registry, CLI output, domain rules | `zellij session`, `codex session` |
| `zelma session id` | Positive integer `id` записи `zelma session`, уникальный внутри repo-local `.zelma/sessions.json` и начинающийся с `1` | Registry, CLI output, future commands | `zellij session`, `codex session`, global database ID |
| `session registry` | Repo-local JSON-файл `.zelma/sessions.json`, который хранит known `zelma sessions` | Persistence, list/create/detect | Codex session log, zellij layout |
| `zellij session` | Runtime session терминального мультиплексора `zellij`, внутри которой существуют panes | Zellij adapter, session refs | `zelma session` |
| `zellij pane` | Pane внутри `zellij session`, в котором может быть запущен Codex | Create/detect/list | Terminal tab, Codex session |
| `codex session` | Идентификатор или ссылка на конкретную Codex-сессию, запущенную в pane | Session identity, detect | Codex process, Codex transcript file |
| `opened path` | Нормализованный абсолютный путь, который открыт в pane и относится к working context Codex-сессии | Registry, filtering, list output | Repo root, shell cwd after later `cd` |
| `repo root` | Корень проекта, относительно которого хранится `.zelma/` | Registry ownership | Current shell directory |
| `managed session` | `zelma session`, созданная через `zelma sessions create` | Origin tracking | Detected/manual session |
| `detected session` | `zelma session`, найденная через `zelma sessions detect` после ручного запуска Codex в `zellij pane` | Origin tracking | Candidate, stale record |
| `candidate session` | Потенциальная Codex-сессия, найденная detect, но еще не имеющая полного набора обязательных свойств | Detection lifecycle | Active `zelma session` |
| `stale session` | Registry record, который больше не подтверждается live `zellij`/Codex state | Reconciliation | Closed session |
| `session origin` | Способ попадания записи в registry: `create`, `detect` или future import/migration | Audit, debugging | Lifecycle state |

## Colloquial Aliases

| Alias | Canonical term | Allowed use | Notes |
| --- | --- | --- | --- |
| `зессия` | `zelma session` | Разговорный обиход, planning notes, informal handoff | Не использовать как CLI command, JSON field или canonical domain term |
| `зешка` | `zelma session` | Короткий сленг для быстрого общения внутри команды | Не использовать как CLI command, JSON field или canonical domain term |

## Naming Rules

- Используй domain terms последовательно в PRD, use cases, features, code comments и ADR.
- Не вводи новый синоним для существующего domain concept без обновления этого glossary.
- UI labels могут отличаться от domain terms, но разница должна быть объяснена в product или UX документах.
- Пиши `zelma`, а не `zelima`.
- Пиши `zelma session`, когда речь о repo-local управляемой записи.
- `зессия` и `зешка` допустимы только как разговорные aliases для `zelma session`; в контрактах, schema, CLI help и governed terminology используй canonical term.
- Пиши `zellij session`, когда речь о runtime session мультиплексора.
- Пиши `codex session`, когда речь о Codex identity внутри pane.

## Ambiguous Terms

| Term | Allowed meaning | Forbidden / overloaded meaning | Replacement |
| --- | --- | --- | --- |
| `session` | Только если локальный контекст явно указывает тип | Общее слово для `zelma session`, `zellij session` и `codex session` одновременно | `zelma session`, `zellij session`, `codex session` |
| `pane` | `zellij pane` | Любая вкладка терминала или окно shell | `zellij pane` |
| `registry` | `.zelma/sessions.json` | Глобальная база данных, Codex log store, zellij layout | `session registry` |
| `path` | `opened path` внутри pane | Произвольный путь запуска CLI, repo root или current shell cwd без уточнения | `opened path`, `repo root` |

## Source Documents

- Исходное описание продукта от пользователя в текущей рабочей сессии
  `2026-07-07`.
- Других domain research документов пока нет.
