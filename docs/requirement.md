# Software Requirements Specification (SRS) — Resume App

## 1. Introduction

### 1.1 Purpose

Resume App is an agentic, chat-first resume builder powered by LLMs and the
orkai MCP platform. Users configure their professional profile once, then chat
with an AI agent to generate tailored resumes and cover letters per job
opportunity. The agent has access to orkai standards, skills, and tools —
including shell execution and artifact reuse — to produce polished PDF output.

This SRS defines the complete product requirements in the IEEE 830 / ISO 29148
style. Every requirement is uniquely identified, testable, prioritized, and
written as a user story with acceptance criteria.

### 1.2 Scope

Resume App covers the full lifecycle: one-time onboarding, opportunity
management, agentic chat-based document generation, structured review and
approval, and PDF export. The app is a single-user, local-first desktop web
application distributed as a single embedded Go binary. It depends on orkai
as a mandatory runtime service for standards, skills, and MCP tool access.
Multi-user, authentication, and cloud sync are explicitly out of scope.

### 1.3 Definitions

| Term | Definition |
|------|------------|
| Resume | A structured, formatted document presenting a user's professional background, tailored to a specific job opportunity |
| Cover Letter | A narrative letter complementing the resume, also tailored to the opportunity |
| Opportunity | A specific job the user is applying for: company, role, and creation date |
| Profile | The user's comprehensive background data: personal info, work history, education, skills |
| Agent | The LLM-powered assistant that generates and revises resumes/cover letters via chat |
| Artifact | A reusable Python or bash script, created by the agent, tested, and marked as reusable after user approval |
| orkai | The MCP platform (sold separately) that provides persistent standards, skills, and MCP tool connectivity |
| MCP | Model Context Protocol — the standard by which the agent connects to orkai for tools and context |

### 1.4 References

- orkai standards: Frontend React + Vite Conventions, Backend Go + Gin Conventions, API Contract Standard, Model Routing Strategy
- AGENTS.md — project restrictions, build commands, branch policy, role workflows

---

## 2. General Description

### 2.1 Product Perspective

Resume App is a standalone web application distributed as a single Go binary
that embeds the React frontend, the Gin backend, and a SQLite database. The
frontend communicates with the backend through a reverse proxy at `/v1/api`.
The backend manages all data in local SQLite storage.

The app depends on **orkai** running as a local service. On startup, the app
performs an orkai health check. If orkai is unreachable, the application blocks
all features and displays a dedicated status page. Once orkai is available, the
agent connects via MCP to access profile standards, PDF generation skills, and
search capabilities.

Optional external dependency: one or more LLM providers (Ollama for local
models, OpenAI, or Anthropic) configured during onboarding.

### 2.2 User Roles

| Role | Description |
|------|-------------|
| User | The sole operator. Onboards once, then creates/manages opportunities and generates documents via the agent. No authentication. |

### 2.3 Operating Environment

- Modern web browser (Chrome, Firefox, Safari, Edge — latest 2 versions)
- Backend runs as a single Go binary on macOS, Linux
- orkai must be installed and running locally
- SQLite database stored alongside the binary (configurable path)
- LLM API access (optional for Ollama/local; required for OpenAI/Anthropic)

### 2.4 Design Constraints

- Backend: Go + Gin, SQLite storage (WAL mode)
- Frontend: React + Vite + TypeScript, Tailwind CSS + shadcn/ui
- API: REST JSON, standard envelope (`{"data": ..., "error": ...}`)
- Distribution: single embedded Go binary (`make run`, `orkai-resume serve`)
- Built-in `/health` and `/metrics` endpoints
- No authentication
- orkai is a mandatory runtime dependency — the app does not function without it

---

## 3. Functional Requirements

Every requirement is uniquely identified, testable, and prioritized:
- **P0 (Critical)** — launch blocker, must ship first
- **P1 (High)** — core value, ship in initial release
- **P2 (Medium)** — important but deferrable
- **P3 (Low)** — nice-to-have, future consideration

---

### 3.1 Build & Run

#### FR-001 — Single Command Run (P0)

**As a** User, **I want to** run `make run` from the cloned repository, **so
that** the entire application compiles and starts with zero manual setup.

Acceptance criteria:
- `make run` compiles the Go backend, embeds the React frontend production build, and bundles SQLite
- A single binary is produced and executed
- The reverse proxy routes frontend requests to `/v1/api` on the backend
- The browser opens automatically to the application URL
- The command works on macOS and Linux

#### FR-002 — Global Install Command (P1)

**As a** User, **I want to** run `make install`, **so that** a global
`orkai-resume serve` command is available anywhere on my system.

Acceptance criteria:
- `make install` compiles the embedded binary and places it in the user's PATH
- Running `orkai-resume serve` starts the application, opens the browser, and displays logs in the terminal
- The installed binary includes the same embedded frontend, backend, and SQLite as `make run`

#### FR-003 — Reverse Proxy (P0)

**As a** Developer, **I want the** backend to serve the embedded frontend and
proxy API calls, **so that** the user accesses a single origin with no CORS
issues.

