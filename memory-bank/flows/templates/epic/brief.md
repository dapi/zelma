---
title: "EP-XXX: Шаблон Brief"
doc_kind: epic
doc_function: template
purpose: "Wrapper-шаблон для легкого intake brief epic. Не заменяет charter/roadmap, а помогает быстро разложить roadmap proposal на feature briefs."
derived_from:
  - ../../epic-flow.md
status: active
audience: humans_and_agents
template_for: epic
template_target_path: ../../../epics/EP-XXX/brief.md
---

# EP-XXX: Шаблон Brief

## Instantiated Frontmatter

```yaml
---
title: "EP-XXX: Brief названия epic"
doc_kind: epic
doc_function: brief
purpose: "Легкий brief для epic: problem, outcome, набросок scope и candidate feature briefs."
derived_from:
  - ../../product/roadmap.md
status: draft
audience: humans_and_agents
must_not_define:
  - implementation_sequence
  - feature_acceptance_contracts
  - selected_solution
---
```

## Instantiated Body

```markdown
# EP-XXX: Brief названия epic

## Проблема

Почему эта инициатива нужна и какой gap она закрывает.

## Результат

Какой наблюдаемый результат должен появиться после исполнения epic.

## Набросок Объема

- `EP-XXX-REQ-01` Что входит в инициативу.

## Что Не Входит

- `EP-XXX-NS-01` Что не должно попадать в инициативу.

## Briefs Фич

- [FT-XXX: Feature Name](../../features/FT-XXX/README.md)

## Заметки О Готовности

- Что нужно уточнить до полного `charter.md` / `roadmap.md`.
```
