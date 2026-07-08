# zelma

`zelma` — CLI-утилита и набор Codex skills для управления Codex-сессиями,
запущенными в panes `zellij`.

Текущий стек реализации: Go CLI на Cobra. Первичная интеграция с `zellij`
планируется через внешний `zellij` binary и его CLI automation surface.

Основная единица продукта — `zelma session`: запись о Codex-сессии в
конкретном `zellij pane`. Реестр таких сессий хранится в `.zelma/sessions.json`
в корне репозитория и содержит как минимум:

- `zellij session`;
- `zellij pane`;
- `codex session`;
- путь, открытый внутри pane.

Цель проекта — сделать работу с несколькими Codex-сессиями в одном репозитории
наблюдаемой, воспроизводимой и управляемой из командной строки и из Codex skills.

## Ожидаемый CLI

Первые команды продукта:

- `zelma sessions create` — создает новый `zellij pane`, запускает в нем Codex и
  сохраняет запись о сессии в `.zelma/sessions.json`.
- `zelma sessions detect` — находит вручную созданные `zellij panes`, в которых
  уже запущен Codex, и регистрирует их в `.zelma/sessions.json`.
- `zelma sessions list` — показывает известные `zelma sessions` для текущего
  репозитория.
- `zelma sessions focus <id>` — переключает `zellij` на tab/pane известной
  `zelma session`.

`sessions create` покрывает controlled workflow, где `zelma` создает pane сама.
`sessions detect` покрывает real-world workflow, где пользователь сначала
вручную открыл pane в `zellij`, запустил Codex, а потом хочет поставить эту
сессию под контроль `zelma`.

## Zellij tab focus

По умолчанию `zelma` ориентируется на запуск Codex в `zellij pane`, а не в новой
tab: текущий `zellij action new-tab` переключает focus в созданную tab и может
мешать вводу пользователя в рабочей вкладке.

