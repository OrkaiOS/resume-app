# Master Plan — Resume App

> Source of truth: orkai plan/milestone/task entities. This file is a mirror —
> if it diverges from orkai, orkai wins. Run the Project Status workflow to
> regenerate from orkai.

## Active Plans

### Glassmorphism Design System (`39d50ea9-45d3-45cc-8964-6261fc3c8c4d`)

| # | Milestone | ID | Tasks |
|---|-----------|-----|-------|
| M1 | Glassmorphism Foundation | `59821748-6525-4793-9b1f-c235894fb1d8` | T1: Dual CSS token system (`c99e9610`) · T2: Style switcher + Zustand store (`544f0dc5`) · T3: GlassCanvas + page integration (`3756b25c`) |

---

## Completed Plans

### Bug fix — opportunity dialog layout (`4baa8cc6-c55a-4e54-9d46-8da072470429`) — done

| # | Milestone | ID | Tasks |
|---|-----------|-----|-------|
| M1 | Dialog layout fixes | `26015b17-17cf-46d8-b18d-a2896c51acde` | T1: Constrain textarea growth and widen dialog (`86995cd0`) ✅ · T2: Add colored accent bar to opportunity cards (`21ec0ef9`) ✅ |

### make dev — configurable ports (`08755da6-4892-4054-9682-5ca86ab0f594`) — done

| # | Milestone | ID | Tasks |
|---|-----------|-----|-------|
| M1 | Configurable ports for make dev | `a3c35a69-4a49-4b33-b93a-80846674ebcd` | T1: vite.config.ts server.port (`3ab50858`) ✅ · T2: Makefile dev target (`6121b456`) ✅ · T3: frontend .env cleanup (`26a9eaee`) ✅ · T4: backend .env.example CORS docs (`e1b15a08`) ✅ · T5: QA workflow port validation (`0c5abc4a`) ✅ |

### UX Improvement — Opportunity form as Dialog modal (`b1a389e9-6d8a-4bb8-9ee2-48c5da9c9c4b`) — done

| # | Milestone | ID | Tasks |
|---|-----------|-----|-------|
| M1 | Modal form migration | `1e0cefc9-f1e0-4e1a-905d-1948a955134f` | T1: Replace inline form with Dialog modal (`4ed0d193`) ✅ · QA: m1 t1 — browser test (`5ce1a1ec`) ✅ · Bug: No error state when backend is down (`3f70199c`) ✅ |

### Chat Improvements — Reasoning, Session Memory, Stop/Continue (`38104bf3-34c0-4877-b33b-fa893ea1b222`) — done

| # | Milestone | ID | Tasks |
|---|-----------|-----|-------|
| M1 | Chat Improvements | `aa94c14e-4834-462a-ac45-4c767ce952e9` | T1: Reasoning Stream + Tool Badges (`805bb79a`) ✅ · T2: Session Memory Tools + User Insights + Overview (`6e9fd358`) ✅ · T3: Stop & Continue with Session Checkpoint (`1cb03532`) ✅ |

### Phase 1 — Foundation (`5a9ad6ea-fca4-4933-966c-f701ad61f6ad`) — done

| # | Milestone | ID |
|---|-----------|-----|
| M1 | Embedded Binary + Build System | `1f78efdd-342e-4375-b2e2-7fe05c9dc389` |
| M2 | Orkai Health Gate | `5e24e1c6-e09b-47c2-99fb-333171e13cd4` |
| M3 | Global Install + Dev Mode | `b66db674-4aa8-4075-838c-00fe5257f8ea` |

### Phase 2 — Data Model + Onboarding (`08de9e43-a822-460b-92cf-c291a376ecb5`) — done

| # | Milestone | ID |
|---|-----------|-----|
| M4 | Data Layer + CRUD API | `6edd5154-86bd-4440-968d-4079b7b0ac61` |
| M5 | Onboarding API | `7a89a6ab-5cd3-416a-b58e-28e6cc985b4e` |
| M6 | Frontend Foundation + Onboarding UI | `631c2534-cf78-4f99-ae8f-a1b9ff2bd637` |

### Standalone Completed Plans

| Plan | ID |
|------|-----|
| Scaffold resume-app subprojects | `23629f7a-49b9-4d8e-8249-b4b5533ae7ac` |
| FR-001 — Browser auto-open on startup | `2366da26-544d-4c67-9a55-dcc4181b2a8f` |
| FR-005 — Go backend live reload via air | `334a1dcb-371f-48ea-ade0-8f7315b16913` |
| FR-020 + FR-021 — Home Page + Empty State | `a142cb9a-5bc4-4cd9-b863-d8d57d33b352` |
| FR-030 — Chat Interface (assistant-ui) | `b8e96556-efd7-41f5-bf14-5e48c85ef8d8` |
| FR-031 — Agent System Prompt | `e12ce11d-12ba-41d5-ab3d-bdeb2f8d55bb` |
| FR-032 — Agent Tools (shell, profile, orkai-search) | `9aa21fff-2de0-4100-9182-e368e06ddadb` |
| Make opportunity description required + paste cleanup | `33b0a100-23a3-4f90-95be-c3d0b9aaa77c` |
| Design System Review | `247dae98-933f-481f-a3d9-38b4bd2ea6e2` |
| make dev — configurable ports | `08755da6-4892-4054-9682-5ca86ab0f594` |