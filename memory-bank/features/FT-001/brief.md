---
title: "FT-001: Go Module Scaffold"
doc_kind: feature
doc_function: canonical
purpose: "Канонический brief для первого delivery slice: создать Go module scaffold и пустой `zelma` binary без registry/zellij side effects."
derived_from:
  - ../../flows/feature-flow.md
  - ../../product/context.md
  - ../../product/roadmap.md
  - ../../epics/EP-001/charter.md
  - ../../adr/ADR-001-mvp-cli-architecture.md
  - ../../engineering/architecture.md
  - ../../engineering/testing-policy.md
status: active
delivery_status: implemented
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - solution_space
---

# FT-001: Go Module Scaffold

## Что

### Проблема

У проекта `zelma` есть принятая архитектура MVP CLI и roadmap, но нет Go module,
entrypoint и проверяемого binary skeleton. Без scaffold нельзя начать
реализацию command tree, help output и последующих registry/zellij features.

### Результат

| ID метрики | Метрика | База | Цель | Способ измерения |
| --- | --- | --- | --- | --- |
| `MET-01` | Go module существует | нет `go.mod` | `go test ./...` обнаруживает module packages | local command |
| `MET-02` | Binary entrypoint существует | нет `cmd/zelma` | `go build ./cmd/zelma` завершается успешно | local command |
| `MET-03` | Scope остается без side effects | кода нет | нет runtime-записей `.zelma/` и live-вызовов `zellij` | code review + tests |

### Объем Работ

- `REQ-01` Создать `go.mod` для репозитория.
- `REQ-02` Создать entrypoint `cmd/zelma/main.go`.
- `REQ-03` Создать минимальный internal package layout, совместимый с ADR-001,
  без реализации registry или zellij behavior.
- `REQ-04` Зафиксировать `go test ./...` и `go build ./cmd/zelma` как
  canonical checks для этого slice.
- `REQ-05` Оставить CLI behavior достаточно минимальным, чтобы `FT-002` владел
  Cobra command tree, а `FT-003` владел agent-first help templates.

### Что Не Входит

- `NS-01` Нет Cobra command tree сверх строго необходимого для компиляции, если
  это вообще понадобится.
- `NS-02` Нет behavior для `sessions list/create/detect`.
- `NS-03` Нет schema, read или write behavior для `.zelma/sessions.json`.
- `NS-04` Нет live-выполнения `zellij`.
- `NS-05` Нет Codex session identification.
- `NS-06` Нет GitHub Actions или release packaging.

### Ограничения И Предположения

- `ASM-01` Go toolchain будет установлен до начала реализации.
- `CON-01` Package layout должен оставаться совместимым с
  [ADR-001](../../adr/ADR-001-mvp-cli-architecture.md).
- `CON-02` Эта feature не должна добавлять runtime side effects.
- `DEC-01` Cobra выбрана в ADR-001, но полный Cobra command tree принадлежит
  следующей feature, если scaffold не потребует минимальную dependency.

## Решение О Необходимости Design

| Решение | Причина | Downstream-владелец |
| --- | --- | --- |
| `Design required: no` | Architecture уже принята в ADR-001, а эта feature только создает scaffold. Здесь не выбирается новый integration contract, schema или runtime side effect. | `none` |

## Проверка

### Критерии Готовности

- `EC-01` `go.mod` существует и объявляет project module.
- `EC-02` `cmd/zelma/main.go` существует и собирается.
- `EC-03` `go test ./...` завершается успешно.
- `EC-04` implementation не вызывает `zellij` и не пишет `.zelma/sessions.json`.

### Матрица Трассировки

| ID требования | Ссылки на проблему | Ссылки на приемку | Проверки | ID доказательств |
| --- | --- | --- | --- | --- |
| `REQ-01` | `ASM-01`, `CON-01` | `EC-01`, `SC-01` | `CHK-01` | `EVID-01` |
| `REQ-02` | `ASM-01`, `CON-01` | `EC-02`, `SC-01` | `CHK-02` | `EVID-02` |
| `REQ-03` | `CON-01`, `CON-02` | `EC-04`, `SC-02` | `CHK-03` | `EVID-03` |
| `REQ-04` | `ASM-01` | `EC-03` | `CHK-01`, `CHK-02` | `EVID-01`, `EVID-02` |
| `REQ-05` | `DEC-01`, `CON-02` | `EC-04`, `SC-02` | `CHK-03` | `EVID-03` |

### Сценарии Приемки

- `SC-01` Разработчик или агент с установленным Go может запустить
  `go test ./...` и `go build ./cmd/zelma` из repo root.
- `SC-02` Scaffold задает package boundaries без попытки обращаться к live
  `zellij`, Codex или `.zelma/sessions.json`.

### Проверки

| ID проверки | Покрывает | Как проверить | Ожидаемый результат | Путь доказательств |
| --- | --- | --- | --- | --- |
| `CHK-01` | `EC-01`, `EC-03`, `SC-01` | `go test ./...` | command exits 0 | `artifacts/ft-001/verify/chk-01/` |
| `CHK-02` | `EC-02`, `SC-01` | `go build ./cmd/zelma` | command exits 0 | `artifacts/ft-001/verify/chk-02/` |
| `CHK-03` | `EC-04`, `SC-02` | code review / `rg -n "zellij|sessions.json|\\.zelma" cmd internal` | нет runtime invocation/write behavior в scaffold | `artifacts/ft-001/verify/chk-03/` |

### Матрица Тестов

| ID проверки | ID доказательств | Путь доказательств |
| --- | --- | --- |
| `CHK-01` | `EVID-01` | `artifacts/ft-001/verify/chk-01/` |
| `CHK-02` | `EVID-02` | `artifacts/ft-001/verify/chk-02/` |
| `CHK-03` | `EVID-03` | `artifacts/ft-001/verify/chk-03/` |

### Доказательства

- `EVID-01` Captured output для `go test ./...`.
- `EVID-02` Captured output для `go build ./cmd/zelma`.
- `EVID-03` Review note или command output, подтверждающий отсутствие runtime side effects.

### Контракт Доказательств

| ID доказательства | Artifact | Producer | Path contract | Используется проверками |
| --- | --- | --- | --- | --- |
| `EVID-01` | Test output | implementer | `artifacts/ft-001/verify/chk-01/` | `CHK-01` |
| `EVID-02` | Build output | implementer | `artifacts/ft-001/verify/chk-02/` | `CHK-02` |
| `EVID-03` | Side-effect review note | implementer / reviewer | `artifacts/ft-001/verify/chk-03/` | `CHK-03` |