Acceptance criteria:
- `GET /` and all non-API routes serve the embedded React app (SPA fallback)
- `GET /v1/api/*` routes are proxied to the Gin backend handlers
- `GET /health` and `GET /metrics` are accessible directly

#### FR-004 — Orkai Health Gate (P0)

**As a** User, **I want to** be informed immediately if orkai is not running,
**so that** I know the application cannot function without it.

Acceptance criteria:
- On application load (and periodically thereafter), the backend checks orkai health
- If orkai is unreachable, the frontend displays a dedicated page: "Orkai is not running" with instructions to start it
- The Home page, Chat, and all other features are inaccessible until orkai is reachable
- The health check re-evaluates on page navigation and recovers automatically when orkai comes back

#### FR-005 — Development Mode (P1)

**As a** Developer, **I want to** run `make dev`, **so that** I get frontend
hot reload and backend live reload during development.

Acceptance criteria:
- `make dev` starts the Vite dev server with HMR and the Go backend with live reload
- The Vite dev server proxies `/v1/api` to the backend
- This mode does NOT use the embedded binary — files are served directly from disk

---

### 3.2 Onboarding

#### FR-010 — One-Page Onboarding (P0)

**As a** first-time User, **I want to** complete onboarding on a single page,
**so that** I can start using the app with minimal friction.

Acceptance criteria:

**LLM Provider Configuration:**
- User selects a provider: Ollama, OpenAI, or Anthropic
- User enters the API token (not required for Ollama)
- User can override the default model name for the selected provider
- Token is validated on save (test call to the provider's API)
- Configuration is persisted to the local SQLite database

**Profile Input:**
- User chooses one of two methods:
  - **Manual entry**: structured form with fields for name, contact info, professional summary, work experience, education, skills, projects, certifications, languages
  - **File upload**: drag-and-drop a PDF or Markdown file; the system parses and extracts structured profile data
- Parsed/extracted data is displayed for review and editing before save
- Profile data is saved as structured data (not raw markdown)

**orkai Setup (auto-executed):**

The backend uses the orkai MCP to create or resolve the following entities in the
user's workspace. All operations run in sequence with real-time progress display.

1. **Category (workspace):** Creates or resolves a category named `personal` to
   scope all orkai entities for this app. The category ID is stored for use in
   all subsequent orkai operations.

2. **Canonical Profile standard:** Creates a standard containing the user's
   profile data (identity, contact, positioning, work history, education,
   skills, languages). The body declares: "When any source document (LinkedIn
   PDF, older resumes, templates) disagrees with this standard, this standard
   wins." This makes the profile the single source of truth for all document
   generation.

3. **Cover Letter Writing Principles standard:** Creates a standard seeded from
   a built-in template encoding content rules: what a cover letter IS and IS
   NOT, tone and copy rules, anti-patterns (no referrer names, no pitching the
   current employer, no title-dropping), a pre-submission checklist, and a
   three-shape contribution framing for referral letters. The user can edit
   this text during onboarding.

4. **PDF Pipeline standard:** Creates a standard seeded from a built-in
   template containing the PDF generation pipeline: Markdown → HTML (pandoc +
   CSS) → PDF (weasyprint). Includes tested macOS install commands, page-count
   verification helpers, and CSS tuned for a professional 2-page resume (A4,
   0.85in margins, 10.5pt body, heading scale, page-break-avoid rules). The
   user can customize page targets and font sizes.

5. **PDF Generation skill:** Creates a skill seeded from a built-in template
   containing the step-by-step how-to: write the markdown source, apply the
   CSS, run WeasyPrint, verify the page count, and produce a PNG preview.

6. **Entity linking:** The PDF pipeline standard is linked to the PDF
   generation skill, and the cover letter principles standard is linked to the
   canonical profile standard, forming a traversable knowledge graph for RAG
   retrieval.

7. **MCP token collection:** The orkai MCP configuration token is auto-detected
   from local tool config files (cursor, cline, or opencode). If detection
   fails, the user is prompted to paste the token manually.

All created entity IDs are stored in the SQLite database for runtime retrieval
during document generation. Re-onboarding (when the user edits their profile
or settings) updates existing entities by ID — never creates duplicates.

#### FR-011 — Onboarding Progress Display (P1)

**As a** User, **I want to** see the real-time status of orkai configuration
steps during onboarding, **so that** I know what is being set up and whether
it succeeded.

Acceptance criteria:
- Each orkai setup step (create category, create profile standard, create cover letter principles standard, create PDF pipeline standard, create PDF generation skill, link entities, collect MCP token) shows a status indicator: pending, in-progress, success, or failed
- If a step fails, a retry button is shown with the error message
- The user can proceed with LLM config + profile even if orkai steps fail (but features remain blocked per FR-004)
- Overall progress bar shows completion percentage

---

### 3.3 Home Page

#### FR-020 — Opportunity Cards (P0)

**As a** User, **I want to** see all my job opportunities as cards on the home
page, **so that** I can quickly access any application's documents or agent.

