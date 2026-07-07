## orkai — Persistent Memory

orkai gives you persistent memory across sessions. Standards, skills, session
summaries, and indexed source code are stored and semantically searchable.
Without it, every session starts from zero. With it, you compound knowledge.

**Call overview() before doing anything else.**

### Project Identity (MANDATORY scoping)

This project is scoped in orkai. **Every orkai tool call that accepts a
category filter MUST be scoped to this project** — otherwise the tool operates
globally and returns entities from unrelated projects.

| Field | Value |
|-------|-------|
| Project name | resume-app |
| Category ID | `233276ae31223e19a727667e45c51d19` |

**Scoping rules — apply on EVERY call:**

- `overview()` → pass `category_id: "233276ae31223e19a727667e45c51d19"` AND
  `project_name: "resume-app"`. This scopes recent sessions and counts to this
  project.
- `session(action: "latest" | "list" | "search")` → pass
  `category_ids: ["233276ae31223e19a727667e45c51d19"]`.
- `session(action: "create")` → pass
  `category_ids: ["233276ae31223e19a727667e45c51d19"]` so the session is filed
  under this project.
- `search_code(query)` → pass
  `category_ids: ["233276ae31223e19a727667e45c51d19"]` (or
  `category: "resume-app"`). Without it, results include code from every
  indexed project.
- `search_document(query)` → same scoping as search_code.
- `standards(action: "search" | "list")`, `skills(action: "search" | "list")`,
  `documents(action: "search" | "list")`, `analytics(action: "search" | "list")`,
  `entity(action: "list" | "search")`, `plan/milestone/tasks(action: "list" |
  "search")` → pass `category_ids: ["233276ae31223e19a727667e45c51d19"]` (or
  `category: "resume-app"`) to restrict results to this project.
- `standards(action: "create")`, `skills(action: "create")`,
  `documents(action: "create")`, `analytics(action: "create")`,
  `plan/milestone/tasks(action: "create")`, `entity(action: "create")` → pass
  `category_ids: ["233276ae31223e19a727667e45c51d19"]` so the new entity is
  filed under this project and not orphaned globally.
- `annotations(action: "list")` → annotations are tied to code entities already
  scoped by search_code; no separate category filter, but always start from a
  category-scoped search_code call.

**Never omit the category on a create/search/list call.** Global results mix
projects and break context; global creates orphan entities outside this
project. `user_preferences` and `agent_preferences` are singletons and are
intentionally global — do NOT scope those.

### Session Startup

1. **Call overview()** with `category_id: "233276ae31223e19a727667e45c51d19"`
   and `project_name: "resume-app"` — returns recent sessions, standards,
   skills, and entity counts scoped to this project.
2. **Load the latest session**: `session(action: "latest",
   category_ids: ["233276ae31223e19a727667e45c51d19"])` — or
   `session(action: "get", id: "<session_id>")` from the overview preview.
3. **Check for open work** — list every plan > milestone > task in orkai that is
   not fully done, and group the pending/in-progress tasks by their assigned
   role workflow so Marco can pick up where he left off:
   - `plan(action: "list", category_ids: ["233276ae31223e19a727667e45c51d19"])`
   - For each plan: `milestone(action: "list", plan_id: <id>, category_ids: [...])`
   - For each milestone whose status ≠ done: `tasks(action: "list",
     milestone_id: <id>, category_ids: [...])` and collect tasks whose
     status is `pending` or `in_progress`.
- Group the collected tasks under one of these headings by which role
      workflow the task body names (or by path: `backend/**` → Backend
      Developer, `frontend/**` → Frontend Developer):
      - **Backend Developer (resume-app)** `2840ff4a-179d-4551-b1d5-c39130533961`
      - **Frontend Developer (resume-app)** `6aee46f4-e39c-4f21-bcbc-3916a49dd464`
      - **Feature Planner (resume-app)** `f87a7d22-8429-459f-8196-63155021ae11`
        (tasks here are "draft the plan" work, not implementation)
      - **Architect (resume-app)** `8def40c2-89cc-47d2-ad16-dcf4adcc59a1`
        (plan mode: before Feature Planner; review mode: after Developer)
    - Tasks with no role-workflow assignment go under "Unassigned".
4. **Brief the user**: summarize (a) what the latest session did, (b) current
   project status and relevant decisions, and (c) the grouped open work from
   step 3. Then ask: *"Which task should we start? Backend Developer / Frontend
   Developer / Product Owner / something else?"* If there is no open work,
   ask what Marco wants to do. Do NOT auto-start implementation — Marco
   picks the task; the agent loads the matching role workflow via
   `workflow(action: "get", id: ...)` and follows it.

If step 3 finds zero open plans/milestones/tasks, skip the grouped list and
just brief + ask what Marco wants to do.

### During the Session

**Search before reading files.** Use
`search_code(query, category_ids: ["233276ae31223e19a727667e45c51d19"])` to
find code by meaning. It returns symbols with file paths, line numbers,
signatures, and AI annotations. Only fall back to Grep/Read if search doesn't
find what you need.

**Force a fresh index before project-wide reads.** If this session is doing
its first project-wide read (onboarding, new feature, bug triage, project
map), run `orkai index .` in the workspace root BEFORE any `search_code` /
`search_document` call so results reflect current source. Targeted
single-file reads inside an already-scoped implementation task skip this —
the role workflows perform their own scoped `search_code` and indexing.

