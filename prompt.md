Прочитай issue {ISSUE_NUMBER} и возьми задачу в работу через новый workflow routing.

<context>
Раньше start-issue всегда запускал `feature-flow`. Теперь новая задача сначала проходит selector из `memory-bank/flows/workflows.md`, а `feature-flow` применяется только если выбран `workflow_profile=feature-package`.

Основная цель: выбрать минимальный workflow/document profile, который не теряет контроль над риском, evidence и traceability.

Терминальная цель start-issue: либо довести задачу до merge-ready PR с зеленым обязательным CI, либо остановиться на явном human gate/blocker. Нельзя завершать работу статусом "готово", если PR/CI/review еще нужно делать отдельным следующим шагом.
</context>

<required_reading>
Перед действиями прочитай:

1. `AGENTS.md`
2. `memory-bank/README.md`
3. `memory-bank/documentation-authority.md`
4. `memory-bank/flows/workflows.md`
5. `memory-bank/flows/task-flow.md`
6. `memory-bank/flows/bugfix-flow.md`
7. `memory-bank/flows/refactor-flow.md`
8. `memory-bank/flows/workflow-decision-log.md`
9. `memory-bank/flows/workflow-metrics.md`
10. `memory-bank/prompts/PROMPT-008-workflow-routing-and-development-kickoff.md`

Затем прочитай issue `{ISSUE_NUMBER}` через `gh`.
</required_reading>

<routing_first>
Сначала выполни routing-only phase:

1. Сформируй Routing Signature:
   - `kind`
   - `size`
   - `risk`
   - `change_surface`
   - `contract_change`
   - `design_need`
   - `evidence_need`
   - `owner_doc_need`
   - `workflow_profile`

2. Проверь promotion triggers из `workflows.md`:
   - contract/API/schema/event/security/financial/integration/rollout changes;
   - alternatives/trade-off reasoning, migration, rollout/backout или failure-mode design;
   - несколько bounded contexts, engines, shared abstractions или runtime topology;
   - новый устойчивый user/operator/system scenario;
   - repeated review findings по scope/design/evidence;
   - `risk=high/critical` или irreversible/external side effects.

3. Если `kind`, `risk`, `change_surface` или `contract_change` остаются `unknown`, не начинай разработку. Остановись и верни human gate с вопросами.

4. Если выбран compact profile, убедись, что нет promotion trigger-а.
</routing_first>

<development_start>
После routing-а начни работу по выбранному profile:

1. `tracker-only`
   - Следуй `memory-bank/flows/task-flow.md`.
   - Не создавай memory-bank package.
   - Зафиксируй routing signature и evidence в issue/PR carrier.
   - Реализуй минимальный scoped diff.

2. `bugfix-compact`
   - Следуй `memory-bank/flows/task-flow.md`.
   - Следуй `memory-bank/flows/bugfix-flow.md`.
   - Используй `memory-bank/flows/templates/task/bugfix.md`.
   - Зафиксируй symptom, reproduction, root cause, fix boundary и regression coverage.
   - Создавай `memory-bank/tasks/TASK-{ISSUE_NUMBER}/README.md` и `bugfix.md` только если issue/PR carrier недостаточен.

3. `refactor-small`
   - Следуй `memory-bank/flows/task-flow.md`.
   - Следуй `memory-bank/flows/refactor-flow.md`.
   - Используй `memory-bank/flows/templates/task/refactor.md`.
   - Зафиксируй intent, behavior invariants, change surface и verification.
   - Создавай `memory-bank/tasks/TASK-{ISSUE_NUMBER}/README.md` и `refactor.md` только если нужен durable managed-task context.

4. `managed-task`
   - Следуй `memory-bank/flows/task-flow.md`.
   - Создай или обнови `memory-bank/tasks/TASK-{ISSUE_NUMBER}/README.md`.
   - Для managed bugfix следуй `memory-bank/flows/bugfix-flow.md`; для managed refactor/chore следуй `memory-bank/flows/refactor-flow.md`.
   - Используй bugfix/refactor template по типу задачи.
   - Если появляется capability или contract trigger, остановись и promote до `feature-package`.

5. `feature-package`
   - Прочитай `memory-bank/flows/feature-flow.md`.
   - Создай или обнови `memory-bank/features/FT-{ISSUE_NUMBER}/`.
   - Сначала `brief.md`, затем `design.md` при необходимости, затем `implementation-plan.md`, потом код.
   - Проведи не более 5 циклов `review-improve` для feature-документов перед execution, если создаётся или существенно меняется feature package.