Acceptance criteria:
- Each card displays: company name, role title, creation date (formatted human-readable)
- Each card shows download links for the latest approved Resume and Cover Letter (if they exist)
- Each card has an "Open Agent" button/link that opens the Chat page with context preloaded for that opportunity
- Cards are displayed in a responsive grid layout
- New opportunity button is prominently visible

#### FR-021 — Empty State (P0)

**As a** first-time User with no opportunities, **I want to** land on the Chat
page, **so that** I can immediately start creating my first application.

Acceptance criteria:
- If the user has zero opportunities, the Home page shows a brief message and a prominent "Start Chat" button
- Clicking "Start Chat" navigates to the Chat page
- After the first opportunity is created, the Home page shows the opportunity card

#### FR-022 — Pagination (P1)

**As a** User with many opportunities, **I want to** paginate through them,
**so that** the home page remains performant and navigable.

Acceptance criteria:
- Opportunities are paginated (default 12 per page)
- Page controls show current page, total pages, and next/previous buttons
- Pagination state is preserved in the URL query string

#### FR-023 — Filters (P1)

**As a** User, **I want to** filter opportunities by company, role, or date
range, **so that** I can find a specific application quickly.

Acceptance criteria:
- Company filter: text input, case-insensitive partial match
- Role filter: text input, case-insensitive partial match
- Date range filter: "from" and "to" date pickers
- Filters can be combined (AND logic)
- Active filters are shown as clearable chips
- Filter state is preserved in the URL query string

#### FR-024 — Search (P1)

**As a** User, **I want to** search across company names and role titles,
**so that** I can find an opportunity without scrolling.

Acceptance criteria:
- Single search input searches both company and role fields (case-insensitive)
- Results update as the user types (debounced 300ms)
- Search works in combination with filters (FR-023)
- Empty search results show a "no results" message with a clear-search option

#### FR-025 — Sorting (P1)

**As a** User, **I want to** sort opportunities, **so that** I can view them in
an order that makes sense to me.

Acceptance criteria:
- Default sort: most recently created first
- Sort options: date created (asc/desc), company name (A-Z / Z-A), role title (A-Z / Z-A)
- Sort selection is a dropdown or button group
- Sort state is preserved in the URL query string

#### FR-026 — Archive Opportunity (P2)

**As a** User, **I want to** archive old opportunities, **so that** the home
page focuses on active applications.

Acceptance criteria:
- Each card has an archive action (icon button or menu item)
- Archived opportunities are hidden from the default view
- A toggle or tab switches between "Active" and "Archived" views
- Archived opportunities can be unarchived

---

### 3.4 Chat / Agent

#### FR-030 — Chat Interface (P0)

**As a** User, **I want to** interact with an AI agent through a chat interface,
**so that** I can generate and refine resumes and cover letters conversationally.

Acceptance criteria:
- Chat UI includes: message history area, text input with send button, and a sidebar/header showing the current opportunity context
- Messages support markdown rendering (headings, lists, bold, italic, links)
- The agent's responses stream incrementally (tokens appear as they are generated)
- User can stop a streaming response
- Chat input supports multi-line text (Shift+Enter for newline, Enter to send)
- The system prompt that configures the agent is pre-loaded from orkai (FR-031)

#### FR-031 — Agent System Prompt (P0)

**As a** User, **I want the** agent to be configured with my orkai standards and
skills, **so that** it generates documents grounded in my actual background and
follows defined generation rules.

Acceptance criteria:
- On chat session start, the system prompt is assembled server-side from:
  - The user's canonical profile standard from orkai (the authoritative source — when any uploaded file or older document disagrees, the profile standard wins)
  - The cover letter writing principles standard from orkai (content rules, tone, anti-patterns, pre-submission checklist)
  - The PDF generation skill from orkai (the step-by-step how-to for producing PDFs)
  - A mandatory rule: "You must use the provided sources. Do not assume or fabricate information. Every claim must be traceable to the profile, the job description, or the writing principles."
- The system prompt is assembled per-session and is never hardcoded in the frontend
- If orkai data has changed since last session, the prompt reflects the latest versions
- The agent is instructed to retrieve additional context via the orkai search tool (FR-032) when the system prompt alone is insufficient

#### FR-032 — Agent Tools (P0)

**As the** agent, **I want to** have access to backend tools, **so that** I can
execute the actions needed to generate documents.

Acceptance criteria:

The agent has access to the following tools via the backend API:

1. **Shell Execution Tool**
   - Executes shell commands in a sandboxed temporary directory (per-session)
   - Supports bash and Python scripts
   - Commands have a configurable timeout (default 30 seconds)
   - Output (stdout, stderr, exit code) is returned to the agent
   - The working directory is isolated per chat session

2. **Artifact Tool**
   - `list_artifacts`: returns metadata for all saved artifacts (name, type, description, creation date, usage count)
   - `save_artifact`: saves a script file as an artifact with a description
   - `get_artifact`: retrieves the content of a specific artifact by ID
   - Artifacts are stored in the SQLite database

