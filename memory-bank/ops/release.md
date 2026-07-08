---
title: Release And Deployment
doc_kind: ops
doc_function: canonical
purpose: Шаблон документа релизного процесса. Читать при адаптации versioning, changelog, deployment и release verification под проект.
derived_from:
  - ../dna/governance.md
status: active
audience: humans_and_agents
---

# Release And Deployment

## Release Flow

`zelma` releases are tag-driven. A GitHub Release is created automatically when a
version tag matching `v*` is pushed.

1. Update [`../../CHANGELOG.md`](../../CHANGELOG.md) if the release changes user
   visible behavior.
2. Verify locally:

   ```bash
   mise exec -- go test ./...
   python3 scripts/check_memory_bank_index.py
   git diff --check
   ```

3. Commit and push all release-prep changes to `main`.
4. Create and push a version tag:

   ```bash
   git tag v0.2.0
   git push origin v0.2.0
   ```

5. GitHub Actions workflow `.github/workflows/release.yml` builds and publishes
   the release.
6. Download binaries from the GitHub Releases page:
   `https://github.com/dapi/zelma/releases`.

## Automated Release Artifacts

For every `v*` tag, CI publishes:

- `zelma_<version>_linux_amd64.tar.gz`
- `zelma_<version>_linux_arm64.tar.gz`
- `zelma_<version>_darwin_amd64.tar.gz`
- `zelma_<version>_darwin_arm64.tar.gz`
- `zelma_<version>_windows_amd64.zip`
- `zelma_<version>_windows_arm64.zip`
- `SHA256SUMS.txt`

The workflow also supports manual dispatch with an existing tag through
`workflow_dispatch`.

## Release Commands

```bash
# local release checks
mise exec -- go test ./...
python3 scripts/check_memory_bank_index.py
git diff --check

# create a release
git tag v0.2.0
git push origin v0.2.0

# inspect release artifacts
gh release view v0.2.0 --repo dapi/zelma
```

## Safety Rules

- Release tags must start with `v`.
- Do not move an already published tag unless explicitly coordinating a broken
  release recovery.
- Do not upload registry contents or `.zelma/` state into release artifacts.
- Release binaries are built from the tag ref, not from local working tree state.
- `contents: write` is required only for the release workflow to publish GitHub
  Release assets.

## Release Test Plan

При каждом релизе полезно создавать отдельный тестовый план.

**Формат:** `release-v{VERSION}-test-plan.md`

**Минимальная структура:**

```markdown
# Тестовый план релиза v{VERSION}

**Дата:** YYYY-MM-DD
**Предыдущая версия:** v{PREV_VERSION}
**Текущая версия:** v{VERSION}
**Стенд:** <environment>

## Обзор изменений

| Issue | Название | Тип | Приоритет |
| --- | --- | --- | --- |
| #XXXX | Описание задачи | Feature/Fix/Refactoring/Tech debt | Высокий/Средний/Низкий |

## Проверка изменений

- [ ] Описан хотя бы один test case для каждого крупного change set

## Smoke-тесты

- [ ] Главная страница открывается
- [ ] Основной пользовательский поток работает
- [ ] Админский или внутренний путь работает
- [ ] Health endpoint отвечает успешно
```

## Rollback

`zelma` is currently a CLI binary release. Rollback means installing an older
GitHub Release asset.

- Rollback unit: one versioned binary archive from GitHub Releases.
- Fastest safe rollback: download the previous release archive for the current
  OS/architecture and replace the local `zelma` binary.
- Registry files remain repo-local and are not migrated by the release workflow.
- If a release asset is broken, publish a new patch tag such as `v0.1.1` instead
  of mutating the old release.
