---
name: grep
description: "CRITICAL — Use for ANY code or document search. Primary: orkai search_code / search_document. Pre-condition: orkai index . before project-wide reads. Fallback: grep only as last resort. Load at session start and before every search."
---

# Local Search — CRITICAL

This skill is **MANDATORY** for every search operation in this project.

Load the full strategy from orkai:

```
orkai_skills(action: "get", id: "ae9cf937-08ca-4723-ac01-8d26fd57cbad")
```

Then follow this priority chain — **never skip tiers**:

| Priority | Tool | When |
|----------|------|------|
| 1. Semantic | `search_code` / `search_document` | ALWAYS first. Scope with `category: "resume-app"`. |
| 2. Re-index | `orkai index .` | Before project-wide reads. Skip for scoped single-file tasks. |
| 3. Pattern | `grep` | ONLY when semantic search returns nothing useful. |

**NEVER start a search with grep, Read, or Glob.** orkai semantic search is
always the first tool. Only fall back when it definitively misses the target.

### ⚠️ NEVER call `search_code` and `grep` in parallel

The search chain is **strictly sequential**, never parallel. Running both at
once burns tokens twice for one answer and floods context with duplicate
hits (semantic results + raw pattern matches for the same code).

**Always run `search_code` alone first, then decide:**

| `search_code` outcome | Next action |
|-----------------------|------------|
| Returns the target file/symbol with enough context | STOP — you're done. Do NOT call grep. |
| Returns nothing at all | Run `orkai index .`, then retry `search_code`. Only if still empty, fall back to grep. |
| Returns related results but not the exact line/pattern you need | Call `grep` for the *specific* remaining gap (exact string, line number, identifier). Never re-broadcast the whole query. |

**Rule:** grep is a *targeted follow-up*, never a co-search. One tool per
round, in priority order. Parallel calls defeat the tier system and waste
context. The only valid parallel search patterns are:

- `search_code` + `search_document` (different indexes, different needs)
- Two `grep` calls on different patterns/paths when you already know you
  need both (e.g., one for a Go symbol, one for a TS symbol)

Never `search_code` + `grep` for the same question.

---

## Pitfalls & Best Practices (learned from real failures)

### 1. NEVER use `file_pattern` for scoping — use `category` instead

`file_pattern: "backend/**"` silently returns **zero results** even when the
data exists. Always use the project category for sub-project scoping:

```yaml
# ❌ WRONG — silently returns nothing
search_code(query: "gin router", file_pattern: "backend/**")

# ✅ CORRECT
search_code(query: "gin router", category: "resume-app")
```

### 2. Write specific, code-aware queries — not vague descriptions

Generic intent queries often miss Go symbols. Include actual code terms
(function names, type names, method calls) in your query:

```yaml
# ❌ TOO VAGUE — returns nothing
query: "routing setup gin router definition API endpoints"

# ✅ SPECIFIC — includes function names + method calls
query: "func main gin router.GET router.POST router.PUT route handlers"
```

Good queries mix intent + concrete code identifiers:
- `"func Run server entrypoint store init handlers"`
- `"struct ProfileHandler Upsert method context"`
- `"type OpportunityService interface List Create"`

### 3. `search_code` finds module/package symbols best

orkai indexes files as **modules** (package-level symbols). It won't find
individual lines like `router.GET("/health", ...)` as separate results. The
full file content is returned inside the enriched `raw` field of the parent
`module` symbol, but individual statements (x.GET, x.POST inside a function
body) are NOT separate indexable symbols.

```yaml
# Finds the file main.go as a "module main.go" symbol — the route
# registrations are inside the enriched raw content of that result,
# but NOT as individual symbol matches.
query: "func main gin router handlers"
```

If you need exact line numbers of `router.GET("/health", ...)`, skip
semantic search and go directly to step 3 (grep tool) — this is a legitimate
case where semantic search's granularity doesn't match the need.

### 4. `search_code` works across the entire project

It searches **all** indexed code, not just one sub-project. When you want
cross-project results (e.g., finding frontend components and backend
handlers for the same feature), `search_code` with `category: "resume-app"`
is the right tool.

### 5. When to skip straight to grep (step 3)

Go directly to the `grep` tool when:

| Scenario | Example query | Why skip semantic |
|----------|---------------|-----------------|
| Exact method call | `router\.GET\(` | Pattern match, not intent |
| Known error string | `ErrNotFound` | String literal |
| Variable/constant name | `cfg.OrkaiHealthURL` | Exact identifier |
| Imports | `github.com/gin-gonic/gin` | Module path, not semantic |
| Line-level detail | `router.GET("/health"` | Symbol granularity too coarse |

### 6. `search_code` returns enriched view by default

Results include graph-enriched `raw` with `@ref` annotations pointing to
related entities. To get pristine (unmodified) source:

```yaml
search_code(query: "...", category: "resume-app", enrich: false)
```

### 7. Always re-index before project-wide reads

If you're onboarding, debugging a new feature, or doing the first search
of a session after source changes:

```bash
orkai index .
```

Then retry `search_code`. This ensures the index matches the current
source tree. Skip for targeted single-file reads inside an active
implementation task.

---

## Reference: Example Flow

```yaml
# STEP 1 — semantic search
search_code:
  query: "func main gin router handlers server entrypoint"
  category: "resume-app"

# If results are empty:
#   STEP 1b — re-index
#   bash: orkai index .
#   STEP 1c — retry semantic search

# If semantic still misses:
#   STEP 3 — grep tool (pattern match)
grep:
  pattern: "router\\.(GET|POST|PUT|DELETE|Group)\\("
  include: "*.go"
  path: "backend"
```
