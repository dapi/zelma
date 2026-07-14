# Rename `sessions` To `instances` Protocol

## Goal

Переименовать публичный продуктовый ресурс `zelma session` в `zelma instance`,
чтобы убрать конфликт с внешними терминами `zellij session` и `codex session`.

Целевой контракт:

- domain concept: `zelma instance`;
- canonical CLI resource: `zelma instances`;
- removed CLI resource: `zelma sessions`;
- registry path: `.zelma/instances.json`;
- schema v1 root key: `instances`;
- technical fields `zellij_session` и `codex_session` не переименовываются.

## Scope

### In Scope

- Удалить `zelma sessions` без deprecated alias.
- Перевести help, diagnostics, recovery hints, README, skill contract и domain
  docs на `zelma instances`.
- Перевести repo-local registry contract с `.zelma/sessions.json` на
  `.zelma/instances.json`.
- Перевести public JSON root key с `sessions` на `instances`.
- Обновить unit/e2e tests на canonical `instances` path.
- Добавить проверку, что `zelma sessions` больше не является командой.

### Out Of Scope

- Переименование external protocol terms: `zellij session`, `codex session`,
  `zellij list-sessions`, `CODEX_HOME/sessions`.
- Полный рефактор внутренних Go type names, если публичный контракт уже
  переименован.

## Compatibility

Backward compatibility intentionally отсутствует:

- `zelma sessions <subcommand>` должен завершаться ошибкой unknown command.
- Старый `.zelma/sessions.json` не читается как fallback.
- Старый JSON root key `sessions` не принимается как fallback schema.
- Skills должны использовать только `zelma instances`.

## Acceptance Criteria

- `zelma instances help` показывает canonical command map.
- `zelma sessions list --json` завершается non-zero как removed command.
- `zelma instances list --json` возвращает schema v1 object с root key
  `instances`.
- Новые registry writes создают `.zelma/instances.json`.
- Diagnostics и recovery hints предлагают только `zelma instances ...`.
- `SKILL.md` запрещает direct registry parsing и ссылается на
  `.zelma/instances.json`.
- README, changelog и engineering skill contract не описывают deprecated alias.
- Технические внешние упоминания `zellij session`, `codex session`,
  `zellij list-sessions` и `CODEX_HOME/sessions` сохранены.
- `go test ./...`, `python3 scripts/check_memory_bank_index.py`,
  `git diff --check` и typo check `rg -n "zelima|Zelima" .` проходят.

## Review

### Findings

- High: additive alias contradicts current product decision. Keeping it would
  preserve the overloaded resource name and keep skills on the old path.
- High: changing only the CLI path while preserving `.zelma/sessions.json` and
  JSON key `sessions` would leave the public machine-readable contract half
  renamed.
- Medium: `sessions` cannot disappear globally because it is part of external
  protocols (`zellij list-sessions`, Codex transcript directory names).
- Low: internal Go identifiers can be renamed later; doing it now is lower
  value than making public CLI/storage/JSON contracts consistent.

### Verdict

Ship this as a breaking rename with no alias and no fallback. Keep only external
technical `session` terms where they describe zellij or Codex, not Zelma's
public resource.

## Verification

Run:

```bash
go test ./...
python3 scripts/check_memory_bank_index.py
git diff --check
rg -n "zelima|Zelima" .
```

Manual smoke:

```bash
./zelma instances help
```

Removed command check:

```bash
./zelma sessions list --json
```

Expected result: `instances help` exits `0`; `sessions list --json` exits
non-zero with unknown command because the old resource was removed.