**Check standards before deciding.** Before architecture or pattern decisions:
  standards(action: "search", query: "<topic>", category_ids: ["233276ae31223e19a727667e45c51d19"])
  skills(action: "search", query: "<topic>", category_ids: ["233276ae31223e19a727667e45c51d19"])
Follow them if they exist — they represent team-agreed conventions.

### Session End

When the user signals the session is ending:
1. Ask: "Would you like me to save this session?"
2. If yes: `session(action: "create", name: "Session: <topic> - <date>",
   text: "<what was done, what's pending, key decisions, files modified>",
   category_ids: ["233276ae31223e19a727667e45c51d19"])`
3. Suggest running "orkai index" if source files were modified.

### Tool Reference

Call `help()` for the full tool reference with examples. It covers all domain
tools (session, standards, skills, preferences, plan/milestone/tasks,
categories, documents, analytics, annotations, events, workflow), the search
tools (search_code, search_document), the entity escape hatch, and utility
tools (overview, help, project_setup).

### CLI Commands

```
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

orkai review is a CLI code-review command: it reviews your changes against the
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
  task may skip the branch rule (e.g. updating AGENTS.md, wiring
  `.orkai.yaml`). When in doubt, branch.
- The Developer workflows enforce this: "Create Branch" runs before
  "Implement Code", "Merge to Main" runs after all tests pass. The Feature
  Planner writes the branch name into each task body so the Developer
  workflow has a deterministic name to use.

### Build & Test Commands

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

Four project-scoped role workflows own the full product lifecycle: from
requirements through planning to implementation. Each enforces its own
standards, skills, test practice, and deterministic gates. Load the full steps
with `workflow(action: "get", id: ...)` only when a task matches the
workflow's scope — do NOT load them during session startup or onboarding.

| Workflow | ID | Trigger | Gates |
|----------|----|---------|-------|
| Product Owner (resume-app) | `33e6b0c6-4603-4747-abd0-423ff16821f2` | Any requirements work: discovery, refinement, prioritization, validation. Owns `docs/requirement.md`. Shapes the product before planning begins. | INVEST quality, traceable acceptance criteria, consistent priorities |
| Architect (resume-app) | `8def40c2-89cc-47d2-ad16-dcf4adcc59a1` | Plan mode: after PO refines requirements, before Feature Planner — identifies architecture docs and standards needed, researches without overengineering. Review mode: after Developer implementation — audits code against plan intent and standards, runs `orkai review`, fixes architecture/standards gaps. | Standards traceable to FRs, no overengineering, review config maintained |
| Feature Planner (resume-app) | `f87a7d22-8429-459f-8196-63155021ae11` | Any new feature or significant change. Searches standards/skills/code, collects orkai entity IDs, drafts design, persists plan > milestone > tasks, updates `docs/plan/MASTER_PLAN.md`, presents for approval, STOP. | P4 plan persistence, task-body contract (standards IDs + role workflow per task) |
| Backend Developer (resume-app) | `2840ff4a-179d-4551-b1d5-c39130533961` | Any implementation task touching `backend/**` (Go + Gin: handlers, services, models, middleware, store, cmd) | `go build`, `go vet`, `gofmt -l`, `go test` |
| Frontend Developer (resume-app) | `6aee46f4-e39c-4f21-bcbc-3916a49dd464` | Any implementation task touching `frontend/**` (React + Vite + TS: components, pages, hooks, api, store, types, lib) | `npm run lint`, `npm run typecheck`, `npm run test`, `npm run build` |
| Frontend QA (resume-app) | `0fc856e1-9fe3-446b-a2de-6ad6319e08e5` | QA task created by Frontend Developer after implementation that introduces or modifies UI. Receives a precise test script (URLs, elements, interactions, assertions) and executes Playwright browser tests against the running dev environment. | Zero console errors, all elements render, all interactions produce expected outcomes, all ACs satisfied |

The Product Owner shapes the product in `docs/requirement.md` (what and why).
The Architect bridges requirements to architecture (plan mode) and implementation
back to standards (review mode). The Feature Planner then turns stable
architecture into plan > milestone > tasks (how). The Developer workflows
implement tasks with full gate enforcement and annotate+tag linking new code to
standards. Full lifecycle: Product Owner → Architect (plan) → Feature Planner →
Backend/Frontend Developer → Frontend QA (when UI changed) → Architect (review).

New features: Product Owner shapes requirements first → Architect (plan mode)
establishes architecture and ensures standards coverage → Feature Planner creates
plan > milestone > tasks → Marco approves → Marco triggers the matching
Developer workflow per task → Architect (review mode) audits code against plan
intent and standards, runs `orkai review`, fixes gaps. Implementation tasks MUST
be handed off to the matching Developer workflow; do not implement backend or
frontend code ad hoc.

### Audit Workflow (lazy-load, not required onboarding)

When Marco asks to audit the project against the 5-step methodology
(ANALYZE, ENCODE, PERSIST, ROUTE, ITERATE), load the workflow via
`workflow(action: "get", id: "f1926ce0-9c37-4328-8568-7f64347a1240")`
and follow its steps. This is a lazy-load reference — do NOT load it during
session startup or onboarding; only when Marco triggers the audit.