6. `epic-package`
   - Прочитай `memory-bank/flows/epic-flow.md`.
   - Создай или обнови `memory-bank/epics/EP-{ISSUE_NUMBER}/`.
   - Code-level execution выноси в отдельные feature/task delivery units.

7. `incident-pir`
   - Используй incident/PIR owner docs и runbooks.
   - Follow-up implementation routing выполняй отдельно для каждой delivery task.
</development_start>

<worktree_and_commands>
Если нужно создавать новую ветку, делай это только через git worktree:

- `git worktree add -b <branch> <path> <base-ref>`
- не используй `git checkout -b`

Все Rails/Rake/RSpec/MySQL команды выполняй только через `./bin/dip`.
Перед DB-dependent командами проверь dip services.
Если пользователь запретил локальные тесты, не запускай даже `./bin/dip rspec`.
</worktree_and_commands>

<review_improve_for_feature_package>
Если выбран `workflow_profile=feature-package` и создан/изменён комплект feature-документов, проведи максимум 5 циклов `review-improve`:

1. Проведи ревью комплекта документов.
   Проверь:
   - целостность между документами;
   - непротиворечивость;
   - полноту обязательных разделов и связей;
   - открытые вопросы, допущения, пробелы в решениях;
   - расхождения между decision log, brief/feature spec, design, implementation plan и verify/evidence.

2. Сформируй замечания:
   - `critical`
   - `important`
   - `minor`

3. Если `critical` и `important` нет, останови цикл досрочно.

4. Для открытых вопросов, которые блокируют устранение `critical` или `important`:
   - закрой вопрос с помощью FPF;
   - обоснуй решение фактами из issue, memory-bank и текущего комплекта документов;
   - не придумывай факты;
   - зафиксируй результат в `decision-log.md`.

5. Если данных недостаточно и решение materially влияет на feature, остановись и оформи human gate.

6. Исправь все `critical` и `important`, которые можно закрыть без human gate.

7. Запусти следующий цикл.
</review_improve_for_feature_package>

<delivery_finish_gate>
После реализации по выбранному profile доведи delivery до одного из терминальных статусов:

1. `merge_ready_pr`
   - все in-scope изменения выполнены;
   - релевантные локальные проверки запущены или явно обоснованно не запускались;
   - изменения закоммичены по project commit policy:
     `fix(issue-{ISSUE_NUMBER}): description` / `feat(issue-{ISSUE_NUMBER}): description` / другой корректный тип,
     в body есть `Fixes #{ISSUE_NUMBER}` и `Issue: {ISSUE_URL}`;
   - branch запушен;
   - PR создан или обновлен через `gh`;
   - PR открыт против `{BASE_BRANCH}` или другой явно обоснованной base branch;
   - PR не draft, если нет blockers и пользователь явно не просил draft;
   - PR не имеет merge conflicts;
   - свежий review/fix loop завершен без `critical`/`high` или `critical`/`important` замечаний;
   - обязательный CI зеленый.

2. `stopped_by_human_gate`
   - нужен выбор человека, потому что данных недостаточно без домысла;
   - нужен product/security/financial/ops approval;
   - нужна ручная Codex `/review`, а агент в текущей surface не может ее вызвать сам;
   - branch protection требует human approval/review, которого агент не может заменить;
   - merge conflict требует содержательного решения, а не механического rebase.

3. `blocked`
   - GitHub/CI/gh недоступны;
   - нет прав на push/PR/check logs;
   - CI не дает получить обязательный статус или логи после повторных попыток;
   - локальное окружение не позволяет выполнить обязательную проверку, и это нельзя безопасно заменить.

4. `max_cycles_reached`
   - лимит review/fix или CI/fix итераций исчерпан, а blockers остались.
</delivery_finish_gate>

<publish_pr>
Если изменения есть и нет human gate:

1. Проверь `git status`, `git diff --check`, staged/untracked files и scope diff.
2. Запусти релевантные проверки по `memory-bank/engineering/testing-policy.md` и feature/task evidence plan.
3. Закоммить только in-scope изменения.
4. Push текущей branch.
5. Найди существующий PR:
   - `gh pr view --json number,url,baseRefName,headRefName,isDraft,mergeStateStatus,mergeable`