3. **Profile Access Tool**
   - Returns the user's structured profile data (all sections)
   - Read-only access

4. **orkai Search Tool**
   - Searches the orkai workspace for standards, skills, and documents
   - Used by the agent to discover additional context beyond the system prompt

#### FR-033 — Artifact Creation Trigger (P1)

**As a** User, **I want** scripts that produced a successful result to be saved
as reusable artifacts, **so that** future resume generation is faster and more
consistent.

Acceptance criteria:
- When the user gives a thumbs up (or equivalent approval) on a message containing a final approved PDF (see FR-047), the agent marks the scripts used in that session as artifact candidates
- The agent evaluates each candidate: if the same script pattern was used successfully, it calls `save_artifact`
- Only scripts that were actually executed and contributed to the output are saved
- The agent checks `list_artifacts` before creating a new script for the same purpose
- Artifacts are associated with descriptive metadata so the agent can determine relevance

#### FR-034 — Chat Session Lifecycle (P0)

**As a** User, **I want to** start a fresh chat session when editing an
opportunity, **so that** the agent has the right context without stale
conversation history.

Acceptance criteria:
- "Open Agent" from an opportunity card starts a new chat session
- The session is preloaded with: the user's profile data, the opportunity details (company, role, date), and any existing resume/cover letter documents for that opportunity
- Sessions are identified by opportunity ID — at most one active session per opportunity
- Opening a previously-edited opportunity starts a new session (no conversation history is retained per FR-035)
- The chat UI displays the opportunity context (company + role) prominently

#### FR-035 — No Conversation History Persistence (P0)

**As a** User, **I want** only the final documents to be saved, **so that** my
storage stays clean and the agent always starts fresh with current data.

Acceptance criteria:
- Chat messages are ephemeral — they are NOT persisted across sessions
- Closing the browser or navigating away discards the conversation
- Only approved documents (resume PDF, cover letter PDF) are persisted to the database
- Draft markdown content is persisted only while the session is active (in-memory or temporary storage)

#### FR-036 — Cover Letter Writing Rules (P0)

**As a** User, **I want the** agent to follow strict cover letter principles,
**so that** every generated letter is professional, non-generic, and free of
common mistakes.

Acceptance criteria:

The agent must follow these rules when generating any cover letter:

**What a cover letter IS:**
- A letter from the candidate to a potential employer, explaining how the candidate can contribute as an employee
- A human-sounding introduction of the candidate's value: craft, standards, hands-on work, range
- Approximately 300–400 words, one page, direct tone, contractions where natural

**What a cover letter IS NOT:**
- NOT a record of private conversations — never mention a friend or referrer's name in the letter body
- NOT a marketing pitch for the candidate's current employer — do not lead with current title or describe the employer's products
- NOT a title-dropping document — the current job title may appear at most once, buried in a role list, never as the headline credential
- NOT a repeat of the resume — the letter complements, it does not restate

