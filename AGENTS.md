# Правила репозитория

## Структура проекта и организация модулей

Корневой `README.md` объясняет назначение `zelma`, ожидаемый CLI и навигацию по
проектной документации.

- `memory-bank/` — durable knowledge layer проекта.
- `memory-bank/product/` — продуктовый контекст, пользователи, метрики и roadmap.
- `memory-bank/domain/` — предметная модель `zelma sessions`, правила, состояния и события.
- `memory-bank/dna/` — governance-ядро документации.
- `memory-bank/flows/` — reusable lifecycle docs и governed templates.
- `memory-bank/prd/` — instantiated Product Requirements Documents.
- `memory-bank/use-cases/` — instantiated канонические сценарии проекта.
- `memory-bank/engineering/`, `memory-bank/ops/` — engineering и операционный контекст.
- `memory-bank/adr/` и `memory-bank/features/` — пустые или минимальные точки назначения для instantiated документов.
- `.zelma/` — будущий runtime-каталог проекта; `sessions.json` в нем является
  локальным реестром `zelma sessions`.

Пока runtime-код не создан, не выдумывайте структуру `src/` без отдельного
решения или feature package. Product/domain-факты фиксируйте в `memory-bank/`
до реализации, чтобы CLI-контракты не расходились с документацией.

## Команды разработки и проверки

Пока у репозитория нет собственного build/runtime-приложения. Перед PR достаточно
легких проверок:

- `rg --files memory-bank` для проверки структуры и имён файлов;
- `python3 scripts/check_memory_bank_index.py` для аудита ссылок, reachability и expected README-индексов внутри `memory-bank/`;
- `git diff --check` для поиска лишних пробелов и conflict markers;
- `sed -n '1,120p' path/to/doc.md` для быстрой проверки frontmatter и заголовков;
- `rg -n "zelima|Zelima" .` для проверки частой опечатки в названии проекта.

## Стиль оформления и соглашения по именованию

Пишите в Markdown: короткие секции, понятные заголовки, относительные ссылки. Governed-документы в `memory-bank/` должны начинаться с YAML frontmatter; поле `status` обязательно всегда, а `derived_from`, `delivery_status` и `decision_status` добавляются, когда этого требует тип документа. См. `memory-bank/dna/frontmatter.md`.

Для обычных документов используйте lowercase kebab-case, например `testing-policy.md`. Для структурированных артефактов сохраняйте шаблонные naming rules, например `features/FT-XXX/` и `ADR-XXX-short-decision-name.md`.

## Правила проверки

Автоматизированного тестового набора у репозитория нет. Проверяйте изменения вручную:

- убедитесь, что индексы и ссылки соответствуют новой структуре;
- не дублируйте один и тот же product/domain-факт в нескольких документах без явного canonical owner;
- при изменении domain rules проверяйте соседние `model.md`, `states.md`, `events.md` и `context-map.md` на противоречия.

## Коммиты и pull request

Следуйте конвенции из `memory-bank/engineering/git-workflow.md`: короткие commit messages в настоящем времени, например `docs: define zelma session domain`.

В pull request опишите:

- что изменено в продуктовой, доменной или инженерной модели;
- какие CLI-контракты, ссылки или naming rules были затронуты.
