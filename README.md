# resume-app

> A lightweight Job Seeker CRM with AI-powered resume and cover letter generation.

resume-app helps you manage your job search end-to-end. Track job opportunities, generate tailored resumes and cover letters with AI, and keep everything organized in one place — locally on your machine.

---

## Why

Job seeking is repetitive. Every application asks for a tailored resume and cover letter that matches the role description. Doing this by hand is slow and error-prone. resume-app automates the document generation while keeping you in control of every output.

## What it does

- **Track job opportunities** — register company name, role, and job description for each application.
- **Generate tailored resumes** — AI produces a resume adapted to the job description.
- **Generate cover letters** — AI produces a cover letter for the specific role and company.
- **Edit via chat** — refine any generated document through a prompt-based chat with the LLM. No need to rewrite from scratch.
- **Open, copy path, edit, delete** — manage generated PDFs from the web UI.
- **Local-first** — all documents are saved to a directory on your machine. No cloud lock-in.

## How it works

1. Create a job opportunity (company, role, description).
2. The AI generates a resume and cover letter tailored to that opportunity.
3. The generated PDFs are saved to your configured local directory.
4. Open the PDF to review, copy the file path, edit via chat, or delete.

## AI providers

resume-app supports multiple LLM providers. Configure yours in `.env`:

| Provider | Token required | Notes |
|----------|---------------|-------|
| Ollama | No | Local-first, private. Recommended for offline use. |
| OpenAI | Yes | Set `OPENAI_API_KEY` in `.env`. |
| Anthropic | Yes | Set `ANTHROPIC_API_KEY` in `.env`. |

Copy `.env.example` to `.env` and fill in your provider and API keys. See [Configuration](#configuration).

## Configuration

All configuration is environment-based. No hardcoded values.

```bash
cp .env.example .env
```

Key variables (documented in `.env.example`):

```env
# AI provider: ollama | openai | anthropic
LLM_PROVIDER=ollama

# OpenAI (required if LLM_PROVIDER=openai)
OPENAI_API_KEY=

# Anthropic (required if LLM_PROVIDER=anthropic)
ANTHROPIC_API_KEY=

# Where generated PDFs are saved
OUTPUT_DIR=./generated

# Backend
BACKEND_PORT=8080
CORS_ALLOWED_ORIGINS=http://localhost:5173

# Frontend
VITE_API_URL=http://localhost:8080
```

## Tech stack

| Layer | Stack |
|-------|-------|
| Backend | Go + [Gin](https://gin-gonic.com/) |
| Frontend | React + Vite + TypeScript |
| UI | Tailwind CSS + [shadcn/ui](https://ui.shadcn.com/) |
| Data fetching | TanStack Query |
| Local state | Zustand |
| AI | Ollama, OpenAI, or Anthropic |

## Project structure

```
resume-app/
├── backend/          # Go + Gin API
│   ├── cmd/          # entrypoint
│   ├── internal/
│   │   ├── handlers/ # HTTP handlers
│   │   ├── services/ # business logic
│   │   ├── models/   # domain types
│   │   ├── store/    # data access
│   │   └── middleware/
│   └── .env.example
├── frontend/         # React + Vite + TypeScript
│   ├── src/
│   │   ├── components/ # reusable UI (shadcn/ui)
│   │   ├── pages/     # route-level views
│   │   ├── hooks/     # custom hooks
│   │   ├── api/       # TanStack Query hooks
│   │   ├── store/     # Zustand stores
│   │   ├── types/     # shared types
│   │   └── lib/       # utilities
│   └── .env.example
└── AGENTS.md         # AI agent instructions (orkai)
```

## Getting started

### Prerequisites

- Go 1.22+
- Node.js 20+
- [air](https://github.com/air-verse/air) for backend live reload during `make dev` — install once: `go install github.com/air-verse/air@latest`
- One of: [Ollama](https://ollama.com/), an OpenAI API key, or an Anthropic API key

### Backend

```bash
cd backend
cp .env.example .env
# edit .env with your provider and keys
go build ./...
go run cmd/main.go
```

### Frontend

```bash
cd frontend
cp .env.example .env
# edit .env with VITE_API_URL
npm install
npm run dev
```

Open http://localhost:5173.

## Development

### `make dev`

Runs the full dev environment with hot reload on both surfaces:

- **Frontend**: Vite dev server with HMR (http://localhost:5173)
- **Backend**: [air](https://github.com/air-verse/air) watches `backend/**/*.go` and rebuilds + restarts the Go server on every save (http://localhost:8080)

Requires `air` in PATH — see [Prerequisites](#prerequisites). The Vite dev
server proxies `/v1/api` and `/health` to the backend, so the frontend talks
to the live-reloaded Go server with no CORS setup.

### Build & test

Backend (run in `backend/`):

```bash
go build ./...      # compile
go vet ./...        # static analysis
gofmt -l .          # formatting check
go test ./...       # tests
```

Frontend (run in `frontend/`):

```bash
npm run lint        # oxlint
npm run typecheck   # tsc --noEmit
npm run test        # vitest
npm run build       # vite build
```

### Built-in endpoints

- `GET /health` — health check, no auth.
- `GET /metrics` — Prometheus metrics, no auth.

## Roadmap

- [ ] Job opportunity CRUD
- [ ] AI resume generation (Ollama / OpenAI / Anthropic)
- [ ] AI cover letter generation
- [ ] Document edit via chat
- [ ] PDF viewer
- [ ] Application status tracking (applied, interview, offer, rejected)
- [ ] Export job opportunities as CSV

## License

[MIT](LICENSE)