**Tone and copy:**
- Direct opening — forbidden opening: "I am writing to express my strong interest in the [Role] position at [Company]"
- Contractions allowed (I'm, I've, don't)
- Varied sentence length — avoid every bullet or sentence starting with the same verb
- Warm, humble, confident tone — never salesy

**For referral or no-specific-role letters:**
- Address a generic hiring team/leadership, never the referrer by name
- Use a three-shape contribution framing: introduce exactly three concrete areas where the candidate could contribute, each 1–2 sentences with technical depth
- Close with explicit flexibility about engagement type (full-time, founding role, starting point)

**Anti-patterns (the agent must never produce):**
- Referrer names anywhere in the letter body
- Current employer as the subject of any paragraph
- Internal product or framework names in the letter body
- The forbidden opening cliché listed above

**Pre-submission:** Before finalizing any cover letter, the agent verifies:
- No referrer name in the letter body
- Current employer is not the subject of any paragraph
- Current title appears at most once and not as the headline
- Internal product or framework names are absent
- Direct opening (no cliché)
- One page, approximately 300–400 words

If any check fails, the agent revises before presenting the draft.

#### FR-037 — Accepted Document Persistence to orkai (P1)

**As a** User, **I want** approved documents to be saved back to my orkai
workspace, **so that** future generations can reference my real history and
improve over time.

Acceptance criteria:
- When a resume or cover letter is approved (via Approve button or chat-based approval), the backend writes the markdown content as a document in the user's orkai workspace
- The document is tagged with metadata: type (resume or cover letter), company, role, and date
- Future chat sessions include these documents in the agent's searchable context via the orkai search tool
- This creates a compounding effect: each generation improves the next by grounding the agent in the user's actual accepted output
- If orkai is unreachable during write-back, the operation fails gracefully — the document is still stored locally and write-back is retried on the next orkai health check

---

### 3.5 Document Review & Approval

#### FR-040 — Draft Mode (P0)

**As a** User, **I want the** agent to produce documents in draft mode, **so
that** I can review them before they become final PDFs.

Acceptance criteria:
- When the agent generates a resume or cover letter, it is marked as "draft"
- Drafts are stored as structured data + markdown (per FR-052, FR-053)
- Drafts are editable — the user can request changes through the chat
- There is no formal revision limit; the chat is the negotiation channel

#### FR-041 — Document Buttons in Chat (P0)

**As a** User, **I want to** see clearly labeled buttons in the chat when the
agent produces a document draft, **so that** I can open it for review with one
click.

Acceptance criteria:
- When the agent generates a resume draft, a `[Resume]` styled button appears in that chat message
- When the agent generates a cover letter draft, a `[Cover Letter]` styled button appears
- Buttons are visually distinct from regular text (button appearance, not inline link)
- Clicking a button opens the document in the review panel (FR-042)

#### FR-042 — Right-Panel Review (P0)

**As a** User, **I want to** preview a document's rendered markdown in a side
panel, **so that** I can read and evaluate the content before approving.

Acceptance criteria:
- Clicking `[Resume]` or `[Cover Letter]` opens a right-side panel
- The panel renders the document's markdown with proper formatting (headings, lists, bold, italic)
- The panel can be resized or collapsed
- If the chat is already displaying a panel and the user clicks a different document button, the panel switches to that document
- The panel includes an "Approve" button in the top-right corner (FR-043)

#### FR-043 — Approve Button (P0)

**As a** User, **I want to** approve a reviewed document with a single click,
**so that** PDF generation begins immediately.

Acceptance criteria:
- The review panel (FR-042) has an "Approve" button in the top-right corner
- Clicking "Approve" triggers PDF generation via the orkai PDF skill
- A loading indicator is shown during PDF generation
- The approved document transitions from "draft" to "approved" status

#### FR-044 — Chat-Based Approval (P1)

**As a** User, **I want to** approve documents by typing in the chat, **so
that** I can stay in the conversation flow without clicking UI buttons.

Acceptance criteria:
- User can type a message like "both documents are approved" or "approve the resume"
- The agent recognizes approval intent and triggers PDF generation
- The agent confirms in chat: "PDF documents created" with download buttons (FR-045)
- This is an alternative to the Approve button (FR-043), not a replacement

#### FR-045 — PDF Download Links (P0)

**As a** User, **I want to** see download links in the chat after PDF
generation, **so that** I can access the final documents.

Acceptance criteria:
- After PDF generation completes, the agent's response includes styled `[Resume]` and `[Cover Letter]` download buttons
- These buttons are visually distinct from draft buttons (e.g., different color or icon indicating final/PDF)
- The filename follows the pattern `{Company}-{Role}-{DocumentType}.pdf`

#### FR-046 — PDF Opens in New Tab (P0)

**As a** User, **I want** clicking a PDF download link to open the PDF in a new
browser tab, **so that** I can review the final rendered output.

Acceptance criteria:
- Clicking a final PDF button opens the PDF in a new browser tab
- The PDF renders directly in the browser (inline Content-Disposition or served with proper MIME type)
- The original chat tab remains open and unchanged

#### FR-047 — Thumbs Up on Final PDF (P1)

**As a** User, **I want to** give a thumbs up on the chat message containing
the final PDF, **so that** the system knows I'm satisfied and the underlying
scripts can be considered for artifact reuse.

Acceptance criteria:
- Each agent message in the chat has a thumbs up reaction button
- Clicking thumbs up on a message with final PDF download links:
  - Records the approval event
  - Triggers the artifact candidate evaluation (FR-033)
  - Shows a subtle confirmation animation
- Thumbs up is only meaningful on messages containing final PDFs; it is available on all messages but its artifact-triggering behavior is scoped to approved output

#### FR-048 — Chat Revision Loop (P0)

**As a** User, **I want to** request changes to drafts through natural
conversation, **so that** I can refine documents iteratively.

Acceptance criteria:
- User can request changes by typing in the chat (e.g., "emphasize my leadership experience", "make it more concise")
- The agent revises the document and produces a new draft
- There is no formal rejection workflow — the chat is the revision channel
- Each revision is a new draft that replaces the previous draft for that document
- The user can request changes to the resume, cover letter, or both in a single message

---

### 3.6 Data Model

#### FR-050 — User Profile (P0)

**As a** User, **I want** my professional background stored as structured data,
**so that** the agent can access specific sections when generating tailored
documents.

Acceptance criteria:
- Profile fields:
  - **Personal Info**: full name, email, phone, location, LinkedIn URL, website URL, GitHub URL
  - **Professional Summary**: free-text (up to 2000 characters)
  - **Work Experience**: ordered list of entries, each with: job title, company, location, start date, end date (or "Present"), description (markdown)
  - **Education**: ordered list of entries, each with: degree, institution, location, start date, end date, GPA (optional), description
  - **Skills**: named categories, each containing an ordered list of skill strings
  - **Projects**: ordered list of entries, each with: name, role, description, technologies, URL
  - **Certifications**: list of entries, each with: name, issuing organization, date obtained, expiry date (optional), credential URL (optional)
  - **Languages**: list of entries, each with: language name, proficiency level
- Profile is editable after onboarding via a settings page
- Profile is accessible to the agent via the Profile Access Tool (FR-032)
- **Authoritative source rule:** The canonical profile is the single source of truth for document generation. When any uploaded file (PDF resume, LinkedIn export) or older document disagrees with the profile, the profile wins. This rule is encoded in the profile standard saved to orkai during onboarding.

#### FR-051 — Opportunity (P0)

**As a** User, **I want to** track each job application as an opportunity with
key metadata, **so that** documents are organized by the job they target.

Acceptance criteria:
- Each opportunity has:
  - Company name (required)
  - Role title (required)
  - Creation date (auto-set to now on creation)
  - Status: active or archived (default: active)
  - Optional: job description (free-text, used as additional context for the agent)
- Opportunities are created by the agent during chat conversation (e.g., "I'm applying to Microsoft as a Backend Developer")
- User can edit opportunity metadata (company, role, description)

#### FR-052 — Resume Per Opportunity (P0)

**As a** User, **I want** each opportunity to have an associated resume with
structured content and a final PDF, **so that** every application has a
tailored document.

Acceptance criteria:
- Each resume is linked to exactly one opportunity
- Resume stores:
  - Structured section data (mirrors the profile structure but may differ — the agent tailors content)
  - Markdown content (the rendered resume text)
  - PDF file path (generated on approval)
  - Status: draft or approved
  - Last modified timestamp
- A resume can be regenerated by the agent at any time
- Only one resume exists per opportunity (regeneration replaces previous)

#### FR-053 — Cover Letter Per Opportunity (P0)

**As a** User, **I want** each opportunity to have an associated cover letter,
**so that** my application package is complete.

Acceptance criteria:
- Each cover letter is linked to exactly one opportunity
- Cover letter stores:
  - Markdown content (the letter text)
  - PDF file path (generated on approval)
  - Status: draft or approved
  - Last modified timestamp
- A cover letter can be regenerated by the agent at any time

#### FR-054 — Artifact Storage (P1)

**As the** agent, **I want to** save and retrieve reusable scripts, **so that**
I don't recreate solutions that have already proven successful.

Acceptance criteria:
- Each artifact stores:
  - Name (filename or descriptive label)
  - Type: Python script or bash script
  - Description (purpose, what it does)
  - Script content (the source code)
  - Creation date
  - Last used date
  - Usage count (number of times successfully reused)
- Artifacts are stored in the SQLite database
- Artifacts can be listed, retrieved by ID, and deleted
- Deletion is allowed (user or agent can remove obsolete artifacts)

---

### 3.7 Design System

#### FR-060 — Grayscale Simplicity Design (P0)

**As a** User, **I want** the UI to follow a grayscale simplicity design
language, **so that** the interface is clean, professional, and distraction-free.

Acceptance criteria:
- Color palette is predominantly grayscale (white, grays, black)
- Accent colors are used sparingly for primary actions and status indicators
- Typography follows a clean, readable hierarchy
- Spacing and layout mirror the simplicity of tools like Google Analytics and Google Cloud Console
- Dark mode is not required for v1
- shadcn/ui components are themed to match the grayscale palette

---

### 3.8 Export

#### FR-070 — PDF Resume Export (P0)

**As a** User, **I want** the agent to generate a PDF of my approved resume,
**so that** I can submit it with my job application.

Acceptance criteria:
- PDF is generated using the pipeline: Markdown → HTML (pandoc) + CSS →
  PDF (weasyprint). No LaTeX dependency.
- The CSS uses the settings from the PDF Pipeline standard (customizable per
  user): A4 page size, configurable margins (default 0.85in), configurable body
  font size (default 10.5pt), heading scale with page-break-avoid rules on
  section headings
- Target page count: 2 pages for a senior candidate resume. Content tactics
  (merging bullets, collapsing older roles, inlining education) are applied
  before reducing font size to hit the target
- Text in the PDF is selectable (not rasterized)
- PDF is stored server-side and accessible via a download URL
- Download filename: `{Company}-{Role}-Resume.pdf`

#### FR-071 — PDF Cover Letter Export (P0)

**As a** User, **I want** the agent to generate a PDF of my approved cover
letter, **so that** I can submit it alongside my resume.

Acceptance criteria:
- PDF is generated using the same pipeline as FR-070: Markdown → HTML (pandoc) + CSS → PDF (weasyprint)
- PDF follows standard business letter formatting (date, salutation, body, closing, signature block)
- Signature block renders with correct vertical rhythm (2em margin-top before "Sincerely,", 1em between sign-off and name)
- Target: 1 page, approximately 300–400 words
- Text in the PDF is selectable (not rasterized)
- PDF is stored server-side and accessible via a download URL
- Download filename: `{Company}-{Role}-CoverLetter.pdf`

---

## 4. Non-Functional Requirements

### 4.1 Performance

- **NFR-01**: Application starts and is ready to serve within 3 seconds of `make run` or `orkai-resume serve` on modern hardware (P0)
- **NFR-02**: Chat agent streaming begins within 1 second of sending a message (P0)
- **NFR-03**: PDF generation completes within 10 seconds for a standard 1–2 page document (P0)
- **NFR-04**: Home page loads and renders opportunity cards within 500ms for up to 100 opportunities (P1)
- **NFR-05**: API responses (excluding PDF generation and LLM calls) under 200ms p95 (P1)

### 4.2 Reliability

- **NFR-06**: Backend exposes `/health` endpoint returning 200 when operational, including orkai connectivity status (P0)
- **NFR-07**: Backend exposes `/metrics` in Prometheus format (P1)
- **NFR-08**: orkai health check runs on application startup and on each page navigation; recovers automatically (P0)
- **NFR-09**: SQLite database uses WAL mode for safe concurrent access (P0)
- **NFR-10**: Agent tool calls (shell, artifact, profile, orkai search) have configurable timeouts and return errors gracefully (P0)
- **NFR-11**: Onboarding form data is persisted incrementally — page refresh does not lose filled fields (P1)

### 4.3 Usability

- **NFR-12**: All form fields use client-side validation with inline error messages (P0)
- **NFR-13**: UI is responsive — usable on screens from 1024px to 2560px wide (P1)
- **NFR-14**: Loading, empty, and error states are handled for every data-fetching view and agent interaction (P0)
- **NFR-15**: Chat messages display timestamps (relative: "2 min ago") (P1)
- **NFR-16**: Keyboard shortcuts: Enter to send chat message, Shift+Enter for newline, Escape to close review panel (P1)

### 4.4 Security

- **NFR-17**: CORS is restricted to the frontend origin only (configurable via env) (P0)
- **NFR-18**: LLM API keys are stored server-side only, never exposed to the frontend (P0)
- **NFR-19**: Shell execution tool runs in an isolated temporary directory per session — no access outside that directory (P0)
- **NFR-20**: The embedded binary does not expose debug endpoints or sensitive configuration in production mode (P0)

### 4.5 Maintainability

- **NFR-21**: Backend follows layered architecture: handlers → services → store (P0)
- **NFR-22**: Frontend follows typed component hierarchy with TanStack Query for data fetching (P0)
- **NFR-23**: All configuration is environment-variable based with a committed `.env.example` (P0)
- **NFR-24**: Agent system prompt is assembled server-side from orkai sources, not hardcoded in the frontend (P0)

---

## 5. External Interfaces

### 5.1 REST API

The backend exposes a REST JSON API. All responses follow a standard envelope:

```json
{
  "data": { ... },
  "error": { "code": "STRING", "message": "Human-readable message" }
}
```

Built-in endpoints (no envelope):
- `GET /health` → `{"status": "ok", "orkai": "connected|unreachable"}`
- `GET /metrics` → Prometheus text format

API endpoints:

**Profile:**
- `GET    /v1/api/profile` — get user profile
- `PUT    /v1/api/profile` — update user profile
- `POST   /v1/api/profile/upload` — upload and parse PDF/Markdown profile

**Opportunities:**
- `GET    /v1/api/opportunities` — list opportunities (supports pagination, filters, search, sort)
- `POST   /v1/api/opportunities` — create opportunity
- `GET    /v1/api/opportunities/:id` — get opportunity with associated documents
- `PUT    /v1/api/opportunities/:id` — update opportunity
- `DELETE /v1/api/opportunities/:id` — delete opportunity (cascades to documents)
- `PUT    /v1/api/opportunities/:id/archive` — toggle archive status

**Resume:**
- `GET    /v1/api/opportunities/:id/resume` — get resume for opportunity
- `GET    /v1/api/opportunities/:id/resume/pdf` — download resume PDF

**Cover Letter:**
- `GET    /v1/api/opportunities/:id/cover-letter` — get cover letter for opportunity
- `GET    /v1/api/opportunities/:id/cover-letter/pdf` — download cover letter PDF

**Agent Chat:**
- `POST   /v1/api/chat` — send message, returns streaming SSE response
- `POST   /v1/api/chat/:sessionId/approve` — approve document, trigger PDF generation

**Agent Tools (called by the agent, not the frontend):**
- `POST   /v1/api/tools/shell` — execute shell command in sandboxed tmp dir
- `GET    /v1/api/tools/artifacts` — list artifacts
- `POST   /v1/api/tools/artifacts` — save artifact
- `GET    /v1/api/tools/artifacts/:id` — get artifact content
- `DELETE /v1/api/tools/artifacts/:id` — delete artifact
- `GET    /v1/api/tools/profile` — get profile data (agent-accessible)
- `POST   /v1/api/tools/orkai-search` — search orkai workspace

**Onboarding:**
- `POST   /v1/api/onboarding/llm-config` — save LLM provider config
- `POST   /v1/api/onboarding/profile` — save profile data
- `POST   /v1/api/onboarding/orkai-setup` — execute orkai setup steps
- `GET    /v1/api/onboarding/orkai-setup/status` — poll orkai setup progress

**System:**
- `GET    /v1/api/orkai/health` — check orkai connectivity

### 5.2 orkai MCP

The app uses the orkai MCP for two purposes: **onboarding** (creating entities
once) and **runtime** (reading entities for generation, writing back results).

**Onboarding — entities created (one-time):**

During onboarding (FR-010), the backend creates the following in the user's
orkai workspace:

| Entity Type | Name Pattern | Content |
|---|---|---|
| Category | `personal` | Scopes all resume-app entities |
| Standard | `{Full Name} — Canonical Profile for Resume & Cover Letter Generation` | User's identity, contact, positioning, work history, education, skills, languages. Declares itself as the authoritative source. |
| Standard | `Cover Letter Writing Principles — Personal Workspace ({Full Name})` | Content rules, tone, anti-patterns, pre-submission checklist, three-shape framing for referrals |
| Standard | `Resume & Cover Letter PDF Pipeline — Tooling & CSS Tuning` | pandoc + weasyprint pipeline, CSS overrides for page targets, verification helpers |
| Skill | `Resume and Cover Letter PDF Generation` | Step-by-step how-to: write markdown, apply CSS, run WeasyPrint, verify page count, produce PNG preview |

The four standard/skill entities are cross-linked with `references` relations
to form a traversable knowledge graph. All entity IDs are stored in the SQLite
`user_settings` table. Re-onboarding updates existing entities by ID — it never
creates duplicates.

**Runtime — generation context:**

At the start of each chat session, the backend fetches the current versions
of the profile standard, cover letter principles standard, and PDF generation
skill. These are assembled into the system prompt (FR-031). During generation,
the agent may search the orkai workspace for additional context, including
previously accepted documents (FR-037).

**Runtime — document write-back:**

When a resume or cover letter is approved, the backend creates a document
in the user's orkai workspace under the `personal` category with metadata
(type, company, role, date). These documents compound over time, grounding
future generations in the user's real accepted output.

The MCP configuration token is collected during onboarding (auto-detected
from cursor, cline, or opencode config files) and stored server-side.

### 5.3 LLM Providers

The backend communicates with one or more LLM providers for the chat agent.
Supported providers:
- **Ollama** — local, no API key required, configurable base URL
- **OpenAI** — API key required, configurable model
- **Anthropic** — API key required, configurable model

Provider selection and configuration happen during onboarding (FR-010) and are
stored in the SQLite database. The backend abstracts provider differences behind
a common chat completion interface.

### 5.4 Storage

SQLite database stored locally. Single-file database at a configurable path
(default: `$HOME/.orkai-resume/data.db` or `./data.db` for development).
WAL mode enabled for safe concurrent access.

Database stores:
- User profile (structured)
- Opportunities with metadata
- Resumes and cover letters (structured data, markdown, PDF paths)
- Artifacts (scripts with metadata)
- LLM provider configuration
- Onboarding state

---

## 6. Appendix

### 6.1 Revision History

| Version | Date | Author | Changes |
|---------|------|--------|---------|
| 0.1.0 | 2026-07-05 | — | Initial SRS draft — form-based resume builder |
| 0.2.0 | 2026-07-05 | — | Complete rewrite — agentic chat-first architecture with orkai MCP integration |
| 0.2.1 | 2026-07-05 | — | Refinement — expanded onboarding with 7 orkai entities, added cover letter writing rules (FR-036), added accepted document write-back to orkai (FR-037), specified pandoc+weasyprint PDF pipeline with CSS tuning, added authoritative profile source rule |

### 6.2 Open Questions

1. Should the app support multiple languages for the UI (i18n)? Currently assumed English-only.
2. Should resume/profile data be exportable/importable via JSON Resume format (jsonresume.org) for interoperability?
3. Should the agent support multiple LLM providers simultaneously (e.g., use Anthropic for writing, Ollama for quick tasks)?
4. Should cover letters support a separate "tone" configuration (formal, conversational, enthusiastic)?
5. What is the exact orkai MCP token collection mechanism across cursor, cline, and opencode — are the config file paths stable?

### 6.3 Priority Summary

| Priority | Count | Key Items |
|----------|-------|-----------|
| P0 | 26 | `make run`, reverse proxy, orkai health gate, onboarding, opportunity cards, empty state, chat interface, agent system prompt, agent tools, chat session lifecycle, no conversation history, cover letter writing rules, draft mode, document buttons, review panel, approve button, PDF download links, PDF in new tab, revision loop, profile, opportunity, resume, cover letter data models, grayscale design, PDF export (resume + cover letter via pandoc+weasyprint), health/metrics, CORS, API key security, sandboxed shell, NFR-01–03 |
| P1 | 12 | Global install, dev mode, onboarding progress, pagination, filters, search, sorting, artifact trigger, accepted document persistence to orkai, chat-based approval, thumbs up, artifact storage, Prometheus metrics, responsive layout, timestamps, keyboard shortcuts, form persistence |
| P2 | 1 | Archive opportunity |