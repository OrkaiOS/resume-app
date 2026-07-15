.PHONY: run install dev build clean

build:
	cd frontend && npm run build
	mkdir -p backend/cmd/static
	cp -R frontend/dist/* backend/cmd/static/
	cd backend && go build -tags prod -o bin/resume-app ./cmd

run: build
	./backend/bin/resume-app

install: build
	@if [ -n "$$GOPATH" ]; then \
		cp backend/bin/resume-app "$$GOPATH/bin/orkai-resume"; \
	else \
		cp backend/bin/resume-app /usr/local/bin/orkai-resume; \
	fi

# Default ports — override with: make dev VITE_DEV_PORT=5180 BACKEND_PORT=5181
VITE_DEV_PORT ?= 5173
BACKEND_PORT ?= 8080

# @orkai:ref(id=a7108b40-a54d-48c6-b464-44a20684e990)
# @orkai:decision air replaces go run for FR-005 live reload; air is dev-only (not in go.mod), inherits env so config.Load + /health + /metrics stay intact
dev:
	@command -v air >/dev/null 2>&1 || { \
		echo "air not found. Install: go install github.com/air-verse/air@latest"; \
		exit 1; \
	}
	@trap 'kill 0' INT TERM; \
	(cd frontend && VITE_DEV_PORT=$(VITE_DEV_PORT) VITE_BACKEND_PORT=$(BACKEND_PORT) npm run dev) & \
	(cd backend && CORS_ALLOWED_ORIGINS=http://localhost:$(VITE_DEV_PORT) BACKEND_PORT=$(BACKEND_PORT) air) & \
	wait

clean:
	rm -rf backend/bin
	find backend/cmd/static -mindepth 1 -not -name .gitkeep -delete 2>/dev/null || true
	rm -rf frontend/dist