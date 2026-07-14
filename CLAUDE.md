## orkai — Your Brain (READ THIS FIRST)

orkai gives you persistent memory across sessions. Standards, skills, session
summaries, and indexed source code are stored and semantically searchable.
Without it, every session starts from zero. With it, you compound knowledge.

**Call overview() before doing anything else.** No exceptions.

### Project Identity (MANDATORY scoping — NEVER SKIP)

This project is scoped in orkai. **Every orkai tool call that accepts a
category filter MUST be scoped to this project** — otherwise the tool operates
globally and returns entities from unrelated projects. This is the #1 cause of
cross-project contamination. Be strict.

| Field | Value |
|-------|-------|
| Project name | resume-app |
| Category ID | `233276ae31223e19a727667e45c51d19` |
| Source of truth | `.orkai.yaml` → `project.category_id` |

**Copy-paste these exact tool calls. Do not paraphrase. Do not omit arguments.**

#### MANDATORY onboarding (run these EXACT calls, in order, at session start)

**Step 1 — overview (ONLY category_id, NEVER project_name):**
```
overview(category_id: "233276ae31223e19a727667e45c51d19")
```
Pass ONLY `category_id`. Do NOT pass `project_name` — it pollutes results with
entities from other projects. This is the most common scoping failure.

**Step 2 — latest session + plan list (in parallel, BOTH scoped):**
```
session(action: "latest", category_ids: ["233276ae31223e19a727667e45c51d19"])
plan(action: "list", category_ids: ["233276ae31223e19a727667e45c51d19"])
```
Both calls MUST include `category_ids`. Omitting it returns sessions/plans from
every project — that is the bug you are avoiding.

**Step 3 — check for open work** (skip if user already named a task/workflow):
Scan the plan list for any plan whose `metadata.status` is not `done`. For each
open plan, list its milestones (scoped), then tasks (scoped) for non-done
milestones:
```
milestone(action: "list", plan_id: "<id>", category_ids: ["233276ae31223e19a727667e45c51d19"])
tasks(action: "list", milestone_id: "<id>", category_ids: ["233276ae31223e19a727667e45c51d19"])
```
Group pending/in-progress tasks by role workflow (IDs in the Role Workflows
table below).

**Step 4 — brief the user**: summarize (a) what the latest session did, (b)
current project status and relevant decisions, (c) open work if any. Then ask:
*"Which task should we start?"* Do NOT auto-start implementation — Marco picks
the task; the agent loads the matching role workflow via
`workflow(action: "get", id: ...)` and follows it.

