---
title: "ADR-001: MVP CLI Architecture"
doc_kind: adr
doc_function: canonical
purpose: "Фиксирует архитектуру MVP CLI: Go, Cobra, zellij adapter через CLI automation и файловый registry."
derived_from:
  - ../product/roadmap.md
  - ../domain/context-map.md
  - ../engineering/architecture.md
  - ../engineering/zellij-integration.md
status: active
decision_status: accepted
date: 2026-07-07
audience: humans_and_agents
must_not_define:
  - current_system_state
  - implementation_plan
---

# ADR-001: MVP CLI Architecture

## Контекст

`zelma` должен управлять Codex-сессиями, запущенными в `zellij panes`, и хранить
repo-local registry в `.zelma/sessions.json`.

MVP должен поддержать:

- `zelma sessions create`;
- `zelma sessions detect`;
- `zelma sessions list`.

Для этого CLI должен координировать три разные зоны ответственности:

- пользовательский command surface;
- runtime state `zellij`;
- файловое состояние `.zelma/`.

## Драйверы решения

- CLI должен быть удобен человеку и пригоден для вызова из Codex skills.
- Help output должен быть оптимизирован для агентов в первую очередь: точные
  команды, safe defaults, JSON flags и recovery hints должны появляться раньше
  человеческого объяснения.
- Интеграция с `zellij` должна быть явной, тестируемой и fixture-friendly.
- `.zelma/sessions.json` должен оставаться единственным repo-local source of
  truth для `zelma sessions`.
- Команды не должны напрямую смешивать parsing flags, вызовы `zellij` и запись
  JSON.
- MVP не должен зависеть от Zellij plugin/WASM lifecycle.

## Рассмотренные варианты

| Вариант | Плюсы | Минусы | Почему выбран / не выбран |
| --- | --- | --- | --- |
| Go CLI + Cobra + internal zellij adapter over `os/exec` | Простой binary, понятная command tree, стандартная модель для nested commands, можно тестировать adapter через fixtures | Зависит от стабильности zellij CLI JSON output | Выбран для MVP как самый прямой путь к `create/list/detect` |
| Go CLI без CLI framework, только standard `flag` | Меньше зависимостей | Неудобнее nested commands вида `sessions create/list/detect`, больше ручного boilerplate | Не выбран для MVP |
| Zellij plugin/WASM как primary integration | Глубже интегрируется с runtime zellij | Добавляет plugin permissions, WASM lifecycle, Rust-oriented API и отдельную доставку plugin artifact | Отложено до отдельного ADR |
| Прямой Go zellij client library | Мог бы дать typed API | Официального Go zellij client library не найдено; риск поддержки выше, чем у CLI automation | Не выбран |

## Решение

MVP `zelma` реализуется как Go CLI.

Command layer:

- использовать `github.com/spf13/cobra` для command tree;
- переопределить Cobra help/usage templates под agent-first output;
- primary command group: `zelma sessions`;
- начальные subcommands: `create`, `detect`, `list`.

Application layer:

- `internal/app` содержит use cases и координирует dependencies;
- CLI handlers не вызывают `zellij` и не пишут registry напрямую.

Adapters:

- `internal/zellij` вызывает внешний `zellij` binary через Go `os/exec`;
- primary zellij commands:
  - `zellij list-sessions --short --no-formatting`;
  - `zellij --session <name> action list-panes --json --all`;
  - `zellij --session <name> run --cwd <path> --name <name> -- codex`;
- `internal/codex` отвечает за Codex runtime/session identification.

Registry:

- `internal/registry` владеет `.zelma/sessions.json`;
- запись делается под lock и через atomic replace;
- registry не вызывает `zellij`, а `zellij-adapter` не пишет registry.

Filesystem/repo:

- `internal/repo` централизованно определяет repo root и `.zelma/` paths.

## Последствия

### Положительные

- CLI structure естественно отражает продуктовый surface.
- Skills смогут вызывать те же команды, что и человек.
- `zelma` и `zelma help` станут discovery surface для агентов, а не только
  справкой для человека.
- Zellij integration можно тестировать fixture JSON без live terminal session.
- Registry остается изолированным persistence boundary.

### Отрицательные

- Нужно поддерживать compatibility с zellij CLI output.
- Cobra добавляет внешнюю зависимость.
- Понадобятся snapshot/contract tests для help output, чтобы дефолтный Cobra
  help случайно не вернулся.
- `os/exec` adapter требует аккуратной обработки timeouts, stderr и exit codes.

### Нейтральные / организационные

- Zellij plugin integration переносится в future ADR.
- Minimum supported zellij version нужно подтвердить tests/fixtures.
- Go toolchain должен быть установлен в dev/CI окружении.

## Риски и mitigation

- Риск: zellij JSON shape меняется.
  Mitigation: fixture tests по поддерживаемым версиям и documented compatibility.

- Риск: команды без `--session` воздействуют на ambient zellij session.
  Mitigation: adapter использует explicit session targeting везде, где это возможно.

- Риск: registry повреждается при concurrent write.
  Mitigation: lock + reload + validate + atomic replace.

- Риск: Codex session id нельзя надежно извлечь.
  Mitigation: не записывать `active` record без `CodexSessionRef`; использовать
  `candidate` или warning до отдельного design решения.

## Follow-up

- Создать Go scaffold.
- Добавить `internal/zellij` с typed adapter interface.
- Добавить registry schema v1 и persistence tests.
- Зафиксировать supported zellij/Codex versions после первых integration checks.
- Позже рассмотреть ADR для plugin/WASM integration, если CLI automation станет
  недостаточной.

## Связанные ссылки

- [Engineering Architecture Patterns](../engineering/architecture.md)
- [Zellij Integration Research](../engineering/zellij-integration.md)
- [Domain Context Map](../domain/context-map.md)
- [Product Roadmap](../product/roadmap.md)