Для безопасного tab workflow ждем upstream поддержку CLI в
[`zellij-org/zellij#5220`](https://github.com/zellij-org/zellij/issues/5220):
там обсуждается API вроде `new-tab --focus false` и attach directly to tab.
Это нужно, чтобы создавать отдельные agent tabs без focus stealing и без
ненадежного workaround "создать tab, затем быстро вернуться назад".

## Документация проекта

- [`memory-bank/`](memory-bank/README.md) — durable knowledge layer проекта.
- [`memory-bank/product/`](memory-bank/product/README.md) — продуктовый контекст,
  аудитории, метрики и roadmap.
- [`memory-bank/domain/`](memory-bank/domain/README.md) — предметная модель
  `zelma sessions`, правила, состояния, события и bounded contexts.
- [`memory-bank/flows/`](memory-bank/flows/README.md) — шаблоны для будущих PRD,
  epics, features и ADR.

## Локальные проверки

- `python3 scripts/check_memory_bank_index.py` — аудит достижимости markdown-документов, broken links и expected README-индексов внутри `memory-bank/`.
- `git diff --check` — проверка лишних пробелов и conflict markers перед PR.

## Releases

Versioned binaries are published from git tags through GitHub Actions.

```bash
git tag v0.2.0
git push origin v0.2.0
```

The release workflow builds Linux, macOS and Windows archives and publishes them
to <https://github.com/dapi/zelma/releases> with `SHA256SUMS.txt`.

## Installation

Download a versioned binary from
<https://github.com/dapi/zelma/releases>. Replace `v0.2.0` with the version you
want to install.

### macOS

Apple Silicon:

```bash
version=v0.2.0
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/zelma_${version}_darwin_arm64.tar.gz"
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/SHA256SUMS.txt"
shasum -a 256 -c SHA256SUMS.txt --ignore-missing
tar -xzf "zelma_${version}_darwin_arm64.tar.gz"
sudo install -m 0755 "zelma_${version}_darwin_arm64/zelma" /usr/local/bin/zelma
zelma help
```

Intel:

```bash
version=v0.2.0
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/zelma_${version}_darwin_amd64.tar.gz"
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/SHA256SUMS.txt"
shasum -a 256 -c SHA256SUMS.txt --ignore-missing
tar -xzf "zelma_${version}_darwin_amd64.tar.gz"
sudo install -m 0755 "zelma_${version}_darwin_amd64/zelma" /usr/local/bin/zelma
zelma help
```

### Linux

x86_64:

```bash
version=v0.2.0
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/zelma_${version}_linux_amd64.tar.gz"
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/SHA256SUMS.txt"
sha256sum -c SHA256SUMS.txt --ignore-missing
tar -xzf "zelma_${version}_linux_amd64.tar.gz"
sudo install -m 0755 "zelma_${version}_linux_amd64/zelma" /usr/local/bin/zelma
zelma help
```

ARM64:

```bash
version=v0.2.0
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/zelma_${version}_linux_arm64.tar.gz"
curl -LO "https://github.com/dapi/zelma/releases/download/${version}/SHA256SUMS.txt"
sha256sum -c SHA256SUMS.txt --ignore-missing
tar -xzf "zelma_${version}_linux_arm64.tar.gz"
sudo install -m 0755 "zelma_${version}_linux_arm64/zelma" /usr/local/bin/zelma
zelma help
```

### Windows PowerShell

x64:

```powershell
$version = "v0.2.0"
Invoke-WebRequest "https://github.com/dapi/zelma/releases/download/$version/zelma_${version}_windows_amd64.zip" -OutFile "zelma.zip"
Expand-Archive .\zelma.zip -DestinationPath .
.\zelma_${version}_windows_amd64\zelma.exe help
```

ARM64:

```powershell
$version = "v0.2.0"
Invoke-WebRequest "https://github.com/dapi/zelma/releases/download/$version/zelma_${version}_windows_arm64.zip" -OutFile "zelma.zip"
Expand-Archive .\zelma.zip -DestinationPath .
.\zelma_${version}_windows_arm64\zelma.exe help
```

Move `zelma.exe` into a directory listed in `PATH` if you want to run it from
any terminal.

### Аудит ссылок и индексации `memory-bank`

Скрипт [`scripts/check_memory_bank_index.py`](scripts/check_memory_bank_index.py) аудирует `memory-bank/` и проверяет:

- broken relative markdown links внутри audit scope;
- orphan-документы, на которые никто не ссылается внутри scope;
- достижимость каждого документа от entrypoint'ов по индексной навигации;
- документы, которые достижимы только глубже порога навигации;
- contract ожидаемых `README.md`-индексов.

Обычный локальный запуск из корня репозитория:

```bash
python3 scripts/check_memory_bank_index.py
```

Что означает результат:

- exit code `0` — errors не найдены; warnings по глубине возможны, но аудит считается пройденным;
- non-zero exit code — найдены проблемы, которые нужно исправить до PR;
- `--json` — структурированный отчёт, пригодный для последующей автоматической доиндексации другим агентом или инструментом.

Параметры запуска:

- `--max-depth N` — порог глубины индексной навигации в прыжках; по умолчанию `3`; документы глубже порога попадают в warning, а не в error;
- `--entrypoint PATH` — явный entrypoint для аудита; параметр repeatable; принимает repo-relative или scope-relative пути; неоднозначные пути без префикса сначала резолвятся внутри `--scope-root`, а для явного repo-root пути используйте `./PATH` или `/PATH`; если передан, используется вместо дефолтного `memory-bank/README.md`;
- `--scope-root DIR` — меняет audit scope; по умолчанию `memory-bank`;
- `--repo-root DIR` — явно задаёт корень репозитория; полезно для сетевого запуска или локально установленной копии скрипта;
- `--json` — печатает только JSON-отчёт.

Примеры:

```bash
python3 scripts/check_memory_bank_index.py --max-depth 4
```

```bash
python3 scripts/check_memory_bank_index.py \
  --entrypoint README.md \
  --entrypoint AGENTS.md \
  --max-depth 4
```

Быстрый запуск по сети без предварительной установки:

```bash
curl -fsSL https://raw.githubusercontent.com/dapi/memory-bank/main/scripts/check_memory_bank_index.py \
  | python3 - --repo-root .
```

Локальная установка или копирование с GitHub:

1. Скопируйте файл со страницы `scripts/check_memory_bank_index.py` на GitHub или скачайте raw-версию:

```bash
mkdir -p ./tools
curl -fsSL \
  -o ./tools/check_memory_bank_index.py \
  https://raw.githubusercontent.com/dapi/memory-bank/main/scripts/check_memory_bank_index.py
chmod +x ./tools/check_memory_bank_index.py
```

2. Запускайте его из корня downstream-репозитория:

```bash
python3 ./tools/check_memory_bank_index.py --repo-root .
```

`macOS` и `Linux`: команды запуска одинаковые. Отличие только в том, куда класть локальную копию, если хочется вызывать скрипт без относительного пути: на Linux чаще используют `~/.local/bin`, на macOS — `~/bin` или любой каталог, добавленный в `PATH`. Если не хотите менять `PATH`, запускайте скрипт через `python3` по полному или относительному пути.

Когда запускать:

- после добавления, удаления или переименования `.md`-файлов в `memory-bank/`;
- после правок `README.md`-индексов и относительных ссылок;
- перед открытием PR с изменениями в template navigation или document structure.

## Настроечные промпты для агента

Запукаются в новых сессиях

```text
Прочитай ./memory-bank и предложи адаптацию AGENTS.md под текущие правила проекта zelma.
```

```text
Прочитай ./memory-bank и помоги уточнить секцию `product`
```

```text
Прочитай ./memory-bank и помоги уточнить секцию `domain`
```

```text
Прочитай ./memory-bank и помоги адаптировать секцию `ops`
```

```text
Прочитай ./memory-bank и помоги адаптировать секцию `engineering`
```

```text
Проведи ревью memory-bank на document governance
```
(внеси правки и повторить до состояния которое вас устроит)


```text
Проведи ревью memory-bank на консистетность, и непротиворечивость
```
(внеси правки и повторить до состояния которое вас устроит)

```text
У нас в проекте подключен memory-bank. Я хочу быть уверен что все страницы в этом memory-bank-а так или иначе доступны через нидексацию начиная с
AGENTS.md. Если страница не упомянются напрямую, то она упомянутся в файле который упомянут в файле который упомянут в AGENTS.md и так далее на глубину до 4-х шагов.
```

```text
Помоги создать PRD
```

```text
Помоги создать глоссарий
```

## Что есть внутри шаблона

- `memory-bank/dna/` — governance-ядро: SSoT, frontmatter, lifecycle, cross-references.
- `memory-bank/flows/` — lifecycle flows и шаблоны для PRD/feature/ADR.
- `memory-bank/product/` — product context, vision, customers, metrics, marketing и roadmap `zelma`.
- `memory-bank/domain/` — glossary, domain model, rules, states, events и context map `zelma`.
- `memory-bank/prd/` — место для instantiated Product Requirements Documents.
- `memory-bank/use-cases/` — место для instantiated project-level use cases.
- `memory-bank/engineering/` — architecture patterns, frontend engineering, testing policy, coding style, autonomy boundaries, git workflow.
- `memory-bank/ops/` — заготовки для development, stages, releases, config и runbooks.
- `memory-bank/adr/` — место для instantiated ADR.
- `memory-bank/features/` — место для instantiated feature packages.