6. Если PR нет, создай его через `gh pr create`:
   - base: `{BASE_BRANCH}` если issue/flow не требует другую base branch;
   - title: project-style title с `issue-{ISSUE_NUMBER}`;
   - body: summary, tests/checks, docs/evidence links, risks/manual gaps, `Fixes #{ISSUE_NUMBER}`.
7. Если PR есть, обнови body/comment так, чтобы в нем были summary, checks и evidence.
8. Если PR draft без blocker-а, переведи в ready for review.
</publish_pr>

<review_fix_loop>
После PR проведи review/fix loop до готовности или terminal gate:

1. Выполни self-review diff against base branch.
2. Если доступен Codex `/review`, проведи свежий review PR/base diff.
3. Если агент не может сам вызвать `/review`, остановись как `stopped_by_human_gate` и попроси пользователя выполнить `/review` against base branch, вернуть результат в thread, затем продолжить loop.
4. Исправь все `critical`/`high` или `critical`/`important` замечания, которые можно закрыть без human gate.
5. Не исправляй unrelated nits, style-only suggestions и optional refactor, если они не несут bug/regression/data/security/maintainability risk.
6. После каждого fix: локальная проверка по scope, commit, push, новый review pass.
7. Лимит: максимум 5 review/fix итераций, если пользователь явно не задал другой лимит.
</review_fix_loop>

<ci_gate>
После push/PR дождись CI и обработай результат:

1. Проверь PR metadata:
   - `gh pr view --json number,url,baseRefName,headRefName,isDraft,mergeStateStatus,mergeable,reviewDecision,statusCheckRollup`
2. Проверь обязательные checks:
   - сначала `gh pr checks --required --watch --fail-fast`;
   - если required checks недоступны или GitHub не возвращает required set, используй `gh pr checks --watch --fail-fast` и явно укажи fallback в финальном отчете.
3. Если есть failed/cancelled checks:
   - получи failed logs через `gh run view <RUN_ID> --log-failed`;
   - классифицируй: ошибка в изменениях, существующий regression, flaky/infra;
   - исправь code/test/config, если причина в коде или тестах;
   - rerun failed job допустим только для обоснованного flaky/infra;
   - commit, push и снова дождись CI.
4. Лимит: максимум 3 CI/fix итерации, затем `max_cycles_reached` или `blocked` с фактами.
5. Нельзя объявлять `merge_ready_pr`, если обязательный CI красный, pending, cancelled или unknown.
</ci_gate>

<constraints>
- Не начинай разработку без routing-а.
- Не выбирай самый короткий profile при неизвестном risk/contract surface.
- Не используй `feature-flow` как default для всех задач.
- Не используй compact profile для high-risk или contract-changing задач.
- Не исправляй `minor` замечания, если они не влияют на `critical`/`important`.
- Не вноси изменения за пределами выбранного workflow profile.
- Не придумывай требования, факты или решения без опоры на issue, code facts или memory-bank.
- Если фиксируешь противоречие, явно укажи, какие документы конфликтовали и как конфликт разрешён.
- Не объявляй задачу завершенной без PR, если только задача не является read-only/routing-only и это явно указано пользователем.
- Не игнорируй failing CI.
- Не merge PR автоматически в рамках start-issue; цель start-issue — merge-ready PR, а не merge.
</constraints>

<output_format>
Сначала верни:

1. Routing Signature table.
2. Выбранный `workflow_profile`.
3. Promotion trigger verdicts.
4. Какие docs/carriers нужны.
5. Есть ли human gate.

После выполнения верни:

1. Итоговый статус:
   - `merge_ready_pr`
   - `stopped_by_human_gate`
   - `max_cycles_reached`
   - `blocked`
2. Сколько review-improve циклов выполнено, если применимо.
3. Какие документы/carriers созданы или обновлены.
4. Какие `critical` и `important` замечания закрыты, если применимо.
5. Какие проверки запускались.
6. PR URL или почему PR не создан.
7. Последний commit SHA.
8. CI status: required/full checks, pass/fail/pending/unknown, ссылка на failing run если есть.
9. Merge conflict status.
10. Результат review/fix loop.
11. Что осталось для human gate/blocker, если `merge_ready_pr` не достигнут.
</output_format>