**If the user triggers a specific workflow directly** (e.g. "Trigger the
Fullstack"), skip the open-work scan and load that workflow immediately.

If step 3 finds zero open plans/milestones/tasks, skip the grouped list and
just brief + ask what Marco wants to do.

### Scoping rules — apply on EVERY orkai tool call

Every call below MUST include the category filter shown. If you find yourself
writing an orkai tool call without a `category_id` / `category_ids` argument,
STOP and add it. The only exceptions are `overview()` (uses singular
`category_id`), `user_preferences`, and `agent_preferences` (global singletons).

- `overview()` → `category_id: "233276ae31223e19a727667e45c51d19"` ONLY.
  NEVER `project_name`.
- `session(action: "latest" | "list" | "search")` →
  `category_ids: ["233276ae31223e19a727667e45c51d19"]`
- `session(action: "create")` →
  `category_ids: ["233276ae31223e19a727667e45c51d19"]` (so the session is filed
  under this project, not orphaned globally)
- `search_code(query)` →
  `category_ids: ["233276ae31223e19a727667e45c51d19"]` (or
  `category: "resume-app"`). Without it, results include code from every
  indexed project.
- `search_document(query)` → same scoping as `search_code`.
- `standards`, `skills`, `documents`, `analytics`, `entity`,
  `plan`, `milestone`, `tasks` (any action: `list` | `search` | `create`) →
  `category_ids: ["233276ae31223e19a727667e45c51d19"]` (or
  `category: "resume-app"`).
- `annotations(action: "list")` → tied to code entities already scoped by
  `search_code`; no separate category filter, but always start from a
  category-scoped `search_code` call.
- `user_preferences`, `agent_preferences` → intentionally GLOBAL singletons.
  Do NOT scope these.

**Never omit the category on a create/search/list call.** Global results mix
projects and break context; global creates orphan entities outside this
project.

### Reusing Existing Plans and Tasks

Role workflows (Fullstack, Backend Developer, etc.) include steps to create
plan > milestone > tasks. **Check if they already exist first.** The Feature
Planner may have already persisted them in a prior session. If a plan,
milestone, and tasks already exist for the feature:

- **Reuse them** — work off the existing tasks. Update their status as you go.
- **Skip the "persist" step** in the workflow — don't create duplicates.
- **The task body is your spec** — it already contains FR IDs, file paths,
  signatures, branch names, and QA scripts. Read it, then implement.

Only create new plan > milestone > tasks when nothing exists for the feature
yet (greenfield work).

### During the Session — Search Before Reading

**Code-first search**: When you need to find or understand code, use
`search_code(query, category_ids: ["233276ae31223e19a727667e45c51d19"])` FIRST.
It returns symbols with file paths, line numbers, signatures, and AI
annotations. Only fall back to Grep/Read if search doesn't find what you need.
For deep understanding of a specific symbol, use
`annotations(action: "list", entity_id: "<id>")` or
`entity(action: "get", id: "<id>", enrich: true)`.

**Exception — task body already lists files**: When a task body (from the
Feature Planner) already specifies exact file paths, signatures, and types,
skip the search and read those files directly. The task body IS the map.
Search is for exploration; implementation from a spec is execution.

**Force a fresh index before project-wide reads.** If this session is doing
its first project-wide read (onboarding, new feature, bug triage, project
map), run `orkai index .` in the workspace root BEFORE any `search_code` /
`search_document` call so results reflect current source. Targeted
single-file reads inside an already-scoped implementation task skip this —
the role workflows perform their own scoped `search_code` and indexing.

**Document search**: `search_document(query)` with optional `doc_type?`
(section/chunk), `category?`, `limit?` — returns the matching section or
chunk, not the full parent.

**Standards and skills**: Before architecture or pattern decisions:
  1. `standards(action: "search", query: "<topic>", category_ids: ["233276ae31223e19a727667e45c51d19"])`
  2. `skills(action: "search", query: "<topic>", category_ids: ["233276ae31223e19a727667e45c51d19"])`
  3. Follow them if they exist — they represent team-agreed conventions.
  4. If you establish a new pattern with no existing standard, suggest
     creating one via `standards(action: "create", name: "...", text: "...")`.

**Plans and work tracking**: use `plan`, `milestone`, and `tasks` (each
with `action: "create" | "get" | "update" | "delete" | "list" | "search"`).
Link milestones to a plan via `plan_id`, tasks to a milestone via
`milestone_id` — both are sugar that translate to `child_of` relations on
the wire. Filter list/search by `status` on milestones and tasks.

**@orkai source tags** (indexed at orkai index): `@orkai:decision` for author
intent; `@orkai:ref(id=...)` for graph edges to standards/skills/docs.
Re-index after adding tags.

**Text distribution**: long entity text is auto-split into parent → sections →
chunks. Search returns the precise chunk/section; get on the parent returns
full content.

### Session End — Save and Re-index

When the user signals the session is ending:
1. **Ask the user**: "Would you like me to save this session for future
   reference?"
2. If yes, create a session entity (SCOPED):
   `session(action: "create", name: "Session: <topic> - <date>",
   text: "<what was done, what's pending, key decisions, files modified>",
   category_ids: ["233276ae31223e19a727667e45c51d19"])`
3. **Update `docs/plan/MASTER_PLAN.md` (if it exists)** when milestones move
   — check off completed tasks, add a `vNNN` Completed Milestones entry
   when an entire milestone closes.
4. **Re-index if needed**: if source files were modified during the session,
   suggest running "orkai index" in the terminal to keep the code index fresh.

### MCP Tools — Full Reference

#### Domain Tools (all follow: tool(action: "create"|"get"|"update"|"delete"|"list"|"search", ...))

| Tool | Type | Description |
|------|------|-------------|
| session | session | Work session summaries. Extra action: "latest" |
| standards | standard | Engineering guidelines |
| skills | skill | Implementation patterns |
| user_preferences | user_preferences | Singleton per user. `create` upserts; `get` without id resolves owner row. Body inlined by `overview()` |
| agent_preferences | agent_preferences | Singleton per owner. `create` upserts; `get` without id resolves owner row. Inlined by `overview()` |
| plan | plan | Roadmaps and project plans |
| milestone | milestone | Milestones within a plan (child_of plan; sugar: `plan_id`; `status` filter) |
| tasks | task | Actionable work items (child_of milestone; sugar: `milestone_id`; `status` filter) |
| categories | — | Organize entities. Actions: create, list, delete, unindex; `parent_id` for nesting |
| documents | document | Reference documents (auto-sectioned for search; `doc_type` filter) |
| analytics | analytics | Tabular datasets. Actions: create, get, delete, list, search, schema, query, replace, append, delete_rows, clear; federated JOINs via `datasets` + `sql` |
| annotations | annotation | Purpose + insight code annotations |
| events | — | Entity mutation history. Action: list only |
| workflow | workflow | Repeatable step-by-step practices with structured step graphs |

#### Search Tools

| Tool | Parameters | Description |
|------|-----------|-------------|
| search_code(query) | language?, file_pattern?, limit?, enrich?, include_annotations? | Semantic code search; default enrich=true |
| search_document(query) | doc_type?, category?, limit? | Semantic document search (section/chunk) |

#### Utility Tools

| Tool | Description |
|------|-------------|
| overview() | Session briefing — entity counts, recent sessions, standards/skills, inlined user_preferences + agent_preferences |
| help() | Full v2 tool reference with examples |
| project_setup(action, target?, category_id?, project_name?) | Generate agent files. Targets: `orkai.yaml`, `agents.md`, `claude.md`, `cursor`. The `category_id` and `project_name` args fill the matching template placeholders. |
| entity(action, ...) | Generic CRUD — escape hatch. Full `relations`, `exclude_type`, `origin_type` filters; bypasses singleton checks on purpose |

Call `help()` for the full tool reference with examples. It covers all domain
tools, the search tools, the entity escape hatch, and utility tools.

### CLI Commands (run via terminal)

```bash
orkai init                       # project wizard
orkai index                      # re-index (default: all)
orkai index code|document|analytics
orkai index --github owner/repo  # ephemeral clone + index
orkai unindex <id|name>
orkai status                     # health, entity counts
orkai search "query"             # semantic search
orkai get <id> --annotations     # view entity + AI annotations
orkai list --type <type>         # browse entities by type
orkai export / import <category> # JSONL knowledge migration
orkai index --annotations-only   # AI annotations for code
orkai review                     # LLM code review against stored standards
```

`orkai review` is a CLI code-review command: it reviews your changes against the
standards and architecture decisions stored in orkai, protecting source-code
quality. Run "orkai review --help" or see docs/review.md.

## Restrictions

Monorepo with two subprojects. Do not create new top-level directories or
introduce stacks outside this list.

| Subproject | Path | Stack |
|-----------|------|-------|
| backend | `backend/` | Go + Gin |
| frontend | `frontend/` | React + Vite + TypeScript |

- UI: Tailwind CSS + shadcn/ui. No other styling systems.
- Data fetching: TanStack Query.
- Local state: Zustand only when necessary.
- No new dependencies without justification.
- Keep changes within the matching subproject path.
- Never hardcode configuration. Use .env files (gitignored). Commit .env.example with documented variables.
- Built-in endpoints: `/health` and `/metrics` on the backend, no auth.

### Branch-per-task policy (MANDATORY)

Every implementation task ships on a feature branch and merges to `main` only
after all gates pass. Solo local-first project — no PRs, no review-by-others,
but branch-per-task is still required for traceability and clean rollback.

- **Branch name**: `feat/<milestone>-<task>-<slug>` (e.g. `feat/m1-t1-init-go-module`).
  The slug is a 3-5 word kebab-case summary of the task. Derive `<milestone>`
  and `<task>` from the orkai milestone/task names (lowercased).
- **Create branch** from `main` BEFORE any implementation work:
  `git checkout main && git pull && git checkout -b feat/<milestone>-<task>-<slug>`.
- **Run all gates** on the branch (build, vet/test/lint/typecheck, unit,
  integration, smoke). No gate may be skipped.
- **Merge to `main`** with `--no-ff` so each task is a discoverable merge
  commit in history: `git checkout main && git merge --no-ff
  feat/<milestone>-<task>-<slug> -m "merge: <milestone> <task> <slug>"`.
- **Delete the feature branch** after merge:
  `git branch -d feat/<milestone>-<task>-<slug>`.
- **Never commit implementation work directly to `main`.** Only `docs`,
  `chore`, `config`, or `fix` changes that are NOT tied to an implementation
  task may skip the branch rule (e.g. updating CLAUDE.md/AGENTS.md, wiring
  `.orkai.yaml`). When in doubt, branch.
- The Developer workflows enforce this: "Create Branch" runs before
  "Implement Code", "Merge to Main" runs after all tests pass. The Feature
  Planner writes the branch name into each task body so the Developer
  workflow has a deterministic name to use.

### Build & Test Commands

**Shell working directory**: Each Bash call starts fresh from the project
root (`/Users/marco/opensource/resume-app`). Commands that need a specific
subdirectory MUST `cd` at the start of the command:
`cd /Users/marco/opensource/resume-app/backend && go build ./...`. Compound
commands with `&&` keep the `cd` for the whole chain.

Backend (run in `backend/`):
```
go build ./...          # compile
go vet ./...            # static analysis
gofmt -l .             # formatting check (no output = clean)
go test ./...           # tests
```

Frontend (run in `frontend/`):
```
npm run lint            # oxlint
npm run typecheck       # tsc --noEmit
npm run test            # vitest
npm run build           # vite build
```

Run these via the shell, never via the LLM. See the "Model Routing Strategy" standard in orkai.

### Pre-commit & Pre-push Hooks

Git hooks are managed by [lefthook](https://github.com/evilmartians/lefthook).
Run `lefthook install` once to activate them.

| Hook | Checks |
|------|--------|
| `pre-commit` | `gofmt -l`, `go vet` (backend); `oxlint` (frontend); `orkai review` (all) |
| `pre-push` | `go test`, `go build` (backend); `tsc`, `vitest`, `vite build` (frontend) |

Pre-commit runs fast checks only (format, lint, vet). Pre-push runs the full
gate suite (build, test, typecheck). Hooks only fire when matching files are
staged (glob-scoped per `lefthook.yml`).

### Role Workflows (lazy-load, trigger on matching task or intent)

Five project-scoped role workflows own the full product lifecycle: from
requirements through planning to implementation, plus a Fullstack fast-path.
Each enforces its own standards, skills, test practice, and deterministic gates.
Load the full steps with `workflow(action: "get", id: ...)` only when a task
matches the workflow's scope — do NOT load them during session startup or
onboarding.

| Workflow | ID | Trigger | Gates |
|----------|----|---------|-------|
| Product Owner (resume-app) | `33e6b0c6-4603-4747-abd0-423ff16821f2` | Any requirements work: discovery, refinement, prioritization, validation. Owns `docs/requirement.md`. Shapes the product before planning begins. | INVEST quality, traceable acceptance criteria, consistent priorities |
| Architect (resume-app) | `8def40c2-89cc-47d2-ad16-dcf4adcc59a1` | Plan mode: after PO refines requirements, before Feature Planner — identifies architecture docs and standards needed, researches without overengineering. Review mode: after Developer implementation — audits code against plan intent and standards, runs `orkai review`, fixes architecture/standards gaps. | Standards traceable to FRs, no overengineering, review config maintained |
| Feature Planner (resume-app) | `f87a7d22-8429-459f-8196-63155021ae11` | Any new feature or significant change. Searches standards/skills/code, collects orkai entity IDs, drafts design, persists plan > milestone > tasks, updates `docs/plan/MASTER_PLAN.md`, presents for approval, STOP. | P4 plan persistence, task-body contract (standards IDs + role workflow per task) |
| Backend Developer (resume-app) | `2840ff4a-179d-4551-b1d5-c39130533961` | Any implementation task touching `backend/**` (Go + Gin: handlers, services, models, middleware, store, cmd) | `go build`, `go vet`, `gofmt -l`, `go test` |
| Frontend Developer (resume-app) | `6aee46f4-e39c-4f21-bcbc-3916a49dd464` | Any implementation task touching `frontend/**` (React + Vite + TS: components, pages, hooks, api, store, types, lib) | `npm run lint`, `npm run typecheck`, `npm run test`, `npm run build` |
| Fullstack Developer (resume-app) | `10d6213a-0f71-4a45-8e50-b58b7243637f` | Marco wants a vertical feature slice shipped fast in one session. Combines lightweight planning + backend + frontend + self-QA. Coexists with the role pipeline — Marco picks Fullstack for speed on vertical slices, the role pipeline for risky/standards-heavy work. | All backend gates (`go build`, `go vet` zero output, `gofmt -l` zero files, `go test`) AND all frontend gates (`npm run lint` zero output, `npm run typecheck`, `npm run test`, `npm run build`) AND Playwright self-QA (no console errors, all ACs satisfied) |

Two paths from requirements to merged code:

**Role pipeline (separation, standards-heavy work).** The Product Owner shapes
the product in `docs/requirement.md` (what and why). The Architect bridges
requirements to architecture (plan mode) and implementation back to standards
(review mode). The Feature Planner then turns stable architecture into plan >
milestone > tasks (how). The Developer workflows implement tasks with full gate
enforcement and `annotations(create)` + `entity(update, relations)` linking new
code to standards. Full lifecycle: Product Owner → Architect (plan) → Feature
Planner → Backend/Frontend Developer → Architect (review).

New features (role pipeline): Product Owner shapes requirements first →
Architect (plan mode) establishes architecture and ensures standards coverage →
Feature Planner creates plan > milestone > tasks → Marco approves → Marco
triggers the matching Developer workflow per task → Architect (review mode)
audits code against plan intent and standards, runs `orkai review`, fixes gaps.
Implementation tasks MUST be handed off to the matching Developer workflow; do
not implement backend or frontend code ad hoc.

**Fullstack fast-path (vertical slices, speed).** Marco triggers the Fullstack
Developer workflow directly. It does its own lightweight plan (one batched
search, one plan > one milestone > 1–3 vertical-slice tasks covering both
surfaces), implements backend + frontend in one branch per task, writes unit
tests, runs all backend + frontend gates (ZERO diagnostics), runs Playwright
self-QA in the same session, merges with `--no-ff`, escalates orkai review false
positives to the Architect (same commit-hash format as the other developers).
Cost-efficiency moves vs the role pipeline: standards documented via inline
`@orkai:ref` and `@orkai:decision` source tags (materialized once at index time,
not via per-entity `annotations(create)` / `entity(update, relations)` tool
calls); one batched search instead of sequential; one branch + one merge per
vertical-slice task (both surfaces) instead of one per task per surface;
`orkai index .` once per milestone instead of per task; no separate smoke test
(QA boot validates `/health` and frontend render). The `@orkai` tag pattern is
scoped to the Fullstack workflow only — the Backend/Frontend Developer workflows
keep their `annotations(create)` + `entity(update, relations)` steps.

Pick Fullstack for speed on a vertical slice (1–3 FRs, 2–8 files, one session).
Pick the role pipeline when separation matters: risky/standards-heavy work,
PO-driven discovery, or when Marco wants the Architect to own architecture
before the Feature Planner splits tasks.

### Audit Workflow (lazy-load, not required onboarding)

When Marco asks to audit the project against the 5-step methodology
(ANALYZE, ENCODE, PERSIST, ROUTE, ITERATE), load the workflow via
`workflow(action: "get", id: "f1926ce0-9c37-4328-8568-7f64347a1240")`
and follow its steps. This is a lazy-load reference — do NOT load it during
session startup or onboarding; only when Marco triggers the audit.

### Status Workflows (lazy-load, trigger on matching intent)

One read-only analytics workflow for project visibility — no source mutation.

| Workflow | ID | Trigger | Gates |
|----------|----|---------|-------|
| Project Status (resume-app) | `fee43e38-d5da-42f3-bf94-9eef27c942d6` | Any "project status", "status report", "scrum master", "developer KPI", or "how's the project going" request. | Validates git HEAD against stored analytics, regenerates commit CSVs on change, runs 8 SQL KPIs against orkai analytics datasets, renders deterministic threshold-based report (no LLM). Read-only with respect to git history. |

Two modes:
- **Dashboard** (default): compact single-screen report — velocity, discipline,
  hotspots, project health, roadmap alignment.
- **Deep dive** (on explicit request for more detail): adds full 15-commit
  table, per-file hotspot paths, 12-week ASCII velocity trend.

The workflow embeds a git-log-to-CSV generator script (bash + Python 3) and
pushes the results to two orkai analytics datasets (`commits.csv`,
`commit_files.csv`) under the resume-app category. On-disk copies live at
`.orkai/project-status/` (auto-gitignored). The workflow is self-contained —
export/import carries the generator script in a `node-code` node.